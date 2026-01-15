# AEX-P1: 通胀与供给控制 - 实现方案分析

> **状态：✅ 已完成**
>
> 本文档分析 AEX 通胀机制的实现方式，区分需要代码开发和仅需 Genesis 配置的功能。

---

## 1. 现有 Sei mint 模块分析

### Sei 的通胀机制

Sei 链的 `x/mint` 模块使用 **TokenReleaseSchedule 时间表释放** 模式：

```
TokenReleaseSchedule[]:
  - StartDate: "2024-01-01"
    EndDate: "2024-12-31"
    TokenReleaseAmount: 10000000  (在此期间每天释放一部分)
  - StartDate: "2025-01-01"
    EndDate: "2025-12-31"  
    TokenReleaseAmount: 5000000
```

- **触发方式**: `AfterEpochEnd` hook（每个 epoch 结束时）
- **释放逻辑**: 按日期范围线性释放预定义总量
- **特点**: 释放量固定，不依赖链上状态

### AEX 需求的通胀机制

根据经济模型文档，AEX 需要 **基于链上指标的动态通胀**：

- 年通胀上限：3%（硬约束）
- 触发条件：交易量、区块稳定性、Gas 使用率
- 净供给约束：任意 12 个月净增发 ≤ 初始量 5%

---

## 2. 功能分类

### ✅ 仅需 Genesis 配置（无需代码开发）

| 功能 | 配置位置 | 说明 |
|------|----------|------|
| **禁用 Sei 时间表释放** | `mint.params.token_release_schedule: []` | 清空释放时间表 |
| **设置 mint denom** | `mint.params.mint_denom: "uaex"` | 确保 mint 使用 uaex |
| **初始供给量** | `bank.supply` + 账户余额 | 500M AEX 初始分配 |

### ✅ 已完成代码开发

| 功能 | 实现 | 状态 |
|------|------|--------|
| **动态通胀触发** | `MintInflation()` - 基于 Gas 使用率触发 | ✅ |
| **年通胀上限 3%** | `calculateInflationAmount()` - 年度预算跟踪 | ✅ |
| **净供给 ≤5% 约束** | `Get12MonthNetSupply()` + `MonthlyBurnData` 滚动窗口 | ✅ |
| **反向刹车机制** | `UpdateReverseBrakeState()` - 连续负净供给时降低销毁 | ✅ |

---

## 3. 实现方案决策

### 方案 A：扩展 x/aexburn 模块（推荐）

将通胀逻辑集成到已有的 `x/aexburn` 模块，重命名为 `x/aexsupply`：

**优点**：
- 销毁和通胀在同一模块，便于计算净供给
- 已有 epoch 集成基础
- 减少模块数量

**缺点**：
- 模块职责扩大

### 方案 B：创建独立 x/aexinflation 模块

**优点**：
- 职责分离清晰

**缺点**：
- 需要跨模块通信获取销毁数据
- 增加复杂度

### 决策：采用方案 A ✅

保留 `x/aexburn` 模块名称，统一管理销毁和通胀。模块已实现：
- 手续费销毁 (`BurnFees`)
- 动态销毁比例 (`CalculateDynamicBurnRate`)
- 通胀铸造 (`MintInflation`)
- 净供给计算 (`Get12MonthNetSupply`)
- 反向刹车 (`UpdateReverseBrakeState`)

---

## 4. Genesis 配置（无需代码）

以下配置在 genesis.json 中设置即可，无需代码修改：

```json
{
  "app_state": {
    "mint": {
      "params": {
        "mint_denom": "uaex",
        "token_release_schedule": []
      },
      "minter": {
        "start_date": "0001-01-01",
        "end_date": "0001-01-01",
        "denom": "uaex",
        "total_mint_amount": "0",
        "remaining_mint_amount": "0",
        "last_mint_amount": "0",
        "last_mint_date": "0001-01-01",
        "last_mint_height": "0"
      }
    }
  }
}
```

**说明**：
- `token_release_schedule: []` - 禁用 Sei 的时间表释放
- `minter` 设为初始值，不执行任何释放
- AEX 通胀完全由 `x/aexsupply` 模块控制

---

## 5. 需要开发的代码变更

### 5.1 扩展 proto 定义

在 `proto/aexburn/` 中添加通胀相关类型：

```protobuf
// 通胀参数
message InflationParams {
  bool inflation_enabled = 1;
  string max_annual_inflation_rate = 2;    // 3%
  string max_net_supply_rate_per_year = 3; // 5%
  uint64 initial_supply = 4;               // 500M * 10^6
}

// 通胀触发条件
message InflationTriggerConfig {
  string min_gas_usage_rate = 1;     // 最低 Gas 使用率阈值
  uint64 min_transaction_count = 2;  // 最低交易数阈值
}

// 年度/月度铸造记录
message MintRecord {
  uint64 epoch = 1;
  string amount = 2;
  int64 timestamp = 3;
}
```

### 5.2 扩展 keeper

```go
// 通胀相关方法
func (k Keeper) CheckInflationTrigger(ctx sdk.Context) bool
func (k Keeper) CalculateInflationAmount(ctx sdk.Context) sdk.Int
func (k Keeper) CheckAnnualInflationLimit(ctx sdk.Context, amount sdk.Int) bool
func (k Keeper) CheckNetSupplyLimit(ctx sdk.Context, amount sdk.Int) bool
func (k Keeper) ExecuteInflation(ctx sdk.Context) error
```

### 5.3 实现 Epoch Hook

```go
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epoch epochTypes.Epoch) {
    // 1. 检查通胀触发条件
    // 2. 计算通胀量
    // 3. 检查年上限和净供给约束
    // 4. 执行铸造
}
```

---

## 6. 实现完成状态

| 步骤 | 状态 | 说明 |
|------|------|------|
| Genesis 配置模板 | ✅ | `aesc_genesis_template.json` 已禁用 Sei mint |
| proto 定义扩展 | ✅ | `proto/aexburn/*.proto` 已完成 |
| 通胀逻辑 | ✅ | `x/aexburn/keeper/inflation.go` |
| 销毁逻辑 | ✅ | `x/aexburn/keeper/burn.go` |
| 反向刹车 | ✅ | `x/aexburn/keeper/burn.go` + `hooks.go` |
| Epoch Hook | ✅ | `x/aexburn/keeper/hooks.go` |
| 单元测试 | ✅ | `x/aexburn/keeper/*_test.go` |

### 关键文件

```
x/aexburn/
├── keeper/
│   ├── keeper.go       # 状态存取方法
│   ├── burn.go         # 销毁逻辑、动态比例、反向刹车
│   ├── inflation.go    # 通胀逻辑、年度上限、净供给约束
│   ├── hooks.go        # Epoch hooks
│   ├── keeper_test.go  # 基础功能测试
│   ├── burn_test.go    # 销毁机制测试
│   └── inflation_test.go # 通胀机制测试
├── types/
│   ├── params.go       # 参数定义与验证
│   └── keys.go         # 存储键
└── genesis.go          # 创世状态导入导出
```

