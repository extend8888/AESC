# Rebrand Sei to AESC - 品牌重塑提案

## 概述

本提案旨在将 AESC Chain 代码库中的所有 Sei 品牌标识系统性地替换为 AESC 品牌标识，建立独立的技术身份。

## 提案文件

- **[proposal.md](./proposal.md)** - 提案概述，说明为什么需要这个变更、变更内容和影响范围
- **[tasks.md](./tasks.md)** - 详细的实施任务清单，分为 11 个阶段
- **[design.md](./design.md)** - 技术设计文档，包含决策、策略和实施细节
- **[validation-checklist.md](./validation-checklist.md)** - 完整的验证清单和测试脚本

## 规格增量

本提案影响以下能力的规格：

1. **[chain-identity](./specs/chain-identity/spec.md)** - 链的基础标识配置
   - 代币面额配置（usei → uaex）
   - 地址前缀配置（sei → aesc）
   - Genesis 配置

2. **[testing-infrastructure](./specs/testing-infrastructure/spec.md)** - 测试基础设施
   - Makefile 测试命令
   - 测试标识符规范
   - 测试工具和辅助函数

3. **[deployment-tools](./specs/deployment-tools/spec.md)** - 部署工具
   - 本地节点部署配置
   - Docker 容器配置
   - 部署脚本和文档

## 核心变更

### 标识符映射

| 类型 | 原标识符 | 新标识符 |
|------|---------|---------|
| 微单位代币 | `usei` | `uaex` |
| 标准代币 | `sei` | `aex` |
| 账户地址前缀 | `sei` | `aesc` |
| 地址示例 | `sei1...` | `aesc1...` |
| 验证者地址 | `seivaloper` | `aescvaloper` |

### 受影响的文件范围

- **核心配置**：`app/params/config.go`（已完成）
- **EVM 模块**：`x/evm/` 下约 20+ 文件
- **其他模块**：`x/mint/`, `x/oracle/`, `x/tokenfactory/` 等
- **测试文件**：约 150+ 测试文件
- **部署脚本**：`poc-deploy/`, `depoly-scripts/`, `docker/`
- **构建系统**：`Makefile`

## 实施阶段

1. **准备和分析** - 扫描和识别需要修改的文件
2. **Makefile 增强** - 添加测试命令
3. **核心代码替换** - 替换核心模块中的标识符
4. **测试文件更新** - 更新所有测试文件
5. **部署配置更新** - 更新部署脚本和配置
6. **工具和脚本更新** - 更新辅助工具
7. **示例和合约更新** - 更新示例代码
8. **子模块更新** - 谨慎处理子模块
9. **文档整理** - 迁移和更新文档
10. **全面测试** - 运行所有测试验证
11. **清理和文档** - 最终检查和文档

## 快速开始

### 验证当前状态

```bash
# 检查残留的 Sei 标识符
./openspec/changes/rebrand-sei-to-aesc/scripts/check-sei-references.sh

# 验证 AESC 标识符配置
./openspec/changes/rebrand-sei-to-aesc/scripts/verify-aesc-identifiers.sh
```

### 运行测试

```bash
# 运行所有测试
make test

# 只运行单元测试
make test-unit

# 只运行集成测试
make test-integration
```

### 启动本地节点

```bash
# 使用新配置启动本地节点
cd poc-deploy/localnode
./scripts/start.sh

# 验证节点使用正确的标识符
seid status
```

## 验收标准

- ✅ 代码库中不再有 `usei` 引用（除了必要的历史注释）
- ✅ 所有测试使用 `aesc1...` 地址前缀
- ✅ Makefile 包含 `test` 相关命令
- ✅ 所有单元测试通过
- ✅ 所有集成测试通过
- ✅ 可以使用新配置成功启动本地节点
- ✅ 技术文档移动到 openspec 目录

## 风险和缓解

### 主要风险

1. **遗漏某些文件中的引用** - 使用自动化脚本全面扫描
2. **测试失败** - 分阶段替换，每个阶段运行测试
3. **配置文件格式错误** - 使用模板和验证工具

### 缓解措施

- 自动化搜索和验证脚本
- 分阶段实施，每阶段验证
- 详细的测试清单
- 准备回滚方案

## 时间估算

- **代码替换和测试**：2-3 天
- **文档整理**：0.5 天
- **验证和修复**：1-2 天
- **总计**：3-5 天

## 依赖关系

- **依赖于**：`aex-gas-token` 提案中定义的代币经济模型
- **被依赖于**：`aesc-staking-contracts` 提案需要正确的基础标识

## 相关资源

- [AEX Gas Token 提案](../aex-gas-token/)
- [AESC Staking Contracts 提案](../aesc-staking-contracts/)
- [项目概览](../../project.md)

## 状态

- **当前状态**：提案阶段（待审批）
- **创建日期**：2026-01-17
- **最后更新**：2026-01-17

## 联系人

如有问题或建议，请联系项目维护者。

