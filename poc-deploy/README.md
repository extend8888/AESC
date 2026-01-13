# POC 部署方案

这个目录包含了 Sei Chain 的 POC（概念验证）部署方案。

## 目录结构

```
poc-deploy/
├── localnode/          # 单节点本地部署方案
│   ├── config/         # 配置文件
│   ├── scripts/        # 部署脚本
│   └── README.md       # 详细文档
├── metrics/            # Prometheus + Grafana 监控方案
│   ├── docker-compose.yml
│   ├── prometheus/     # Prometheus 配置
│   ├── grafana/        # Grafana 配置和 Dashboards
│   └── README.md       # 详细文档
└── tools/              # 批量订单测试工具
    ├── generate-test-orders.sh  # 生成测试订单
    ├── batch-submit.go          # 并发提交脚本
    ├── quick-test.sh            # 快速测试
    ├── Makefile                 # Make 命令
    └── README.md                # 详细文档
```

## 快速开始

### 单节点本地部署

```bash
# 进入 localnode 目录
cd poc-deploy/localnode

# 查看详细文档
cat README.md

# 运行部署
chmod +x scripts/*.sh
./scripts/deploy.sh
```

## 方案说明

### localnode - 单节点本地部署

- **用途**: 本地开发和测试
- **特点**:
  - 单个验证人节点
  - 无 Docker 依赖
  - 无 Price Feeder
  - 与 `docker/localnode` 配置一致
- **文档**: [localnode/README.md](localnode/README.md)

### metrics - Prometheus + Grafana 监控

- **用途**: 监控 Sei Chain 节点性能和状态
- **特点**:
  - Prometheus 时序数据库
  - Grafana 可视化面板
  - 预配置的 Sei Chain Dashboard
  - 自动数据源配置
  - Docker Compose 一键部署
- **文档**: [metrics/README.md](metrics/README.md)
- **快速开始**: [metrics/QUICKSTART.md](metrics/QUICKSTART.md)

### tools - 批量订单测试工具

- **用途**: 生成和提交大量测试订单
- **特点**:
  - 自动生成测试订单 JSON 文件
  - Go 并发提交脚本
  - 支持自定义并发数和批量大小
  - 实时统计和进度显示
  - 性能测试场景支持
- **文档**: [tools/README.md](tools/README.md)

## 完整工作流

### 1. 部署节点

```bash
cd poc-deploy/localnode
chmod +x scripts/*.sh
./scripts/deploy.sh
```

### 2. 启动监控

```bash
cd poc-deploy/metrics
chmod +x *.sh
./start.sh
```

### 3. 批量测试

```bash
cd poc-deploy/tools
chmod +x *.sh

# 快速测试（5 个文件，每个 50 订单）
./quick-test.sh

# 或使用 Make
make quick-test

# 或自定义参数
make generate NUM_FILES=100 ORDERS_PER_FILE=1000
make submit CONCURRENCY=10
```

### 4. 访问服务

- **Sei Chain RPC**: http://localhost:26657
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

## 未来扩展

可以在此目录下添加更多部署方案：

- `testnet/` - 测试网部署
- `devnet/` - 开发网部署
- `multinode/` - 多节点本地部署
- 等等...

## 相关文档

- [docker/localnode](../docker/localnode) - 原始的 Docker 多节点部署方案
- [localnode/README.md](localnode/README.md) - 单节点部署详细文档

