# x402-relayer 运维手册

## 1. Relayer 钱包管理

### 1.1 AEX (Gas) 余额监控

Relayer 钱包需要足够的 AEX 来支付 Gas：

```bash
# 查询余额
seid query bank balances <relayer_sei_address> --node http://localhost:26657

# 或通过 EVM RPC
cast balance 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --rpc-url http://localhost:8545
```

### 1.2 AEX 充值

```bash
# 从其他账户转账
seid tx bank send <from_address> <relayer_sei_address> 1000000000uaex \
  --from <key_name> \
  --keyring-backend file \
  --chain-id aesc-poc \
  --fees 1000uaex \
  --node http://localhost:26657
```

### 1.3 告警阈值建议

| 指标 | 警告阈值 | 严重阈值 |
|------|---------|---------|
| AEX 余额 | < 10 AEX | < 1 AEX |
| 日交易量 | > 10000 | > 50000 |
| 失败率 | > 5% | > 10% |

## 2. USDT 收入管理

### 2.1 查询 USDT 余额

```bash
# 通过 Bank 模块
seid query bank balances <relayer_sei_address> --denom usdt

# 通过 EVM (预编译)
cast call 0x0000000000000000000000000000000000001010 \
  "balanceOf(address)(uint256)" \
  0x70997970C51812dc3A010C7d01b50e0d17dc79C8 \
  --rpc-url http://localhost:8545
```

### 2.2 提取 USDT

```bash
# 通过 Cosmos 转账
seid tx bank send <relayer_address> <target_address> 1000000usdt \
  --from <key_name> \
  --keyring-backend file \
  --chain-id aesc-poc
```

## 3. 数据库维护

### 3.1 数据库位置

默认：`./x402-relayer.db`

### 3.2 备份

```bash
# 在线备份
sqlite3 x402-relayer.db ".backup backup-$(date +%Y%m%d).db"
```

### 3.3 查询统计

```bash
# 查询总记录数
sqlite3 x402-relayer.db "SELECT COUNT(*) FROM relay_records;"

# 查询成功率
sqlite3 x402-relayer.db "
SELECT 
  relay_status,
  COUNT(*) as count,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM relay_records), 2) as pct
FROM relay_records 
GROUP BY relay_status;
"

# 查询日交易量
sqlite3 x402-relayer.db "
SELECT 
  DATE(created_at) as date,
  COUNT(*) as tx_count,
  SUM(payment_amount) as total_usdt
FROM relay_records 
GROUP BY DATE(created_at)
ORDER BY date DESC
LIMIT 7;
"
```

## 4. 日志管理

### 4.1 日志位置

- 服务日志：`stdout/stderr`
- 推荐使用 systemd 或 Docker 管理

### 4.2 日志级别

通过环境变量控制：

```bash
export LOG_LEVEL=debug  # debug, info, warn, error
./x402-relayer --config config.toml
```

### 4.3 关键日志模式

```bash
# 查找支付失败
grep "settlement failed" x402-relayer.log

# 查找广播失败
grep "broadcast failed" x402-relayer.log

# 查找签名验证失败
grep "signature verification failed" x402-relayer.log
```

## 5. 故障排查

### 5.1 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| 402 支付被拒 | 签名无效/过期 | 检查客户端签名逻辑 |
| 500 结算失败 | AEX 不足 / nonce 冲突 | 充值 AEX / 重试 |
| 500 广播失败 | Gas 估算错误 / 链不可用 | 检查 RPC / Gas 设置 |
| 连接超时 | RPC 节点问题 | 检查节点状态 |

### 5.2 健康检查

```bash
# 服务健康
curl -s http://localhost:8402/health | jq .

# 节点健康
curl -s http://localhost:26657/status | jq .result.sync_info.catching_up

# RPC 健康
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

### 5.3 重启服务

```bash
# systemd
sudo systemctl restart x402-relayer

# Docker
docker restart x402-relayer
```

## 6. 监控指标

### 6.1 Prometheus 指标 (建议实现)

```
# 请求计数
x402_relay_requests_total{status="success|failed"}

# Gas 消耗
x402_gas_used_total

# 支付金额
x402_payment_amount_total

# 延迟
x402_request_duration_seconds
```

### 6.2 Grafana 仪表盘

建议监控：
- 请求 QPS
- 成功率
- 平均延迟
- AEX 余额趋势
- USDT 收入趋势

## 7. 安全建议

1. **私钥保护**：使用环境变量，不要硬编码
2. **网络隔离**：x402-relayer 只暴露给可信客户端
3. **日志脱敏**：不记录私钥和完整签名
4. **定期轮换**：定期更换 Relayer 钱包
5. **限流**：实现请求限流，防止 DoS

