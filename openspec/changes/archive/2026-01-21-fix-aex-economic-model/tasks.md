# 任务清单

## 前置确认
- [x] 0.1 确认 STAEX 代币参数（display: `staex`, base: `ustaex`, 精度 6）

## 1. 质押/治理载体隔离
- [x] 1.1 在 `app/params/config.go` 中新增 STAEX 代币配置（HumanStakeCoinUnit、BaseStakeCoinUnit、UstaexExponent）
- [x] 1.2 修改 `DefaultBondDenom` 从 `uaex` 改为 `ustaex`
- [x] 1.3 在 `RegisterDenoms()` 中注册 STAEX 代币
- [x] 1.4 更新 `depoly-scripts/localnode/aesc_genesis_template.json` 中的 `bond_denom` 为 `ustaex`
- [x] 1.5 更新 `poc-deploy/localnode/scripts/` 相关脚本：
  - [x] 1.5.1 `step1_configure_init.sh` - gentx 质押使用 ustaex
  - [x] 1.5.2 `step2_genesis.sh` - bond_denom、crisis constant_fee、mint denom、gov deposit 等
  - [x] 1.5.3 `add_validator_to_genesis.sh` - 解析质押金额的 denom
  - [x] 1.5.4 `verify_genesis.sh` - 验证脚本
- [x] 1.6 更新 `poc-deploy/localnode/multi-node-scripts/` 相关脚本
- [x] 1.7 更新 `poc-deploy/localnode/stress-test-scripts/` 相关脚本
- [x] 1.8 更新 `poc-deploy/localnode/exmaple_genesis.json`
- [x] 1.9 更新 `poc-deploy/tools/` 相关工具（batch-submit.go、test-command.sh - 无需修改，仅用于 Gas 费用）
- [x] 1.10 更新 `depoly-scripts/localnode/` 相关脚本：
  - [x] 1.10.1 `scripts/step1_configure_init.sh`
  - [x] 1.10.2 `scripts/step2_genesis.sh`
  - [x] 1.10.3 `scripts/add_validator_to_genesis.sh`
  - [x] 1.10.4 `scripts/verify_genesis.sh`
  - [x] 1.10.5 `exmaple_genesis.json`
  - [x] 1.10.6 `GENESIS_CONFIG.md`
- [x] 1.11 更新配置文件中的 denom 引用：
  - [x] 1.11.1 `poc-deploy/localnode/config/app.toml` - 无需修改，仅包含 Gas 价格配置
  - [x] 1.11.2 `depoly-scripts/localnode/config/app.toml` - 无需修改，仅包含 Gas 价格配置
- [x] 1.12 更新文档文件（README.md、deploy_a.md、deploy_b.md）
- [ ] 1.13 更新相关测试文件中的 bond_denom 引用（待验证）

## 2. Gas 使用率真实采集
- [x] 2.1 在 `x/aexburn/types/` 中定义 `EpochGasData` 结构（累计 gas_used、gas_limit）
- [x] 2.2 在 `x/aexburn/keeper/keeper.go` 中添加 `EpochGasData` 的存取方法
- [x] 2.3 创建 BeginBlock/EndBlock hook 累计每个区块的 gas 数据
- [x] 2.4 修改 `hooks.go` 中的 `calculateGasUsageRate` 使用真实累计数据
- [x] 2.5 在 epoch 结束时重置累计数据
- [x] 2.6 添加 Gas 使用率采集的单元测试

## 3. 月度销毁数据记录
- [x] 3.1 修改 `BurnFees()` 在销毁后更新 `MonthlyBurnData.BurnedAmount`
- [x] 3.2 实现当前月份索引的计算逻辑
- [ ] 3.3 添加月度销毁数据记录的单元测试（复用现有测试）

## 4. 高使用率销毁方向修正
- [x] 4.1 修改 `CalculateDynamicBurnRate` 逻辑：高使用率 → 降低销毁比例
- [x] 4.2 更新相关注释和文档
- [x] 4.3 更新现有销毁比例测试用例

## 5. 参数硬边界约束
- [x] 5.1 在 `params.go` 中添加硬边界常量（MaxAllowedInflationRate = 3% 等）
- [x] 5.2 在 `Params.Validate()` 中添加硬约束检查
- [ ] 5.3 添加参数硬边界的单元测试（可选）

## 6. 规格更新
- [ ] 6.1 更新 `openspec/specs/aex-gas-token/spec.md` 反映代码修正
- [ ] 6.2 添加 STAEX 质押代币相关规格

## 7. 集成验证
- [x] 7.1 运行所有 aexburn 模块测试 ✅ 全部通过
- [x] 7.2 主程序编译成功 ✅
- [ ] 7.3 启动 localnode 验证配置生效（待执行）

