# 设计文档：重构 Gas 使用率统计与月度数据口径

## 背景
当前 aexburn 模块的 gas 使用率统计和月度数据跟踪存在设计缺陷，导致动态销毁机制无法按规格运行。

### 约束
- 必须保持与现有状态存储的兼容性
- 必须支持 EVM 和 Cosmos 交易的合并 gas 口径
- 不能影响现有的区块处理性能

### 利益相关者
- 验证者：销毁比例直接影响收入
- 代币持有者：销毁/通胀影响供给

## 目标 / 非目标

### 目标
- 实现真实的交易 gas 使用率统计（EVM+Cosmos 合并口径）
- 保存上一周期使用率供下一周期计算使用
- 统一月度数据口径为 epoch
- 实现 12 个月滚动环形缓冲
- 修复本地开发环境配置
- 补充测试覆盖

### 非目标
- 不改变销毁/通胀的业务规则
- 不修改参数边界（30%-60%销毁比例，3%年通胀上限等）
- 不重构整体模块架构
- **不改变 Income Smoother 逻辑**：当前 income smoother 使用硬编码阈值，本次重构不涉及；如需使用真实 gas 使用率，应单独提案

## 决策

### 决策 1：从 txResults 获取真实 GasUsed
- **选择**：在 `app.ProcessBlock` 中遍历 `txResults` 计算 `sum(GasUsed)`
- **原因**：txResults 已包含每笔交易的真实 GasUsed（EVM 已转换到 Sei gas 单位），是最准确的来源
- **替代方案**：继续使用 `ctx.GasMeter()` → 拒绝，因为 EndBlock 的 GasMeter 不代表交易 gas

### 决策 2：Gas Limit Fallback 链（与 checkTotalBlockGas 一致）
- **选择**：`MaxGas` → `MaxGasWanted` → `sum(effectiveGas)`
- **原因**：避免 `max_gas=-1` 配置导致回落 50% 默认值
- **口径一致性**：最终 fallback `sum(effectiveGas)` 与 `app/app.go:checkTotalBlockGas` 完全一致
  - 排除 gasless 交易（`isGasless(tx) == true`）
  - 排除 associate tx（`isAssociateTx(tx) == true`）
  - EVM 交易：`effectiveGas = tx.GetGasEstimate()` 如果 `>= 21000 && <= gasWanted`，否则 `effectiveGas = gasWanted`
  - Cosmos 交易：`effectiveGas = feeTx.GetGas()`
- **注意**：`sum(effectiveGas)` 不等于简单的 `sum(gasWanted)`，而是混合口径
- **替代方案**：仅使用 `MaxGas` → 拒绝，无法处理无限 gas 配置

### 决策 3：通过 Context Value 传递 Gas 数据并在 App 级累积
- **选择**：使用 `ctx.WithValue` 注入 blockGasUsed/Limit，并在 app 级 EndBlock **所有模块之前**调用 `AccumulateBlockGas`
- **原因**：
  - 模块 EndBlock 顺序中 `distr` 在 `aexburn` 前，BurnFees（通过 distr hook 触发）会先于 aexburn EndBlock 执行
  - 如果在 aexburn EndBlock 才累积，BurnFees 读取的 EpochGasData 会缺少当前区块数据
  - 在 app.EndBlocker 开始处（模块循环之前）统一调用 `AccumulateBlockGas`，确保所有模块都能读到最新累积
- **执行位置**：`app/app.go` 的 `EndBlocker` 函数，在 `mm.EndBlock` 之前
- **替代方案**：在 aexburn EndBlock 累积 → 拒绝，无法保证模块顺序

### 决策 4：新增 LastGasUsageRate 状态存储
- **选择**：在 aexburn keeper 中新增 `LastGasUsageRate sdk.Dec` 状态
- **存储位置**：`x/aexburn/types/keys.go` 新增 `LastGasUsageRateKey`
- **更新时机**：epoch 结束时，在重置 EpochGasData 之前保存当前使用率
- **使用场景**：
  - `CalculateDynamicBurnRate`：epoch 内无新数据时使用上一周期使用率
  - 初始状态（无历史数据）：返回 TargetBurnRate
- **替代方案**：不保存，每次无数据返回 50% → 拒绝，违反规格要求

### 决策 5："无数据"的判断标准
- **选择**：以 `EpochGasData.BlockCount == 0` 或 `TotalGasLimit == 0` 判断"无数据"
- **重要**：`gasUsageRate == 0` 是有效值（空块/低活跃期的真实使用率），不能视为"无数据"
- **逻辑**（在 `CalculateDynamicBurnRate` 内部）：
  ```
  epochData := GetEpochGasData(ctx)
  if epochData.BlockCount == 0 || epochData.TotalGasLimit == 0:
      // 真正无数据
      lastRate := GetLastGasUsageRate(ctx)
      if lastRate == 0:
          return TargetBurnRate  // 完全无历史，使用默认
      gasUsageRate = lastRate    // 使用上周期
  else:
      gasUsageRate = epochData.CalculateUsageRate()  // 使用当前真实值（可能为 0）
  // 继续正常计算
  ```
- **原因**：
  - `gasUsageRate == 0` 代表真实的低活跃期，应触发高销毁比例（接近 60%）
  - 误判会掩盖低活跃期的高销毁，违反规格

### 决策 6：直接读取 Epoch Keeper 获取当前 Epoch
- **选择**：通过 `epochKeeper.GetEpoch(ctx).CurrentEpoch` 获取当前 epoch
- **原因**：
  - epoch 模块在 BeginBlock 中先调用 `AfterEpochEnd`，再 SetEpoch 新值
  - 如果 aexburn 在 `AfterEpochEnd` hook 中更新 currentEpoch，会读到旧值（滞后一拍）
  - 直接读取 epoch keeper 可获取最新值，避免自己维护状态
- **依赖注入**：aexburn keeper 需要持有 epochKeeper 引用
- **替代方案**：在 `BeforeEpochStart` hook 更新 → 可行但增加复杂度

### 决策 7：统一月度口径为 Epoch
- **选择**：所有月度计算基于 `epochsPerMonth = epochsPerYear / 12`
- **原因**：与现有 epoch 参数一致，避免 block 高度估算的不确定性
- **替代方案**：继续使用 block 高度 → 拒绝，口径不一致问题持续存在

### 决策 8：12 槽位环形缓冲（通用 helper）
- **选择**：MonthlyBurnData 作为固定 12 个槽位，写入时检测并轮转清零
- **通用化**：抽取 `getOrResetMonthlySlot(ctx, monthIndex, currentEpoch)` helper 函数
  - 检测槽位 epoch 范围是否属于当前月
  - 不属于则清零 BurnedAmount/MintedAmount，重置 StartEpoch/EndEpoch
  - 返回可写入的槽位
- **共用路径**：`updateMonthlyBurnData` 和 `updateMonthlyMintData` 都调用此 helper
- **原因**：
  - 保证 `Get12MonthNetSupply` 始终返回连续 12 个月数据
  - 避免某月只有通胀没有销毁时，旧数据未被清零导致净供给污染
- **替代方案**：保留所有历史数据并过滤 → 拒绝，存储开销大且逻辑复杂

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| Context Value 性能开销 | 仅传递两个 int64 值，开销可忽略 |
| Fallback 链可能返回不准确的 limit | 添加日志记录实际使用的 limit 来源 |
| epochKeeper 依赖注入 | 在 app.go 中正确注入，与其他 keeper 一致 |

## 迁移计划

**无需迁移**：代码尚未上线，不存在存量数据，直接部署新代码即可。

- Genesis 初始化时 `LastGasUsageRate` 默认为 0
- `MonthlyBurnData` 槽位从空开始，随着交易执行自然填充

## 已解决的问题
- ✅ epoch 模块提供 `GetEpoch(ctx).CurrentEpoch` API 可直接使用
- ✅ 代码未上线，无需 upgrade handler

