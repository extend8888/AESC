## MODIFIED Requirements

### Requirement: 手续费销毁机制
系统必须（SHALL）销毁部分交易手续费，销毁比例在 30%-60% 之间动态调节。

#### Scenario: 动态销毁比例计算
- **GIVEN** 当前 Gas 使用率
- **WHEN** 计算销毁比例
- **THEN** 使用率低时销毁比例接近 30%
- **AND** 使用率正常时销毁比例约 50%
- **AND** 使用率过高时销毁比例下调（不高于目标值），更多分配给验证者

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

## ADDED Requirements

### Requirement: 质押/治理载体隔离
系统必须（MUST）确保质押/治理载体与 Gas 代币隔离，不得使用 AEX/uaex 作为 `bond_denom`。

#### Scenario: bond_denom 不得为 uaex
- **GIVEN** 系统已初始化
- **WHEN** 查询 staking 参数
- **THEN** `bond_denom` 不得为 `uaex`
- **AND** `bond_denom` 必须与 Gas 代币面额不同

### Requirement: 经济模型参数硬边界
系统必须（MUST）拒绝突破以下硬边界的参数变更：年通胀上限 3%、12 个月净供给上限 5%、销毁比例区间 30%-60%。

#### Scenario: 拒绝超限治理
- **GIVEN** 治理提案将年通胀上限设置为 4%
- **WHEN** 参数校验执行
- **THEN** 提案必须被拒绝
