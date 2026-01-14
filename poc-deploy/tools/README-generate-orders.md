# 订单生成工具

## 概述

提供两种订单生成方式：
- **Go 版本**（推荐）：速度快，性能高
- **Shell 版本**：兼容旧版，功能相同

## 性能对比

### Go 版本 (generate-orders.go)
- ✅ **速度快**：约 10,000+ 订单/秒
- ✅ **内存高效**：批量生成，一次性写入
- ✅ **真正并发**：订单级别并发生成
- ✅ **无外部依赖**：不需要 bc, shuf, date 等命令

### Shell 版本 (generate-test-orders.sh)
- ⚠️ **速度慢**：约 100-200 订单/秒
- ⚠️ **外部命令多**：每个订单调用 6 次外部命令
- ⚠️ **文件 I/O 频繁**：每个订单多次文件追加
- ⚠️ **并发受限**：只有文件级别并发

### 实际测试数据

生成 20,000 个订单（2 账户 × 50 文件 × 200 订单）：

| 版本 | 耗时 | 速度 |
|------|------|------|
| Go 版本 | ~2 秒 | ~10,000 订单/秒 |
| Shell 版本 | ~120 秒 | ~167 订单/秒 |

**性能提升：约 60 倍**

## 使用方法

### 方法 1：使用 Go 直接运行（推荐）

```bash
cd poc-deploy/tools

# 基本用法（位置参数）
go run generate-orders.go <账户数> <每账户文件数> <每文件订单数>

# 示例
go run generate-orders.go 2 50 200

# 使用命名参数
go run generate-orders.go --accounts 2 --files 50 --orders 200 --pair ATOM/USDC
```

### 方法 2：使用 Makefile

```bash
cd poc-deploy/tools

# 使用默认参数（2 账户，50 文件，200 订单）
make generate

# 自定义参数
make generate NUM_ACCOUNTS=3 FILES_PER_ACCOUNT=100 ORDERS_PER_FILE=500

# 使用 Shell 版本
make generate-shell NUM_ACCOUNTS=2 FILES_PER_ACCOUNT=50 ORDERS_PER_FILE=200
```

### 方法 3：编译后使用

```bash
cd poc-deploy/tools

# 编译
go build generate-orders.go

# 运行
./generate-orders 2 50 200
```

## 参数说明

### 位置参数
```bash
go run generate-orders.go <账户数> <每账户文件数> <每文件订单数>
```

### 命名参数
```bash
--accounts <数量>      # 账户数量（默认：2）
--files <数量>         # 每账户文件数（默认：50）
--orders <数量>        # 每文件订单数（默认：200）
--pair <交易对>        # 交易对（默认：ATOM/USDC）
```

## 输出示例

```
==========================================
生成测试订单文件
==========================================
账户数量: 2
每账户文件数: 50
每文件订单数: 200
交易对: ATOM/USDC
总订单数: 20000

获取账户信息...
账户 1: admin1 (aesc15ehe3ffhp6wun6lhlxrh3xkvpgta6v4h302rq7)
账户 2: admin2 (sei1mzv7acyaj37zags3aupu3fuzw50avrjhq0y66q)

[账户 1] 开始生成 50 个文件...
[账户 2] 开始生成 50 个文件...
[账户 1] ✓ 完成
[账户 2] ✓ 完成

==========================================
✓ 生成完成！
==========================================

order1/: 50 个文件
order2/: 50 个文件

总账户数: 2
总文件数: 100
总订单数: 20000
耗时: 1.85 秒
速度: 10811 订单/秒

查看示例文件:
  cat order1/orders-0001.json | jq .

使用 batch-submit 提交:
  go run batch-submit.go --count 2
```

## 生成的文件结构

```
poc-deploy/tools/
├── order1/
│   ├── orders-0001.json
│   ├── orders-0002.json
│   └── ...
├── order2/
│   ├── orders-0001.json
│   ├── orders-0002.json
│   └── ...
└── ...
```

## JSON 文件格式

```json
{
  "pair": "ATOM/USDC",
  "orders": [
    {
      "order_id": "order1-1-0-1699876543210",
      "owner": "aesc15ehe3ffhp6wun6lhlxrh3xkvpgta6v4h302rq7",
      "side": "buy",
      "price": "45.23",
      "quantity": "123.45",
      "order_type": "limit"
    },
    ...
  ]
}
```

## 完整工作流

```bash
# 1. 部署节点（创建 admin 账户）
cd poc-deploy/localnode
./scripts/clean.sh
./scripts/deploy.sh

# 2. 生成订单（Go 版本）
cd ../tools
go run generate-orders.go 2 50 200

# 3. 提交订单
go run batch-submit.go --count 2

# 4. 查看结果
cat order1/orders-0001.json | jq .
```

## 常见问题

### Q: 为什么 Go 版本这么快？

A: 主要原因：
1. **无外部命令调用**：随机数、时间戳都在内存中生成
2. **批量 I/O**：在内存中生成完整 JSON，一次性写入
3. **真正并发**：多个文件同时生成，每个文件内部也是批量处理
4. **编译型语言**：Go 是编译型语言，执行效率高

### Q: Shell 版本会被移除吗？

A: 不会。Shell 版本会保留用于兼容性，但推荐使用 Go 版本。

### Q: 如何验证生成的订单是否正确？

A: 使用 jq 查看：
```bash
cat order1/orders-0001.json | jq .
cat order1/orders-0001.json | jq '.orders | length'  # 查看订单数量
```

### Q: 生成失败怎么办？

A: 检查：
1. 节点是否运行：`seid status`
2. admin 账户是否存在：`echo 12345678 | seid keys list`
3. Go 是否安装：`go version`

## 性能优化建议

### 大量订单生成

如果需要生成大量订单（如 100,000+），建议：

```bash
# 增加账户数，减少每账户文件数
go run generate-orders.go 10 100 1000  # 10 账户，每账户 100 文件，每文件 1000 订单

# 而不是
go run generate-orders.go 2 500 1000   # 2 账户，每账户 500 文件，每文件 1000 订单
```

原因：更多账户 = 更高的并发度

### 磁盘 I/O 优化

如果磁盘较慢，可以：
1. 使用 SSD
2. 减少文件数，增加每文件订单数
3. 使用内存盘（tmpfs）

## 技术细节

### Go 版本实现要点

1. **并发控制**：使用 semaphore 限制并发数（20）
2. **随机数生成**：每个 goroutine 独立的 rand.Rand 实例
3. **JSON 序列化**：使用 encoding/json 标准库
4. **错误处理**：使用 channel 收集错误

### Shell 版本性能瓶颈

1. **外部命令**：120,000 次（bc, shuf, date, cat）
2. **文件 I/O**：40,100 次（每次打开-写入-关闭）
3. **进程创建**：每次外部命令都需要 fork + exec

## 相关文件

- `generate-orders.go` - Go 版本生成脚本（推荐）
- `generate-test-orders.sh` - Shell 版本生成脚本（兼容）
- `batch-submit.go` - 批量提交脚本
- `Makefile` - Make 命令集合
- `quick-test.sh` - 快速测试脚本

## 更新日志

### v2.0 (Go 版本)
- ✅ 性能提升 60 倍
- ✅ 支持命名参数
- ✅ 更好的错误处理
- ✅ 实时进度显示

### v1.0 (Shell 版本)
- ✅ 基本功能
- ✅ 并行文件生成
- ⚠️ 性能较慢

