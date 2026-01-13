# 多节点部署脚本说明

本目录包含用于多服务器部署 Sei Chain 的脚本。

## 脚本列表

| 脚本 | 执行节点 | 说明 |
|------|---------|------|
| `step0_build.sh` | 所有节点 | 编译 seid 程序 |
| `step1_configure_init.sh` | 所有节点 | 初始化节点配置，生成 gentx |
| `step2_genesis.sh` | 所有节点 | 准备 genesis 参数 |
| `step3_add_validator_to_genesis.sh` | 仅协调节点 | 收集 gentx 并生成 genesis.json |
| `step4_config_override.sh` | 所有节点 | 应用配置文件 |
| `step5_start_sei.sh` | 所有节点 | 启动节点 |

## 部署流程

### 1. 所有节点执行 step0 和 step1

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # 每个节点使用不同的名称

./poc-deploy/localnode/multi-node-scripts/step0_build.sh
./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
```

**记录每个节点的 Node ID**（用于配置 persistent_peers）

### 2. 所有节点执行 step2

```bash
./poc-deploy/localnode/multi-node-scripts/step2_genesis.sh
```

### 3. 收集 gentx 到协调节点（手动）

将 validator1、validator2、validator3 的 gentx 文件复制到 validator0：

```bash
# 在 validator0 上执行
scp root@<validator1_ip>:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/
scp root@<validator2_ip>:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/
scp root@<validator3_ip>:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/
```

### 4. 协调节点执行 step3（仅 validator0）

```bash
# 仅在 validator0 上执行
./poc-deploy/localnode/multi-node-scripts/step3_add_validator_to_genesis.sh
```

**记录 Genesis Hash**（用于验证）

### 5. 分发 genesis.json 到所有节点（手动）

```bash
# 在 validator0 上执行
scp build/generated/genesis.json root@<validator1_ip>:~/sei-chain/build/generated/
scp build/generated/genesis.json root@<validator2_ip>:~/sei-chain/build/generated/
scp build/generated/genesis.json root@<validator3_ip>:~/sei-chain/build/generated/
```

### 6. 所有节点执行 step4

```bash
./poc-deploy/localnode/multi-node-scripts/step4_config_override.sh
```

然后手动配置 `persistent_peers`：

```bash
# 编辑 ~/.sei/config/config.toml
# 设置 persistent_peers = "node_id1@ip1:26656,node_id2@ip2:26656,..."
```

### 7. 所有节点执行 step5

```bash
./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

## 环境变量

每个脚本都使用以下环境变量：

- `CHAIN_ID`: 链 ID（默认：sei-testnet）
- `VALIDATOR_NAME`: 验证者名称（validator0, validator1, validator2, validator3）
- `GOBIN`: Go 二进制文件目录（默认：$HOME/go/bin）

## 注意事项

1. **手动步骤**：
   - 收集 gentx 文件（步骤 3）
   - 分发 genesis.json（步骤 5）
   - 配置 persistent_peers（步骤 6）

2. **验证**：
   - 确保所有节点的 genesis hash 一致
   - 确保至少 3 个节点在线才能出块

3. **备份**：
   - 备份 `~/.sei/config/priv_validator_key.json`
   - 备份 `~/.sei/config/node_key.json`

## 故障排查

### 查看日志

```bash
tail -f build/generated/logs/seid.log
```

### 检查节点状态

```bash
curl http://localhost:26657/status | jq
```

### 检查 peers

```bash
curl http://localhost:26657/net_info | jq '.result.n_peers'
```

### 重启节点

```bash
kill $(cat build/generated/seid.pid)
./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

## 完整文档

详细的部署文档请参考：`poc-deploy/localnode/deploy.md`

