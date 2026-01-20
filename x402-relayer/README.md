# x402-relayer

基于 [Coinbase x402](https://github.com/coinbase/x402) 协议的交易中继服务。

用户可以用 **USDT 支付中继费**，由 x402-relayer 代为广播交易并支付 AEX Gas。

## 快速开始

### 1. 配置文件

创建 `config.toml`:

```toml
[x402-relayer]
enabled = true
port = 8402
pay_to_address = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
usdt_precompile = "0x0000000000000000000000000000000000001010"
usdt_denom = "usdt"
network_id = "eip155:713715"
private_key = "${X402_RELAYER_KEY}"
relay_fee_per_tx = "10000"
evm_rpc = "http://localhost:8545"
db_path = "./x402-relayer.db"
```

### 2. 启动服务

```bash
# 设置私钥环境变量
export X402_RELAYER_KEY="59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

# 启动服务
./x402-relayer --config config.toml
```

### 3. 测试健康检查

```bash
curl http://localhost:8402/health
# {"status":"ok","timestamp":1737385601}
```

## 配置说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `enabled` | 是否启用服务 | `false` |
| `port` | HTTP 服务端口 | `8402` |
| `pay_to_address` | 收款钱包地址 (EVM) | 必填 |
| `usdt_precompile` | USDT 预编译地址 | `0x...1010` |
| `usdt_denom` | Bank 模块 denom | `usdt` |
| `network_id` | CAIP-2 网络 ID | 必填 |
| `private_key` | Relayer 私钥 | 必填 |
| `relay_fee_per_tx` | 每笔中继费 (USDT 最小单位) | `10000` (0.01 USDT) |
| `evm_rpc` | EVM RPC 地址 | `http://localhost:8545` |
| `db_path` | SQLite 数据库路径 | `./x402-relayer.db` |

### 私钥配置

支持两种方式：

```toml
# 方式 1: 直接配置 (不推荐)
private_key = "59c6995e..."

# 方式 2: 环境变量引用 (推荐)
private_key = "${X402_RELAYER_KEY}"
```

## API 接口

### GET /health

健康检查。

```bash
curl http://localhost:8402/health
```

### GET /payment-requirements

获取支付要求。

```bash
curl http://localhost:8402/payment-requirements
```

响应：
```json
{
  "accepts": [{
    "scheme": "exact",
    "network": "eip155:713715",
    "maxAmountRequired": "10000",
    "asset": "0x0000000000000000000000000000000000001010",
    "payTo": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    "requiredDeadlineSeconds": 300
  }]
}
```

### POST /relay

提交交易中继请求。

**无支付头**：返回 HTTP 402

```bash
curl -X POST http://localhost:8402/relay \
  -H "Content-Type: application/json" \
  -d '{"signedTx": "0x..."}'
# HTTP 402 Payment Required
```

**有支付头**：执行中继

```bash
curl -X POST http://localhost:8402/relay \
  -H "Content-Type: application/json" \
  -H "X-PAYMENT: <base64_payment_payload>" \
  -d '{"signedTx": "0x..."}'
```

响应：
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

```bash
curl "http://localhost:8402/records?limit=10&offset=0"
```

### GET /stats

获取统计信息。

```bash
curl http://localhost:8402/stats
```

## 运行测试

```bash
# 运行 E2E 测试
./x402-relayer/e2e/run_test.sh
```

## 架构

```
x402-relayer
├── config/       # 配置管理
├── facilitator/  # 支付层：EIP-712 验证、USDT 结算
├── handler/      # HTTP 处理器
├── middleware/   # 支付中间件
├── relayer/      # 业务层：交易广播
├── storage/      # SQLite 存储
└── types/        # 类型定义
```

## 许可证

Apache-2.0

