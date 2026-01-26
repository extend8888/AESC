## ADDED Requirements

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

