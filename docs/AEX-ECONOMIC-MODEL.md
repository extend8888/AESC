# AEX 经济模型实现文档

## 概述

AEX 是 AESC Chain 的原生 Gas 代币，用于支付交易手续费。本文档详细说明了 AEX 代币的经济模型实现，包括动态手续费销毁和通胀机制。

## 代币基本信息

| 属性 | 值 |
|------|-----|
| 代币符号 | AEX |
| 基础单位 | uaex |
| 精度 | 6 位（1 AEX = 1,000,000 uaex） |
| 初始供给 | 500,000,000 AEX |
| Bech32 前缀 | aesc |

## 经济模型核心机制

### 1. 动态手续费销毁（AEX-P2）

每个区块的交易手续费会根据网络拥堵情况动态销毁一定比例。

#### 销毁比例计算规则

| Gas 使用率 | 销毁比例 |
|-----------|---------|
| < 30% | 趋向 30%（最小值） |
| 30% - 70% | 保持 50%（目标值） |
| > 70% | 趋向 60%（最大值） |

#### 参数配置

```json
{
  "aexburn": {
    "params": {
      "burn_enabled": true,
      "min_burn_rate": "0.300000000000000000",
      "max_burn_rate": "0.600000000000000000",
      "target_burn_rate": "0.500000000000000000",
      "low_gas_threshold": "0.300000000000000000",
      "high_gas_threshold": "0.700000000000000000"
    }
  }
}
```

#### 执行流程

```
BeginBlock
    ↓
distribution.AllocateTokens()
    ↓
FeeBurnHook.BurnFees()
    ├── 获取 FeeCollector 中的手续费
    ├── 计算当前 Gas 使用率
    ├── 根据使用率计算动态销毁比例
    ├── 从 FeeCollector 销毁 AEX
    ├── 更新销毁统计
    └── 发出 aex_burn 事件
    ↓
剩余手续费分配给验证者
```

### 2. 动态通胀机制（AEX-P1）

通胀基于链上活动指标动态触发，而非固定时间表。

#### 通胀触发条件

- **Gas 使用率阈值**：≥ 50%
- 低于阈值时不触发通胀

#### 通胀约束

| 约束 | 限制 |
|-----|-----|
| 年度通胀上限 | 3%（相对于初始供给） |
| 12 月净供给上限 | 5%（铸造 - 销毁 ≤ 5%） |

#### 参数配置

```json
{
  "aexburn": {
    "params": {
      "inflation_enabled": true,
      "max_annual_inflation_rate": "0.030000000000000000",
      "max_net_supply_rate_per_year": "0.050000000000000000",
      "initial_supply": "500000000000000",
      "min_gas_usage_for_inflation": "0.500000000000000000",
      "epochs_per_year": "365"
    }
  }
}
```

#### 通胀计算公式

```
每 epoch 最大通胀 = (初始供给 × 年度上限) / 每年 epoch 数
                 = (500M × 3%) / 365
                 ≈ 41,096 AEX/epoch

实际通胀 = 每 epoch 最大通胀 × 缩放因子
缩放因子 = (Gas 使用率 - 50%) / (100% - 50%)
```

#### 执行流程

```
Epoch End
    ↓
AfterEpochEnd() hook
    ↓
MintInflation()
    ├── 检查 inflation_enabled
    ├── 检查 Gas 使用率 ≥ 50%
    ├── 计算缩放后的通胀量
    ├── 检查年度 3% 上限
    ├── 检查 12 月净供给 ≤5%
    ├── 铸造 AEX 到 FeeCollector
    ├── 更新统计
    └── 发出 aex_mint 事件
```

## 模块结构

```
x/aexburn/
├── keeper/
│   ├── keeper.go      # Keeper 主体，参数和统计管理
│   ├── burn.go        # 销毁逻辑
│   ├── inflation.go   # 通胀逻辑
│   └── hooks.go       # Epoch hooks 实现
├── types/
│   ├── params.go      # 参数定义和验证
│   ├── keys.go        # 存储键定义
│   ├── genesis.go     # Genesis 状态
│   ├── expected_keepers.go  # 依赖接口
│   ├── codec.go       # 编解码器
│   └── *.pb.go        # Protobuf 生成文件
├── genesis.go         # InitGenesis/ExportGenesis
└── module.go          # AppModule 实现
```

## Genesis 配置示例

完整的 genesis 配置位于 `depoly-scripts/localnode/aesc_genesis_template.json`。

### aexburn 模块配置

```json
{
  "aexburn": {
    "params": {
      "burn_enabled": true,
      "min_burn_rate": "0.300000000000000000",
      "max_burn_rate": "0.600000000000000000",
      "target_burn_rate": "0.500000000000000000",
      "low_gas_threshold": "0.300000000000000000",
      "high_gas_threshold": "0.700000000000000000",
      "inflation_enabled": true,
      "max_annual_inflation_rate": "0.030000000000000000",
      "max_net_supply_rate_per_year": "0.050000000000000000",
      "initial_supply": "500000000000000",
      "min_gas_usage_for_inflation": "0.500000000000000000",
      "epochs_per_year": "365"
    },
    "burn_stats": {
      "total_burned": "0",
      "last_burn_rate": "0",
      "last_epoch_number": "0",
      "last_block_height": "0"
    },
    "inflation_stats": {
      "total_minted": "0",
      "annual_minted": "0",
      "last_annual_reset_epoch": "0",
      "last_mint_epoch": "0",
      "last_mint_block_height": "0"
    },
    "monthly_burn_data": []
  }
}
```

### mint 模块配置（禁用 Sei 原生释放）

```json
{
  "mint": {
    "minter": {
      "last_mint_amount": "0",
      "last_mint_date": "0001-01-01",
      "last_mint_height": "0",
      "denom": "uaex"
    },
    "params": {
      "mint_denom": "uaex",
      "token_release_schedule": []
    }
  }
}
```

## 代码修改说明

### 1. Distribution 模块集成

在 `sei-cosmos/x/distribution/keeper/keeper.go` 中添加了 `FeeBurnHook` 接口：

```go
type FeeBurnHook interface {
    BurnFees(ctx sdk.Context) error
}

func (k *Keeper) SetFeeBurnHook(hook FeeBurnHook) {
    k.feeBurnHook = hook
}
```

在 `sei-cosmos/x/distribution/keeper/allocation.go` 的 `AllocateTokens` 中调用：

```go
func (k Keeper) AllocateTokens(...) {
    // 在分配手续费之前先销毁一部分
    if k.feeBurnHook != nil {
        if err := k.feeBurnHook.BurnFees(ctx); err != nil {
            k.Logger(ctx).Error("failed to burn fees", "error", err)
        }
    }
    // ... 继续原有的分配逻辑
}
```

### 2. Epoch Hooks 注册

在 `app/app.go` 中注册 aexburn 的 epoch hooks：

```go
app.EpochKeeper = *epochmodulekeeper.NewKeeper(
    appCodec,
    keys[epochmoduletypes.StoreKey],
    keys[epochmoduletypes.MemStoreKey],
    app.GetSubspace(epochmoduletypes.ModuleName),
).SetHooks(epochmoduletypes.NewMultiEpochHooks(
    app.MintKeeper.Hooks(),
    app.AexburnKeeper.Hooks(),  // 添加 aexburn hooks
))
```

## 事件

### aex_burn 事件

每次销毁时触发：

| 属性 | 说明 |
|-----|-----|
| epoch | Epoch 编号 |
| denom | 代币单位 (uaex) |
| amount | 销毁数量 |
| burn_rate | 销毁比例 |
| gas_usage | Gas 使用率 |

### aex_mint 事件

每次铸造时触发：

| 属性 | 说明 |
|-----|-----|
| epoch | Epoch 编号 |
| amount | 铸造数量 |
| gas_usage | Gas 使用率 |

## 查询接口

### 查询参数

```bash
aescd query aexburn params
```

### 查询销毁统计

```bash
aescd query aexburn burn-stats
```

### 查询通胀统计

```bash
aescd query aexburn inflation-stats
```

### 查询净供给变化

```bash
aescd query aexburn net-supply
```

## 测试建议

### 单元测试

1. 测试不同 Gas 使用率下的销毁比例计算
2. 测试通胀触发条件（Gas 使用率阈值）
3. 测试年度通胀上限约束
4. 测试 12 月净供给约束
5. 测试 epoch hooks 正确触发

### 集成测试

1. 启动本地测试网
2. 发送交易产生手续费
3. 验证销毁事件和统计
4. 模拟高 Gas 使用率触发通胀
5. 验证通胀事件和统计

## 相关文件

| 文件 | 说明 |
|-----|-----|
| `proto/aexburn/params.proto` | 参数 protobuf 定义 |
| `proto/aexburn/burn.proto` | 销毁/通胀类型定义 |
| `proto/aexburn/genesis.proto` | Genesis 状态定义 |
| `proto/aexburn/query.proto` | 查询接口定义 |
| `x/aexburn/` | 模块完整实现 |
| `app/app.go` | 模块注册和初始化 |
| `sei-cosmos/x/distribution/keeper/` | FeeBurnHook 集成 |
| `depoly-scripts/localnode/aesc_genesis_template.json` | Genesis 配置模板 |
| `depoly-scripts/localnode/GENESIS_CONFIG.md` | Genesis 配置说明 |

## Git 提交历史

```
a9a70c5 feat(aexburn): implement AEX-P2 fee burning mechanism
c2d37cd feat(aexburn): add AEX-P1 inflation mechanism
2d34fb2 chore: update codebase for AESC chain branding
```

