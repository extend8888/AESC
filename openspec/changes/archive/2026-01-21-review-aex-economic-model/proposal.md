# 变更：Review AEX 经济模型实现一致性

## 为什么
当前链层实现已包含 AEX Gas 相关模块，但尚未系统对照经济模型需求文档进行一致性审查。该 Review 用于识别关键偏差与风险，避免经济模型失真导致的长期激励与安全问题。

## 变更内容
- 对照 `tmp/AESC 公链 Gas (AEX)经济模型.md` 与现有代码/配置，形成差异清单
- 明确修复建议与待决策问题，形成任务清单
- 增补缺失的规格要求，作为后续实现依据

## Review 结果（仅经济模型，排除 x402-server）

### P0 关键偏差
1. **Gas 代币与质押/治理未隔离**
   - 证据：`app/params/config.go` 中 `DefaultBondDenom = uaex`；`depoly-scripts/localnode/aesc_genesis_template.json` 中 `bond_denom` 为 `uaex`
   - 影响：Gas 代币参与质押与治理权重，违背“Gas 不参与节点/质押/治理”的约束
   - 决定：链未启动，主网启动前直接调整 `bond_denom`（无需迁移）

2. **12 个月净供给统计未计入销毁**
   - 证据：`x/aexburn/keeper/burn.go` 未更新 `MonthlyBurnData`；`x/aexburn/keeper/keeper.go` 的 `Get12MonthNetSupply()` 仅统计 `MonthlyBurnData`
   - 影响：净供给约束与反向刹车判断失真

3. **Gas 使用率采集为占位值**
   - 证据：`x/aexburn/keeper/burn.go` 的 `CalculateDynamicBurnRate()` 固定 50%；`x/aexburn/keeper/hooks.go` 与 `x/aexburn/keeper/income_smoother.go` 依赖默认 50%
   - 影响：销毁/通胀/收入平滑不随真实负载变化
   - 决定：使用率口径合并（EVM + Cosmos）

### P1 重要偏差
4. **参数硬边界未强制**
   - 证据：`x/aexburn/types/params.go` 仅验证 0-100%，未限制 3%/5%/30%-60%
   - 影响：治理可突破经济模型硬约束

5. **高使用率下销毁方向与需求不一致**
   - 文档：高使用率下调销毁比例，更多分配给验证者
   - 实现：`CalculateDynamicBurnRate()` 在高使用率上调至 `MaxBurnRate`

### P2 观察项
6. **Gasless 路径削弱“无免 Gas 行为”**
   - 证据：`app/antedecorators/gasless.go` 在 CheckTx 阶段允许特定消息绕过费用校验
   - 决定：保持现状

### 排除项（由 x402-server 实现，不在本次 Review 范围）
- Gas Credit 兜底机制
- 目标成本区间 0.01-0.05 USD 的收费逻辑

## 影响
- 受影响规格：`openspec/specs/aex-gas-token/spec.md`
- 受影响代码：`x/aexburn/`, `app/params/`, `app/antedecorators/`, `depoly-scripts/localnode/`, `x/evm/`
