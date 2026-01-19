# 变更提案：AEX Gas 代币（链层完整实现）

## 状态：核心功能已完成

> **说明**：本变更提案涵盖 AEX 链层系统的完整功能实现。
> AEX 与 AESC 是两个独立系统，先完成 AEX 链层，再实现 AESC 合约层。

---

## 1. 概述

将 AESC 链的 Gas 代币从 `uaex` 改为 `uaex`，并实现完整的 AEX 经济模型，包括：
- 基础配置 ✅
- 通胀机制 ✅
- 手续费销毁与动态调节 ✅
- 净供给硬约束 ✅
- 动态 Gas 价格 ✅（使用现有 EIP-1559 实现）

## 2. 动机

AESC 链 fork 自 Sei Chain，需要：
1. 将默认的 `uaex` 代币替换为 `uaex` (AEX) 作为 Gas 代币
2. 实现 AEX 完整经济模型，确保代币供给可控、激励合理

## 3. 代币参数

| 参数 | 值 |
|------|-----|
| 名称 | AEX |
| 符号 | AEX |
| 最小单位 | uaex |
| 精度 | 6 (1 AEX = 10^6 uaex) |
| 初始发行量 | 500,000,000 AEX |
| 用途 | 交易手续费（Gas） |
| 地址前缀 | aesc |

---

## 4. 变更内容

### 4.1 AEX-P0：基础配置 ✅ 已完成

| 工作项 | 说明 | 状态 |
|--------|------|------|
| 修改 `app/params/config.go` | BaseCoinUnit → "uaex", Bech32Prefix → "aesc" | ✅ |
| 修改 `x/evm/keeper/params.go` | BaseDenom → "uaex" | ✅ |
| 修改 `cmd/seid/cmd/root.go` | MinGasPrices → "0.02uaex" | ✅ |
| 全局搜索替换 | 30+ 文件的 "uaex"/"sei" 替换 | ✅ |
| Genesis 配置模板 | `depoly-scripts/localnode/aesc_genesis_template.json` | ✅ |
| poc-deploy 适配 | 脚本、配置、文档更新 | ✅ |

### 4.2 AEX-P1：通胀与供给控制 ✅ 已完成

| 工作项 | 说明 | 状态 |
|--------|------|------|
| 通胀机制 | 年通胀上限 3%，基于 Gas 使用率触发 | ✅ |
| 通胀触发条件 | Gas 使用率超过阈值（默认 50%）时触发 | ✅ |
| 净供给硬约束 | 任意 12 个月净增发 ≤ 初始量 5% | ✅ |

**技术实现：**
- 在 `x/aexburn` 模块中实现通胀机制（销毁和通胀统一管理，便于计算净供给）
- 使用 `x/epoch` 的 `AfterEpochEnd` hook 触发
- 12 个月滚动窗口数据存储（`MonthlyBurnData`）
- 禁用 Sei 原有的 `x/mint` 时间表释放

### 4.3 AEX-P2：手续费销毁与动态调节 ✅ 已完成

| 工作项 | 说明 | 状态 |
|--------|------|------|
| 手续费销毁 | 通过 `FeeBurnHook` 接口在 distribution 模块中销毁 | ✅ |
| 动态销毁比例 | 30% - 60%，基于 Gas 使用率自动调节 | ✅ |
| 验证者分配 | 70% - 40%，剩余部分 | ✅ |
| 净供给反向刹车 | 连续 3 周期净供给为负时，自动下调销毁 10% | ✅ |

**技术实现：**
- 在 `x/distribution/keeper` 中添加 `FeeBurnHook` 接口
- `x/aexburn` 模块实现 hook，在 `AllocateTokens` 时销毁部分 fee
- 反向刹车机制在 epoch 结束时检查并更新状态

### 4.4 AEX-P3：动态 Gas 价格 ✅ 已评估

| 工作项 | 说明 | 状态 |
|--------|------|------|
| 评估现有实现 | Sei 链已有完整 EIP-1559 实现 | ✅ |
| 参数评估 | 当前参数满足目标 $0.01 - $0.05 | ✅ 无需调整 |

**现有实现：**
- `x/evm/keeper/fee.go` - `AdjustDynamicBaseFeePerGas` 动态调整 base fee
- MinimumFeePerGas: 1 gwei，MaximumFeePerGas: 1000 gwei
- 上调 1.89%/块，下调 0.39%/块

### 4.5 AEX-P4：辅助功能 ✅ 已完成

| 工作项 | 说明 | 状态 |
|--------|------|------|
| Gas Credit | ~~兜底机制~~ → 使用 Paymaster 代付替代 | ✅ 不需要 |
| 验证者收入平滑 | 缓冲池机制，默认关闭 | ✅ 已实现 |

**验证者收入平滑机制实现：**
- 实现位置：`x/aexburn/keeper/income_smoother.go`
- 集成方式：在 `BurnFees` 方法中调用，零侵入 distribution 模块
- 高活跃期（Gas 使用率 > 70%）：抽取 10% 收入到缓冲池
- 低活跃期（Gas 使用率 < 30%）：从缓冲池释放 5% 补贴
- 缓冲池上限：初始供给的 1%
- 默认关闭，可通过治理启用

---

## 5. 影响范围

### 5.1 已修改的模块

| 模块 | 修改内容 | 状态 |
|------|----------|------|
| `app/params` | 基础配置 (uaex, aesc 前缀) | ✅ |
| `x/aexburn` | 新模块：通胀、销毁、净供给控制 | ✅ |
| `x/distribution` | 添加 FeeBurnHook 接口 | ✅ |
| `x/epoch` | 周期触发 hook (AfterEpochEnd) | ✅ |
| `app/app.go` | 注册 aexburn 模块账户权限 | ✅ |

### 5.2 不影响

- AESC 合约层（独立系统）
- Tendermint 共识机制
- EVM 兼容性
- 跨链桥接

---

## 6. 风险

| 风险 | 缓解措施 |
|------|----------|
| 通胀/销毁逻辑错误 | 充分的单元测试和集成测试 |
| 净供给计算偏差 | 12 个月滚动窗口数据验证 |
| 动态调节过于激进 | 参数可配置，渐进调整 |
| 验证者收入不稳定 | 反向刹车机制保护 |

---

## 7. 测试计划

| 测试项 | 说明 |
|--------|------|
| 单元测试 | 通胀、销毁、动态调节逻辑 |
| 集成测试 | 多模块联动验证 |
| 本地测试网 | 完整功能验收 |
| 压力测试 | 高负载下的动态调节 |

---

## 8. 参考文档

- `tmp/AESC 链开发技术方案.md`
- `tmp/AESC 公链 Gas (AEX)经济模型.md`

