# x402-relayer 使用指南

本文档介绍如何部署和使用 x402-relayer 服务，实现用户使用 ERC20 代币支付中继费，由 Relayer 代付 Gas 广播交易。

## 目录

- [概述](#概述)
- [工作原理](#工作原理)
- [服务端部署](#服务端部署)
- [客户端接入](#客户端接入)
- [API 参考](#api-参考)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)

---

## 概述

x402-relayer 是基于 [Coinbase x402](https://github.com/coinbase/x402) 协议的交易中继服务，解决用户没有原生代币（AEX）无法支付 Gas 的问题。

**核心价值**：
- 用户只需持有支付代币，无需 AEX 即可发送交易
- 基于 EIP-3009 标准，支付授权无需 Gas
- 支持任何实现 EIP-3009 的 ERC20 代币

**代币要求**：

支付代币必须实现 [EIP-3009](https://eips.ethereum.org/EIPS/eip-3009) 标准，支持 `transferWithAuthorization` 方法。

---

## 工作原理

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户端                                    │
├─────────────────────────────────────────────────────────────────┤
│  1. 构建并签名交易 (signedTx)                                    │
│  2. 创建 EIP-3009 支付授权签名                                   │
│  3. 发送请求到 x402-relayer                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      x402-relayer                                │
├─────────────────────────────────────────────────────────────────┤
│  4. 验证支付授权签名 (EIP-712)                                   │
│  5. 验证收款地址和金额                                           │
│  6. 调用 transferWithAuthorization 结算代币                       │
│  7. 广播用户交易 (Relayer 支付 Gas)                              │
│  8. 返回交易回执                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**关键安全机制**：
- 支付授权的 `to` 地址必须等于 Relayer 的 `pay_to_address`
- 支付金额必须 >= `relay_fee_per_tx`
- 授权必须在 `validBefore` 时间之前使用

---

## 服务端部署

### 1. 环境要求

- Go 1.21+
- 运行中的 AESC 节点（EVM RPC 可用）
- Relayer 钱包（需要 AEX 余额支付 Gas）

### 2. 编译

```bash
cd x402-relayer
go build -o x402-relayer ./cmd/x402-relayer
```

### 3. 配置文件

创建 `config.toml`：

```toml
[x402-relayer]
enabled = true
port = 8402
pay_to_address = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
token_contract = "0x5FbDB2315678afecb367f032d93F642f64180aa3"  # ERC20 合约地址
token_name = "Test USDT"      # 必须与合约 name() 一致
token_version = "1"           # 必须与合约 version 一致
network_id = "eip155:71603"
private_key = "${X402_RELAYER_KEY}"
relay_fee_per_tx = "10000000000000000"  # 0.01 token (18 decimals)
evm_rpc = "http://localhost:8545"
db_path = "./x402-relayer.db"
```

### 4. 启动服务

```bash
# 设置私钥环境变量（推荐）
export X402_RELAYER_KEY="59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

# 启动
./x402-relayer --config config.toml
```

### 5. 验证服务

```bash
# 健康检查
curl http://localhost:8402/health
# {"status":"ok","timestamp":1737385601}

# 查看支付要求
curl http://localhost:8402/payment-requirements
```

---

## 客户端接入

### 完整流程示例 (JavaScript/ethers.js)

```javascript
const { ethers } = require('ethers');

async function relayTransaction() {
  const RELAYER_URL = 'http://localhost:8402';
  const CHAIN_ID = 71603;
  
  // 用户钱包
  const wallet = new ethers.Wallet(PRIVATE_KEY);
  
  // ========== 步骤 1: 获取支付要求 ==========
  const reqRes = await fetch(`${RELAYER_URL}/payment-requirements`);
  const { accepts } = await reqRes.json();
  const paymentReq = accepts[0];
  
  console.log('Payment requirements:', paymentReq);

  // ========== 步骤 2: 构建并签名用户交易 ==========
  const provider = new ethers.JsonRpcProvider('http://localhost:8545');
  const nonce = await provider.getTransactionCount(wallet.address);

  const tx = {
    to: '0x0000000000000000000000000000000000000001',
    value: 0n,
    data: '0x',
    nonce: nonce,
    gasLimit: 21000,
    gasPrice: ethers.parseUnits('1', 'gwei'),
    chainId: CHAIN_ID
  };

  const signedTx = await wallet.signTransaction(tx);
  console.log('Signed transaction:', signedTx);

  // ========== 步骤 3: 创建 EIP-3009 支付授权 ==========
  const from = wallet.address;
  const to = paymentReq.payTo;  // 必须使用 Relayer 返回的地址
  const value = BigInt(paymentReq.maxAmountRequired);  // 支付金额
  const validAfter = 0n;
  const validBefore = BigInt(Math.floor(Date.now() / 1000) + 300);  // 5分钟后过期
  const paymentNonce = ethers.randomBytes(32);  // 随机 nonce

  // EIP-712 Domain（必须与代币合约一致）
  const domain = {
    name: 'Test USDT',  // 必须与合约的 name() 返回值一致
    version: '1',       // 必须与合约的 version 一致
    chainId: CHAIN_ID,
    verifyingContract: paymentReq.asset
  };

  // EIP-712 Types
  const types = {
    TransferWithAuthorization: [
      { name: 'from', type: 'address' },
      { name: 'to', type: 'address' },
      { name: 'value', type: 'uint256' },
      { name: 'validAfter', type: 'uint256' },
      { name: 'validBefore', type: 'uint256' },
      { name: 'nonce', type: 'bytes32' }
    ]
  };

  const message = { from, to, value, validAfter, validBefore, nonce: paymentNonce };
  const signature = await wallet.signTypedData(domain, types, message);
  const sig = ethers.Signature.from(signature);

  // ========== 步骤 4: 构建支付 Payload ==========
  const payload = {
    x402Version: 1,
    scheme: 'exact',
    network: paymentReq.network,
    payload: {
      from: from,
      to: to,
      value: value.toString(),
      validAfter: validAfter.toString(),
      validBefore: validBefore.toString(),
      nonce: ethers.hexlify(paymentNonce),
      v: sig.v,
      r: sig.r,
      s: sig.s
    }
  };

  const paymentBase64 = btoa(JSON.stringify(payload));

  // ========== 步骤 5: 发送中继请求 ==========
  const response = await fetch(`${RELAYER_URL}/relay`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-PAYMENT': paymentBase64
    },
    body: JSON.stringify({ signedTx })
  });

  if (response.status === 402) {
    const error = await response.json();
    throw new Error(`Payment failed: ${error.error}`);
  }

  const result = await response.json();
  console.log('Transaction hash:', result.txHash);
  console.log('Gas used:', result.gasUsed);
  console.log('Record ID:', result.recordId);

  return result;
}

relayTransaction().catch(console.error);
```

---

## API 参考

### GET /health

健康检查端点。

**响应**：
```json
{
  "status": "ok",
  "timestamp": 1737385601
}
```

### GET /payment-requirements

获取支付要求，客户端据此构建支付授权。

**响应**：
```json
{
  "accepts": [{
    "scheme": "exact",
    "network": "eip155:71603",
    "maxAmountRequired": "10000000000000000",
    "asset": "0x5FbDB2315678afecb367f032d93F642f64180aa3",
    "payTo": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    "requiredDeadlineSeconds": 300,
    "resource": "/relay",
    "description": "Transaction relay service"
  }]
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| `network` | CAIP-2 格式的链 ID |
| `maxAmountRequired` | 中继费用（代币最小单位） |
| `asset` | 支付代币合约地址 |
| `payTo` | Relayer 收款地址（**支付授权的 to 必须等于此值**） |
| `requiredDeadlineSeconds` | 建议的授权有效期（秒） |

### POST /relay

提交交易中继请求。

**请求头**：
- `Content-Type: application/json`
- `X-PAYMENT: <base64_encoded_payment_payload>`（可选）

**请求体**：
```json
{
  "signedTx": "0xf86c..."
}
```

**无 X-PAYMENT 头**：返回 HTTP 402
```json
{
  "accepts": [...],
  "error": "payment required"
}
```

**有 X-PAYMENT 头且验证通过**：返回 HTTP 200
```json
{
  "success": true,
  "txHash": "0x9e255391...",
  "gasUsed": 21000,
  "recordId": "3b48b036-..."
}
```

### GET /records

查询中继记录。

**查询参数**：
- `limit`: 每页数量（默认 10）
- `offset`: 偏移量（默认 0）
- `status`: 过滤状态（可选：`success`, `failed`）

**响应**：
```json
{
  "records": [...],
  "total": 100
}
```

### GET /records/stats

获取统计信息。

**响应**：
```json
{
  "total_records": 100,
  "successful_relays": 95,
  "failed_relays": 5,
  "total_payments": "950000",
  "total_gas_used": 1995000
}
```

---

## 错误处理

### HTTP 状态码

| 状态码 | 含义 | 处理建议 |
|--------|------|---------|
| 200 | 成功 | 保存 txHash |
| 402 | 支付验证失败 | 检查签名/余额/过期时间/收款地址 |
| 400 | 请求格式错误 | 检查请求体格式 |
| 500 | 服务器错误 | 查看错误信息，可能是链上问题 |

### 常见错误及解决方案

| 错误信息 | 原因 | 解决方案 |
|----------|------|---------|
| `payment recipient mismatch` | 支付授权的 `to` 与 `payTo` 不匹配 | 使用 `/payment-requirements` 返回的 `payTo` |
| `insufficient payment amount` | 支付金额不足 | 增加 `value` 至 >= `maxAmountRequired` |
| `authorization expired` | 授权已过期 | 增大 `validBefore` 时间戳 |
| `signature verification failed` | 签名无效 | 检查 EIP-712 Domain 配置 |
| `insufficient balance` | 用户代币余额不足 | 充值代币 |
| `nonce already used` | nonce 已被使用 | 重新生成随机 nonce |

---

## 最佳实践

### 服务端

1. **私钥安全**：使用环境变量，不要硬编码在配置文件中
2. **AEX 余额监控**：设置告警，确保 Relayer 钱包有足够 Gas
3. **数据库备份**：定期备份 SQLite 数据库
4. **日志监控**：监控 `settlement failed` 和 `broadcast failed` 日志
5. **限流保护**：在生产环境添加请求限流

### 客户端

1. **动态获取配置**：每次请求前调用 `/payment-requirements` 获取最新配置
2. **合理设置过期时间**：`validBefore` 建议设置为当前时间 + 5 分钟
3. **随机 nonce**：使用 `crypto.randomBytes(32)` 生成随机 nonce
4. **错误重试**：对于 500 错误，可以适当重试
5. **保存 recordId**：用于后续查询交易状态

### 代币精度

ERC20 代币通常使用 18 位精度。0.01 代币 = `10000000000000000` (1e16)。

---

## 相关文档

- [客户端接入指南](./client-guide.md) - 详细的客户端代码示例
- [运维手册](./operations.md) - 服务监控和故障排查
- [README](../README.md) - 快速开始

