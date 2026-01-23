# 变更：重构 Gas 使用率统计与月度数据口径

## 为什么
当前实现存在以下严重问题：
1. **动态销毁使用硬编码 50% gas 使用率**：`CalculateDynamicBurnRate` 始终返回基于 50% 的固定销毁比例，未接入真实数据
2. **Gas 使用率采集逻辑错误**：`AccumulateBlockGas` 依赖 EndBlock 的 `ctx.GasMeter()` 和 `MaxGas>0`，这既不等于交易 gas（EVM+Cosmos 合并口径），也会在 `max_gas=-1` 时完全跳过并回落默认 50%
3. **月度数据口径不一致**：销毁用 block 高度估算月份，铸造用 epoch 估算月份，且 `Get12MonthNetSupply` 直接汇总所有月度数据，可能把多年数据叠加成"12个月"
4. **本地配置不一致**：docker 启动脚本把 `bond_denom` 设为 `uaex`，与 `ustaex` 质押配置不一致

## 变更内容

### 1. 从真实交易结果统计 Gas（EVM+Cosmos 合并口径）
- **破坏性**：移除 `AccumulateBlockGas` 中使用 `ctx.GasMeter()` 的错误逻辑
- 在 `app.ProcessBlock` 中计算 `blockGasUsed`（sum 所有 txResults.GasUsed）
- 通过 `ctx.WithValue` 注入到 EndBlock 供 aexburn 模块使用
- **App 级累积**：在 `app.EndBlocker` 的模块循环之前调用 `AccumulateBlockGas`，确保 BurnFees（distr hook）能读取最新数据
- 移除 50% 硬编码默认值，无数据时返回 0

### 2. Gas Limit 来源 Fallback 链
- 优先使用 `ctx.ConsensusParams().Block.MaxGas`
- 如果 `MaxGas <= 0`，退化为 `MaxGasWanted`
- 如果仍为 0，退化为 `sum(gasWanted)`

### 3. 新增持久化状态：LastGasUsageRate
- **新增状态**：`LastGasUsageRate sdk.Dec`，保存上一 epoch 的 gas 使用率
- **用途**：当前 epoch 无数据时使用上周期使用率计算销毁比例
- **更新时机**：epoch 结束时，在重置 EpochGasData 之前保存

### 4. 新增依赖注入：epochKeeper
- **新增依赖**：aexburn keeper 需要持有 `epochKeeper` 引用
- **用途**：通过 `epochKeeper.GetEpoch(ctx).CurrentEpoch` 获取当前 epoch，用于月度数据口径统一

### 5. 统一月度数据口径为 Epoch
- **破坏性**：`updateMonthlyBurnData` 改为基于 epoch 计算月份（与 `updateMonthlyMintData` 一致）
- `monthIndex = (currentEpoch / epochsPerMonth) % 12`
- 实现 12 槽位环形缓冲，写入时自动轮转清零
- 抽取通用 helper `getOrResetMonthlySlot`，供 burn/mint 两条路径共用

### 6. 修复本地 Docker 配置
- 将 `docker/localnode/scripts/step2_genesis.sh` 中 `bond_denom` 改为 `ustaex`

### 7. 补充测试覆盖
- 添加真实 gas 采集的行为测试
- 添加月度销毁写入/轮转的测试
- 添加硬边界参数验证测试（30%-60% 范围）
- 添加 App 级累积顺序测试

## 影响
- 受影响的规格：`aex-gas-token`、`deployment-tools`
- **新增持久化状态**：`LastGasUsageRate`（x/aexburn/types/keys.go）
- **新增依赖**：aexburn keeper → epochKeeper
- 受影响的代码：
  - `x/aexburn/keeper/burn.go` - 动态销毁计算、月度数据更新
  - `x/aexburn/keeper/keeper.go` - Gas 累计逻辑、12个月净供给计算、LastGasUsageRate 存取、通用 helper
  - `x/aexburn/keeper/hooks.go` - Gas 使用率计算
  - `x/aexburn/keeper/inflation.go` - 月度铸造数据更新（使用通用 helper）
  - `x/aexburn/types/keys.go` - 新增 LastGasUsageRateKey
  - `app/app.go` - 注入 blockGasUsed/Limit 到 context、App 级 EndBlocker 累积、epochKeeper 注入
  - `docker/localnode/scripts/step2_genesis.sh` - 修复 bond_denom

## 风险评估
- **高风险**：改变 gas 统计来源可能影响销毁比例计算，需要充分测试
- **中风险**：月度数据口径变更，代码尚未上线无需迁移
- **低风险**：docker 配置变更仅影响本地开发环境

