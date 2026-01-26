# 链标识规格

> 归档自变更：`rebrand-sei-to-aesc` (2026-01-19)

## Purpose

定义 AESC 链的品牌标识配置，包括代币面额、地址前缀、链名称和 EVM Chain ID 映射。
## Requirements
### Requirement: 链基础标识配置
系统必须（SHALL）使用 AESC 品牌标识作为链的基础配置，包括代币面额、地址前缀和链名称。

#### Scenario: 代币面额配置正确
- **GIVEN** 系统已初始化
- **WHEN** 查询链的基础代币配置
- **THEN** 基础代币单位应为 `uaex`
- **AND** 标准代币单位应为 `aex`
- **AND** 代币精度应为 6（1 aex = 10^6 uaex）

#### Scenario: 地址前缀配置正确
- **GIVEN** 系统已初始化
- **WHEN** 生成新的账户地址
- **THEN** 地址应以 `aesc1` 开头
- **AND** 验证者地址应以 `aescvaloper` 开头
- **AND** 共识地址应以 `aescvalcons` 开头

#### Scenario: Bech32 地址验证
- **GIVEN** 一个 `aesc1` 开头的地址
- **WHEN** 系统验证该地址
- **THEN** 验证应该通过
- **AND** 地址长度应符合 Bech32 规范

#### Scenario: 拒绝旧的 Sei 地址格式
- **GIVEN** 一个 `sei1` 开头的地址
- **WHEN** 系统验证该地址
- **THEN** 验证应该失败
- **AND** 返回地址前缀不匹配的错误

### Requirement: Genesis 配置
系统必须（SHALL）在 genesis 配置中使用正确的 AESC 标识符。

#### Scenario: Genesis 代币配置
- **GIVEN** 一个新的 genesis 配置文件
- **WHEN** 检查代币元数据配置
- **THEN** base denom 应为 `uaex`
- **AND** display denom 应为 `aex`
- **AND** 代币名称应为 "AESC Gas Token"
- **AND** 代币符号应为 "AEX"

#### Scenario: Genesis 账户地址
- **GIVEN** genesis 配置中的账户列表
- **WHEN** 检查所有账户地址
- **THEN** 所有地址都应以 `aesc1` 开头
- **AND** 不应存在 `sei1` 开头的地址

#### Scenario: Genesis 余额配置
- **GIVEN** genesis 配置中的余额列表
- **WHEN** 检查余额的代币面额
- **THEN** 所有余额都应使用 `uaex` 作为面额
- **AND** 不应存在 `usei` 面额的余额

### Requirement: 最小 Gas 价格配置
系统必须（SHALL）使用 `uaex` 作为最小 Gas 价格的面额单位。

#### Scenario: 默认最小 Gas 价格
- **GIVEN** 节点使用默认配置启动
- **WHEN** 查询最小 Gas 价格配置
- **THEN** 最小 Gas 价格应为 `0.02uaex`

#### Scenario: 自定义最小 Gas 价格
- **GIVEN** 节点配置了自定义最小 Gas 价格
- **WHEN** 配置值为 `0.05uaex`
- **THEN** 系统应接受该配置
- **AND** 交易的 Gas 价格必须不低于此值

#### Scenario: 拒绝错误的面额
- **GIVEN** 节点配置了最小 Gas 价格
- **WHEN** 配置值使用 `usei` 面额
- **THEN** 系统应拒绝该配置
- **OR** 在启动时显示警告

### Requirement: EVM 模块代币配置
系统必须（SHALL）在 EVM 模块中使用 `uaex` 作为基础代币。

#### Scenario: EVM 交易 Gas 费用
- **GIVEN** 用户发起一笔 EVM 交易
- **WHEN** 计算 Gas 费用
- **THEN** Gas 费用应以 `uaex` 计价
- **AND** 从用户账户扣除的代币应为 `uaex`

#### Scenario: EVM 余额查询
- **GIVEN** 用户通过 EVM RPC 查询余额
- **WHEN** 调用 eth_getBalance
- **THEN** 返回的余额应对应 `uaex` 数量
- **AND** 转换率应为 1 wei = 1 uaex

#### Scenario: EVM 代币转账
- **GIVEN** 用户通过 EVM 发起代币转账
- **WHEN** 转账 1 ETH（在 AESC 中对应 1 AEX）
- **THEN** 实际转账的代币应为 1000000 uaex
- **AND** 接收方余额增加 1000000 uaex

### Requirement: AESC EVM Chain ID 配置

系统必须（MUST）为 AESC 链的不同网络环境提供独立的 EVM Chain ID 配置。

EVM Chain ID 映射规则：

| Cosmos Chain ID | EVM Chain ID | 用途 |
|-----------------|--------------|------|
| `aesc-mainnet-1` | `71600` | 主网（生产环境） |
| `aesc-testnet-1` | `71601` | 测试网（公开测试环境） |
| `aesc-devnet-1` | `71602` | 开发网（内部开发环境） |
| `aesc-poc` | `71603` | 本地测试（保持现有脚本兼容） |

默认 Chain ID（当 Cosmos Chain ID 未在映射表中时）应为 `71600`。

**环境边界说明**：
- `aesc-poc`：用于本地单节点测试，与现有 poc-deploy 脚本兼容
- `aesc-devnet-1`：用于多节点开发网络部署

#### Scenario: 主网 EVM Chain ID

- **GIVEN** Cosmos Chain ID 为 `aesc-mainnet-1`
- **WHEN** 调用 `GetEVMChainID("aesc-mainnet-1")`
- **THEN** 返回 `71600`

#### Scenario: 测试网 EVM Chain ID

- **GIVEN** Cosmos Chain ID 为 `aesc-testnet-1`
- **WHEN** 调用 `GetEVMChainID("aesc-testnet-1")`
- **THEN** 返回 `71601`

#### Scenario: 开发网 EVM Chain ID

- **GIVEN** Cosmos Chain ID 为 `aesc-devnet-1`
- **WHEN** 调用 `GetEVMChainID("aesc-devnet-1")`
- **THEN** 返回 `71602`

#### Scenario: 本地测试 EVM Chain ID（兼容现有脚本）

- **GIVEN** Cosmos Chain ID 为 `aesc-poc`
- **WHEN** 调用 `GetEVMChainID("aesc-poc")`
- **THEN** 返回 `71603`

#### Scenario: 未知 Chain ID 使用默认值

- **GIVEN** Cosmos Chain ID 不在映射表中
- **WHEN** 调用 `GetEVMChainID("unknown-chain")`
- **THEN** 返回默认值 `71600`

#### Scenario: EVM RPC 返回正确的 Chain ID（本地测试）

- **GIVEN** 节点以 `aesc-poc` Cosmos Chain ID 启动
- **WHEN** 调用 EVM RPC `eth_chainId`
- **THEN** 返回 `0x117B3`（71603 的十六进制表示）

#### Scenario: 主网防护阻止 mockBalance 调用

- **GIVEN** 节点以 `aesc-mainnet-1` Cosmos Chain ID 启动
- **WHEN** 调用 `mockBalance` 函数
- **THEN** 触发 panic 阻止执行，保护主网安全

#### Scenario: 未知链不触发主网防护

- **GIVEN** 节点以未知 Cosmos Chain ID（如 `custom-test`）启动
- **WHEN** 调用 `mockBalance` 函数
- **THEN** 正常执行，不触发 panic（因为仅判断 Cosmos Chain ID 是否为 `aesc-mainnet-1`）

