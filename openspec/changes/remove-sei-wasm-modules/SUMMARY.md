# WASM 模块移除总结

> **分支**: `remove-wasm`
> **日期**: 2026-01-19
> **状态**: ✅ 完成

---

## 概述

本次变更从 AESC Chain 中完整移除了 CosmWasm/WASM 相关模块，简化代码库，专注于 EVM 兼容性。

### 变更统计

| 类型 | 数量 |
|------|------|
| 删除文件 | 701 |
| 修改文件 | 68 |
| 总变更 | 769 |

---

## 一、删除的目录和文件

### 1.1 核心 WASM 模块 (565 文件)

| 目录 | 文件数 | 说明 |
|------|--------|------|
| `sei-wasmd/` | 450 | CosmWasm 模块 (fork of wasmd) |
| `sei-wasmvm/` | 115 | WASM 虚拟机 (fork of wasmvm) |

### 1.2 WASM 集成代码 (90 文件)

| 目录 | 文件数 | 说明 |
|------|--------|------|
| `wasmbinding/` | 10 | WASM 与 Cosmos SDK 绑定 |
| `precompiles/wasmd/` | 32 | WASM 预编译合约 |
| `aclmapping/wasm/` | 2 | WASM ACL 映射 |
| `parallelization/wasm/` | 4 | WASM 并行化代码 |

### 1.3 WASM 合约和示例 (53 文件)

| 目录 | 文件数 | 说明 |
|------|--------|------|
| `example/cosmwasm/` | 46 | CosmWasm 示例项目 (cw20, cw721, cw1155) |
| `contracts/wasm/` | 7 | 预编译 WASM 合约文件 (.wasm) |

### 1.4 测试和工具 (6 文件)

| 目录/文件 | 说明 |
|-----------|------|
| `integration_test/wasm_module/` | WASM 集成测试 (4 文件) |
| `testutil/processblock/genesiswasm.go` | WASM genesis 工具 |
| `cmd/seid/cmd/genwasm.go` | WASM genesis 命令 |

### 1.5 模块内 WASM 客户端 (27 文件)

| 目录 | 说明 |
|------|------|
| `x/tokenfactory/client/wasm/` | TokenFactory WASM 客户端 |
| `x/oracle/client/wasm/` | Oracle WASM 客户端 |
| `x/evm/client/wasm/` | EVM WASM 客户端 |
| `x/epoch/client/wasm/` | Epoch WASM 客户端 |

---

## 二、修改的文件 (68 文件)

### 2.1 核心应用层 (11 文件)

| 文件 | 修改内容 |
|------|----------|
| `app/app.go` | 移除 WasmKeeper、wasm 模块注册、wasm 相关导入 |
| `app/ante.go` | 移除 wasm ante handler |
| `app/precompiles.go` | 移除 wasmd 预编译注册 |
| `app/test_helpers.go` | 更新 Setup() 签名，移除 wasm 参数 |
| `app/receipt.go` | 移除 wasm 相关逻辑 |
| `app/receipt_test.go` | 移除 wasm 测试用例 |
| `app/ante_test.go` | 移除 wasm 测试 |
| `app/antedecorators/gas_test.go` | 移除 wasm 相关测试 |
| `app/upgrades/v0/upgrade.go` | 移除 wasm 升级逻辑 |

### 2.2 命令行工具 (4 文件)

| 文件 | 修改内容 |
|------|----------|
| `cmd/seid/cmd/root.go` | 移除 AddGenesisWasmMsgCmd，更新 app.New() 调用 |
| `cmd/seid/cmd/blocktest.go` | 更新 app.New() 调用签名 |
| `cmd/seid/cmd/ethreplay.go` | 更新 app.New() 调用签名 |
| `cmd/seid/cmd/snapshot.go` | 更新 app.New() 调用签名 |

### 2.3 预编译合约 (15 文件)

| 文件 | 修改内容 |
|------|----------|
| `precompiles/setup.go` | 移除 wasmd 预编译注册 |
| `precompiles/pointer/*.go` | 移除 CW20/CW721/CW1155 指针逻辑 (11 版本) |
| `precompiles/solo/*.go` | 移除 WASM 相关逻辑 (3 文件) |

### 2.4 EVM 模块 (14 文件)

| 文件 | 修改内容 |
|------|----------|
| `x/evm/keeper/keeper.go` | 移除 WasmKeeper 引用 |
| `x/evm/keeper/msg_server.go` | 移除 wasm 调用逻辑 |
| `x/evm/keeper/pointer.go` | 移除 CW 指针逻辑 |
| `x/evm/keeper/precompile.go` | 移除 wasm 预编译 |
| `x/evm/types/params.go` | 移除 DeliverTxHookWasmGasLimit 等参数 |
| `x/evm/types/params_test.go` | 移除 wasm 参数测试 |
| `x/evm/module.go` | 移除 wasm 模块依赖 |
| 其他 keeper/client 文件 | 移除 wasm 相关导入和逻辑 |

### 2.5 EVM RPC (9 文件)

| 文件 | 修改内容 |
|------|----------|
| `evmrpc/send.go` | 移除 wasm 相关逻辑 |
| `evmrpc/simulate.go` | 移除 wasm 模拟逻辑 |
| `evmrpc/block.go` | 移除 wasm 引用 |
| `evmrpc/tx.go` | 移除 wasm 交易处理 |
| `evmrpc/utils.go` | 移除 wasm 工具函数 |
| `evmrpc/tests/*.go` | 移除 wasm 测试 mock |

### 2.6 ACL 映射 (2 文件)

| 文件 | 修改内容 |
|------|----------|
| `aclmapping/dependency_generator.go` | 移除 wasm 依赖生成 |
| `aclmapping/utils/resource_type.go` | 移除 wasm 资源类型 |

### 2.7 依赖管理 (3 文件)

| 文件 | 修改内容 |
|------|----------|
| `go.mod` | 移除 CosmWasm/wasmd, CosmWasm/wasmvm 依赖 |
| `go.sum` | 更新依赖校验 |
| `go.work.sum` | 更新工作区依赖 |

### 2.8 测试和工具 (10 文件)

| 文件 | 修改内容 |
|------|----------|
| `testutil/network/network.go` | 移除 wasm 网络配置 |
| `testutil/processblock/presets.go` | 移除 wasm 预设 |
| `occ_tests/messages/test_msgs.go` | 移除 wasm 测试消息 |
| `occ_tests/utils/utils.go` | 移除 wasm 工具 |
| `loadtest/main.go` | 移除 wasm 负载测试 |
| `tools/migration/sc/migrator.go` | 移除 wasm 迁移 |

---

## 三、关键技术变更

### 3.1 app.New() 签名变更

**变更前**:
```go
func New(
    ...
    wasmEnabledProposals []wasm.ProposalType,
    wasmOpts []wasm.Option,
    wasmGasRegisterConfig *wasmkeeper.WasmGasRegisterConfig,
    ...
) *App
```

**变更后**:
```go
func New(
    ...
    // 移除了 wasmEnabledProposals, wasmOpts, wasmGasRegisterConfig 参数
    ...
) *App
```

### 3.2 Keeper 依赖变更

**移除的 Keeper**:
- `WasmKeeper` - CosmWasm 合约管理
- `WasmMsgServer` - WASM 消息处理

### 3.3 预编译合约变更

**移除的预编译**:
- `wasmd` - WASM 合约执行预编译
- CW20/CW721/CW1155 指针预编译

**保留的预编译**:
- `bank`, `staking`, `gov`, `distribution` 等核心预编译
- `pointer` (仅 ERC20/ERC721 指针)
- `solo` (仅 EVM 相关)

### 3.4 模块注册变更

**移除的模块**:
```go
// 从 ModuleBasics 和 ModuleAccountPermissions 中移除
wasmtypes.ModuleName
```

---

## 四、验证结果

### 4.1 编译验证
```bash
go build ./app/... ./cmd/... ./x/... ./precompiles/... ./evmrpc/... ./aclmapping/...
# 退出码: 0 (成功)
```

### 4.2 E2E 测试
| 测试项 | 状态 |
|--------|------|
| 节点启动 | ✅ 通过 |
| 区块生产 | ✅ 2400+ 区块 |
| AEX 经济模型 | ✅ 全部通过 |
| EVM RPC | ✅ 正常响应 |
| 验证者状态 | ✅ BONDED |

### 4.3 单元测试
| 模块 | 状态 |
|------|------|
| x/aexburn | ✅ 全部通过 |
| x/evm/types | ✅ 全部通过 |
| precompiles/bank | ⚠️ 1 失败 (已知问题) |

---

## 五、注意事项

1. **不兼容变更**: 移除 WASM 后，所有依赖 CosmWasm 的功能将不可用
2. **数据迁移**: 如有现存 WASM 合约数据，需要在升级前处理
3. **API 变更**: app.New() 签名变更，所有调用点需更新

---

## 六、文档位置

| 文档 | 路径 |
|------|------|
| 变更提案 | `openspec/changes/remove-sei-wasm-modules/proposal.md` |
| 技术设计 | `openspec/changes/remove-sei-wasm-modules/design.md` |
| 任务清单 | `openspec/changes/remove-sei-wasm-modules/tasks.md` |
| 变更日志 | `openspec/changes/remove-sei-wasm-modules/changelog.md` |
| 本总结 | `openspec/changes/remove-sei-wasm-modules/SUMMARY.md` |

