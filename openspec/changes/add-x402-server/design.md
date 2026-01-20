# x402-relayer 设计文档

## 背景

x402 是基于 HTTP 402 状态码的支付协议，由 Coinbase 开发并开源。

AESC Chain 需要提供交易中继服务，让没有 AEX 的用户可以用 USDT 支付中继费，由服务代为广播交易。

## 术语说明

在标准 x402 协议中有三个角色，在本设计中合并为一个服务：

```
+-------------------------------------------------------------------+
|                      x402-relayer (:8402)                          |
|                                                                    |
|  +------------------+  +------------------+  +------------------+  |
|  |   x402-server    |  |     Relayer      |  |   Facilitator    |  |
|  |   (协议层)        |  |    (业务层)       |  |    (支付层)       |  |
|  +------------------+  +------------------+  +------------------+  |
|  | - HTTP 服务       |  | - 验证用户交易    |  | - 验证 EIP-712   |  |
|  | - 返回 402       |  | - 广播交易        |  | - 检查 USDT 余额  |  |
|  | - 解析 X-PAYMENT |  | - 返回交易回执    |  | - 调用结算合约    |  |
|  +------------------+  +------------------+  +------------------+  |
|                                                                    |
+-------------------------------------------------------------------+
```

| 角色 | 职责 | 本设计中 |
|------|------|----------|
| **x402-server** | HTTP 402 协议框架 | 合并 |
| **Relayer** | 业务逻辑：广播用户交易 | 合并 |
| **Facilitator** | 支付逻辑：验证签名、结算 USDT | 合并 |

> **命名**：本服务统一命名为 **x402-relayer**，强调其核心功能是交易中继。

## 设计决策

| 决策项 | 选择 | 说明 |
|--------|------|------|
| 服务名称 | **x402-relayer** | 三合一服务 |
| 支付代币 | **USDT** | 从 BSC 跨链到 AESC |
| Facilitator | **自建** | 集成在同一服务中 |
| USDT 实现 | **预编译合约** | Bank Denom 封装，Cosmos-EVM 互通 |
| USDT Denom | **usdt** | Bank 模块统一管理 |
| 跨链桥 | **BSC → AESC** | 需要单独实现 |

## USDT 预编译方案

```
+------------------------------------------------------------------+
|                        Bank 模块 (x/bank)                         |
|                                                                   |
|  ┌─────────────────────────────────────────────────────────────┐ |
|  │                     余额存储 (唯一数据源)                      │ |
|  │   denom: "usdt"                                              │ |
|  │   address_1: 500                                             │ |
|  │   address_2: 1000                                            │ |
|  └─────────────────────────────────────────────────────────────┘ |
|                              │                                    |
|              ┌───────────────┴───────────────┐                   |
|              ▼                               ▼                   |
|  +---------------------+         +------------------------+      |
|  |  Cosmos 端           |         |  EVM 端                 |      |
|  +---------------------+         +------------------------+      |
|  | seid tx bank send   |         | USDT 预编译 (0x1010)    |      |
|  | --denom usdt        |         | - transfer()           |      |
|  |                     |         | - balanceOf()          |      |
|  |                     |         | - transferWithAuth()   |      |
|  +---------------------+         +------------------------+      |
|                                                                   |
+------------------------------------------------------------------+
```

### 预编译优势

| 对比项 | Solidity 合约 | 预编译合约 |
|--------|--------------|-----------|
| Gas 消耗 | ~80,000 | ~3,000 (降低 96%) |
| Cosmos 可见 | ❌ | ✅ 原生可见 |
| CLI 转账 | ❌ | ✅ `seid tx bank send --denom usdt` |
| IBC 跨链 | ❌ 需要适配 | ✅ 原生支持 |

## 目标

- 实现 x402-relayer 服务（x402-server + Relayer + Facilitator 三合一）
- 实现 USDT 预编译合约（ERC-20 + EIP-3009，封装 Bank Denom）
- 提供可配置的启用/禁用开关
- 与 AESC Chain 节点生命周期无缝集成

## 非目标

- 不实现 x402 客户端（买方）功能
- 不支持 Solana 网络（AESC 为 EVM 链）
- 不处理复杂的定价逻辑（由应用层决定）
- 跨链桥不在本提案范围内（单独提案）

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AESC Chain Node                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐    ┌──────────────────────────────────┐    ┌─────────────┐ │
│  │   EVM RPC   │    │        x402-relayer              │    │  Tendermint │ │
│  │   :8545     │    │          :8402                   │    │    RPC      │ │
│  └─────────────┘    ├──────────────────────────────────┤    └─────────────┘ │
│                     │ ┌──────────────────────────────┐ │                    │
│                     │ │  x402-server (协议层)         │ │                    │
│                     │ │  - HTTP 服务 / 返回 402      │ │                    │
│                     │ ├──────────────────────────────┤ │                    │
│                     │ │  Relayer (业务层)            │ │                    │
│                     │ │  - 验证用户交易 / 广播交易    │ │                    │
│                     │ ├──────────────────────────────┤ │                    │
│                     │ │  Facilitator (支付层)        │ │                    │
│                     │ │  - 验证签名 / 结算 USDT      │ │                    │
│                     │ └──────────────┬───────────────┘ │                    │
│                     └────────────────┼─────────────────┘                    │
│                                      │                                      │
│                                      ▼                                      │
│                          ┌───────────────────────────┐                      │
│                          │  USDT 合约 (EIP-3009)     │                      │
│                          │  transferWithAuthorization │                      │
│                          └───────────────────────────┘                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                       ▲
                                       │ 跨链桥
                                       │
                              ┌────────┴────────┐
                              │   BSC Chain     │
                              │   USDT          │
                              └─────────────────┘
```

## 用户使用流程

```
┌─────────────┐                                    ┌─────────────────────────┐
│    用户      │                                    │      x402-relayer       │
│ (有USDT)    │                                    │                         │
│ (无AEX)     │                                    │                         │
└──────┬──────┘                                    └────────────┬────────────┘
       │                                                        │
       │  1. POST /relay {signedTx}                             │
       │───────────────────────────────────────────────────────▶│
       │                                                        │
       │  2. HTTP 402 + X-PAYMENT-REQUIRED                      │
       │     {amount: "10000", token: "USDT", ...}              │
       │◀───────────────────────────────────────────────────────│
       │                                                        │
       │  [用户离线签名 EIP-712 授权，不需要 AEX]                   │
       │                                                        │
       │  3. POST /relay {signedTx}                             │
       │     + X-PAYMENT: {signature, authorization}            │
       │───────────────────────────────────────────────────────▶│
       │                                                        │
       │                    ┌──────────────────────────────────┐│
       │                    │ 4. [Facilitator] 验证 EIP-712    ││
       │                    │ 5. [Facilitator] 调用合约:        ││
       │                    │    USDT.transferWithAuth()       ││
       │                    │    用户 USDT → x402-relayer      ││
       │                    │ 6. [Relayer] 广播用户交易         ││
       │                    │    (x402-relayer 付 AEX Gas)     ││
       │                    └──────────────────────────────────┘│
       │                                                        │
       │  7. HTTP 200 {txHash, status}                          │
       │◀───────────────────────────────────────────────────────│
       │                                                        │
```

## 核心组件

### 1. USDT 合约 (contracts/usdt/)

自行部署的 USDT 合约，支持 EIP-3009 标准：

```solidity
interface IUSDT_EIP3009 {
    // ERC-20 标准接口
    function transfer(address to, uint256 value) external returns (bool);
    function approve(address spender, uint256 value) external returns (bool);
    function transferFrom(address from, address to, uint256 value) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
    function allowance(address owner, address spender) external view returns (uint256);

    // EIP-3009: 授权转账 (核心功能)
    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;

    // EIP-3009: 取消授权
    function cancelAuthorization(
        address authorizer,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;

    // EIP-3009: 查询 nonce 是否已使用
    function authorizationState(address authorizer, bytes32 nonce) external view returns (bool);

    // EIP-2612: Permit (可选，增强兼容性)
    function permit(
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;

    // 跨链桥铸造/销毁 (仅桥合约可调用)
    function mint(address to, uint256 amount) external;
    function burn(address from, uint256 amount) external;
}
```

### 2. x402-relayer 配置 (services/x402-relayer/config/config.go)

```go
type Config struct {
    // 是否启用 x402 服务
    Enabled bool `mapstructure:"enabled"`

    // HTTP 服务端口
    Port int `mapstructure:"port"`

    // 接收支付的钱包地址 (Relayer 收款地址)
    PayToAddress string `mapstructure:"pay_to_address"`

    // USDT 合约地址 (AESC 链上)
    USDTContract string `mapstructure:"usdt_contract"`

    // 网络标识符 (CAIP-2 格式)
    NetworkID string `mapstructure:"network_id"`

    // Facilitator 私钥 (用于调用 transferWithAuthorization)
    FacilitatorPrivateKey string `mapstructure:"facilitator_private_key"`

    // 交易中继定价 (单位: USDT 最小单位)
    RelayFeePerTx string `mapstructure:"relay_fee_per_tx"`

    // EVM RPC 地址 (用于广播交易)
    EVMRPC string `mapstructure:"evm_rpc"`
}
```

### 3. x402-relayer 核心流程

```
接收请求
    │
    ▼
是否有 X-PAYMENT 头？ ──否──▶ 返回 HTTP 402 + 支付要求
    │
    是
    ▼
验证 EIP-712 签名 ──失败──▶ 返回 401 Unauthorized
    │
    成功
    ▼
检查用户 USDT 余额 ──不足──▶ 返回 402 + 余额不足
    │
    充足
    ▼
调用 USDT.transferWithAuthorization() ──失败──▶ 返回 500
    │
    成功
    ▼
广播用户原交易 ──失败──▶ 返回 500 (可能需要退款逻辑)
    │
    成功
    ▼
返回交易回执
```

### 4. 目录结构

```
precompiles/usdt/                    # USDT 预编译合约 (新增)
├── USDT.sol                         # Solidity 接口定义
├── abi.json                         # ABI 文件
├── usdt.go                          # Go 实现 (调用 bankKeeper)
├── eip3009.go                       # EIP-3009 签名验证
├── usdt_test.go                     # 单元测试
└── setup.go                         # 版本管理

services/x402-relayer/               # x402-relayer 服务
├── config/
│   └── config.go                    # 配置定义
├── server.go                        # HTTP 服务入口
├── types.go                         # x402 协议类型定义
├── handler/
│   ├── relay.go                     # 交易中继处理器
│   └── health.go                    # 健康检查
├── facilitator/
│   ├── verifier.go                  # EIP-712 签名预验证
│   ├── settler.go                   # 支付结算 (调用 USDT 预编译)
│   └── balance.go                   # 余额检查 (查 Bank 或预编译)
├── relayer/
│   ├── broadcaster.go               # 交易广播
│   └── gas_estimator.go             # Gas 估算 & 定价
└── middleware/
    └── payment.go                   # 支付中间件
```

## 配置模板

```toml
[x402-relayer]
# 是否启用 x402-relayer 服务
enabled = false

# x402-relayer 服务端口
port = 8402

# 接收支付的钱包地址 (收款地址)
pay_to_address = "0x..."

# USDT 预编译合约地址 (固定)
usdt_precompile = "0x0000000000000000000000000000000000001010"

# USDT denom (Bank 模块)
usdt_denom = "usdt"

# 网络标识符 (CAIP-2)
network_id = "eip155:CHAIN_ID"

# 服务私钥 (用于调用预编译和广播交易，敏感，建议使用环境变量)
# private_key = "${X402_RELAYER_KEY}"

# 每笔交易中继费用 (单位: USDT 最小单位，6位小数)
# 0.01 USDT = 10000
relay_fee_per_tx = "10000"

# EVM RPC 地址
evm_rpc = "http://localhost:8545"
```

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| Relayer 私钥泄露 | 使用 HSM/KMS；最小权限；定期轮换 |
| AEX Gas 耗尽 | 监控告警；自动补充；设置最低余额阈值 |
| 支付成功但广播失败 | 实现退款逻辑；或延迟结算 |
| 重放攻击 | EIP-3009 nonce 机制已内置防护 |
| 预编译合约漏洞 | 完整测试覆盖；代码审计 |
| 跨链桥风险 | 限制单笔/单日额度；多签治理 |

## 依赖项

| 依赖 | 说明 | 状态 |
|------|------|------|
| USDT 合约 | 支持 EIP-3009 | 需要开发 |
| BSC → AESC 跨链桥 | USDT 桥接 | 需要单独提案 |
| EVM RPC | 交易广播 | 已有 |

## 开放问题 (已解决)

| 问题 | 决策 |
|------|------|
| ~~是否需要支持多个 Facilitator？~~ | 否，自建单一服务 |
| ~~是否需要本地支付验证模式？~~ | 是，自建 Facilitator 就是本地验证 |
| ~~使用 USDC 还是 USDT？~~ | USDT，从 BSC 跨链 |

## 待确认问题

1. 交易中继费用定价策略？（固定费用 / 动态 Gas 相关）
2. 跨链桥方案选择？（第三方 / 自建）
3. USDT 合约是否需要支持升级？

