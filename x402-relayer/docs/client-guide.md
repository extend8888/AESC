# x402-relayer 客户端接入指南

本指南介绍如何从客户端调用 x402-relayer 服务进行交易中继。

## 概述

用户（没有 AEX）可以通过以下流程完成交易：

```
1. 用户构建并签名交易
2. 用户创建 EIP-3009 授权签名（USDT 支付授权）
3. 发送请求到 x402-relayer
4. x402-relayer 验证支付 → 结算 USDT → 广播交易
5. 返回交易回执
```

## 前置条件

- 用户钱包有足够的 **USDT** 余额
- 用户钱包有 **已签名的交易**（无需 AEX）
- 了解当前 **中继费用**（通过 `/payment-requirements` 查询）

## 步骤详解

### 步骤 1: 查询支付要求

```bash
curl http://localhost:8402/payment-requirements
```

响应：
```json
{
  "accepts": [{
    "scheme": "exact",
    "network": "eip155:71603",
    "maxAmountRequired": "10000",
    "asset": "0x0000000000000000000000000000000000001010",
    "payTo": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    "requiredDeadlineSeconds": 300
  }]
}
```

关键字段：
- `maxAmountRequired`: 中继费用（0.01 USDT = 10000，假设 6 decimals）
- `payTo`: Relayer 收款地址（**必须与支付授权的 `to` 字段匹配**）
- `network`: 链 ID（用于签名）
- `asset`: USDT 合约地址

### 步骤 2: 构建用户交易

使用任何 EVM 库构建并签名交易：

```javascript
const { ethers } = require('ethers');

const wallet = new ethers.Wallet(PRIVATE_KEY);
const tx = {
  to: '0x...',
  value: ethers.parseEther('0'),
  data: '0x...',
  nonce: 0,
  gasLimit: 21000,
  gasPrice: ethers.parseUnits('1', 'gwei'),
  chainId: 71603
};

const signedTx = await wallet.signTransaction(tx);
// signedTx = "0xf86c..."
```

### 步骤 3: 创建 EIP-3009 授权签名

EIP-3009 允许用户授权第三方转移自己的代币，无需 Gas。

```javascript
const { ethers } = require('ethers');

// 参数
const from = wallet.address;
const to = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';  // Relayer
const value = 10000n;  // 0.01 USDT
const validAfter = 0n;
const validBefore = BigInt(Math.floor(Date.now() / 1000) + 300);  // 5分钟后过期
const nonce = ethers.randomBytes(32);  // 随机 nonce

// EIP-712 Domain (必须与 Relayer 配置的 token 信息一致)
// 默认 USDT 预编译使用 name: "Tether USD", version: "1"
// 自定义 ERC20 需要使用合约定义的 name 和 version
const domain = {
  name: 'Tether USD',  // 必须与合约的 name() 一致
  version: '1',        // 必须与合约的 version 一致
  chainId: 71603,      // AESC 链 ID
  verifyingContract: '0x0000000000000000000000000000000000001010'  // USDT 预编译地址
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

// 签名数据
const message = { from, to, value, validAfter, validBefore, nonce };

// 签名
const signature = await wallet.signTypedData(domain, types, message);
const { v, r, s } = ethers.Signature.from(signature);
```

### 步骤 4: 构建支付 Payload

```javascript
const payload = {
  x402Version: 1,
  scheme: 'exact',
  network: 'eip155:71603',  // 必须与 Relayer 的 network_id 一致
  payload: {
    from: from,
    to: to,           // 必须与 payTo 地址一致！
    value: value.toString(),  // 必须 >= maxAmountRequired
    validAfter: validAfter.toString(),
    validBefore: validBefore.toString(),
    nonce: ethers.hexlify(nonce),
    v: v,
    r: ethers.hexlify(r),
    s: ethers.hexlify(s)
  }
};

const paymentBase64 = btoa(JSON.stringify(payload));
```

### 步骤 5: 发送中继请求

```javascript
const response = await fetch('http://localhost:8402/relay', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-PAYMENT': paymentBase64
  },
  body: JSON.stringify({ signedTx })
});

if (response.status === 402) {
  // 支付验证失败
  const error = await response.json();
  console.error('Payment failed:', error);
} else if (response.ok) {
  // 成功
  const result = await response.json();
  console.log('Transaction hash:', result.txHash);
  console.log('Gas used:', result.gasUsed);
}
```

## 完整示例 (JavaScript)

```javascript
const { ethers } = require('ethers');

async function relayTransaction() {
  const RELAYER_URL = 'http://localhost:8402';
  const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
  const CHAIN_ID = 71603;  // AESC 链 ID
  
  const wallet = new ethers.Wallet(PRIVATE_KEY);
  
  // 1. 获取支付要求
  const reqRes = await fetch(`${RELAYER_URL}/payment-requirements`);
  const { accepts } = await reqRes.json();
  const paymentReq = accepts[0];
  
  // 2. 签名用户交易
  const tx = {
    to: '0x0000000000000000000000000000000000000001',
    value: 0n,
    nonce: 0,
    gasLimit: 21000,
    gasPrice: ethers.parseUnits('1', 'gwei'),
    chainId: CHAIN_ID
  };
  const signedTx = await wallet.signTransaction(tx);
  
  // 3. 创建 EIP-3009 签名 (见步骤3)
  // ...
  
  // 4. 发送中继请求
  const res = await fetch(`${RELAYER_URL}/relay`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-PAYMENT': paymentBase64
    },
    body: JSON.stringify({ signedTx })
  });
  
  return await res.json();
}
```

## 错误处理

| HTTP 状态 | 含义 | 处理建议 |
|-----------|------|---------|
| 200 | 成功 | 保存 txHash |
| 402 | 支付验证失败 | 检查签名/余额/过期时间 |
| 400 | 请求格式错误 | 检查请求体格式 |
| 500 | 服务器错误 | 重试或联系运维 |

## 常见问题

**Q: 如何获取 USDT？**
A: 通过 BSC → AESC 跨链桥转入。

**Q: 签名过期怎么办？**
A: 重新创建 EIP-3009 签名，增大 `validBefore`。

**Q: nonce 冲突怎么办？**
A: EIP-3009 的 nonce 是随机的 bytes32，冲突概率极低。如发生冲突，重新生成即可。

