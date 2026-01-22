# 变更提案：Review add-e2e-test-commands 实现完备性

## 状态：📝 提案中

---

## 为什么
`add-e2e-test-commands` 已完成开发，但任务清单与规格要求尚未进行系统性对照审查。本 Review 用于确认完成度、识别缺口并给出后续补齐建议。

## 变更内容
- 对照提案、tasks、spec 增量与测试用例清单
- 检查 Makefile 与两类 E2E 脚本的实现
- 输出完成度结论与缺口列表

## Review 结果
### 已完成
- Makefile 新增 `test-tokenomics-e2e`/`test-consensus-e2e` 目标
- 两个脚本具备 PASS/FAIL 汇总与失败退出码
- 共识脚本提供清理选项（`--cleanup`/`--no-cleanup`）

### 未完成/不一致
- 阶段 4 验证任务未执行（E2E 未跑、README 未更新）
- Tokenomics 脚本未覆盖：高 Gas 使用率→低销毁比例、Epoch gas_used/gas_limit 统计、USDT transfer/approve/transferFrom 的真实交易与 REST 对照
- 共识脚本未覆盖：跨节点交易同步、chain-id 一致性断言、区块 hash 一致性、bank/staking 状态一致性
- 区块高度差异阈值与规格不一致（脚本 ≤2，规格/任务 ≤1）
- 环境依赖检查不足（jq/bc/curl/docker）

## 影响
- 受影响变更：`openspec/changes/add-e2e-test-commands/`
- 受影响能力：`testing-infrastructure`
- 受影响代码：`Makefile`、`poc-deploy/localnode/scripts/run_tokenomics_e2e_tests.sh`、`docker/localnode/scripts/run_consensus_e2e_tests.sh`

## 结论
当前实现 **未达到完成标准**。需补齐缺失用例与文档，并执行实际 E2E 验证后再标记完成。
