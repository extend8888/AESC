# 变更：更新 AESC EVM Chain ID 配置

## 为什么

当前 AESC 链的 EVM Chain ID 配置存在问题：
1. `aesc-poc` 使用的是 Sei 的 `arctic-1` 测试网 Chain ID `713715`，这不符合 AESC 独立品牌的要求
2. 缺少 `aesc-mainnet-1` 和 `aesc-testnet-1` 的 EVM Chain ID 映射
3. 需要建立一套完整的、逻辑一致的 AESC EVM Chain ID 体系

## 变更内容

### 1. EVM Chain ID 映射更新

建立 AESC 专属的 EVM Chain ID 体系：

| Cosmos Chain ID | EVM Chain ID | 用途 |
|-----------------|--------------|------|
| `aesc-mainnet-1` | `71600` | 主网 |
| `aesc-testnet-1` | `71601` | 测试网 |
| `aesc-devnet-1` | `71602` | 开发网 |
| `aesc-poc` | `71603` | 本地测试（保持现有脚本兼容） |

### 2. 代码文件更新

- `x/evm/config/config.go`: 更新 `ChainIDMapping` 和 `DefaultChainID`
- `x/evm/keeper/keeper_test.go`: 更新测试用例以反映新的 Chain ID

### 3. 主网防护逻辑更新

- `x/evm/state/mock_balances.go`: 更新 `mockBalance` 函数中的主网 Chain ID 检查
  - 当前硬编码检查 EVM Chain ID `1329`（Sei 主网）
  - **改为直接判断 Cosmos Chain ID**：`if s.ctx.ChainID() == "aesc-mainnet-1" { panic }`
  - 这样避免误伤未知链（因为 DefaultChainID=71600，未映射链会被误判为主网）

### 4. 移除 Sei 遗留配置（仅 ChainIDMapping）

- 移除 `x/evm/config/config.go` 中 `ChainIDMapping` 的 Sei 遗留映射：
  - `pacific-1`、`atlantic-2`、`arctic-1`
- 更新 `evmrpc/rpcutils/sig_test.go` 测试用例中的 Chain ID

**范围说明**：
- 本次变更仅移除 `ChainIDMapping` 中的 Sei 映射
- 测试夹具文件（如 `evmrpc/rpcutils/testdata/transaction_test_data.json`）中的 Sei Chain ID 不在本次范围内，这些是签名验证测试数据，不影响 AESC 链的运行

## 影响

- **受影响的规格**: `chain-identity`
- **受影响的代码**:
  - `x/evm/config/config.go`
  - `x/evm/keeper/keeper_test.go`
  - `x/evm/keeper/params.go`（调用 `config.GetEVMChainID`）
  - `x/evm/state/mock_balances.go`（主网防护逻辑）
  - `evmrpc/rpcutils/sig_test.go`（测试用例）
- **破坏性变更**: 是 - EVM Chain ID 变更会影响所有 EVM 客户端（如 MetaMask）的网络配置
- **迁移要求**: 现有测试网/开发网需要重新配置 EVM 客户端的 Chain ID 和 Cosmos Chain ID

