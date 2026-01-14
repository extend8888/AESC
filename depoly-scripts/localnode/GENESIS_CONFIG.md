# AESC Chain Genesis 配置说明

## 文件说明

- `aesc_genesis_template.json` - AESC 链 Genesis 模板文件
- `exmaple_genesis.json` - 原始 Sei 示例文件（保留供参考）

## 关键配置项

### 1. Chain ID

```json
"chain_id": "aesc-mainnet"
```

对于测试网可改为 `aesc-testnet`。

### 2. AEX 代币 (Gas 代币)

**Denom Metadata:**

```json
{
  "description": "The native gas token of AESC Chain",
  "denom_units": [
    {"denom": "uaex", "exponent": 0, "aliases": ["microaex"]},
    {"denom": "maex", "exponent": 3, "aliases": ["milliaex"]},
    {"denom": "aex", "exponent": 6, "aliases": ["AEX"]}
  ],
  "base": "uaex",
  "display": "aex",
  "name": "AEX",
  "symbol": "AEX"
}
```

**精度:** 1 AEX = 1,000,000 uaex (10^6)

### 3. 初始发行量

根据需求文档，AEX 总量为 **5 亿 (500,000,000 AEX)**。

在 `bank.balances` 中配置金库地址的初始余额：

```json
{
  "address": "aesc1...<金库地址>",
  "coins": [
    {
      "denom": "uaex",
      "amount": "500000000000000"  // 5亿 AEX = 500,000,000 * 10^6 uaex
    }
  ]
}
```

### 4. 需要修改的关键位置

| 位置 | 字段 | 说明 |
|------|------|------|
| `bank.balances` | 各账户余额 | 配置初始代币分配 |
| `crisis.constant_fee.denom` | uaex | 危机模块费用 |
| `gov.deposit_params` | uaex | 治理提案押金 |
| `mint.params.mint_denom` | uaex | 铸币 denom |
| `staking.params.bond_denom` | uaex | 质押 denom |

### 5. Mint 模块配置（通胀）

AEX 链**禁用 Sei 的 TokenReleaseSchedule 时间表释放**，通胀完全由 `aexburn` 模块控制：

```json
"mint": {
  "params": {
    "mint_denom": "uaex",
    "token_release_schedule": []  // 空数组，禁用时间表释放
  },
  "minter": {
    "total_mint_amount": "0",
    "remaining_mint_amount": "0"
    // ... 初始值，不执行任何释放
  }
}
```

**重要**: `token_release_schedule` 必须为空数组，否则 Sei 的 mint 模块会按时间表释放代币。

### 6. AEX Burn 模块配置（销毁与通胀控制）

```json
"aexburn": {
  "params": {
    "burn_enabled": true,
    "min_burn_rate": "0.300000000000000000",    // 最低销毁比例 30%
    "max_burn_rate": "0.600000000000000000",    // 最高销毁比例 60%
    "target_burn_rate": "0.500000000000000000", // 目标销毁比例 50%
    "low_gas_threshold": "0.300000000000000000",  // 低 Gas 使用率阈值
    "high_gas_threshold": "0.700000000000000000"  // 高 Gas 使用率阈值
  },
  "burn_stats": {
    "total_burned": "0",
    "last_burn_height": "0",
    "last_burn_time": "0"
  },
  "monthly_burn_data": []
}
```

**销毁比例动态调节规则**:
- Gas 使用率 < 30%: 销毁比例趋向 30%
- Gas 使用率 30%-70%: 销毁比例保持 50%
- Gas 使用率 > 70%: 销毁比例趋向 60%

### 7. 验证者配置

在 `validators` 数组中配置初始验证者：

```json
{
  "address": "<验证者 Tendermint 地址>",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "<Base64 公钥>"
  },
  "power": "100",
  "name": "validator-1"
}
```

## 使用方法

1. 复制模板文件：
   ```bash
   cp aesc_genesis_template.json genesis.json
   ```

2. 修改金库地址和初始余额

3. 配置验证者信息

4. 将 genesis.json 放置到节点的 `config/` 目录

## 地址前缀

AESC 链使用 `aesc` 作为 Bech32 地址前缀：

- 账户地址: `aesc1...`
- 验证者操作地址: `aescvaloper1...`
- 验证者共识地址: `aescvalcons1...`

