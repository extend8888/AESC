# 变更：移除 sei-wasm 模块

## 为什么
sei-wasm 模块（包括 sei-wasmd 和 sei-wasmvm）对于当前项目来说是多余的。移除这些模块可以：
- 减少代码库的复杂度和维护成本
- 缩小编译产物体积
- 简化依赖管理
- 专注于 EVM 功能

## 变更内容
- **破坏性** 移除 `sei-wasmd/` 目录（CosmWasm 智能合约模块）
- **破坏性** 移除 `sei-wasmvm/` 目录（CosmWasm 虚拟机）
- **破坏性** 移除 `wasmbinding/` 目录（wasm 绑定代码）
- **破坏性** 移除 `aclmapping/wasm/` 目录（wasm ACL 映射）
- **破坏性** 移除 `precompiles/wasmd/` 目录（wasm 预编译合约）
- **破坏性** 移除 `contracts/wasm/` 目录（wasm 合约文件）
- **破坏性** 移除 `integration_test/wasm_module/` 目录
- **破坏性** 移除 `parallelization/wasm/` 目录
- **破坏性** 移除 `example/cosmwasm/` 目录
- 修改 `app/app.go` 移除 wasm 模块注册
- 修改 `app/ante.go` 移除 wasm 相关装饰器
- 修改 `app/precompiles.go` 移除 wasm 预编译
- 修改 `precompiles/setup.go` 移除 wasm 预编译设置
- 修改 `precompiles/pointer/` 移除对 wasm 的引用
- 修改 `precompiles/solo/` 移除对 wasm 的引用
- 修改 `aclmapping/dependency_generator.go` 移除 wasm 依赖
- 修改 `go.mod` 移除 CosmWasm 依赖
- 修改 `go.work` 如需要
- 修改 `openspec/project.md` 更新技术栈描述

## 影响
- 受影响的规格：无现有规格（wasm 功能未在 specs 中定义）
- 受影响的代码：
  - `app/` - 主应用初始化
  - `precompiles/` - 预编译合约
  - `wasmbinding/` - 整个目录删除
  - `aclmapping/` - wasm 相关映射
  - `go.mod` / `go.work` - 依赖配置
- 破坏性变更：所有 CosmWasm 智能合约功能将不可用
- 迁移路径：无法迁移，wasm 合约需要在移除前迁移到其他链或 EVM 合约

## 前置条件
- 确认没有生产环境的 CosmWasm 合约需要保留
- 切换到新分支进行操作

## 风险
- 移除后无法执行任何 CosmWasm 合约
- 需要确保 EVM 相关功能不受影响
- 编译和测试验证工作量较大

