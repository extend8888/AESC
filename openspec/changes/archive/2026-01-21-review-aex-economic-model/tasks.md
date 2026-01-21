## 1. 决策与澄清
- [x] 1.1 AESC 质押/治理载体需与 Gas 隔离（链未启动，主网前调整 `bond_denom`，无需迁移）
- [x] 1.2 高使用率下销毁比例按文档实现（下调）
- [x] 1.3 Gas Credit 由 x402-server 实现
- [x] 1.4 目标成本区间由 x402-server 负责收费逻辑
- [x] 1.5 确认 AESC 质押/治理代币 display denom 为 `staex`
- [x] 1.6 确认 AESC 质押/治理代币 base denom 为 `ustaex`（精度 6）

## 2. 实现修复
- [ ] 2.1 实现 epoch 级 Gas 使用率统计（累计 gas_used / gas_limit）并持久化
- [ ] 2.2 将真实使用率接入 `BurnFees()` / `MintInflation()` / `SmoothIncome()`（EVM + Cosmos 合并口径）
- [ ] 2.3 在 `BurnFees()` 中写入 `MonthlyBurnData.BurnedAmount`，修正净供给统计
- [ ] 2.4 修复反向刹车与净供给约束对齐新的统计口径
- [ ] 2.5 在参数验证/治理入口增加硬边界（年通胀≤3%，净供给≤5%，销毁比例 30%-60%）
- [ ] 2.6 高使用率下调销毁比例并补充测试
- [ ] 2.7 主网前更新 `bond_denom` 与相关配置/代码（无需迁移）

## 3. 配置与验证
- [ ] 3.1 更新 genesis/配置模板以符合“初始发行 + Gas Treasury”约束
- [ ] 3.2 增加单测/集成测试覆盖使用率统计、净供给约束与反向刹车
- [ ] 3.3 同步 AESC staking denom 到 genesis/配置/文档
