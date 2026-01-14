# Sei Chain 监控方案

使用 Prometheus + Grafana 监控 Sei Chain 节点。

## 目录结构

```
poc-deploy/metrics/
├── docker-compose.yml              # Docker Compose 配置
├── prometheus/
│   └── prometheus.yml              # Prometheus 配置
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml      # Prometheus 数据源配置
│   │   └── dashboards/
│   │       └── default.yml         # Dashboard 自动加载配置
│   └── dashboards/
│       └── sei-chain-overview.json # Sei Chain 概览 Dashboard
└── README.md                       # 本文档
```

## 快速开始

### 1. 启动监控服务

```bash
cd poc-deploy/metrics

# 启动 Prometheus + Grafana
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 2. 访问服务

- **Grafana**: http://localhost:3000
  - 用户名: `admin`
  - 密码: `admin`
  
- **Prometheus**: http://localhost:9090

### 3. 配置 Sei Chain 节点

确保 Sei Chain 节点启用了 Prometheus metrics：

#### 方法 1: 修改 config.toml

编辑 `~/.sei/config/config.toml`:

```toml
#######################################################
###       Instrumentation Configuration Options     ###
#######################################################
[instrumentation]

# When true, Prometheus metrics are served under /metrics on
# PrometheusListenAddr.
# Check out the documentation for the list of available metrics.
prometheus = true

# Address to listen for Prometheus collector(s) connections
prometheus_listen_addr = ":26660"
```

#### 方法 2: 使用环境变量

```bash
export INSTRUMENTATION_PROMETHEUS=true
export INSTRUMENTATION_PROMETHEUS_LISTEN_ADDR=":26660"
```

#### 方法 3: 启动时指定参数

```bash
seid start --instrumentation.prometheus=true --instrumentation.prometheus-listen-addr=":26660"
```

### 4. 验证 Metrics 端点

```bash
# 检查 Tendermint metrics
curl http://localhost:26660/metrics

# 应该看到类似输出：
# tendermint_consensus_height 12345
# tendermint_consensus_validators 1
# tendermint_mempool_size 0
# ...
```

## 配置说明

### Prometheus 配置

编辑 `prometheus/prometheus.yml` 来添加或修改监控目标：

```yaml
scrape_configs:
  - job_name: 'sei-tendermint'
    static_configs:
      - targets: ['host.docker.internal:26660']
        labels:
          instance: 'aesc-node-poc'
```

**注意**: 
- 在 Docker 中访问宿主机使用 `host.docker.internal`
- 在 Linux 上可能需要使用 `172.17.0.1` 或宿主机 IP

### Grafana 配置

#### 修改管理员密码

编辑 `docker-compose.yml`:

```yaml
environment:
  - GF_SECURITY_ADMIN_USER=admin
  - GF_SECURITY_ADMIN_PASSWORD=your-secure-password
```

#### 添加自定义 Dashboard

1. 在 Grafana UI 中创建 Dashboard
2. 导出为 JSON
3. 保存到 `grafana/dashboards/` 目录
4. 重启 Grafana 或等待自动加载

## 监控指标说明

### Tendermint 核心指标

| 指标名称 | 说明 |
|---------|------|
| `tendermint_consensus_height` | 当前区块高度 |
| `tendermint_consensus_validators` | 验证人数量 |
| `tendermint_consensus_validators_power` | 验证人总投票权 |
| `tendermint_consensus_missing_validators` | 缺失的验证人数量 |
| `tendermint_consensus_missing_validators_power` | 缺失验证人的投票权 |
| `tendermint_consensus_byzantine_validators` | 拜占庭验证人数量 |
| `tendermint_consensus_block_interval_seconds` | 区块间隔时间 |
| `tendermint_consensus_rounds` | 共识轮次 |
| `tendermint_consensus_num_txs` | 区块中的交易数量 |
| `tendermint_mempool_size` | 内存池大小 |
| `tendermint_mempool_tx_size_bytes` | 内存池交易大小 |
| `tendermint_mempool_failed_txs` | 失败的交易数量 |
| `tendermint_mempool_recheck_times` | 重新检查次数 |
| `tendermint_p2p_peers` | P2P 对等节点数量 |
| `tendermint_p2p_peer_receive_bytes_total` | 接收字节总数 |
| `tendermint_p2p_peer_send_bytes_total` | 发送字节总数 |

### Cosmos SDK 指标

| 指标名称 | 说明 |
|---------|------|
| `cosmos_bank_supply` | 代币总供应量 |
| `cosmos_staking_bonded_tokens` | 质押的代币数量 |
| `cosmos_staking_not_bonded_tokens` | 未质押的代币数量 |
| `cosmos_distribution_community_pool` | 社区池余额 |

## 常用查询示例

### PromQL 查询

```promql
# 区块生产速率（每分钟）
rate(tendermint_consensus_height[1m])

# 平均区块时间
rate(tendermint_consensus_block_interval_seconds_sum[5m]) / rate(tendermint_consensus_block_interval_seconds_count[5m])

# 内存池增长率
rate(tendermint_mempool_size[5m])

# P2P 网络流量
rate(tendermint_p2p_peer_receive_bytes_total[5m])
rate(tendermint_p2p_peer_send_bytes_total[5m])
```

## 故障排查

### 1. Prometheus 无法抓取 metrics

**检查步骤**:

```bash
# 1. 确认 Sei 节点正在运行
ps aux | grep seid

# 2. 确认 metrics 端点可访问
curl http://localhost:26660/metrics

# 3. 检查 Prometheus 配置
docker-compose exec prometheus cat /etc/prometheus/prometheus.yml

# 4. 查看 Prometheus 日志
docker-compose logs prometheus

# 5. 检查 Prometheus targets 状态
# 访问 http://localhost:9090/targets
```

**常见问题**:

- **Connection refused**: 检查 Sei 节点是否启用了 Prometheus
- **No route to host**: 检查防火墙设置
- **Docker 网络问题**: 使用正确的主机地址（`host.docker.internal` 或 IP）

### 2. Grafana 无法连接 Prometheus

**检查步骤**:

```bash
# 1. 确认 Prometheus 容器正在运行
docker-compose ps

# 2. 测试容器间网络
docker-compose exec grafana ping prometheus

# 3. 检查数据源配置
# Grafana UI -> Configuration -> Data Sources -> Prometheus
```

### 3. Dashboard 不显示数据

**检查步骤**:

1. 确认时间范围正确（右上角）
2. 检查查询语句是否有数据返回
3. 在 Prometheus UI 中手动执行查询
4. 检查 Grafana 数据源连接状态

## 高级配置

### 添加告警规则

创建 `prometheus/alerts/sei-alerts.yml`:

```yaml
groups:
  - name: sei_alerts
    interval: 30s
    rules:
      - alert: NodeDown
        expr: up{job="sei-tendermint"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Sei node is down"
          description: "Sei node {{ $labels.instance }} has been down for more than 1 minute."

      - alert: HighMempool
        expr: tendermint_mempool_size > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High mempool size"
          description: "Mempool size is {{ $value }} on {{ $labels.instance }}"
```

然后在 `prometheus.yml` 中添加：

```yaml
rule_files:
  - "alerts/*.yml"
```

### 数据持久化

数据已自动持久化到 Docker volumes：
- `prometheus-data`: Prometheus 时序数据
- `grafana-data`: Grafana 配置和 dashboards

查看 volumes:
```bash
docker volume ls | grep metrics
```

备份数据:
```bash
docker run --rm -v metrics_prometheus-data:/data -v $(pwd):/backup alpine tar czf /backup/prometheus-backup.tar.gz -C /data .
docker run --rm -v metrics_grafana-data:/data -v $(pwd):/backup alpine tar czf /backup/grafana-backup.tar.gz -C /data .
```

## 参考资料

- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [Tendermint Metrics](https://docs.tendermint.com/master/nodes/metrics.html)
- [Cosmos SDK Telemetry](https://docs.cosmos.network/main/core/telemetry)

