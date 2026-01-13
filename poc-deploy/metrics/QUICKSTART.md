# 快速开始指南

## 5 分钟启动监控

### 步骤 1: 启用 Sei 节点 Metrics

编辑 `~/.sei/config/config.toml`:

```toml
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

重启 Sei 节点：

```bash
# 停止节点
pkill seid

# 启动节点
cd poc-deploy/localnode
./scripts/step4_start_sei.sh
```

### 步骤 2: 测试 Metrics 端点

```bash
cd poc-deploy/metrics
./test-metrics.sh
```

应该看到类似输出：

```
✓ Metrics 端点可访问

关键指标：

区块高度:
tendermint_consensus_height 12345

验证人数量:
tendermint_consensus_validators 1

内存池大小:
tendermint_mempool_size 0
```

### 步骤 3: 启动监控服务

```bash
./start.sh
```

### 步骤 4: 访问 Grafana

1. 打开浏览器访问: http://localhost:3000
2. 登录:
   - 用户名: `admin`
   - 密码: `admin`
3. 首次登录会提示修改密码（可跳过）
4. 点击左侧菜单 "Dashboards" → "Browse"
5. 选择 "Sei Chain Overview"

### 步骤 5: 查看监控数据

Dashboard 会显示：
- 当前区块高度
- 验证人数量
- 区块生产速率
- 内存池大小

## 常用命令

```bash
# 启动监控
./start.sh

# 停止监控
./stop.sh

# 查看日志
./logs.sh

# 查看特定服务日志
./logs.sh prometheus
./logs.sh grafana

# 测试 metrics
./test-metrics.sh
```

## 故障排查

### 问题 1: Grafana 显示 "No Data"

**解决方案**:

1. 检查 Prometheus 是否能抓取数据:
   ```bash
   # 访问 http://localhost:9090/targets
   # 确认 sei-tendermint 状态为 UP
   ```

2. 检查 Sei 节点 metrics:
   ```bash
   ./test-metrics.sh
   ```

3. 检查时间范围（Grafana 右上角）

### 问题 2: 无法访问 Metrics 端点

**解决方案**:

1. 确认 Sei 节点正在运行:
   ```bash
   ps aux | grep seid
   ```

2. 确认 config.toml 配置正确:
   ```bash
   grep -A 5 "\[instrumentation\]" ~/.sei/config/config.toml
   ```

3. 重启 Sei 节点

### 问题 3: Docker 容器无法启动

**解决方案**:

1. 检查端口是否被占用:
   ```bash
   lsof -i :3000  # Grafana
   lsof -i :9090  # Prometheus
   ```

2. 查看容器日志:
   ```bash
   ./logs.sh
   ```

3. 重新启动:
   ```bash
   ./stop.sh
   ./start.sh
   ```

## 下一步

- 查看 [README.md](README.md) 了解详细配置
- 自定义 Prometheus 查询
- 创建自定义 Dashboard
- 配置告警规则

