## MODIFIED Requirements

### Requirement: 真实 Gas 使用率计算
系统必须（MUST）基于真实的链上交易 Gas 使用数据计算使用率，使用 EpochGasData 结构跟踪，合并 EVM 和 Cosmos 交易的 gas 统计，并保存上一周期使用率供无数据时使用。

#### Scenario: 基于 Epoch Gas 统计计算使用率
- **GIVEN** 一个 epoch 周期结束
- **WHEN** 计算该周期的 Gas 使用率
- **THEN** 从 EpochGasData 获取累计 gas_used 和 gas_limit
- **AND** 使用率 = gas_used / gas_limit
- **AND** 返回介于 0 和 1 之间的使用率

#### Scenario: 从交易结果统计真实 GasUsed
- **GIVEN** ProcessBlock 执行完所有交易
- **WHEN** 计算区块 gas 使用量
- **THEN** 遍历所有 txResults 累加 GasUsed（EVM+Cosmos 合并口径）
- **AND** 将 blockGasUsed 通过 context 注入传递给 EndBlock

#### Scenario: Gas Limit Fallback 链
- **GIVEN** 需要获取区块 gas limit
- **WHEN** `ctx.ConsensusParams().Block.MaxGas > 0`
- **THEN** 使用 MaxGas 作为 gas limit
- **WHEN** `MaxGas <= 0` 且 `MaxGasWanted > 0`
- **THEN** 使用 MaxGasWanted 作为 fallback
- **WHEN** `MaxGas <= 0` 且 `MaxGasWanted = 0`
- **THEN** 使用 `sum(gasWanted)` 作为最终 fallback，口径与 checkTotalBlockGas 一致（排除 gasless/associate tx，EVM 优先使用 gasEstimate）

#### Scenario: EndBlocker 更新 Gas 统计
- **GIVEN** 每个区块结束时
- **WHEN** EndBlocker 被调用
- **THEN** 从 context 获取注入的 blockGasUsed 和 blockGasLimit
- **AND** 调用 AccumulateBlockGas 累加到 EpochGasData

#### Scenario: Epoch 结束时保存使用率并重置统计
- **GIVEN** 当前 epoch 周期结束
- **WHEN** 开始新的 epoch
- **THEN** 先保存当前使用率到 LastGasUsageRate 状态
- **AND** 然后重置 EpochGasData 的 gas_used 和 gas_limit 为 0

#### Scenario: "无数据"判断标准
- **GIVEN** 需要判断 epoch 是否有 gas 数据
- **WHEN** 判断"无数据"条件
- **THEN** 使用 `EpochGasData.BlockCount == 0` 或 `TotalGasLimit == 0` 判断
- **AND** `gasUsageRate == 0` 视为有效值（代表低活跃期），不视为"无数据"

#### Scenario: 无数据时使用上周期使用率
- **GIVEN** 当前 epoch 满足"无数据"条件（BlockCount == 0 或 TotalGasLimit == 0）
- **WHEN** 计算动态销毁比例
- **THEN** CalculateDynamicBurnRate 使用 LastGasUsageRate 进行计算

#### Scenario: 完全无数据时使用 TargetBurnRate
- **GIVEN** 满足"无数据"条件且 LastGasUsageRate 也为 0（如链刚启动）
- **WHEN** 计算动态销毁比例
- **THEN** 直接返回 TargetBurnRate（而非硬编码的 50% 对应的比例）

#### Scenario: 低活跃期使用真实使用率
- **GIVEN** 当前 epoch 有数据（BlockCount > 0 且 TotalGasLimit > 0）但 gasUsageRate == 0（空块）
- **WHEN** 计算动态销毁比例
- **THEN** 使用真实的 gasUsageRate = 0 进行计算
- **AND** 触发高销毁比例（接近 60%），符合规格预期

### Requirement: 手续费销毁机制
系统必须（SHALL）销毁部分交易手续费，销毁比例在 30%-60% 之间动态调节（硬约束边界）。

#### Scenario: 动态销毁比例计算（保护验证者收益）
- **GIVEN** 当前 Gas 使用率（从真实交易数据计算）
- **WHEN** 计算销毁比例
- **THEN** 使用率低时销毁比例接近 60%（低活跃期多销毁）
- **AND** 使用率正常时销毁比例约 45%
- **AND** 使用率高时销毁比例接近 30%（高活跃期少销毁，保留更多给验证者）
- **AND** 销毁比例边界 30%-60% 为硬约束

#### Scenario: 销毁执行
- **GIVEN** 一笔交易支付手续费
- **WHEN** 手续费处理完成
- **THEN** 按计算比例销毁部分手续费
- **AND** 剩余部分分配给验证者

#### Scenario: 销毁时更新月度数据
- **GIVEN** 一笔手续费销毁交易
- **WHEN** 销毁执行成功
- **THEN** 基于当前 epoch 计算月份索引
- **AND** 当前月份的 `BurnedAmount` 累加销毁量
- **AND** 更新 `MonthlyBurnData` 状态

### Requirement: 净供给硬约束
系统必须（SHALL）确保任意连续 12 个月的净增发不超过初始供给的 5%（硬约束）。

#### Scenario: 12个月滚动窗口约束
- **GIVEN** 过去 12 个月的通胀和销毁数据（基于 epoch 的 12 槽位环形缓冲）
- **WHEN** 计算净供给变化
- **THEN** 净增发量不得超过初始供给的 5%
- **AND** 如超出限制，阻止进一步通胀
- **AND** 此为硬约束，不可通过治理修改

#### Scenario: 月度数据统一使用 Epoch 口径
- **GIVEN** 需要更新月度销毁或铸造数据
- **WHEN** 计算月份索引
- **THEN** 通过 `epochKeeper.GetEpoch(ctx).CurrentEpoch` 获取当前 epoch
- **AND** 使用 `monthIndex = (currentEpoch / epochsPerMonth) % 12`
- **AND** `epochsPerMonth = epochsPerYear / 12`
- **AND** 销毁和铸造使用相同口径

#### Scenario: 环形缓冲轮转清零（通用 helper）
- **GIVEN** 写入 MonthlyBurnData 到某个槽位（销毁或铸造）
- **WHEN** 该槽位记录的 epoch 范围不属于当前月
- **THEN** 调用通用 helper `getOrResetMonthlySlot(ctx, monthIndex, currentEpoch)`
- **AND** helper 清零该槽位的 BurnedAmount 和 MintedAmount
- **AND** helper 重置 StartEpoch/EndEpoch 为当前 epoch
- **AND** 然后累加本次数据
- **AND** `updateMonthlyBurnData` 和 `updateMonthlyMintData` 共用此 helper

#### Scenario: 12个月净供给计算
- **GIVEN** 调用 Get12MonthNetSupply
- **WHEN** 汇总月度数据
- **THEN** 只汇总 12 个槽位的数据
- **AND** 不会累加多年历史数据

## ADDED Requirements

### Requirement: LastGasUsageRate 状态存储
系统必须（MUST）保存上一周期的 gas 使用率，供当前周期无数据时使用。

#### Scenario: Epoch 结束时保存使用率
- **GIVEN** 当前 epoch 结束
- **WHEN** 触发 AfterEpochEnd hook
- **THEN** 将当前 EpochGasData 计算的使用率保存到 LastGasUsageRate
- **AND** 然后再重置 EpochGasData

#### Scenario: 获取上周期使用率
- **GIVEN** LastGasUsageRate 已保存
- **WHEN** 当前周期无累积数据
- **THEN** CalculateDynamicBurnRate 使用 LastGasUsageRate 计算销毁比例

### Requirement: 单元测试覆盖 - Gas 统计
aexburn 模块必须（MUST）具备真实 gas 统计逻辑的单元测试。

#### Scenario: 真实 Gas 累积测试
- **GIVEN** aexburn keeper 测试环境
- **WHEN** 调用 `AccumulateBlockGas(ctx, used, limit)` 传入真实数据
- **THEN** 验证 EpochGasData 正确累加

#### Scenario: Gas Limit Fallback 测试
- **GIVEN** 不同的 MaxGas/MaxGasWanted 配置
- **WHEN** 获取 gas limit
- **THEN** 验证 fallback 链按预期工作

#### Scenario: 无数据时返回零使用率测试
- **GIVEN** EpochGasData 为空或全零
- **WHEN** 调用 calculateGasUsageRate
- **THEN** 返回 0 而非 50%

#### Scenario: LastGasUsageRate 存取测试
- **GIVEN** 保存 LastGasUsageRate 值
- **WHEN** 调用 GetLastGasUsageRate
- **THEN** 返回之前保存的值

#### Scenario: 无数据时使用 LastGasUsageRate 测试
- **GIVEN** 当前周期无数据，LastGasUsageRate = 0.6
- **WHEN** 调用 CalculateDynamicBurnRate
- **THEN** 使用 0.6 计算销毁比例

#### Scenario: 完全无数据时使用 TargetBurnRate 测试
- **GIVEN** 当前周期无数据，LastGasUsageRate = 0
- **WHEN** 调用 CalculateDynamicBurnRate
- **THEN** 直接返回 TargetBurnRate

### Requirement: 单元测试覆盖 - 月度数据
aexburn 模块必须（MUST）具备月度数据口径和轮转的单元测试。

#### Scenario: 月度数据基于 Epoch 测试
- **GIVEN** 当前 epoch 号
- **WHEN** 调用 updateMonthlyBurnData
- **THEN** 验证使用 epoch 计算月份而非 block 高度

#### Scenario: 环形缓冲轮转测试
- **GIVEN** 已有旧月份的数据在某槽位
- **WHEN** 新月份写入同一槽位
- **THEN** 验证旧数据被清零后再累加新数据

#### Scenario: 跨年净供给计算测试
- **GIVEN** 超过 12 个月的历史数据
- **WHEN** 调用 Get12MonthNetSupply
- **THEN** 只返回最近 12 个月数据，不累加历史

