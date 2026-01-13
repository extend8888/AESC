# Oracle 配置说明

## 问题背景

Sei Chain 的 Oracle 模块要求验证者定期提交价格投票。如果没有 Price Feeder，验证者会因为缺少有效投票而被惩罚（jail），导致单节点链停止出块。

### Oracle 工作机制

```
每 2 个区块 → 验证者需要提交价格投票
              ↓
         没有 Price Feeder
              ↓
         累积 MissCount
              ↓
    2天后（SlashWindow 结束）
              ↓
   有效投票率 < 5% (MinValidPerWindow)
              ↓
      验证者被 Jail（监禁）
              ↓
     单节点链无法继续出块
              ↓
          链停止 🛑
```

### 关键参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `vote_period` | 2 | 每 2 个区块投票一次 |
| `slash_window` | 108,000 | 惩罚窗口（约 24 小时） |
| `min_valid_per_window` | 0.05 | 最小有效投票率（5%） |
| `slash_fraction` | 0 | 惩罚比例（默认为 0） |

## 解决方案

### 方案 A：添加 Price Feeder（完整功能）

**优点**：
- ✅ 完整的 Oracle 功能
- ✅ 支持真实的价格数据
- ✅ 与生产环境一致

**缺点**：
- ❌ 需要编译 price-feeder 程序
- ❌ 需要配置文件和账户设置
- ❌ 增加部署复杂度

**实施步骤**：
1. 复制 `docker/localnode/config/price_feeder_config.toml` → `poc-deploy/localnode/config/`
2. 复制 `docker/localnode/scripts/step6_start_price_feeder.sh` → `poc-deploy/localnode/scripts/step5_start_price_feeder.sh`
3. 修改 `deploy.sh` 添加 step5 调用
4. 编译 price-feeder：`cd oracle/price-feeder && make install`

---

### 方案 B：禁用 Oracle 惩罚（当前方案）⭐

**优点**：
- ✅ 简单快速
- ✅ 不需要额外程序
- ✅ 适合测试环境
- ✅ 避免链停止问题

**缺点**：
- ❌ Oracle 功能不完整（无真实价格数据）

**实施方法**：

在 `step2_genesis.sh` 中添加：

```bash
# Disable Oracle slashing to prevent validator from being jailed without price feeder
override_genesis '.app_state["oracle"]["params"]["min_valid_per_window"]="0"'
```

**原理**：

参考代码 `x/oracle/keeper/slash.go:34`：

```go
// 只有当 validVoteRate < minValidPerWindow 时才惩罚
if validVoteRate.LT(minValidPerWindow) {
    k.StakingKeeper.Slash(...)
    k.StakingKeeper.Jail(ctx, consAddr)
}
```

当 `minValidPerWindow = 0` 时：
- 判断条件变成：`validVoteRate < 0`
- 由于 `validVoteRate` 最小值为 0，条件永远为 `false`
- 惩罚代码永远不会执行

## 验证配置

运行验证脚本：

```bash
./poc-deploy/localnode/scripts/verify_oracle_config.sh
```

预期输出：

```
✅ Oracle slashing is DISABLED (min_valid_per_window = 0)
   → Validator will NOT be jailed for missing Oracle votes
   → Chain can run indefinitely without price feeder

🎯 Configuration: SAFE for testing without price feeder
```

或者手动检查：

```bash
cat ~/.sei/config/genesis.json | jq '.app_state.oracle.params'
```

应该看到：

```json
{
  "vote_period": "2",
  "min_valid_per_window": "0.000000000000000000",  ← 应该是 0
  "slash_fraction": "0.000000000000000000",
  "slash_window": "108000",
  ...
}
```

## 使用建议

| 场景 | 推荐方案 | 理由 |
|------|---------|------|
| **POC 测试** | 方案 B | 简单快速，满足订单测试需求 |
| **压力测试** | 方案 B | 避免 Oracle 干扰测试结果 |
| **功能演示** | 方案 A | 展示完整功能 |
| **生产环境** | 方案 A | 必须有真实价格数据 |

## 相关文件

- `poc-deploy/localnode/scripts/step2_genesis.sh` - Genesis 配置脚本（已修改）
- `poc-deploy/localnode/scripts/verify_oracle_config.sh` - 验证脚本
- `x/oracle/keeper/slash.go` - Oracle 惩罚逻辑
- `x/oracle/types/params.go` - Oracle 参数定义

## 参考资料

- [Oracle 模块代码](../../x/oracle/)
- [Price Feeder 实现](../../oracle/price-feeder/)
- [Docker LocalNode 配置](../../docker/localnode/)

