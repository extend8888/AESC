# AEX Gas Token Review 发现的问题

> 本文档记录对 aex-gas-token 实现的 Review 发现，作为后续修复的规格依据。

---

## ADDED Requirements

### Requirement: 净供给反向刹车机制

系统必须（MUST）实现净供给反向刹车机制，当连续3个统计周期出现净供给为负（销毁量 > 通胀量）时，系统应自动下调销毁比例，直至净供给恢复至安全区间。

#### Scenario: 连续净负时自动下调销毁

- **GIVEN** 过去3个统计周期的净供给数据
- **WHEN** 连续3个周期净供给均为负
- **THEN** 系统自动将销毁比例下调至区间下限（30%）
- **AND** 记录调整事件

#### Scenario: 净供给恢复后维持正常销毁

- **GIVEN** 净供给已恢复为正或零
- **WHEN** 计算下一周期销毁比例
- **THEN** 恢复使用基于 Gas 使用率的动态计算逻辑

---

### Requirement: 销毁数据月度记录

系统必须（MUST）在每次销毁手续费时，将销毁量记录到月度数据中，以便正确计算12个月滚动窗口的净供给。

#### Scenario: 销毁时更新月度数据

- **GIVEN** 一笔手续费销毁交易
- **WHEN** 销毁执行成功
- **THEN** 当前月份的 `BurnedAmount` 累加销毁量
- **AND** 更新 `MonthlyBurnData` 状态

---

### Requirement: 真实 Gas 使用率计算

系统必须（MUST）基于真实的链上 Gas 使用数据计算使用率，而非使用固定的默认值。

#### Scenario: 基于实际区块 Gas 计算使用率

- **GIVEN** 一个 epoch 周期结束
- **WHEN** 计算该周期的 Gas 使用率
- **THEN** 使用该周期内所有区块的累计 Gas 消耗
- **AND** 除以该周期内所有区块的累计 Gas 上限
- **AND** 返回介于 0 和 1 之间的使用率

---

## MODIFIED Requirements

### Requirement: 单元测试覆盖

aexburn 模块必须（MUST）具备完整的单元测试覆盖，包括销毁逻辑、通胀逻辑、动态调节逻辑和 epoch hooks。

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

