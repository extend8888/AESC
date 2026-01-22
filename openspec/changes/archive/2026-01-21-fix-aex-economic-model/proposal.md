# 变更：修复 AEX 经济模型实现问题

## 为什么

根据 `review-aex-economic-model` 代码审查，发现当前 AEX 经济模型实现存在多个与规格文档不一致的问题，需要在主网发布前进行修复。

## 变更内容

### 1. 质押/治理载体隔离（**破坏性**）
- 将 `bond_denom` 从 `uaex` 改为 `ustaex`
- AEX 仅用于 Gas，不参与质押/治理
- 新增 STAEX 代币配置（display: `staex`, base: `ustaex`, 精度 6）
- 更新 genesis 模板

### 2. Gas 使用率真实采集
- 移除 `calculateGasUsageRate` 中的占位值（50%）
- 实现 epoch 周期内累计 gas_used/gas_limit 统计
- 合并 EVM + Cosmos 交易的 Gas 数据

### 3. 月度销毁数据记录
- `BurnFees` 执行时写入 `MonthlyBurnData.BurnedAmount`
- 确保月度净供给约束数据准确

### 4. 高使用率销毁方向修正
- 当前：高使用率 → 提高销毁比例（错误）
- 修正：高使用率 → 降低销毁比例
- 理由：高活跃度时应保留更多代币给验证者

### 5. 参数硬边界约束
- 强制 `MaxAnnualInflationRate` ≤ 3%
- 强制 `MaxNetSupplyRatePerYear` ≤ 5%
- 强制 `MinBurnRate` ≥ 30%，`MaxBurnRate` ≤ 60%
- 在 `Params.Validate()` 中添加硬约束检查

### 6. 初始发行去向（暂不处理）
- 需要确认 Gas Treasury 地址
- 后续单独变更处理

## 影响

- 受影响的规格：`aex-gas-token`
- 受影响的代码：
  - `app/params/config.go`（新增 STAEX 配置）
  - `x/aexburn/keeper/burn.go`（销毁方向、月度数据）
  - `x/aexburn/keeper/hooks.go`（Gas 使用率统计）
  - `x/aexburn/types/params.go`（硬边界约束）
  - `depoly-scripts/localnode/aesc_genesis_template.json`（bond_denom）
  - 相关测试文件

## 依赖

- 需要确认：STAEX display denom 为 `staex`，base denom 为 `ustaex`（精度 6）

