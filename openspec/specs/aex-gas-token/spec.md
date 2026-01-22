# AEX Gas Token 规格

> 归档自变更：`aex-gas-token`, `review-aex-gas-token` (2026-01-19), `aex-staex-token-split` (2026-01-21)

## Requirements

### Requirement: 双代币架构
系统必须（SHALL）实现 AEX 和 STAEX 双代币架构，分离 Gas 和质押功能。

#### Scenario: AEX Gas 代币配置
- **GIVEN** 系统已初始化
- **WHEN** 查询 Gas 代币配置
- **THEN** 代币名称应为 "AEX"
- **AND** 最小单位应为 `uaex`
- **AND** 精度应为 6 (1 AEX = 10^6 uaex)
- **AND** 初始发行量应为 500,000,000 AEX

#### Scenario: AEX 代币用途限定
- **GIVEN** AEX 代币
- **WHEN** 检查代币用途
- **THEN** 仅用于交易手续费（Gas）
- **AND** 不参与质押、治理或验证者激励

#### Scenario: STAEX 质押代币配置
- **GIVEN** 系统已初始化
- **WHEN** 查询质押代币配置
- **THEN** 代币名称应为 "STAEX"
- **AND** 最小单位应为 `ustaex`
- **AND** 精度应为 6 (1 STAEX = 10^6 ustaex)
- **AND** 链的 `bond_denom` 应设置为 `ustaex`

#### Scenario: STAEX 代币用途限定
- **GIVEN** STAEX 代币
- **WHEN** 检查代币用途
- **THEN** 用于验证者质押（staking）
- **AND** 用于治理投票
- **AND** 验证者激励以 STAEX 发放

### Requirement: 通胀机制
系统必须（SHALL）实现可控的通胀机制，年通胀上限为 3%（硬约束）。

#### Scenario: 年通胀上限约束
- **GIVEN** 当前年度已通胀量
- **WHEN** 计算本次通胀量
- **THEN** 累计年通胀不得超过初始供给的 3%
- **AND** 此为硬约束，不可通过治理修改

#### Scenario: 通胀触发条件
- **GIVEN** 一个 epoch 周期结束
- **WHEN** Gas 使用率超过阈值（默认 50%）
- **THEN** 系统触发通胀铸造
- **AND** 铸造量基于使用率动态计算

### Requirement: 净供给硬约束
系统必须（SHALL）确保任意连续 12 个月的净增发不超过初始供给的 5%（硬约束）。

#### Scenario: 12个月滚动窗口约束
- **GIVEN** 过去 12 个月的通胀和销毁数据
- **WHEN** 计算净供给变化
- **THEN** 净增发量不得超过初始供给的 5%
- **AND** 如超出限制，阻止进一步通胀
- **AND** 此为硬约束，不可通过治理修改

### Requirement: 手续费销毁机制
系统必须（SHALL）销毁部分交易手续费，销毁比例在 30%-60% 之间动态调节（硬约束边界）。

#### Scenario: 动态销毁比例计算（保护验证者收益）
- **GIVEN** 当前 Gas 使用率
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
- **THEN** 当前月份的 `BurnedAmount` 累加销毁量
- **AND** 更新 `MonthlyBurnData` 状态

### Requirement: 净供给反向刹车机制
系统必须（MUST）实现净供给反向刹车机制，防止过度通缩。

#### Scenario: 连续净负时自动下调销毁
- **GIVEN** 过去3个统计周期的净供给数据
- **WHEN** 连续3个周期净供给均为负
- **THEN** 系统自动将销毁比例下调至区间下限（30%）
- **AND** 记录调整事件

#### Scenario: 净供给恢复后维持正常销毁
- **GIVEN** 净供给已恢复为正或零
- **WHEN** 计算下一周期销毁比例
- **THEN** 恢复使用基于 Gas 使用率的动态计算逻辑

### Requirement: 真实 Gas 使用率计算
系统必须（MUST）基于真实的链上 Gas 使用数据计算使用率，使用 EpochGasData 结构跟踪。

#### Scenario: 基于 Epoch Gas 统计计算使用率
- **GIVEN** 一个 epoch 周期结束
- **WHEN** 计算该周期的 Gas 使用率
- **THEN** 从 EpochGasData 获取累计 gas_used 和 gas_limit
- **AND** 使用率 = gas_used / gas_limit
- **AND** 返回介于 0 和 1 之间的使用率

#### Scenario: EndBlocker 更新 Gas 统计
- **GIVEN** 每个区块结束时
- **WHEN** EndBlocker 被调用
- **THEN** 累加当前区块的 gas_used 到 EpochGasData
- **AND** 累加当前区块的 gas_limit 到 EpochGasData

#### Scenario: Epoch 结束时重置统计
- **GIVEN** 当前 epoch 周期结束
- **WHEN** 开始新的 epoch
- **THEN** 重置 EpochGasData 的 gas_used 和 gas_limit 为 0
- **AND** 保存上一周期的使用率用于计算

### Requirement: 验证者收入平滑机制
系统应（SHOULD）提供可选的验证者收入平滑机制，默认关闭。

#### Scenario: 高活跃期收入缓冲
- **GIVEN** 收入平滑机制已启用
- **WHEN** Gas 使用率超过 70%
- **THEN** 抽取 10% 验证者收入到缓冲池

#### Scenario: 低活跃期收入补贴
- **GIVEN** 收入平滑机制已启用
- **WHEN** Gas 使用率低于 30%
- **THEN** 从缓冲池释放 5% 补贴给验证者

### Requirement: 单元测试覆盖
aexburn 模块必须（MUST）具备完整的单元测试覆盖。

#### Scenario: 销毁逻辑测试
- **GIVEN** aexburn keeper 测试环境
- **WHEN** 调用 `BurnFees()` 方法
- **THEN** 验证销毁金额符合动态比例计算
- **AND** 验证状态正确更新

#### Scenario: 通胀逻辑测试
- **GIVEN** aexburn keeper 测试环境
- **WHEN** 调用 `MintInflation()` 方法
- **THEN** 验证年上限约束生效
- **AND** 验证12个月净供给约束生效

#### Scenario: 反向刹车测试
- **GIVEN** 连续3个周期净供给为负的测试数据
- **WHEN** 计算下一周期销毁比例
- **THEN** 验证销毁比例自动下调

