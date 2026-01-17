# 品牌重塑设计文档

> **文档来源**：本文档整合自 `tmp/AESC 链开发技术方案.md`，并针对品牌重塑工作进行了调整和补充。

## 背景

AESC Chain 是基于 Sei Protocol 分叉的高性能区块链项目。虽然技术架构继承自 Sei，但 AESC 需要建立独立的品牌标识和技术身份。当前代码库中保留了大量 Sei 的标识符，包括：

- 代币面额：`usei`（微单位）、`sei`（标准单位）
- 地址前缀：`sei`、`sei1...`
- 链名称和配置中的各种 `sei` 引用

这些标识符需要系统性地替换为 AESC 的品牌标识，以确保：
1. 用户体验的一致性
2. 技术文档的准确性
3. 开发和测试环境的正确性

### 项目技术背景

AESC Chain 基于 Sei Protocol，具有以下技术特点：
- **高性能**：400ms 区块最终确认
- **并行执行**：EVM 和 CosmWasm 的乐观并行执行
- **双代币体系**：
  - **AEX**：Gas 代币（链层原生）
  - **AESC**：权益代币（智能合约实现）

## 目标 / 非目标

### 目标
1. ✅ 将所有用户可见的 Sei 标识替换为 AESC 标识
2. ✅ 更新所有测试文件使用正确的 AESC 标识符
3. ✅ 更新部署和配置脚本
4. ✅ 增强 Makefile，添加便捷的测试命令
5. ✅ 整理技术文档到正确的位置（openspec 目录）
6. ✅ 确保所有测试通过

### 非目标
1. ❌ 不修改第三方依赖的核心代码（如 go-ethereum、sei-cosmos 等）
2. ❌ 不改变底层技术架构和共识机制
3. ❌ 不修改已经正确的 `app/params/config.go`（已完成）
4. ❌ 不在此阶段实现 AEX 的通胀和销毁机制（属于 aex-gas-token 提案）

## 核心标识符映射

### 代币面额
| 原标识符 | 新标识符 | 说明 |
|---------|---------|------|
| `usei` | `uaex` | 微单位（1 AEX = 10^6 uaex） |
| `sei` | `aex` | 标准单位 |
| `UseiExponent` | `UaexExponent` | 指数常量 |

### 地址前缀
| 原标识符 | 新标识符 | 说明 |
|---------|---------|------|
| `sei` | `aesc` | Bech32 账户地址前缀 |
| `sei1...` | `aesc1...` | 账户地址示例 |
| `seivaloper` | `aescvaloper` | 验证者地址前缀 |
| `seivalcons` | `aescvalcons` | 共识地址前缀 |

### 链标识
| 原标识符 | 新标识符 | 说明 |
|---------|---------|------|
| `sei-chain` | `aesc-chain` | 链名称 |
| `seid` | `seid` | **保持不变**（二进制名称，避免大规模重构） |

## 技术决策

### 决策 1：保留 `seid` 二进制名称
**决策**：不修改 `seid` 二进制名称和相关的命令行工具名称。

**原因**：
- 修改二进制名称需要大量的构建脚本、文档和工具链修改
- `seid` 作为内部工具名称，对用户不可见
- 可以在未来的重大版本升级时考虑修改

**替代方案**：
- 方案 A：修改为 `aescd`（被拒绝，工作量大，收益小）
- 方案 B：保持 `seid`（采用）

### 决策 2：分阶段替换策略
**决策**：采用分阶段替换策略，每个阶段完成后运行测试验证。

**阶段划分**：
1. Makefile 增强（独立，低风险）
2. 核心代码替换（高优先级）
3. 测试文件更新（依赖核心代码）
4. 配置和脚本更新（独立）
5. 文档整理（独立）
6. 全面测试和验证

**原因**：
- 降低风险，每个阶段可以独立验证
- 便于定位问题
- 可以在发现问题时快速回滚

### 决策 3：子模块处理策略
**决策**：只更新子模块（sei-cosmos、sei-tendermint 等）中的测试和配置文件，不修改核心逻辑。

**原因**：
- 这些是 fork 的上游依赖，修改核心逻辑可能导致兼容性问题
- 测试和配置的修改足以支持 AESC 的品牌标识
- 保持与上游的技术兼容性

### 决策 4：Makefile 测试命令设计
**决策**：添加三个测试命令：
- `make test`：运行所有测试
- `make test-unit`：只运行单元测试
- `make test-integration`：只运行集成测试

**实现方式**：
```makefile
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v -race -timeout=10m ./x/... ./app/... ./evmrpc/... ./precompiles/...

test-integration:
	@echo "Running integration tests..."
	cd integration_test && go test -v -timeout=30m ./...
```

**原因**：
- 提供便捷的测试入口
- 支持快速的单元测试反馈循环
- 集成测试可以单独运行，节省时间

## 实施策略

### 自动化替换脚本
为了确保替换的一致性和完整性，将使用以下策略：

1. **搜索和识别**：
   ```bash
   # 查找所有包含 usei 的文件
   find . -type f -name "*.go" -o -name "*.json" -o -name "*.sh" | \
     xargs grep -l "usei" > files_with_usei.txt
   
   # 查找所有包含 sei1 地址的文件
   find . -type f -name "*.go" | \
     xargs grep -l "sei1" > files_with_sei1.txt
   ```

2. **批量替换**（需要人工审查）：
   ```bash
   # 替换 usei -> uaex
   find . -type f -name "*.go" -exec sed -i 's/usei/uaex/g' {} +
   
   # 替换 sei1 -> aesc1（需要更谨慎）
   # 建议手动处理或使用更精确的正则表达式
   ```

3. **验证**：每次替换后运行测试确保没有破坏功能

### 特殊处理的文件类型

#### 1. Genesis 配置文件
- 位置：`poc-deploy/localnode/exmaple_genesis.json`
- 需要更新：
  - `denom` 字段：`usei` → `uaex`
  - `address` 字段：`sei1...` → `aesc1...`
  - `denom_metadata` 配置

#### 2. 测试文件
- 测试地址生成
- 测试金额和面额
- 断言中的字符串匹配

#### 3. 部署脚本
- `poc-deploy/localnode/scripts/`
- `depoly-scripts/localnode/`
- Shell 脚本中的环境变量和配置

#### 4. Docker 配置
- `docker/localnode/`
- `docker/rpcnode/`
- 环境变量和启动参数

## 风险和缓解措施

### 风险 1：遗漏某些文件中的引用
**影响**：中等 - 可能导致部分功能使用错误的标识符

**缓解措施**：
1. 使用自动化脚本全面扫描
2. 在多个阶段进行验证
3. 运行完整的测试套件
4. 手动检查关键配置文件

### 风险 2：测试失败
**影响**：高 - 可能表明替换破坏了功能

**缓解措施**：
1. 分阶段替换，每个阶段运行测试
2. 保留详细的测试日志
3. 准备回滚方案
4. 优先修复核心功能的测试

### 风险 3：地址格式不兼容
**影响**：高 - 可能导致地址验证失败

**缓解措施**：
1. 验证 Bech32 前缀配置正确
2. 测试地址生成和验证功能
3. 确保所有地址工具使用正确的前缀

### 风险 4：Genesis 配置错误
**影响**：严重 - 可能导致链无法启动

**缓解措施**：
1. 使用 JSON 验证工具检查格式
2. 在测试环境中验证 genesis 配置
3. 保留原始配置作为参考
4. 详细记录所有修改

## 验证策略

### 单元测试验证
```bash
# 运行核心模块的单元测试
make test-unit

# 预期结果：所有测试通过
```

### 集成测试验证
```bash
# 运行集成测试
make test-integration

# 预期结果：所有测试通过
```

### 本地节点验证
```bash
# 启动单节点
cd poc-deploy/localnode
./scripts/start.sh

# 验证项目：
# 1. 节点成功启动
# 2. 可以创建 aesc1... 地址
# 3. 可以使用 uaex 进行转账
# 4. Gas 费用正确计算
```

### 多节点集群验证
```bash
# 启动 4 节点集群
make docker-cluster-start

# 验证项目：
# 1. 所有节点成功启动
# 2. 共识正常工作
# 3. 跨节点交易正常
```

## 迁移计划

### 开发环境迁移
1. 清理旧的链数据：`rm -rf ~/.sei`
2. 重新构建二进制：`make install`
3. 使用新配置初始化：`seid init ...`
4. 验证配置正确

### 测试环境迁移
1. 停止所有运行的节点
2. 清理测试数据
3. 更新测试脚本
4. 重新运行测试套件

### 回滚计划
如果发现严重问题：
1. 使用 git 回滚到替换前的提交
2. 恢复备份的配置文件
3. 重新构建和测试
4. 分析失败原因，调整策略

## 文档整理

### 技术文档迁移
将 `tmp/` 目录下的技术文档整理到 openspec 目录：

1. **AESC 链开发技术方案.md**
   - 目标位置：`openspec/changes/rebrand-sei-to-aesc/design.md`
   - 处理方式：整合到本设计文档中

2. **AESC 公链 Gas (AEX)经济模型.md**
   - 建议位置：`openspec/changes/aex-gas-token/` 或 `docs/`
   - 作为 AEX 代币经济模型的参考文档

3. **AESC 节点系统.md**
   - 建议位置：`openspec/changes/aesc-staking-contracts/` 或 `docs/`
   - 作为 AESC 权益系统的参考文档

### 文档更新清单
- [ ] 更新 README.md 中的示例地址
- [ ] 更新 docs/ 中的技术文档
- [ ] 更新 openspec/project.md 中的项目描述
- [ ] 添加迁移指南（如果需要）

## 开放问题

1. **是否需要更新 Makefile 中的 ldflags？**
   - 当前：`-X github.com/cosmos/cosmos-sdk/version.Name=sei`
   - 建议：保持不变或更新为 `aesc`
   - 决策：待讨论

2. **是否需要更新 Docker 镜像名称？**
   - 当前：`sei-chain/localnode`
   - 建议：更新为 `aesc-chain/localnode`
   - 决策：待讨论

3. **子模块目录名称是否需要重命名？**
   - 当前：`sei-cosmos/`, `sei-tendermint/` 等
   - 建议：保持不变（避免大规模重构）
   - 决策：保持不变

## 参考资料

### 相关提案
- `openspec/changes/aex-gas-token/`：AEX Gas 代币实现
- `openspec/changes/aesc-staking-contracts/`：AESC 权益合约

### 技术文档
- `tmp/AESC 公链 Gas (AEX)经济模型.md`
- `tmp/AESC 节点系统.md`
- `tmp/AESC 链开发技术方案.md`

### 代码参考
- `app/params/config.go`：已完成的基础配置
- `x/evm/keeper/params.go`：EVM 模块配置
- `cmd/seid/cmd/root.go`：命令行工具配置

