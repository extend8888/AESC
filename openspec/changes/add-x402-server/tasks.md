# x402-relayer 实现任务清单

> **设计决策**：
> - USDT 作为 Bank 模块的子 denom (`usdt`)，Cosmos-EVM 完全互通
> - USDT 预编译合约封装 ERC-20 + EIP-3009 接口
> - x402-server / Relayer / Facilitator 三合一服务

---

## 阶段 0: 前置依赖

### 0.1 USDT 预编译合约 (Bank Denom 封装)

> **架构**：USDT 预编译不存储余额，调用 bankKeeper 操作 `usdt` denom

- [ ] 0.1.1 创建预编译目录结构
  ```
  precompiles/usdt/
  ├── USDT.sol              # Solidity 接口定义
  ├── abi.json              # ABI 文件
  ├── usdt.go               # Go 实现 (核心)
  ├── eip3009.go            # EIP-3009 签名验证
  ├── usdt_test.go          # 单元测试
  └── setup.go              # 版本管理
  ```

- [ ] 0.1.2 定义 Solidity 接口 (`precompiles/usdt/USDT.sol`)
  - 预编译地址: `0x0000000000000000000000000000000000001010`
  - ERC-20 标准: `transfer`, `balanceOf`, `approve`, `transferFrom`, `allowance`
  - EIP-3009: `transferWithAuthorization`, `cancelAuthorization`, `authorizationState`
  - 元数据: `name`, `symbol`, `decimals`, `totalSupply`

- [ ] 0.1.3 实现预编译核心 (`precompiles/usdt/usdt.go`)
  - `PrecompileExecutor` 结构体
  - 注入 `bankKeeper`, `evmKeeper`
  - `transfer` → `bankKeeper.SendCoins(ctx, from, to, sdk.NewCoin("usdt", amount))`
  - `balanceOf` → `bankKeeper.GetBalance(ctx, account, "usdt")`
  - `totalSupply` → `bankKeeper.GetSupply(ctx, "usdt")`

- [ ] 0.1.4 实现 EIP-3009 (`precompiles/usdt/eip3009.go`)
  - EIP-712 域分隔符 (DOMAIN_SEPARATOR)
  - `transferWithAuthorization` 签名验证
  - `authorizationState` nonce 状态存储
  - `cancelAuthorization` 取消授权

- [ ] 0.1.5 注册预编译 (`precompiles/setup.go`)
  - 添加到 `GetCustomPrecompiles()`
  - 添加到 `InitializePrecompiles()`

- [ ] 0.1.6 编写单元测试 (`precompiles/usdt/usdt_test.go`)
  - ERC-20 标准功能测试
  - EIP-3009 签名验证测试
  - Cosmos-EVM 互操作测试

- [ ] 0.1.7 注册 `usdt` denom 元数据
  - 在创世配置或链启动时注册
  - `name: "Tether USD"`, `symbol: "USDT"`, `decimals: 6`

### 0.2 跨链桥铸造/销毁权限

> ⚠️ 此部分需要单独提案，这里仅列出接口依赖

- [ ] 0.2.1 定义桥模块接口 (mint/burn `usdt` denom)
- [ ] 0.2.2 预编译预留管理员方法 (可选)
  - `mint(address to, uint256 amount)` - 仅桥模块可调用
  - `burn(address from, uint256 amount)` - 仅桥模块可调用

---

## 阶段 1: 基础设施

### 1.1 配置模块

- [ ] 1.1.1 创建 `services/x402-relayer/config/config.go`
  - Config 结构体定义
  - 配置验证函数
- [ ] 1.1.2 实现 `ReadConfig()` 配置读取
- [ ] 1.1.3 添加 TOML 配置模板 `ConfigTemplate`
- [ ] 1.1.4 集成配置到链启动流程 (`app/seid/cmd/root.go`)

### 1.2 类型定义

- [ ] 1.2.1 创建 `services/x402-relayer/types.go`
  - PaymentPayload 结构
  - PaymentRequired 响应结构
  - EIP-3009 Authorization 结构
  - RelayRequest/Response 结构

---

## 阶段 2: Facilitator 模块 (支付层)

> **说明**：Facilitator 调用 USDT 预编译合约完成支付结算

### 2.1 签名验证

- [ ] 2.1.1 创建 `services/x402-relayer/facilitator/verifier.go`
- [ ] 2.1.2 实现 EIP-712 签名恢复 (复用预编译的验证逻辑)
- [ ] 2.1.3 实现 EIP-3009 授权预验证 (调用前校验)
- [ ] 2.1.4 单元测试

### 2.2 余额检查

- [ ] 2.2.1 创建 `services/x402-relayer/facilitator/balance.go`
- [ ] 2.2.2 实现 USDT 余额查询
  - 方式 A: 直接查 Bank 模块 `bankKeeper.GetBalance(ctx, addr, "usdt")`
  - 方式 B: 调用 USDT 预编译 `balanceOf(addr)`
- [ ] 2.2.3 实现 nonce 状态查询 (调用预编译 `authorizationState`)
- [ ] 2.2.4 单元测试

### 2.3 支付结算

- [ ] 2.3.1 创建 `services/x402-relayer/facilitator/settler.go`
- [ ] 2.3.2 构建 `transferWithAuthorization` 调用
  - 目标: USDT 预编译地址 `0x1010`
  - 参数: from, to, value, validAfter, validBefore, nonce, v, r, s
- [ ] 2.3.3 实现 EVM 交易构建与签名
- [ ] 2.3.4 实现交易广播与确认
- [ ] 2.3.5 Gas 估算 (预编译 Gas 固定且低)
- [ ] 2.3.6 单元测试 + 集成测试

---

## 阶段 3: Relayer 模块 (业务层)

### 3.1 交易广播

- [ ] 3.1.1 创建 `services/x402-relayer/relayer/broadcaster.go`
- [ ] 3.1.2 实现用户交易验证
- [ ] 3.1.3 实现交易广播
- [ ] 3.1.4 实现交易回执查询
- [ ] 3.1.5 单元测试

### 3.2 定价模块

- [ ] 3.2.1 创建 `services/x402-relayer/relayer/gas_estimator.go`
- [ ] 3.2.2 实现 Gas 估算
- [ ] 3.2.3 实现中继费用计算
- [ ] 3.2.4 单元测试

---

## 阶段 4: HTTP 服务 (协议层)

### 4.1 服务入口

- [ ] 4.1.1 创建 `services/x402-relayer/server.go`
- [ ] 4.1.2 实现 HTTP 服务器启动/停止
- [ ] 4.1.3 实现优雅关闭
- [ ] 4.1.4 集成到 `seid start` 命令

### 4.2 处理器

- [ ] 4.2.1 创建 `services/x402-relayer/handler/relay.go`
  - POST /relay - 交易中继接口
- [ ] 4.2.2 创建 `services/x402-relayer/handler/health.go`
  - GET /health - 健康检查
  - GET /info - 服务信息 (支持的代币、费用等)

### 4.3 中间件

- [ ] 4.3.1 创建 `services/x402-relayer/middleware/payment.go`
- [ ] 4.3.2 实现 402 响应生成
- [ ] 4.3.3 实现 X-PAYMENT 头解析
- [ ] 4.3.4 实现请求/响应日志

---

## 阶段 5: 集成测试

- [ ] 5.1 端到端测试：完整支付 + 中继流程
- [ ] 5.2 压力测试：并发交易处理
- [ ] 5.3 失败场景测试
  - 余额不足
  - 签名无效
  - 交易广播失败
- [ ] 5.4 Gas 监控告警测试

---

## 阶段 6: 文档与部署

- [ ] 6.1 更新配置文档
- [ ] 6.2 编写运维手册
  - Facilitator 钱包 AEX 充值
  - 监控指标说明
  - 故障排查指南
- [ ] 6.3 编写用户接入指南 (客户端如何调用)

---

## 任务依赖关系

```
阶段 0 (USDT 预编译)
    │
    ├──▶ 阶段 1 (基础设施)
    │         │
    │         ├──▶ 阶段 2 (Facilitator)
    │         │         │
    │         │         ├──▶ 阶段 3 (Relayer)
    │         │         │         │
    │         │         │         ├──▶ 阶段 4 (HTTP 服务)
    │         │         │         │         │
    │         │         │         │         ├──▶ 阶段 5 (测试)
    │         │         │         │         │         │
    │         │         │         │         │         ├──▶ 阶段 6 (文档)
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

