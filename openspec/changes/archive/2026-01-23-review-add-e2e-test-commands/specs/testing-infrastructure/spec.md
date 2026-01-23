# testing-infrastructure Review 发现的问题

> 本文档记录对 add-e2e-test-commands 的 Review 发现，作为后续补齐的规格依据。

---

## ADDED Requirements

### Requirement: Tokenomics E2E 必须执行真实交易与对照
系统必须（MUST）在 Tokenomics E2E 中执行至少一次真实交易，并对照 REST API 结果。

#### Scenario: USDT 预编译转账对照
- **GIVEN** 测试账户有 USDT 余额
- **WHEN** 通过 EVM 调用 `transfer(to, amount)` 并等待上链
- **THEN** 发送方/接收方余额变化应与 REST API 查询一致

#### Scenario: USDT 预编译授权转账对照
- **GIVEN** owner 有 USDT 余额且 spender 存在
- **WHEN** owner 调用 `approve(spender, amount)`，spender 调用 `transferFrom(owner, to, amount)`
- **THEN** allowance 与余额变化应与 REST API 查询一致

### Requirement: Tokenomics E2E 必须验证 EpochGasData 统计
系统必须（MUST）在 Tokenomics E2E 中验证 epoch gas_used/gas_limit 统计为正值且可读。

#### Scenario: 消耗 Gas 后统计非零
- **GIVEN** 测试过程中发送了多笔交易
- **WHEN** 查询 EpochGasData
- **THEN** `gas_used > 0` 且 `gas_limit > 0`

### Requirement: 共识 E2E 必须验证跨节点交易同步与区块一致性
系统必须（MUST）在共识 E2E 中验证交易传播与区块一致性。

#### Scenario: 跨节点交易同步
- **GIVEN** 4 节点集群正常运行
- **WHEN** 在 node0 广播交易并等待出块
- **THEN** node1/2/3 均可查询到该交易

#### Scenario: 高度与区块哈希一致性
- **GIVEN** 4 节点集群正常运行
- **WHEN** 查询同一高度的区块
- **THEN** 高度差异 ≤ 1 且区块 hash 一致
