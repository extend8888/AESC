# AEX 经济模型测试 - 问题追踪

> **文档版本**：v1.0
> **创建日期**：2026-01-19

---

## 问题清单

| ID | 状态 | 问题描述 | 解决方案 | 相关文件 |
|----|------|----------|----------|----------|
| ISS-001 | ✅ 已解决 | `types.pb.go` 缺少 `,string` JSON tag | 手动添加 `,string` | `sei-cosmos/x/params/types/types.pb.go` |
| ISS-002 | ✅ 已解决 | `fee_collector` 缺少 Burner 权限 | 添加权限 + 补充测试 | `app/app.go`, `x/aexburn/keeper/burn_test.go` |

---

## ISS-001: CosmosGasParams JSON 反序列化失败

### 问题描述

链启动后在第一个 epoch 后崩溃，报错 `integer divide by zero`。

**根本原因**：`sei-cosmos/x/params/types/types.pb.go` 中 `CosmosGasParams` 结构体的 JSON tag 缺少 `,string` 标记。

- Amino JSON 编码 `uint64` 为字符串（如 `"1"`）
- 标准 `json.Unmarshal` 无法将字符串解析为 `uint64`（需要 `,string` 标记）
- 导致 `CosmosGasMultiplierDenominator = 0`，后续除法运算 panic

### 问题对比

| 版本 | JSON tag |
|------|----------|
| yellow-chain (正确) | `json:"cosmos_gas_multiplier_numerator,string,omitempty"` |
| AESC (修复前) | `json:"cosmos_gas_multiplier_numerator,omitempty"` |

### 解决方案

手动修改 `sei-cosmos/x/params/types/types.pb.go` 第 82-83 行：

```go
// 修复后
CosmosGasMultiplierNumerator   uint64 `json:"cosmos_gas_multiplier_numerator,string,omitempty"`
CosmosGasMultiplierDenominator uint64 `json:"cosmos_gas_multiplier_denominator,string,omitempty"`
```

### 后续建议

1. 调查 `buf` 或 `protoc` 版本差异，确认为何生成时丢失 `,string`
2. 考虑在 CI 中添加检查，防止此文件被意外覆盖
3. 或在文档中记录此文件需要手动维护

---

## ISS-002: fee_collector 模块账户缺少 Burner 权限

### 问题描述

链运行到第 693 个区块时崩溃，报错：

```
panic: module account fee_collector does not have permissions to burn tokens: unauthorized
```

**根本原因**：`aexburn` 模块的 `BurnFees` 函数尝试从 `fee_collector` 账户销毁代币，但 `fee_collector` 在 `maccPerms` 中没有配置 `Burner` 权限。

### 问题代码位置

`app/app.go` 第 222 行：
```go
// 修复前
authtypes.FeeCollectorName: nil,

// 修复后
authtypes.FeeCollectorName: {authtypes.Burner},
```

`x/aexburn/keeper/burn.go` 第 80 行：
```go
err = k.bankKeeper.BurnCoins(ctx, authtypes.FeeCollectorName, burnCoins)
```

### 解决方案

1. **修改权限配置**：在 `app/app.go` 的 `maccPerms` 中为 `FeeCollectorName` 添加 `Burner` 权限
2. **添加集成测试**：在 `x/aexburn/keeper/burn_test.go` 中添加 `TestBurnFees_Integration` 测试

### 测试覆盖建议

现有的 `burn_test.go` 只测试了计算逻辑（`CalculateDynamicBurnRate`、`UpdateReverseBrakeState`），没有测试 `BurnFees` 的完整执行流程。

**已添加测试**：`TestBurnFees_Integration`
- 验证 `fee_collector` 有资金时能正常销毁
- 验证 `BurnStats` 正确更新
- 验证剩余余额计算正确

---

## 经验教训

1. **模块间依赖权限检查**：当一个模块（aexburn）操作另一个模块账户（fee_collector）时，必须确保目标账户有相应权限

2. **集成测试覆盖**：单元测试应覆盖完整的执行路径，特别是涉及跨模块操作的场景

3. **Protobuf 生成文件维护**：注意 `.pb.go` 文件可能需要手动维护某些标记，不能完全依赖自动生成

---

## 验证步骤

修复后验证：

```bash
# 重新编译
go build -o build/seid ./cmd/seid

# 清理并重启
pkill seid; rm -rf ~/.aesc build/generated
./poc-deploy/localnode/scripts/deploy.sh

# 运行测试
cd x/aexburn && go test -v ./keeper/... -run TestBurnFees_Integration
```

