# 批量订单测试工具

用于生成和提交大量测试订单的工具集。

## 工具说明

### 1. generate-test-orders.sh

生成测试订单 JSON 文件。

**功能**:
- 生成 N 个 JSON 文件
- 每个文件包含 M 个订单
- 自动生成随机价格、数量、订单方向
- 支持自定义交易对和所有者地址

**用法**:

```bash
./generate-test-orders.sh [文件数量] [每文件订单数] [输出目录] [交易对] [所有者地址]
```

**参数**:
- `文件数量`: 生成的 JSON 文件数量（默认: 10）
- `每文件订单数`: 每个文件包含的订单数量（默认: 100）
- `输出目录`: 输出目录路径（默认: test-orders）
- `交易对`: 交易对标识符（默认: ATOM/USDC）
- `所有者地址`: 订单所有者地址（默认: 自动从 seid keys 获取 validator 地址）

**示例**:

```bash
# 使用默认参数（10 个文件，每个 100 订单）
./generate-test-orders.sh

# 生成 50 个文件，每个 200 订单
./generate-test-orders.sh 50 200

# 自定义所有参数
./generate-test-orders.sh 100 500 my-orders BTC/USDT sei1abc...xyz
```

**输出格式**:

```json
{
  "pair": "ATOM/USDC",
  "orders": [
    {
      "order_id": "order-1-0-1234567890123",
      "owner": "aesc1shavtcw5w6rem6mtje5z3meuy889hj0yuplp4z",
      "side": "buy",
      "price": "45.23",
      "quantity": "123.45",
      "order_type": "limit"
    },
    ...
  ]
}
```

### 2. batch-submit.go

并发提交订单文件到链上。

**功能**:
- 并发提交多个订单文件
- 支持自定义并发数
- 实时显示提交进度
- 统计成功/失败数量

**用法**:

```bash
go run batch-submit.go [选项]
```

**选项**:

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `-dir` | `test-orders` | 订单文件目录 |
| `-concurrency` | `5` | 并发数 |
| `-from` | `validator` | 发送账户名称 |
| `-chain-id` | `aesc-poc` | 链 ID |
| `-node` | `tcp://localhost:26657` | 节点地址 |
| `-fees` | `1000uaex` | 交易费用 |
| `-gas` | `auto` | Gas 限制 |
| `-gas-adjustment` | `1.5` | Gas 调整系数 |
| `-broadcast-mode` | `sync` | 广播模式 (sync, async, block) |
| `-dry-run` | `false` | 只打印命令不执行 |

**示例**:

```bash
# 使用默认参数
go run batch-submit.go

# 自定义并发数和目录
go run batch-submit.go -dir my-orders -concurrency 10

# Dry run 模式（只打印命令）
go run batch-submit.go -dry-run

# 完整参数
go run batch-submit.go \
  -dir test-orders \
  -concurrency 10 \
  -from validator \
  -chain-id aesc-poc \
  -fees 2000uaex \
  -gas auto \
  -broadcast-mode sync
```

## 完整工作流

### 步骤 1: 生成测试订单

```bash
cd poc-deploy/tools

# 赋予执行权限
chmod +x generate-test-orders.sh

# 生成 100 个文件，每个 1000 订单
./generate-test-orders.sh 100 1000
```

### 步骤 2: 提交订单

```bash
# 使用 10 个并发提交
go run batch-submit.go -dir test-orders -concurrency 10
```

### 步骤 3: 查看结果

```bash
# 查询订单
seid query execution orders ATOM/USDC --limit 10

# 查看区块高度
seid status | jq .SyncInfo.latest_block_height
```

## 性能测试场景

### 场景 1: 小批量高频

```bash
# 生成 1000 个文件，每个 10 订单
./generate-test-orders.sh 1000 10

# 使用 20 并发提交
go run batch-submit.go -concurrency 20
```

### 场景 2: 大批量低频

```bash
# 生成 10 个文件，每个 10000 订单
./generate-test-orders.sh 10 10000

# 使用 5 并发提交
go run batch-submit.go -concurrency 5
```

### 场景 3: 压力测试

```bash
# 生成 500 个文件，每个 5000 订单（总计 250 万订单）
./generate-test-orders.sh 500 5000 stress-test

# 使用 50 并发提交
go run batch-submit.go -dir stress-test -concurrency 50
```

## 监控和分析

### 使用 Grafana 监控

启动监控服务：

```bash
cd poc-deploy/metrics
./start.sh
```

访问 Grafana: http://localhost:3000

关键指标：
- 区块高度增长速率
- 内存池大小
- 交易处理速度

### 查询统计信息

```bash
# 查询特定交易对的订单数量
seid query execution orders ATOM/USDC --count-total

# 查询特定用户的订单
seid query execution orders-by-owner <address> --limit 100
```

## 故障排查

### 问题 1: 生成脚本失败

**错误**: `无法获取账户地址`

**解决方案**:
```bash
# 检查账户是否存在
seid keys list

# 手动指定地址
./generate-test-orders.sh 10 100 test-orders ATOM/USDC $(seid keys show validator -a)
```

### 问题 2: 提交失败

**错误**: `insufficient fees`

**解决方案**:
```bash
# 增加费用
go run batch-submit.go -fees 5000uaex
```

**错误**: `account sequence mismatch`

**解决方案**:
```bash
# 降低并发数
go run batch-submit.go -concurrency 1

# 或使用 async 模式
go run batch-submit.go -broadcast-mode async
```

### 问题 3: Gas 估算失败

**错误**: `out of gas`

**解决方案**:
```bash
# 增加 gas adjustment
go run batch-submit.go -gas-adjustment 2.0

# 或手动指定 gas
go run batch-submit.go -gas 500000
```

## 高级用法

### 编译 Go 程序

```bash
# 编译为可执行文件
go build -o batch-submit batch-submit.go

# 使用编译后的程序
./batch-submit -dir test-orders -concurrency 10
```

### 批量清理测试数据

```bash
# 删除生成的订单文件
rm -rf test-orders/

# 重置链（如果需要）
cd poc-deploy/localnode
./scripts/clean.sh
./scripts/deploy.sh
```

### 自定义订单生成

如果需要更复杂的订单生成逻辑，可以修改 `generate-test-orders.sh` 中的函数：

```bash
# 自定义价格生成
generate_price() {
    # 你的逻辑
    echo "100.00"
}

# 自定义数量生成
generate_quantity() {
    # 你的逻辑
    echo "50.00"
}
```

## 性能优化建议

### 1. 调整并发数

- **低配置机器**: 并发数 5-10
- **中等配置**: 并发数 10-20
- **高配置**: 并发数 20-50

### 2. 选择广播模式

- **sync**: 等待交易进入内存池（推荐）
- **async**: 立即返回（最快，但可能丢失错误）
- **block**: 等待交易打包（最慢，但最可靠）

### 3. 批量大小

- **小批量** (10-100 订单/文件): 适合高频测试
- **中批量** (100-1000 订单/文件): 平衡性能和可靠性
- **大批量** (1000+ 订单/文件): 最大化吞吐量

## 参考资料

- [Sei Chain 文档](https://docs.sei.io/)
- [Cosmos SDK 交易](https://docs.cosmos.network/main/core/transactions)
- [Go 并发编程](https://go.dev/doc/effective_go#concurrency)

