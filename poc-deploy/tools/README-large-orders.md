# 大订单数据生成工具使用说明

## 功能说明

`generate-orders.go` 已升级支持生成大数据量的订单文件,用于测试区块链的大交易处理能力。

### 新增功能

- 支持指定每个 msg (Order) 的目标大小 (MB)
- 通过填充 `Side` 字段来达到目标大小
- `Quantity` 字段保持为随机生成的正常数量
- 保持原有的小数据生成模式 (默认)

## 使用方法

### 方式 1: 位置参数 (推荐)

```bash
go run generate-orders.go <accounts> <files> <orders> <targetMsgSizeMB>
```

**示例:**

```bash
# 10个账户,每个账户500个文件,每个文件1个订单,每个订单约2MB
go run generate-orders.go 10 500 1 2
```

### 方式 2: 命令行参数

```bash
go run generate-orders.go --accounts 10 --files 500 --orders 1 --size 2
```

## 参数说明

| 参数 | 位置 | 标志 | 默认值 | 说明 |
|------|------|------|--------|------|
| accounts | 1 | --accounts | 2 | 账户数量 |
| files | 2 | --files | 50 | 每账户文件数 |
| orders | 3 | --orders | 200 | 每文件订单数 |
| targetMsgSizeMB | 4 | --size | 0 | 每个 msg 的目标大小 (MB), 0 表示使用默认小数据 |

## 使用场景

### 场景 1: 测试大交易 (2MB/订单)

```bash
# 生成 10 个账户, 每个账户 500 个文件, 每个文件 1 个订单, 每个订单 2MB
go run generate-orders.go 10 500 1 2

# 预期结果:
# - 总文件数: 10 × 500 = 5,000 个文件
# - 总订单数: 5,000 个订单
# - 总数据量: 5,000 × 2MB ≈ 10 GB
```

### 场景 2: 测试超大交易 (10MB/订单)

```bash
# 生成 5 个账户, 每个账户 100 个文件, 每个文件 1 个订单, 每个订单 10MB
go run generate-orders.go 5 100 1 10

# 预期结果:
# - 总文件数: 5 × 100 = 500 个文件
# - 总订单数: 500 个订单
# - 总数据量: 500 × 10MB ≈ 5 GB
```

### 场景 3: 默认小数据模式

```bash
# 不指定 size 参数,使用默认小数据
go run generate-orders.go 2 50 200

# 预期结果:
# - 每个订单约 200-300 字节
# - 适合常规测试
```

## 数据结构

### Order 结构

```json
{
  "order_id": "order1-1-0-1234567890",
  "owner": "sei1...",
  "side": "0123456789...",  // 大数据填充在这里
  "price": "50.25",
  "quantity": "123.45",     // 正常的随机数量
  "order_type": "limit"
}
```

### 文件结构

```json
{
  "pair": "ATOM/USDC",
  "orders": [
    {
      "order_id": "...",
      "owner": "...",
      ...
    }
  ]
}
```

## 实现原理

### 大小计算

```
目标大小 (targetMsgSizeMB) = 2 MB
目标字节数 = 2 × 1024 × 1024 = 2,097,152 bytes

其他字段大小估算:
- OrderID: ~30 bytes
- Owner: ~45 bytes
- Price: ~10 bytes
- Quantity: ~10 bytes
- OrderType: ~10 bytes
- JSON 开销: ~100 bytes
总计: ~200 bytes

Side 字段大小 = 2,097,152 - 200 = 2,096,952 bytes
```

### 填充策略

- 使用数字字符 (0-9) 填充 `Side` 字段
- `Quantity` 字段保持为正常的随机数量 (1.00 - 1000.00)
- 保证 JSON 序列化后仍然是有效字符串
- 随机生成,避免压缩优化

## 注意事项

### 1. 磁盘空间

生成大数据文件需要足够的磁盘空间:

```
示例: 10 账户 × 500 文件 × 1 订单 × 2MB = 10 GB
```

### 2. 内存使用

生成过程会占用一定内存,建议:
- 单次生成不超过 20 GB 数据
- 如需更多数据,分批生成

### 3. 生成时间

大数据生成需要时间:
- 2MB/订单: 约 1-2 秒/文件
- 10MB/订单: 约 5-10 秒/文件

### 4. 区块链限制

当前配置下,单笔交易最大约 21 MB (受 `max_bytes` 限制):
- 建议测试 2MB, 5MB, 10MB, 20MB 等不同大小
- 超过 21MB 的交易会被拒绝

## 测试建议

### 渐进式测试

```bash
# 1. 小规模测试 (1 账户, 10 文件, 2MB/订单)
go run generate-orders.go 1 10 1 2

# 2. 中等规模测试 (5 账户, 100 文件, 2MB/订单)
go run generate-orders.go 5 100 1 2

# 3. 大规模测试 (10 账户, 500 文件, 2MB/订单)
go run generate-orders.go 10 500 1 2
```

### 性能测试

```bash
# 测试不同大小的订单
go run generate-orders.go 1 10 1 1   # 1MB
go run generate-orders.go 1 10 1 2   # 2MB
go run generate-orders.go 1 10 1 5   # 5MB
go run generate-orders.go 1 10 1 10  # 10MB
go run generate-orders.go 1 10 1 20  # 20MB
```

## 故障排查

### 问题 1: 生成的文件大小不准确

**原因**: JSON 序列化会增加额外开销 (缩进、换行等)

**解决**: 实际文件大小会比目标大小略大 10-20%,这是正常的

### 问题 2: 内存不足

**原因**: 一次性生成太多大文件

**解决**: 减少 `files` 参数或分批生成

### 问题 3: 磁盘空间不足

**原因**: 生成的数据量超过可用空间

**解决**: 清理旧文件或增加磁盘空间

## 清理

```bash
# 清理生成的订单文件
rm -rf order*
```

