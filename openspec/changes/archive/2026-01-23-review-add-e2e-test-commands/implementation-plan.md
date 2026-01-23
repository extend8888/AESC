# 修改方案：add-e2e-test-commands 补齐与验证

本文档用于指导补齐 `add-e2e-test-commands` 的实现缺口并完成验证与文档更新。

## 目标

- 对齐规格与任务清单，补齐缺失的 E2E 验证点。
- 所有测试脚本输出清晰 PASS/FAIL 且失败返回非零退出码。
- 完成阶段 4 验证与 README 更新。

## 涉及文件

- `poc-deploy/localnode/scripts/run_tokenomics_e2e_tests.sh`
- `docker/localnode/scripts/run_consensus_e2e_tests.sh`
- `Makefile`
- `README.md`
- `openspec/changes/add-e2e-test-commands/tasks.md`
- `openspec/changes/add-e2e-test-commands/test-cases.md`

## Tokenomics E2E 脚本修改

### 1) 环境依赖检查

- 在脚本入口增加依赖检查：`curl`、`jq`、`bc`、`seid`。
- 任一缺失即 FAIL 并退出。

### 2) 真实交易与对照（USDT 预编译）

目的：覆盖 `transfer/approve/transferFrom` 的真实交易与 REST 对照。

- 需要准备：
  - 可用的 EVM 私钥/地址（建议写入环境变量）。
  - 该地址具备足够的 USDT 余额与 Gas 余额。
- 建议路径（二选一）：
  - 方案 A：使用 `cast`（Foundry）发送交易。
  - 方案 B：构造原始交易并使用 `eth_sendRawTransaction`。
- 验证要求：
  - 交易成功（code=0 或 RPC 返回成功）。
  - REST API 余额变化与 EVM 调用一致。
  - allowance 变化与 REST API 一致。

### 3) 高 Gas 使用率导致低销毁比例

- 记录低负载下的 `burn_rate`。
- 通过批量交易制造高负载（循环发送交易或高 gas 消耗交易）。
- 再次查询 `burn_rate`，断言高负载时 `burn_rate` 下降。

### 4) EpochGasData 统计验证

- 查询包含 `gas_used`、`gas_limit` 的接口。
- 断言 `gas_used > 0` 且 `gas_limit > 0`。
- 若字段名不同，以实际接口为准并更新断言逻辑。

### 5) AEX/STAEX 基础验证增强

- 增加一次真实 `uaex` 转账。
- 余额前后对比，确认 Gas 使用 `uaex` 扣除。

## 共识 E2E 脚本修改

### 1) chain-id 一致性

- 从 4 个节点 `status` 提取 `network/chain-id`。
- 断言完全一致。

### 2) 跨节点交易同步

- 在 node0 发送一笔交易（建议 `docker exec` 调用 `aescd tx`）。
- 获取 tx hash。
- 在 node1/2/3 查询 tx，必须可查。

### 3) 区块 hash 一致性

- 选择同一高度，查询 4 个节点区块 hash。
- 断言 hash 一致。

### 4) 模块状态一致性补齐

- 在 4 节点查询 bank 余额（同一账户）。
- 在 4 节点查询 staking 状态（如 validators 列表或 params）。
- 断言一致。

### 5) 高度差异阈值对齐

- 将高度差异阈值从 `<= 2` 调整为 `<= 1`（与规格一致）。

## Makefile

- 保持现有目标不变。
- 可选：补充提示信息（Tokenomics 需 localnode，Consensus 自动启动/停止集群）。

## 文档与验证

### 1) 运行验证

- `make test-tokenomics-e2e`
- `make test-consensus-e2e`

### 2) 更新用例结果

- `openspec/changes/add-e2e-test-commands/test-cases.md` 填写实际结果。

### 3) 更新 README

- 增加新命令说明、依赖与运行前置条件。

### 4) 更新任务状态

- 将 `openspec/changes/add-e2e-test-commands/tasks.md` 中 T-4.1/T-4.2/T-4.3 置为完成。

## 执行顺序建议

1. 先补齐 Tokenomics 脚本缺口。
2. 再补齐共识脚本缺口。
3. 运行两类 E2E。
4. 更新用例与 README。
5. 更新任务状态。

## 验收标准

- Tokenomics 与共识脚本均覆盖规格中的所有场景。
- 真实交易与对照可重复通过。
- 阶段 4 验证任务与 README 更新完成。
