# 实施任务清单

> **状态**: 大部分已完成 ✅ (2025-01-17)
>
> **提交历史**:
> - `da2d659` feat(makefile): add test commands for unit and integration testing
> - `8d11998` docs(openspec): add rebrand-sei-to-aesc proposal
> - `c7bc124` feat(rebrand): replace usei with uaex in config files, scripts, and docs
> - `214fb51` fix(rebrand): rename SplitUseiWeiAmount to SplitUaexWeiAmount
> - `ea12d69` fix(rebrand): rename remaining Usei identifiers to Uaex
> - `3eb2552` docs(rebrand): update USEI references to UAEX in documentation
> - `8005028` fix(rebrand): adjust creator max length test case for aesc1 prefix

## 阶段 1：准备和分析 ✅
- [x] 1.1 扫描代码库，生成所有包含 `usei` 的文件清单
- [x] 1.2 扫描代码库，生成所有包含 `sei1` 地址的文件清单
- [x] 1.3 识别需要特殊处理的文件（如 vendor、第三方库等）
- [x] 1.4 备份当前的测试基准结果

## 阶段 2：Makefile 增强 ✅
- [x] 2.1 添加 `test` 命令到 Makefile（运行所有测试）
- [x] 2.2 添加 `test-unit` 命令（只运行单元测试）
- [x] 2.3 添加 `test-integration` 命令（只运行集成测试）
- [x] 2.4 验证新命令可以正常工作

## 阶段 3：核心代码替换 ✅
- [x] 3.1 替换 `x/evm/` 模块中的所有 `usei` 引用
- [x] 3.2 替换 `x/mint/` 模块中的所有 `usei` 引用
- [x] 3.3 替换 `x/oracle/` 模块中的所有 `usei` 引用
- [x] 3.4 替换 `x/tokenfactory/` 模块中的所有 `usei` 引用
- [x] 3.5 替换 `cmd/seid/` 中的所有 `usei` 引用
- [x] 3.6 替换 `app/` 目录中的所有 `usei` 引用（除了已完成的 config.go）
- [x] 3.7 运行单元测试验证核心模块

## 阶段 4：测试文件更新 ✅
- [x] 4.1 更新 `x/evm/` 的测试文件
- [x] 4.2 更新 `x/mint/` 的测试文件
- [x] 4.3 更新 `x/oracle/` 的测试文件
- [x] 4.4 更新 `app/` 的测试文件
- [x] 4.5 更新 `integration_test/` 中的所有测试
- [x] 4.6 更新 `testutil/` 中的测试工具
- [x] 4.7 运行所有单元测试

## 阶段 5：部署和配置文件更新 ✅
- [x] 5.1 更新 `poc-deploy/localnode/` 中的配置文件
- [x] 5.2 更新 `poc-deploy/localnode/scripts/` 中的脚本
- [x] 5.3 更新 `depoly-scripts/` 中的部署脚本
- [x] 5.4 更新 `docker/` 中的 Docker 配置
- [x] 5.5 更新 genesis 示例文件
- [ ] 5.6 验证可以使用新配置启动本地节点 ⏳ (需要编译 EVM 合约)

## 阶段 6：工具和脚本更新 ✅
- [x] 6.1 更新 `scripts/` 目录中的所有脚本
- [x] 6.2 更新 `loadtest/` 中的负载测试工具
- [x] 6.3 更新 `oracle/price-feeder/` 中的配置
- [x] 6.4 更新 `poc-deploy/tools/` 中的工具

## 阶段 7：示例和合约更新 ✅
- [x] 7.1 更新 `example/` 中的示例代码
- [x] 7.2 更新 `contracts/` 中的测试配置
- [x] 7.3 更新 `evmrpc/` 中的测试文件

## 阶段 8：子模块更新（谨慎处理）✅
- [x] 8.1 检查 `sei-cosmos/` 中需要更新的测试文件
- [x] 8.2 检查 `sei-tendermint/` 中需要更新的配置
- [x] 8.3 检查 `sei-wasmd/` 中需要更新的配置
- [x] 8.4 注意：只更新测试和配置，不修改核心逻辑

## 阶段 9：文档整理 ✅
- [x] 9.1 将 `tmp/AESC 链开发技术方案.md` 复制到 `openspec/changes/rebrand-sei-to-aesc/design.md`
- [x] 9.2 更新 design.md 中的所有 Sei 引用为 AESC
- [x] 9.3 在 design.md 中添加说明，指出这是从 tmp 目录迁移的文档
- [ ] 9.4 考虑是否删除 tmp 目录中的原文件（或添加说明指向新位置）⏳ (待用户决定)

## 阶段 10：全面测试和验证 ⏳
- [x] 10.1 运行完整的单元测试套件：`make test-unit`
- [ ] 10.2 运行集成测试：`make test-integration` ⏳ (需要编译 EVM 合约)
- [ ] 10.3 启动本地单节点验证基本功能 ⏳ (需要编译 EVM 合约)
- [ ] 10.4 启动本地多节点集群验证共识 ⏳ (需要编译 EVM 合约)
- [ ] 10.5 验证 EVM RPC 端点正常工作 ⏳ (需要编译 EVM 合约)
- [ ] 10.6 验证代币转账功能 ⏳ (需要编译 EVM 合约)
- [ ] 10.7 验证 Gas 费用计算正确 ⏳ (需要编译 EVM 合约)

## 阶段 11：清理和文档 ✅
- [x] 11.1 检查是否有遗漏的 `usei` 引用 (仅剩 base64 编码的历史交易数据)
- [x] 11.2 检查是否有遗漏的 `sei1` 地址 (已清理完毕)
- [x] 11.3 更新 README.md（如果需要）
- [x] 11.4 生成变更摘要文档
- [x] 11.5 准备提交信息

## 验证检查清单
- [x] ✅ 搜索 `usei` 只返回必要的历史注释或第三方代码
- [x] ✅ 搜索 `sei1` 只返回必要的历史注释
- [x] ✅ `make test` 命令存在且可以运行
- [ ] ⏳ 所有单元测试通过 (EVM artifacts 需要先编译)
- [ ] ⏳ 所有集成测试通过 (EVM artifacts 需要先编译)
- [ ] ⏳ 本地节点可以成功启动 (EVM artifacts 需要先编译)
- [x] ✅ 可以创建 `aesc1...` 格式的地址
- [ ] ⏳ 可以使用 `uaex` 进行交易 (需要节点运行)
- [x] ✅ 技术文档已移动到 openspec 目录

## 已知问题

### EVM Artifacts 缺失
测试运行时报错：`pattern CW1155ERC1155Pointer.abi: no matching files found`

这是预先存在的环境问题，需要先编译 EVM 合约生成 ABI 文件：
```bash
# 编译合约 (示例命令，具体取决于项目配置)
cd contracts && npm install && npm run build
```

缺失的 artifacts:
- `CW1155ERC1155Pointer.abi`
- `CW20ERC20Pointer.abi`
- `CW721ERC721Pointer.abi`
- `NativeSeiTokensERC20.abi`
- `WSEI.abi`

### 保留的 Base64 编码数据
4 个文件中包含 base64 编码的历史交易数据，其中包含 "usei"：
- `evmrpc/tests/mock_data/transactions/evm_transaction_by_hash.json`
- `evmrpc/tests/mock_data/transactions/evm_transaction_by_hash_error.json`
- `evmrpc/tests/mock_data/transactions/evm_transaction_by_hash_pending.json`
- `evmrpc/tests/mock_data/transactions/internal_transaction_by_hash.json`

这些是真实交易的编码记录，保留作为历史测试数据。

## 注意事项
1. **不要修改第三方依赖**：`go-ethereum/`, `sei-cosmos/`, `sei-tendermint/` 等子模块的核心代码不应修改，只更新测试和配置
2. **保留必要的注释**：如果某些注释中提到 Sei 是为了说明历史或技术来源，应该保留
3. **分阶段提交**：建议每完成一个阶段就提交一次，方便回滚
4. **测试优先**：每个阶段完成后都应该运行相关测试
5. **文档同步**：确保所有文档和注释都使用正确的术语

