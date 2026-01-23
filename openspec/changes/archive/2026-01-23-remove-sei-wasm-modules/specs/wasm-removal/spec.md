# 规格增量：移除 WASM 模块

## REMOVED Requirements

### Requirement: CosmWasm 智能合约支持
系统不再支持 CosmWasm 智能合约的部署和执行。

**原因**：项目决定专注于 EVM 生态，移除 CosmWasm 支持以简化代码库和减少维护成本。

**迁移**：无迁移路径。任何依赖 CosmWasm 的功能需要重新实现为 EVM 合约或在其他支持 CosmWasm 的链上运行。

#### Scenario: CosmWasm 合约不可用
- **WHEN** 用户尝试部署 CosmWasm 合约
- **THEN** 系统返回错误，表示不支持此功能

### Requirement: WASM 预编译合约
系统不再提供 WASM 相关的 EVM 预编译合约（位于 `0x1002`）。

**原因**：移除 CosmWasm 支持后，WASM 预编译合约失去意义。

**迁移**：EVM 合约不应再调用 WASM 预编译地址。

#### Scenario: WASM 预编译地址不可用
- **WHEN** EVM 合约调用 WASM 预编译地址
- **THEN** 调用失败或返回空结果

### Requirement: WASM-EVM 指针合约
系统不再支持 WASM 和 EVM 合约之间的指针映射功能。

**原因**：移除 CosmWasm 支持后，跨运行时指针功能失去意义。

**迁移**：使用纯 EVM 合约替代。

#### Scenario: 指针功能不可用
- **WHEN** 尝试创建 WASM-EVM 指针
- **THEN** 操作失败

### Requirement: WASM 绑定查询
系统不再提供通过 CosmWasm 合约查询链上状态的绑定接口。

**原因**：随 CosmWasm 支持一同移除。

**迁移**：使用 EVM 预编译合约或直接 gRPC 查询替代。

#### Scenario: WASM 绑定查询不可用
- **WHEN** CosmWasm 合约尝试使用自定义绑定查询
- **THEN** 查询失败

## MODIFIED Requirements

### Requirement: 支持的智能合约类型
系统必须（MUST）仅支持 EVM（Solidity）智能合约。

**变更说明**：原支持 EVM 和 CosmWasm 双运行时，现仅保留 EVM。

#### Scenario: 部署 EVM 合约
- **WHEN** 用户提交 EVM 字节码部署交易
- **THEN** 系统成功部署并返回合约地址

#### Scenario: 执行 EVM 合约
- **WHEN** 用户调用已部署的 EVM 合约方法
- **THEN** 系统正确执行并返回结果

