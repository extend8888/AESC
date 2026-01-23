# 任务清单：移除 sei-wasm 模块

## 0. 准备工作
- [ ] 0.1 创建新分支 `remove-wasm`
- [ ] 0.2 记录当前 wasm 相关文件清单（备份参考）

## 1. 阶段一：移除独立测试和示例文件
- [ ] 1.1 移除 `contracts/wasm/` 目录
- [ ] 1.2 移除 `integration_test/wasm_module/` 目录
- [ ] 1.3 移除 `parallelization/wasm/` 目录
- [ ] 1.4 移除 `example/cosmwasm/` 目录
- [ ] 1.5 移除 `tests/mars.wasm` 文件
- [ ] 1.6 验证：`go build ./...` 应该仍能编译

## 2. 阶段二：移除业务代码模块
- [ ] 2.1 移除 `wasmbinding/` 目录
- [ ] 2.2 移除 `aclmapping/wasm/` 目录
- [ ] 2.3 移除 `precompiles/wasmd/` 目录
- [ ] 2.4 清理 `aclmapping/dependency_generator.go` 中的 wasm 引用
- [ ] 2.5 验证：`go build ./...`（预期失败，需要继续清理依赖）

## 3. 阶段三：清理预编译合约中的 wasm 引用
- [ ] 3.1 修改 `precompiles/setup.go` 移除 wasmd 预编译注册
- [ ] 3.2 清理 `precompiles/pointer/` 中所有版本的 wasm 引用
- [ ] 3.3 清理 `precompiles/solo/` 中所有版本的 wasm 引用
- [ ] 3.4 验证：`go build ./...`

## 4. 阶段四：清理 app 核心模块
- [ ] 4.1 修改 `app/app.go` 移除 wasm 模块注册和初始化
- [ ] 4.2 修改 `app/ante.go` 移除 wasm 相关装饰器
- [ ] 4.3 修改 `app/precompiles.go` 移除 wasm 预编译设置
- [ ] 4.4 修改 `app/receipt.go` 移除 wasm 相关收据处理（如有）
- [ ] 4.5 清理 `app/antedecorators/accesscontrol_wasm_dependency.go`
- [ ] 4.6 清理 `app/upgrades/` 中的 wasm 相关升级逻辑
- [ ] 4.7 清理 `app/test_helpers.go` 中的 wasm 相关测试辅助
- [ ] 4.8 清理 `x/epoch/client/wasm/` 目录
- [ ] 4.9 验证：`go build ./...`

## 5. 阶段五：更新构建配置
- [ ] 5.1 修改 `go.mod` 移除 CosmWasm 依赖项
- [ ] 5.2 检查并更新 `go.work`（如需要）
- [ ] 5.3 运行 `go mod tidy`
- [ ] 5.4 验证：`go build ./...`

## 6. 阶段六：删除子模块目录
- [ ] 6.1 删除 `sei-wasmd/` 目录
- [ ] 6.2 删除 `sei-wasmvm/` 目录
- [ ] 6.3 验证：`go build ./...`

## 7. 阶段七：清理残留引用
- [ ] 7.1 使用 `rg -l "wasm|cosmwasm|wasmd"` 搜索残留引用
- [ ] 7.2 清理搜索到的残留代码
- [ ] 7.3 验证：`go build ./...`

## 8. 阶段八：测试验证
- [ ] 8.1 运行单元测试 `go test ./...`（移除 wasm 相关失败测试）
- [ ] 8.2 运行 EVM 相关测试确保功能正常
- [ ] 8.3 检查是否有构建警告

## 9. 阶段九：文档和配置更新
- [ ] 9.1 更新 `openspec/project.md` 移除 CosmWasm 相关描述
- [ ] 9.2 更新 `README.md` 移除 wasm 相关说明（如有）
- [ ] 9.3 检查并更新 `Makefile` 中的 wasm 相关目标
- [ ] 9.4 检查并更新 `Dockerfile` 中的 wasm 相关配置

## 10. 完成
- [ ] 10.1 代码审查
- [ ] 10.2 提交变更
- [ ] 10.3 归档此变更提案

