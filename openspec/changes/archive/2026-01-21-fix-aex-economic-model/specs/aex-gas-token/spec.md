## MODIFIED Requirements

### Requirement: Gas 代币基础配置
系统必须（SHALL）使用 AEX 作为链的原生 Gas 代币，STAEX 作为质押/治理代币。

#### Scenario: AEX 代币参数配置
- **GIVEN** 系统已初始化
- **WHEN** 查询 Gas 代币配置
- **THEN** 代币名称应为 "AEX"
- **AND** 最小单位应为 `uaex`
- **AND** 精度应为 6 (1 AEX = 10^6 uaex)
- **AND** 初始发行量应为 500,000,000 AEX

#### Scenario: STAEX 代币参数配置
- **GIVEN** 系统已初始化
- **WHEN** 查询质押代币配置
- **THEN** 代币名称应为 "STAEX"
- **AND** 最小单位应为 `ustaex`
- **AND** 精度应为 6 (1 STAEX = 10^6 ustaex)
- **AND** `bond_denom` 应为 `ustaex`

#### Scenario: Gas 代币用途限定
- **GIVEN** AEX 代币
- **WHEN** 检查代币用途
- **THEN** 仅用于交易手续费（Gas）
- **AND** 不参与节点/质押/裂变激励

### Requirement: 真实 Gas 使用率计算
系统必须（MUST）基于真实的链上 Gas 使用数据计算使用率。

#### Scenario: Epoch 累计 Gas 统计
- **GIVEN** 一个 epoch 周期进行中
- **WHEN** 每个区块结束时
- **THEN** 累加该区块的 gas_used 到 epoch 累计值
- **AND** 累加该区块的 gas_limit 到 epoch 累计值
- **AND** 包含 EVM 和 Cosmos 交易的 Gas 数据

#### Scenario: 基于实际区块 Gas 计算使用率
- **GIVEN** 一个 epoch 周期结束
- **WHEN** 计算该周期的 Gas 使用率
- **THEN** 使用该周期内所有区块的累计 Gas 消耗
- **AND** 除以该周期内所有区块的累计 Gas 上限
- **AND** 返回介于 0 和 1 之间的使用率

### Requirement: 手续费销毁机制
系统必须（SHALL）销毁部分交易手续费，销毁比例在 30%-60% 之间动态调节。

#### Scenario: 动态销毁比例计算（修正方向）
- **GIVEN** 当前 Gas 使用率
- **WHEN** 计算销毁比例
- **THEN** 使用率低时销毁比例接近 60%（高销毁）
- **AND** 使用率正常时销毁比例约 50%
- **AND** 使用率高时销毁比例接近 30%（低销毁，保留更多给验证者）

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

### Requirement: 参数硬边界约束
系统必须（MUST）强制执行经济参数的硬边界约束，防止治理突破安全限制。

#### Scenario: 通胀率硬边界
- **GIVEN** 治理提案修改 `MaxAnnualInflationRate`
- **WHEN** 验证参数
- **THEN** 必须 ≤ 3%，否则拒绝

#### Scenario: 净供给率硬边界
- **GIVEN** 治理提案修改 `MaxNetSupplyRatePerYear`
- **WHEN** 验证参数
- **THEN** 必须 ≤ 5%，否则拒绝

#### Scenario: 销毁比例硬边界
- **GIVEN** 治理提案修改销毁比例参数
- **WHEN** 验证参数
- **THEN** `MinBurnRate` 必须 ≥ 30%
- **AND** `MaxBurnRate` 必须 ≤ 60%
- **AND** 否则拒绝

