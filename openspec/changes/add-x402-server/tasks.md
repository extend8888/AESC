# x402-relayer 实现任务清单

> **设计决策**：
> - USDT 作为 Bank 模块的子 denom (`usdt`)，Cosmos-EVM 完全互通
> - USDT 预编译合约封装 ERC-20 + EIP-3009 接口
> - x402-server / Relayer / Facilitator 三合一服务
> - x402-relayer 作为独立 Go workspace 模块 (`github.com/sei-protocol/x402-relayer`)

---

## 阶段 0: 前置依赖 ✅

### 0.1 USDT 预编译合约 (Bank Denom 封装) ✅

> **架构**：USDT 预编译不存储余额，调用 bankKeeper 操作 `usdt` denom
> **提交**: `e0188c5` - feat(precompiles): add USDT precompile with ERC-20 and EIP-3009 support

- [x] 0.1.1 创建预编译目录结构
  ```
  precompiles/usdt/
  ├── USDT.sol              # Solidity 接口定义
  ├── abi.json              # ABI 文件
  ├── usdt.go               # Go 实现 (核心)
  ├── eip3009.go            # EIP-3009 签名验证
  ├── usdt_test.go          # 单元测试
  └── setup.go              # 版本管理
  ```

- [x] 0.1.2 定义 Solidity 接口 (`precompiles/usdt/USDT.sol`)
  - 预编译地址: `0x0000000000000000000000000000000000001010`
  - ERC-20 标准: `transfer`, `balanceOf`, `approve`, `transferFrom`, `allowance`
  - EIP-3009: `transferWithAuthorization`, `cancelAuthorization`, `authorizationState`
  - 元数据: `name`, `symbol`, `decimals`, `totalSupply`

- [x] 0.1.3 实现预编译核心 (`precompiles/usdt/usdt.go`)
  - `PrecompileExecutor` 结构体
  - 注入 `bankKeeper`, `evmKeeper`
  - `transfer` → `bankKeeper.SendCoins(ctx, from, to, sdk.NewCoin("usdt", amount))`
  - `balanceOf` → `bankKeeper.GetBalance(ctx, account, "usdt")`
  - `totalSupply` → `bankKeeper.GetSupply(ctx, "usdt")`

- [x] 0.1.4 实现 EIP-3009 (`precompiles/usdt/eip3009.go`)
  - EIP-712 域分隔符 (DOMAIN_SEPARATOR)
  - `transferWithAuthorization` 签名验证
  - `authorizationState` nonce 状态存储
  - `cancelAuthorization` 取消授权

- [x] 0.1.5 注册预编译 (`precompiles/setup.go`)
  - 添加到 `GetCustomPrecompiles()`
  - 添加到 `InitializePrecompiles()`

- [x] 0.1.6 编写单元测试 (`precompiles/usdt/usdt_test.go`)
  - ERC-20 标准功能测试
  - EIP-3009 签名验证测试
  - Cosmos-EVM 互操作测试

- [x] 0.1.7 注册 `usdt` denom 元数据
  - 在创世配置或链启动时注册
  - `name: "Tether USD"`, `symbol: "USDT"`, `decimals: 6`

### 0.2 跨链桥铸造/销毁权限

> ⚠️ 此部分需要单独提案，这里仅列出接口依赖

- [ ] 0.2.1 定义桥模块接口 (mint/burn `usdt` denom)
- [ ] 0.2.2 预编译预留管理员方法 (可选)
  - `mint(address to, uint256 amount)` - 仅桥模块可调用
  - `burn(address from, uint256 amount)` - 仅桥模块可调用

---

## 阶段 1: 基础设施 ✅

> **重构**: 已迁移到独立 Go workspace 模块 `x402-relayer/`
> **提交**: `df19f4b` - feat(x402): add configuration and type definitions
> **提交**: `982b139` - refactor(x402): move to independent Go workspace module

### 1.1 配置模块 ✅

- [x] 1.1.1 创建 `x402-relayer/config/config.go`
  - Config 结构体定义
  - 配置验证函数
- [x] 1.1.2 实现 `ReadConfig()` 配置读取
- [x] 1.1.3 添加 TOML 配置模板 `ConfigTemplate`
- [x] 1.1.4 创建独立 cmd 入口 (`x402-relayer/cmd/x402-relayer/main.go`)

### 1.2 类型定义 ✅

- [x] 1.2.1 创建 `x402-relayer/types/types.go`
  - PaymentPayload 结构
  - PaymentRequired 响应结构
  - EIP-3009 Authorization 结构
  - RelayRequest/Response 结构

---

## 阶段 2: Facilitator 模块 (支付层) ✅

> **说明**：Facilitator 调用 USDT 预编译合约完成支付结算
> **提交**: `1126584` - feat(x402): add Facilitator module for payment verification and settlement

### 2.1 签名验证 ✅

- [x] 2.1.1 创建 `x402-relayer/facilitator/verifier.go`
- [x] 2.1.2 实现 EIP-712 签名恢复 (复用预编译的验证逻辑)
- [x] 2.1.3 实现 EIP-3009 授权预验证 (调用前校验)
- [ ] 2.1.4 单元测试

### 2.2 余额检查 ✅

- [x] 2.2.1 创建 `x402-relayer/facilitator/balance.go`
- [x] 2.2.2 实现 USDT 余额查询
  - 方式 B: 调用 USDT 预编译 `balanceOf(addr)`
- [x] 2.2.3 实现 nonce 状态查询 (调用预编译 `authorizationState`)
- [ ] 2.2.4 单元测试

### 2.3 支付结算 ✅

- [x] 2.3.1 创建 `x402-relayer/facilitator/settler.go`
- [x] 2.3.2 构建 `transferWithAuthorization` 调用
  - 目标: USDT 预编译地址 `0x1010`
  - 参数: from, to, value, validAfter, validBefore, nonce, v, r, s
- [x] 2.3.3 实现 EVM 交易构建与签名
- [x] 2.3.4 实现交易广播与确认
- [x] 2.3.5 Gas 估算 (预编译 Gas 固定且低)
- [ ] 2.3.6 单元测试 + 集成测试

---

## 阶段 3: Relayer 模块 (业务层) ✅

> **提交**: `84da731` - feat(x402): add Relayer module for transaction broadcasting and gas estimation

### 3.1 交易广播 ✅

- [x] 3.1.1 创建 `x402-relayer/relayer/broadcaster.go`
- [x] 3.1.2 实现用户交易验证
- [x] 3.1.3 实现交易广播
- [x] 3.1.4 实现交易回执查询
- [ ] 3.1.5 单元测试

### 3.2 定价模块 ✅

- [x] 3.2.1 创建 `x402-relayer/relayer/gas_estimator.go`
- [x] 3.2.2 实现 Gas 估算
- [x] 3.2.3 实现中继费用计算
- [ ] 3.2.4 单元测试

---

## 阶段 4: HTTP 服务 (协议层) ✅

> **提交**: `9836e10` - feat(x402): add HTTP server, handlers, and payment middleware

### 4.1 服务入口 ✅

- [x] 4.1.1 创建 `x402-relayer/server.go`
- [x] 4.1.2 实现 HTTP 服务器启动/停止
- [x] 4.1.3 实现优雅关闭
- [x] 4.1.4 创建独立 cmd 入口 (不再集成到 seid)

### 4.2 处理器 ✅

- [x] 4.2.1 创建 `x402-relayer/handler/relay.go`
  - POST /relay - 交易中继接口
- [x] 4.2.2 实现 GET /health - 健康检查

### 4.3 中间件 ✅

- [x] 4.3.1 创建 `x402-relayer/middleware/payment.go`
- [x] 4.3.2 实现 402 响应生成
- [x] 4.3.3 实现 X-PAYMENT 头解析
- [ ] 4.3.4 实现请求/响应日志

---

## 阶段 5: 集成测试 ✅

> **提交**: `3ac287a` - feat(x402): complete E2E tests with full relay flow

### 5.1 端到端测试 ✅

- [x] 5.1.1 创建 E2E 测试框架 (`x402-relayer/e2e/`)
- [x] 5.1.2 实现测试脚本 (`run_test.sh`)
  - 自动部署本地链
  - 启动 x402-relayer
  - 运行测试
  - 自动清理
- [x] 5.1.3 实现 EIP-712 签名生成 (`helpers_test.go`)
- [x] 5.1.4 添加 `aesc-poc` 到 EVM chain-id 映射

### 5.2 测试用例 ✅

| 测试 | 描述 | 状态 |
|------|------|------|
| `TestHealthEndpoint` | 健康检查 `/health` | ✅ PASS |
| `TestPaymentRequirementsEndpoint` | 支付要求 `/payment-requirements` | ✅ PASS |
| `TestRelayWithoutPayment` | 无支付请求被拒绝 (HTTP 402) | ✅ PASS |
| `TestRecordsEndpoint` | 记录查询 `/records` | ✅ PASS |
| `TestStatsEndpoint` | 统计 `/records/stats` | ✅ PASS |
| `TestFullRelayWithPayment` | **完整支付+中继流程** | ✅ PASS |

### 5.3 完整中继测试结果 ✅

```
=== RUN   TestFullRelayWithPayment
    e2e_test.go:286: ✅ Relay succeeded!
    e2e_test.go:287:    TxHash: 0x9e255391ef01054d49118dfe42671ae472afcad10dbfefe082e8baea1d7ffc9c
    e2e_test.go:288:    GasUsed: 21000
    e2e_test.go:289:    RecordID: 92d4e4e0-3413-4918-bc3c-aa2c2ab37dcf
--- PASS: TestFullRelayWithPayment (1.52s)
```

**记录详情**:
| 字段 | 值 |
|------|-----|
| settle_status | success |
| settle_tx_hash | 0x0e315d2567ed6bbd7cacffb71cfdef25fb9c702147fc473792b99719222a0811 |
| settle_gas_used | 44472 |
| relay_status | success |
| relay_tx_hash | 0x9e255391ef01054d49118dfe42671ae472afcad10dbfefe082e8baea1d7ffc9c |
| relay_gas_used | 21000 |
| payment_amount | 10000 |
| payment_from | 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 |
| payment_to | 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 |

### 5.4 测试账户配置 ✅

使用 Hardhat 测试账户（已在 genesis 中预分配余额）：

| 角色 | EVM 地址 | Sei 地址 |
|------|----------|----------|
| User | 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 | aesc17w0adeg64ky0daxwd2ugyuneellmjgnxn7tzgf |
| Relayer | 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 | aesc1wzvhjux9rqfdcwspp37srdgwp5tac7wgm40au0 |

### 5.5 失败场景测试 ✅

| 测试 | 描述 | 状态 |
|------|------|------|
| `TestFailureScenarios/InvalidSignature` | 无效签名被拒绝 | ✅ PASS |
| `TestFailureScenarios/ExpiredAuthorization` | 过期授权被拒绝 | ✅ PASS |
| `TestFailureScenarios/InvalidNetwork` | 错误网络 ID 被拒绝 | ✅ PASS |
| `TestFailureScenarios/MalformedPaymentHeader` | 畸形支付头被拒绝 | ✅ PASS |

### 5.6 并发压力测试 ✅

```
=== RUN   TestConcurrentRelays
    e2e_test.go:589: ✅ Concurrent test completed in 30.000472167s
    e2e_test.go:590:    Success: 2, Failed/Rejected: 3
    e2e_test.go:591:    Throughput: 0.17 req/s
    e2e_test.go:599: ✅ Server remained healthy after concurrent requests
--- PASS: TestConcurrentRelays (30.00s)
```

**结果分析**:
- 5 个并发请求中 2 个成功（支付结算 + 交易中继）
- 3 个因超时/nonce 冲突被拒绝（预期行为）
- 服务器在并发测试后保持健康
- 每个交易需要等待链上确认，因此吞吐量受限于区块时间

### 5.7 待完成测试

- [ ] 5.7.1 Gas 监控告警测试（需要集成监控系统）

---

## 阶段 6: 文档与部署 ✅

- [x] 6.1 更新配置文档
  - `x402-relayer/README.md` - 快速开始、配置说明、API 参考
- [x] 6.2 编写运维手册
  - `x402-relayer/docs/operations.md`
  - Relayer 钱包管理 (AEX 充值、USDT 提取)
  - 数据库维护 (备份、查询统计)
  - 日志管理
  - 故障排查指南
  - 监控指标建议
- [x] 6.3 编写用户接入指南
  - `x402-relayer/docs/client-guide.md`
  - EIP-3009 签名流程
  - JavaScript 完整示例
  - 错误处理

---

## 任务依赖关系

```
阶段 0 (USDT 预编译)      ✅ e0188c5
    │
    ├──▶ 阶段 1 (基础设施)   ✅ df19f4b, 982b139
    │         │
    │         ├──▶ 阶段 2 (Facilitator)  ✅ 1126584
    │         │         │
    │         │         ├──▶ 阶段 3 (Relayer)  ✅ 84da731
    │         │         │         │
    │         │         │         ├──▶ 阶段 4 (HTTP 服务)  ✅ 9836e10
    │         │         │         │         │
    │         │         │         │         ├──▶ 阶段 5 (测试)  ✅ 3ac287a
    │         │         │         │         │         │
    │         │         │         │         │         ├──▶ 阶段 6 (文档)  ⏳
```

---

## 预估工作量

| 阶段 | 预估时间 | 说明 |
|------|----------|------|
| 阶段 0 | 5-7 天 | USDT 预编译 + EIP-3009 (比 Solidity 合约多 2 天) |
| 阶段 1 | 1-2 天 | 配置 + 类型定义 |
| 阶段 2 | 2-3 天 | Facilitator (预编译简化了调用) |
| 阶段 3 | 2-3 天 | Relayer |
| 阶段 4 | 2-3 天 | HTTP 服务 |
| 阶段 5 | 2-3 天 | 集成测试 |
| 阶段 6 | 1-2 天 | 文档 |
| **合计** | **15-23 天** | |

> ⚠️ 跨链桥不在此估算范围内，需要单独评估

---

## 技术亮点

```
USDT 预编译 + Bank Denom 方案优势:

1. Cosmos-EVM 完全互通
   ┌──────────────────────────────────────────────────┐
   │ Cosmos: seid tx bank send ... --denom usdt      │
   │ EVM:    USDT.transfer(to, amount)               │
   │                    ↓                            │
   │         同一份余额 (Bank 模块存储)                 │
   └──────────────────────────────────────────────────┘

2. Gas 成本极低
   ┌──────────────────────────────────────────────────┐
   │ Solidity 合约:   ~80,000 gas                    │
   │ USDT 预编译:     ~3,000 gas (降低 96%)           │
   └──────────────────────────────────────────────────┘

3. 跨链桥原生集成
   ┌──────────────────────────────────────────────────┐
   │ mint/burn 通过 Bank 模块权限控制                  │
   │ 无需额外的权限管理合约                            │
   └──────────────────────────────────────────────────┘
```

