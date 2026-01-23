# 实现任务清单

## 1. 真实 Gas 使用率统计

### 1.1 定义 Context Key 和辅助函数
- [x] 1.1.1 在 `x/aexburn/types/` 下定义 context key 类型（BlockGasUsedKey, BlockGasLimitKey）
- [x] 1.1.2 添加 `WithBlockGasData(ctx, used, limit)` 和 `GetBlockGasData(ctx)` 辅助函数

### 1.2 在 app.ProcessBlock 中注入 Gas 数据
- [x] 1.2.1 在 `app/app.go` 的 `ProcessBlock` 函数中计算 `blockGasUsed = sum(txResults.GasUsed)`
- [x] 1.2.2 获取 `blockGasLimit`：
  - 优先 `MaxGas`（如果 > 0）
  - fallback 到 `MaxGasWanted`（如果 > 0）
  - 最终 fallback 到 `sum(effectiveGas)`，口径与 `checkTotalBlockGas` 完全一致：
    - 排除 gasless 交易（`isGasless(tx) == true`）
    - 排除 associate tx（`isAssociateTx(tx) == true`）
    - EVM 交易：`effectiveGas = tx.GetGasEstimate()` 如果 `>= 21000 && <= gasWanted`，否则 `effectiveGas = gasWanted`
    - Cosmos 交易：`effectiveGas = feeTx.GetGas()`
- [x] 1.2.3 调用 `aexburntypes.WithBlockGasData(ctx, blockGasUsed, blockGasLimit)` 传递给 EndBlock

### 1.3 在 App 级 EndBlocker 累积 Gas 数据（模块之前）
- [x] 1.3.1 修改 `app/app.go` 的 `EndBlocker` 函数，在 `mm.EndBlock` **之前**：
  - 调用 `aexburntypes.GetBlockGasData(ctx)` 获取注入的数据
  - 调用 `app.AexburnKeeper.AccumulateBlockGas(ctx, blockGasUsed, blockGasLimit)`
- [x] 1.3.2 确保 AccumulateBlockGas 在所有模块 EndBlock 之前执行，保证 `distr` 模块的 BurnFees hook 能读取到包含当前区块的 EpochGasData

### 1.4 移除 aexburn 模块 EndBlock 中的旧累积调用
- [x] 1.4.1 修改 `x/aexburn/module.go` 的 `EndBlock`，移除对 `AccumulateBlockGas` 的调用（已前移到 App 级）
- [x] 1.4.2 保留 EndBlock 中其他逻辑（如有），或改为空实现

### 1.5 重构 AccumulateBlockGas 方法
- [x] 1.5.1 修改 `x/aexburn/keeper/keeper.go` 中 `AccumulateBlockGas` 签名，接收 used/limit 参数
- [x] 1.5.2 移除 `ctx.GasMeter()` 和 `ctx.ConsensusParams()` 的依赖
- [x] 1.5.3 使用传入的参数累加到 `EpochGasData`

### 1.6 新增 LastGasUsageRate 状态存储
- [x] 1.6.1 在 `x/aexburn/types/keys.go` 新增 `LastGasUsageRateKey`
- [x] 1.6.2 在 keeper 中添加 `SetLastGasUsageRate(ctx, rate sdk.Dec)` 方法
- [x] 1.6.3 在 keeper 中添加 `GetLastGasUsageRate(ctx) (sdk.Dec, bool)` 方法（返回存在性标志）
- [x] 1.6.4 在 keeper 中添加 `HasLastGasUsageRate(ctx) bool` 辅助方法
- [x] 1.6.5 在 epoch 结束 hook 中，重置 EpochGasData 前保存当前使用率到 LastGasUsageRate

### 1.7 重构动态销毁比例计算
- [x] 1.7.1 修改 `x/aexburn/keeper/hooks.go` 中 `calculateGasUsageRate`，无数据时返回 0 而非 50%
- [x] 1.7.2 修改 `x/aexburn/keeper/burn.go` 中 `CalculateDynamicBurnRate`：
  - 接收 `epochGasData EpochGasData` 参数（而非 gasUsageRate）
  - 判断"无数据"：`epochGasData.BlockCount == 0 || epochGasData.TotalGasLimit == 0`
  - 无数据时：使用 `GetLastGasUsageRate(ctx)` 获取 `(rate, exists)`
    - `exists == true`：使用历史 rate（0 是有效值，代表低活跃期，触发 MaxBurnRate）
    - `exists == false`：无历史数据，回退到 TargetBurnRate 作为基础值
  - 有数据时：使用 `epochGasData.CalculateUsageRate()` 计算真实使用率
  - **注意**：使用存在性判断（`exists`）而非值判断（`IsZero()`）来区分"无数据"和"数据为 0"
  - **注意**：反向刹车逻辑对所有代码路径统一应用（包括回退到 TargetBurnRate 时）

### 1.8 修改 BurnFees 调用点
- [x] 1.8.1 修改 `x/aexburn/keeper/burn.go` 中 `BurnFees`：
  - 调用 `GetEpochGasData(ctx)` 获取当前 epoch 累积数据（已包含当前区块，因 App 级 EndBlocker 先累积）
  - 将 epochGasData 传入 `CalculateDynamicBurnRate`

## 2. 统一月度数据口径为 Epoch

### 2.1 注入 Epoch Keeper 依赖
- [x] 2.1.1 修改 `x/aexburn/keeper/keeper.go`，添加 `epochKeeper` 字段
- [x] 2.1.2 修改 keeper 构造函数，接收 epochKeeper 参数
- [x] 2.1.3 在 `app/app.go` 中正确注入 epochKeeper 到 aexburnKeeper
- [x] 2.1.4 添加 `getCurrentEpoch(ctx) uint64` 方法，调用 `epochKeeper.GetEpoch(ctx).CurrentEpoch`

### 2.2 实现环形缓冲通用 helper
- [x] 2.2.1 在 `x/aexburn/keeper/keeper.go` 中添加 `getOrResetMonthlySlot(ctx, monthIndex, currentEpoch)` 通用函数：
  - 检测槽位的 StartEpoch/EndEpoch 是否属于当前月
  - 不属于则清零 BurnedAmount 和 MintedAmount，重置 StartEpoch/EndEpoch
  - 返回可写入的槽位
- [x] 2.2.2 确保 `MonthlyBurnData` 结构使用 `StartEpoch`/`EndEpoch` 字段

### 2.3 重构 updateMonthlyBurnData
- [x] 2.3.1 修改 `x/aexburn/keeper/burn.go` 中 `updateMonthlyBurnData`，使用 `getCurrentEpoch(ctx)` 计算月份
- [x] 2.3.2 调用 `getOrResetMonthlySlot` 获取槽位后累加销毁数据

### 2.4 重构 updateMonthlyMintData
- [x] 2.4.1 修改 `x/aexburn/keeper/inflation.go` 中 `updateMonthlyMintData`，使用 `getCurrentEpoch(ctx)` 计算月份
- [x] 2.4.2 调用 `getOrResetMonthlySlot` 获取槽位后累加铸造数据

### 2.5 重构 Get12MonthNetSupply
- [x] 2.5.1 确保只汇总 12 个槽位数据，不会累加多年历史

## 3. 修复本地 Docker 配置

- [x] 3.1 修改 `docker/localnode/scripts/step2_genesis.sh` 中 `bond_denom` 从 `uaex` 改为 `ustaex`

## 4. 补充测试覆盖

### 4.1 Gas 采集测试
- [x] 4.1.1 添加 `AccumulateBlockGas` 使用真实 used/limit 参数的测试（TestAccumulateBlockGas_WithRealParameters）
- [x] 4.1.2 添加 Gas Limit 无效值跳过的测试（TestAccumulateBlockGas_InvalidLimit_Skipped）
- [x] 4.1.3 添加零 gas 使用率区块的测试（TestAccumulateBlockGas_ZeroGasBlock）
- [x] 4.1.4 添加 EpochGasData.CalculateUsageRate 测试（TestEpochGasData_CalculateUsageRate）

### 4.2 LastGasUsageRate 测试
- [x] 4.2.1 添加 `SetLastGasUsageRate`/`GetLastGasUsageRate` 存取测试（TestGetSetLastGasUsageRate）
- [x] 4.2.2 添加零值存取测试验证存在性判断（TestGetSetLastGasUsageRate_ZeroValue）
- [x] 4.2.3 添加无数据时 `CalculateDynamicBurnRate` 使用 LastGasUsageRate 的测试
- [x] 4.2.4 添加完全无数据时返回 TargetBurnRate 的测试

### 4.3 动态销毁测试
- [x] 4.3.1 修改现有 `TestCalculateDynamicBurnRate` 测试，使用可控的 `EpochGasData` 输入
- [x] 4.3.2 添加低/中/高 gas 使用率对应不同销毁比例的测试
- [x] 4.3.3 添加硬边界参数验证测试（TestCalculateDynamicBurnRate_HardBoundaries）
- [x] 4.3.4 添加反向刹车不破坏最小边界的测试（TestCalculateDynamicBurnRate_ReverseBrake_RespectsMinBoundary）

### 4.4 月度数据测试
- [x] 4.4.1 添加环形缓冲 12 槽位数据测试（TestMonthlyBurnData_RingBufferRotation）
- [x] 4.4.2 添加净供应量计算测试（TestGet12MonthNetSupply_Calculation）
- [x] 4.4.3 添加通胀/通缩场景净供应量测试（TestGet12MonthNetSupply_InflationaryCase）
- [x] 4.4.4 添加只返回最近 12 个月数据的测试（TestGet12MonthNetSupply_OnlyReturns12Months）
- [x] 4.4.5 添加 GetOrResetMonthlySlot 轮转清零测试（覆盖旧月份数据重置逻辑）
  - TestGetOrResetMonthlySlot_CurrentMonth_NoReset - 当前月份数据不重置
  - TestGetOrResetMonthlySlot_OldMonth_Resets - 旧月份数据清零（关键测试）
  - TestGetOrResetMonthlySlot_EmptySlot_CreatesNew - 空槽位创建新数据
  - TestGetOrResetMonthlySlot_YearRollover - 跨年轮转测试

### 4.5 集成测试
- [x] 4.5.1 添加 epoch 间数据重置测试（TestEpochGasData_ResetBetweenEpochs）
- [x] 4.5.2 添加多种 gas 使用率场景测试（TestGasUsageRate_VariousScenarios）
- [x] 4.5.3 添加完整 BurnFees 集成测试（TestBurnFees_Integration）

### 4.6 CalculateGasUsageRate 测试（直接调用函数）
- [x] 4.6.1 添加"无数据返回 0"分支测试（TestCalculateCurrentGasUsageRate_NoData_ReturnsZero）
  - 直接调用 Keeper.CalculateCurrentGasUsageRate 验证无数据时返回 0
- [x] 4.6.2 添加有数据时返回正确率测试（TestCalculateCurrentGasUsageRate_WithData_ReturnsCorrectRate）
- [x] 4.6.3 添加所有条件分支覆盖测试（TestCalculateCurrentGasUsageRate_AllConditions）
  - 验证 BlockCount=0 或 TotalGasLimit=0 时返回零
  - 验证正常情况返回正确使用率

### 4.7 App 级 CalculateBlockGasLimit Fallback 链测试
- [x] 4.7.1 添加 MaxGas 优先级测试（TestCalculateBlockGasLimit_MaxGasPriority）
- [x] 4.7.2 添加 MaxGas=-1 时回退到 MaxGasWanted 测试（TestCalculateBlockGasLimit_MaxGasWantedFallback）
- [x] 4.7.3 添加 MaxGas=0 时回退测试（TestCalculateBlockGasLimit_MaxGasZeroFallback）
- [x] 4.7.4 添加两者都为 0 时回退到交易 gas 总和测试（TestCalculateBlockGasLimit_BothZero_EmptyTxs）
- [x] 4.7.5 添加 nil ConsensusParams 处理测试（TestCalculateBlockGasLimit_NilConsensusParams）

### 4.8 模块顺序测试
- [x] 4.8.1 添加 AexburnKeeper 正确初始化验证（TestEndBlockerModuleOrder_AexburnBeforeMint）

## 5. 问题修复（补充）

### 5.1 区分"无历史值"和"历史值=0"
- [x] 5.1.1 修改 `GetLastGasUsageRate` 返回 `(rate, exists)` 模式，使用存储存在性判断
- [x] 5.1.2 添加 `HasLastGasUsageRate(ctx) bool` 辅助方法
- [x] 5.1.3 修改 `CalculateDynamicBurnRate` 使用 `exists` 判断而非 `IsZero()`
- [x] 5.1.4 添加测试验证 0% 历史值触发 MaxBurnRate（TestCalculateDynamicBurnRate_ZeroHistoryRate_DistinguishFromNoHistory）

### 5.2 无历史时仍应用反向刹车
- [x] 5.2.1 重构 `CalculateDynamicBurnRate` 确保反向刹车逻辑在所有代码路径末尾统一应用
- [x] 5.2.2 添加测试验证无历史+反向刹车场景（TestCalculateDynamicBurnRate_NoHistory_WithReverseBrake）

### 5.3 LastGasUsageRate 纳入 Genesis 导出/导入
- [x] 5.3.1 修改 `proto/aexburn/genesis.proto` 添加 `last_gas_usage_rate` 字段（nullable）
- [x] 5.3.2 运行 `make proto-gen` 重新生成 pb.go
- [x] 5.3.3 修改 `x/aexburn/genesis.go` 的 `InitGenesis` 导入 LastGasUsageRate
- [x] 5.3.4 修改 `x/aexburn/genesis.go` 的 `ExportGenesis` 导出 LastGasUsageRate

