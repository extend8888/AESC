# AEX 经济模型实现 Review（链层）

## 已实现内容（确认符合）

| 项目 | 现状 | 证据 | 备注 |
|---|---|---|---|
| Gas 代币基础配置 | 已配置 AEX/uaex/6 位精度 | `app/params/config.go` | 规格一致 |
| 手续费销毁流程接入分配 | Burn hook 已接入分配流程 | `sei-cosmos/x/distribution/keeper/allocation.go` | 先销毁再分配 |
| 动态销毁与通胀模块 | aexburn 模块实现销毁/通胀/统计 | `x/aexburn/keeper/burn.go`, `x/aexburn/keeper/inflation.go` | 逻辑齐全 |
| 反向刹车机制 | epoch 结束调用反向刹车 | `x/aexburn/keeper/hooks.go` | 已接入 |
| 收入平滑机制（默认关闭） | 可选收入平滑模块 | `x/aexburn/keeper/income_smoother.go`, `x/aexburn/types/params.go` | 默认关闭 |
| 查询与测试 | 提供统计查询与单测 | `x/aexburn/keeper/grpc_query.go`, `x/aexburn/keeper/*_test.go` | 覆盖核心路径 |
| Gasless 例外 | 现有 gasless 逻辑保持现状 | `app/antedecorators/gasless.go` | 不调整 |

## 应调整内容（需实现变更）

| 项目 | 需调整点 | 证据 | 影响 |
|---|---|---|---|
| 质押/治理载体隔离 | 主网前调整 `bond_denom`（无需迁移），不得为 `uaex` | `app/params/config.go`, `depoly-scripts/localnode/aesc_genesis_template.json` | 违背 Gas 隔离约束 |
| Gas 使用率采集 | 由占位值改为真实统计（epoch 累计 gas_used/gas_limit，EVM + Cosmos 合并口径） | `x/aexburn/keeper/burn.go`, `x/aexburn/keeper/hooks.go`, `x/aexburn/keeper/income_smoother.go` | 销毁/通胀/平滑失真 |
| 月度净供给统计 | `BurnFees` 需写入 `MonthlyBurnData.BurnedAmount` | `x/aexburn/keeper/burn.go`, `x/aexburn/keeper/keeper.go` | 净供给约束失准 |
| 高使用率销毁方向 | 高使用率下应下调销毁比例 | `x/aexburn/keeper/burn.go` | 与文档方向不一致 |
| 参数硬边界 | 强制 3%/5%/30-60% 硬约束 | `x/aexburn/types/params.go` | 防止治理突破约束 |
| 初始发行去向 | 初始发行应进入 Gas Treasury（非多地址分散） | `depoly-scripts/localnode/aesc_genesis_template.json` | 与文档不一致 |

## 需确认内容（决策输入）

| 项目 | 需要确认的问题 | 备注 |
|---|---|---|
| 质押/治理 denom | AESC display denom 为 `staex`，base denom 为 `ustaex`（精度 6） | 用于替换 `bond_denom` |
