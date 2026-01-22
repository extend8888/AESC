# 变更提案：添加 E2E 测试命令

## 状态：📝 提案中

---

## 1. 为什么

当前 Makefile 已有基础测试命令（`make test`、`make test-unit`、`make test-integration`），但缺少专门的 E2E 测试命令来验证：
1. **Tokenomics 经济模型** - AEX/STAEX 双代币体系、aexburn 销毁机制、参数硬约束
2. **多节点共识** - 4 节点 Docker 集群的共识行为、跨节点交易同步、状态一致性

这些 E2E 测试对于确保生产部署前的完整性验证至关重要。

---

## 2. 变更内容

### 2.1 新增 Makefile 命令

| 命令 | 描述 |
|------|------|
| `make test-tokenomics-e2e` | 运行 Tokenomics 经济模型 E2E 测试 |
| `make test-consensus-e2e` | 运行多节点共识 E2E 测试 |

### 2.2 新增测试脚本

| 脚本 | 位置 | 用途 |
|------|------|------|
| `run_tokenomics_e2e_tests.sh` | `poc-deploy/localnode/scripts/` | Tokenomics E2E 测试入口 |
| `run_consensus_e2e_tests.sh` | `docker/localnode/scripts/` | 多节点共识 E2E 测试入口 |

### 2.3 测试内容

**Tokenomics E2E 测试**:
- TC-TK-01: AEX Gas 代币功能验证
- TC-TK-02: STAEX 质押代币功能验证
- TC-TK-03: aexburn 销毁机制验证
- TC-TK-04: 参数硬约束验证 (3%/5%/30-60%)
- TC-TK-05: Epoch Gas 数据采集验证

**共识 E2E 测试**:
- TC-MC-01: 4 节点集群启动验证
- TC-MC-02: 跨节点交易同步验证
- TC-MC-03: 区块高度一致性验证
- TC-MC-04: 模块状态一致性验证

---

## 3. 不在范围内

- 修改现有 `make test` / `make test-unit` / `make test-integration` 命令
- 性能基准测试
- 压力测试脚本

---

## 4. 依赖

- Docker 和 Docker Compose 已安装（共识测试）
- `poc-deploy/localnode/scripts/deploy.sh` 可正常运行（Tokenomics 测试）
- AEX 经济模型修复已完成（`fix-aex-economic-model` 已归档）

---

## 5. 影响

- **受影响规格**: `testing-infrastructure`
- **受影响代码**: `Makefile`、`poc-deploy/localnode/scripts/`、`docker/localnode/scripts/`

---

## 6. 验收标准

1. ✅ `make test-tokenomics-e2e` 可成功运行并输出测试结果
2. ✅ `make test-consensus-e2e` 可成功启动 4 节点集群并运行测试
3. ✅ 所有测试用例有清晰的 PASS/FAIL 输出
4. ✅ 测试失败时返回非零退出码

---

## 7. 参考

- 参考项目: `/Users/liaofeng/part-time/study/AevirDaoTech/aevir-chain`
  - `openspec/changes/archive/2026-01-20-test-multi-node-consensus`
  - `openspec/changes/archive/2026-01-22-test-aevir-tokenomics`

