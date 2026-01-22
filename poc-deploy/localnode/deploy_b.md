# Sei Chain 多节点部署指南 - 方案 B（单节点 + 动态加入）

## 概述

**方案 B** 采用"单节点启动 + 动态加入验证者"的方式部署多节点测试网络。

## 代币说明

本链使用双代币模型：

| 代币 | Denom | 用途 |
|------|-------|------|
| **AEX** | `uaex` | Gas 代币，用于交易手续费 |
| **STAEX** | `ustaex` | 质押代币，用于质押/治理/投票 |

- **1 AEX = 1,000,000 uaex**
- **1 STAEX = 1,000,000 ustaex**

### 核心思路

1. **启动单节点**：使用 `deploy.sh` 脚本启动 1 个 genesis 验证者
2. **其他节点同步**：其他节点复制 genesis.json 并同步区块
3. **动态成为验证者**：通过 `create-validator` 交易将全节点升级为验证者
4. **RPC 节点**：部署专门的 RPC 节点，不参与共识

### 优势

- ✅ **部署简单**：validator0 直接运行 `deploy.sh`，无需手动配置
- ✅ **灵活性高**：可以随时添加或删除验证者
- ✅ **真实模拟**：模拟生产环境中验证者加入流程
- ✅ **快速测试**：适合快速搭建测试环境

### 劣势

- ⚠️ 需要手动执行 create-validator 交易
- ⚠️ 需要确保账户有足够的代币用于质押
- ⚠️ 初始只有 1 个验证者（单点）

---

## 架构概述

### 网络拓扑

```
┌─────────────────────────────────────────────────────────────────┐
│                     Sei Testnet (方案 B)                          │
│                   4 Validators + 1 RPC Node                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐                                                │
│  │  validator0  │  ← Genesis 验证者（deploy.sh 启动）            │
│  │  (Genesis)   │     CHAIN_ID: aesc-poc                          │
│  └──────┬───────┘                                                │
│         │                                                         │
│         │ P2P 连接                                                │
│         │                                                         │
│  ┌──────┴───────┬──────────────┬──────────────┬──────────────┐  │
│  │              │              │              │              │  │
│  ▼              ▼              ▼              ▼              ▼  │
│ ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐   │
│ │validator1│ │validator2│ │validator3│ │   rpc   │ │(全节点)│   │
│ │(全节点) │  │(全节点) │  │(全节点) │  │(RPC节点)│ │        │   │
│ └────┬───┘  └────┬───┘  └────┬───┘  └────────┘            │   │
│      │           │           │                              │   │
│      │ create-validator 交易 │                              │   │
│      ▼           ▼           ▼                              │   │
│ ┌────────┐  ┌────────┐  ┌────────┐                         │   │
│ │validator1│ │validator2│ │validator3│                      │   │
│ │(验证者) │  │(验证者) │  │(验证者) │                       │   │
│ └────────┘  └────────┘  └────────┘                         │   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 服务器规划

| 节点名称 | IP 地址 | 角色 | 说明 |
|---------|---------|------|------|
| validator0 | 192.168.1.10 | Genesis 验证者 | 使用 deploy.sh 启动 |
| validator1 | 192.168.1.11 | 全节点 → 验证者 | 通过 create-validator 加入 |
| validator2 | 192.168.1.12 | 全节点 → 验证者 | 通过 create-validator 加入 |
| validator3 | 192.168.1.13 | 全节点 → 验证者 | 通过 create-validator 加入 |
| rpc | 192.168.1.14 | RPC 节点 | 只同步区块，不参与共识 |

---

## 快速开始

如果你已经完成了前置准备，可以按照以下步骤快速部署：

### validator0（192.168.1.10）

```bash
cd ~/sei-chain
./poc-deploy/localnode/scripts/deploy.sh
# 记录 Node ID 和 IP
seid tendermint show-node-id
```

### validator1-3 + rpc（192.168.1.11-14）

```bash
cd ~/sei-chain
make install

# 初始化（每个节点使用不同的名称）
seid init validator1 --chain-id aesc-poc

# 复制 genesis.json
scp root@192.168.1.10:~/.sei/config/genesis.json ~/.sei/config/genesis.json

# 配置 persistent_peers（替换为实际的 Node ID）
sed -i "s/persistent_peers = \"\"/persistent_peers = \"<node_id>@192.168.1.10:26656\"/" ~/.sei/config/config.toml

# 启动节点
mkdir -p build/generated/logs
nohup seid start --chain-id aesc-poc > build/generated/logs/seid.log 2>&1 &
echo $! > build/generated/seid.pid
```

### validator1-3 成为验证者

```bash
# 创建账户
printf "12345678\n" | seid keys add validator1

# 在 validator0 上转账 AEX (Gas 代币) 和 STAEX (质押代币)
seid tx bank send admin <validator1_address> 100000000uaex --chain-id aesc-poc --fees 2000uaex -y
seid tx bank send admin <validator1_address> 100000000ustaex --chain-id aesc-poc --fees 2000uaex -y

# 创建验证者（使用 STAEX 质押）
printf "12345678\n" | seid tx staking create-validator \
  --amount=10000000ustaex \
  --pubkey=$(seid tendermint show-validator) \
  --moniker="validator1" \
  --chain-id="aesc-poc" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from="validator1" \
  --fees=2000uaex \
  -y
```

---

## 前置准备

### 1. 软件依赖

在**所有 5 个服务器**上安装以下软件：

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装基础工具
sudo apt install -y build-essential git jq curl wget bc

# 安装 Go 1.24.9（使用 snap）
sudo snap install go --classic --channel=1.24/stable

# 验证 Go 版本
go version  # 应该显示 go version go1.24.x

# 配置 Go 环境变量（添加到 ~/.bashrc）
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc
```

### 2. 克隆代码

在**所有 5 个服务器**上克隆代码：

```bash
# 克隆仓库
cd ~
git clone https://github.com/sei-protocol/sei-chain.git
cd sei-chain

# 切换到目标分支（如果需要）
# git checkout <branch_name>
```

### 3. 网络配置

确保所有节点之间可以互相访问：

```bash
# 测试网络连通性
ping -c 3 192.168.1.10  # validator0
ping -c 3 192.168.1.11  # validator1
ping -c 3 192.168.1.12  # validator2
ping -c 3 192.168.1.13  # validator3
ping -c 3 192.168.1.14  # rpc

# 确保端口 26656 (P2P) 和 26657 (RPC) 开放
# 如果有防火墙，需要开放这些端口
sudo ufw allow 26656/tcp
sudo ufw allow 26657/tcp
```

---

## 部署步骤

### 步骤概览

```
阶段 1: validator0 启动单节点（deploy.sh）
  ↓
阶段 2: validator1-3 + rpc 同步区块（作为全节点）
  ↓
阶段 3: validator1-3 创建验证者账户
  ↓
阶段 4: validator0 给 validator1-3 转账
  ↓
阶段 5: validator1-3 执行 create-validator 交易
  ↓
阶段 6: 验证多节点共识
```

---

### 阶段 1：启动单节点（validator0）

在 **validator0** 上执行：

```bash
cd ~/sei-chain

# 直接运行部署脚本（使用默认配置）
./poc-deploy/localnode/scripts/deploy.sh
```

**脚本会自动完成**：
- 编译 seid
- 初始化节点（CHAIN_ID=aesc-poc, MONIKER=aesc-node-poc）
- 创建 validator 账户和 admin 账户
- 配置 genesis 参数（包括禁用 Oracle 惩罚）
- 启动链

**等待链启动成功**：

```bash
# 查看日志
tail -f build/generated/logs/seid.log

# 检查节点状态
curl http://localhost:26657/status | jq

# 应该看到区块高度在增长
```

**记录关键信息**（其他节点需要）：

```bash
# 1. Node ID
seid tendermint show-node-id
# 输出示例：7c3b1849414937f8d538b2761909bba34961cb99

# 2. IP 地址
hostname -I | awk '{print $1}'
# 输出示例：192.168.1.10

# 3. Genesis Hash（用于验证）
sha256sum ~/.sei/config/genesis.json
```

---

### 阶段 2：其他节点同步区块（validator1-3 + rpc）

在 **validator1, validator2, validator3, rpc** 上分别执行：

#### 2.1 编译 seid

```bash
cd ~/sei-chain

# 编译（与 validator0 相同）
make install

# 验证
seid version
```

#### 2.2 初始化节点

```bash
# 设置节点名称（每个节点不同）
NODE_NAME="validator1"  # validator1, validator2, validator3, rpc

# 初始化节点（使用与 validator0 相同的 CHAIN_ID）
seid init "$NODE_NAME" --chain-id aesc-poc
```

#### 2.3 复制 genesis.json

```bash
# 从 validator0 复制 genesis.json
scp root@192.168.1.10:~/.sei/config/genesis.json ~/.sei/config/genesis.json

# 验证 genesis hash（应该与 validator0 一致）
sha256sum ~/.sei/config/genesis.json
```

#### 2.4 配置 persistent_peers

```bash
# 配置连接到 validator0
VALIDATOR0_NODE_ID="7c3b1849414937f8d538b2761909bba34961cb99"  # 替换为实际值
VALIDATOR0_IP="192.168.1.10"  # 替换为实际值

sed -i "s/persistent_peers = \"\"/persistent_peers = \"$VALIDATOR0_NODE_ID@$VALIDATOR0_IP:26656\"/" ~/.sei/config/config.toml
```

#### 2.5 复制配置文件（可选）

```bash
# 如果需要与 validator0 相同的配置，可以复制
scp root@192.168.1.10:~/sei-chain/poc-deploy/localnode/config/app.toml ~/.sei/config/app.toml
scp root@192.168.1.10:~/sei-chain/poc-deploy/localnode/config/config.toml ~/.sei/config/config.toml

# 重新配置 persistent_peers（因为 config.toml 被覆盖了）
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$VALIDATOR0_NODE_ID@$VALIDATOR0_IP:26656\"/" ~/.sei/config/config.toml
```

#### 2.6 启动节点

```bash
# 创建日志目录
mkdir -p build/generated/logs

# 启动节点（作为全节点同步）
nohup seid start --chain-id aesc-poc > build/generated/logs/seid.log 2>&1 &

# 保存 PID
echo $! > build/generated/seid.pid
```

#### 2.7 验证同步状态

```bash
# 查看日志
tail -f build/generated/logs/seid.log

# 检查同步状态
curl http://localhost:26657/status | jq '.result.sync_info'

# 等待 catching_up 变为 false
```

---

### 阶段 3：创建验证者账户（validator1-3）

**注意**：只在 **validator1, validator2, validator3** 上执行，**rpc 节点不需要**。

在 **validator1, validator2, validator3** 上分别执行：

```bash
# 设置验证者名称（每个节点不同）
VALIDATOR_NAME="validator1"  # validator1, validator2, validator3

# 创建验证者账户
printf "12345678\n" | seid keys add "$VALIDATOR_NAME"

# 记录地址（重要！后续转账需要）
VALIDATOR_ADDRESS=$(printf "12345678\n" | seid keys show "$VALIDATOR_NAME" -a)
echo "Validator Address: $VALIDATOR_ADDRESS"

# 导出验证者公钥（后续 create-validator 需要）
seid tendermint show-validator
```

---

### 阶段 4：转账给验证者账户（validator0）

在 **validator0** 上执行：

```bash
# 给每个验证者账户转账（用于 Gas 和质押）
# 注意：使用 admin 账户转账，CHAIN_ID 是 aesc-poc
# AEX (uaex) 用于交易手续费，STAEX (ustaex) 用于质押

# 转账 AEX (Gas 代币)
seid tx bank send admin <validator1_address> 100000000uaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

seid tx bank send admin <validator2_address> 100000000uaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

seid tx bank send admin <validator3_address> 100000000uaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

# 转账 STAEX (质押代币)
seid tx bank send admin <validator1_address> 100000000ustaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

seid tx bank send admin <validator2_address> 100000000ustaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

seid tx bank send admin <validator3_address> 100000000ustaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y
```

**验证转账成功**：

在各个验证者节点上检查余额：

```bash
seid query bank balances <validator_address>
```

---

### 阶段 5：执行 create-validator 交易（validator1-3）

**注意**：只在 **validator1, validator2, validator3** 上执行，**rpc 节点不需要**。

在 **validator1, validator2, validator3** 上分别执行：

```bash
# 设置验证者名称（每个节点不同）
VALIDATOR_NAME="validator1"  # validator1, validator2, validator3

# 创建验证者（使用 STAEX 质押，AEX 支付手续费）
printf "12345678\n" | seid tx staking create-validator \
  --amount=10000000ustaex \
  --pubkey=$(seid tendermint show-validator) \
  --moniker="$VALIDATOR_NAME" \
  --chain-id="aesc-poc" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from="$VALIDATOR_NAME" \
  --fees=2000uaex \
  -y

# 等待交易被打包（约 2-4 秒）
sleep 5

# 验证验证者状态
seid query staking validator $(seid keys show "$VALIDATOR_NAME" --bech val -a)
```

---

### 阶段 6：验证多节点共识

在任意节点上执行：

```bash
# 查看所有验证者
seid query staking validators --output json | jq '.validators[] | {moniker, status, tokens}'

# 应该看到 4 个验证者：
# - validator (Genesis 验证者，来自 validator0)
# - validator1 (Bonded)
# - validator2 (Bonded)
# - validator3 (Bonded)

# 查看最新区块的签名数量
curl http://localhost:26657/block | jq '.result.block.last_commit.signatures | length'
# 应该看到 4 个签名（表示 4 个验证者在共识）

# 查看验证者集合
curl http://localhost:26657/validators | jq '.result.validators[] | {address, voting_power}'
```

### RPC 节点验证

在 **rpc** 节点上验证：

```bash
# 检查节点状态（应该已同步）
curl http://localhost:26657/status | jq '.result.sync_info'

# 查看验证者列表（应该看到 4 个验证者）
seid query staking validators --output json | jq '.validators[] | .description.moniker'

# RPC 节点不应该在验证者列表中
```

---

## 验证和测试

### 1. 检查节点状态

```bash
# 查看节点信息
curl http://localhost:26657/status | jq

# 查看验证者集合
curl http://localhost:26657/validators | jq '.result.validators[] | {address, voting_power}'
```

### 2. 测试交易

```bash
# 创建测试账户
printf "12345678\n" | seid keys add test_user

# 转账 AEX 测试（Gas 代币）
seid tx bank send admin $(seid keys show test_user -a) 1000000uaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

# 转账 STAEX 测试（质押代币）
seid tx bank send admin $(seid keys show test_user -a) 1000000ustaex \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y

# 查询余额
seid query bank balances $(seid keys show test_user -a)
```

### 3. 验证共识

```bash
# 查看最新区块的签名数量
for i in {1..10}; do
  HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
  SIGS=$(curl -s http://localhost:26657/block?height=$HEIGHT | jq '.result.block.last_commit.signatures | length')
  echo "Block $HEIGHT: $SIGS signatures"
  sleep 2
done

# 应该看到每个区块都有多个签名
```

---

## 故障排查

### 问题 1：节点无法同步

**症状**：`catching_up: true` 一直不变

**解决方法**：

```bash
# 检查 persistent_peers 配置
grep "persistent_peers" ~/.sei/config/config.toml

# 检查网络连接
telnet 192.168.1.10 26656

# 查看日志
tail -f build/generated/logs/seid.log | grep -i "error\|peer"
```

### 问题 2：create-validator 交易失败

**症状**：交易返回错误

**常见原因**：

1. **余额不足**：
   ```bash
   seid query bank balances <validator_address>
   ```

2. **验证者已存在**：
   ```bash
   seid query staking validator $(seid keys show "$VALIDATOR_NAME" --bech val -a)
   ```

3. **公钥已被使用**：
   ```bash
   # 检查是否有其他验证者使用了相同的公钥
   seid query staking validators --output json | jq '.validators[] | .consensus_pubkey'
   ```

### 问题 3：验证者未参与共识

**症状**：验证者状态为 Bonded，但没有签名

**解决方法**：

```bash
# 检查验证者状态
seid query staking validator $(seid keys show "$VALIDATOR_NAME" --bech val -a) | jq '.status'

# 检查是否被 jail
seid query slashing signing-info $(seid tendermint show-validator)

# 重启节点
kill $(cat build/generated/seid.pid)
nohup seid start --chain-id sei-testnet > build/generated/logs/seid.log 2>&1 &
echo $! > build/generated/seid.pid
```

---

## 维护和监控

### 停止节点

```bash
# 停止节点
kill $(cat build/generated/seid.pid)

# 验证已停止
ps aux | grep seid
```

### 重启节点

```bash
# 重启节点
nohup seid start --chain-id sei-testnet > build/generated/logs/seid.log 2>&1 &
echo $! > build/generated/seid.pid
```

### 添加更多验证者

重复阶段 2-5 即可添加新的验证者。

### 删除验证者

```bash
# 解绑验证者（需要等待 21 天，使用 STAEX）
seid tx staking unbond $(seid keys show "$VALIDATOR_NAME" --bech val -a) 10000000ustaex \
  --from="$VALIDATOR_NAME" \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y
```

### RPC 节点管理

RPC 节点只需要同步区块，不需要执行 create-validator：

```bash
# 停止 RPC 节点
kill $(cat build/generated/seid.pid)

# 重启 RPC 节点
nohup seid start --chain-id aesc-poc > build/generated/logs/seid.log 2>&1 &
echo $! > build/generated/seid.pid

# 检查同步状态
curl http://localhost:26657/status | jq '.result.sync_info'
```

---

## 快速参考

### 常用命令

```bash
# 查看所有验证者
seid query staking validators --output json | jq '.validators[] | {moniker, status, tokens}'

# 查看节点状态
curl http://localhost:26657/status | jq

# 查看日志
tail -f build/generated/logs/seid.log

# 停止节点
kill $(cat build/generated/seid.pid)

# 启动节点
nohup seid start --chain-id aesc-poc > build/generated/logs/seid.log 2>&1 &
echo $! > build/generated/seid.pid

# 查看账户余额
seid query bank balances <address>

# 查看验证者详情
seid query staking validator <validator_address>
```

### 重要配置

| 配置项 | 值 | 说明 |
|--------|-----|------|
| CHAIN_ID | aesc-poc | 链 ID（deploy.sh 默认值） |
| MONIKER | aesc-node-poc | validator0 的节点名称 |
| P2P 端口 | 26656 | 节点间通信端口 |
| RPC 端口 | 26657 | RPC 服务端口 |
| Genesis 验证者质押 | 100 STAEX | validator0 的初始质押（ustaex）|
| 动态验证者质押 | 10 STAEX | validator1-3 的质押（ustaex）|
| Gas 代币 | uaex | 用于交易手续费 |
| 质押代币 | ustaex | 用于质押/治理/投票 |

### 脚本位置

- **单节点部署脚本**：`poc-deploy/localnode/scripts/deploy.sh`
- **配置文件模板**：`poc-deploy/localnode/config/`
- **日志目录**：`build/generated/logs/`
- **Genesis 文件**：`~/.sei/config/genesis.json`
- **配置文件**：`~/.sei/config/config.toml`, `~/.sei/config/app.toml`

---

## 常见问题

### Q1: 为什么 validator0 的验证者名称是 "validator" 而不是 "validator0"？

**A**: 因为 `deploy.sh` 脚本中默认的账户名称是 "validator"。如果需要修改，可以在运行脚本前设置环境变量：

```bash
export MONIKER="validator0"
./poc-deploy/localnode/scripts/deploy.sh
```

### Q2: 如何查看 validator0 的 admin 账户密码？

**A**: `deploy.sh` 脚本中所有账户的密码都是 `12345678`（仅用于测试）。

### Q3: RPC 节点需要多少存储空间？

**A**: RPC 节点需要存储完整的区块链数据，建议至少 100GB 的磁盘空间。

### Q4: 如何增加验证者的质押金额？

**A**: 使用 `delegate` 命令（使用 STAEX 质押代币）：

```bash
seid tx staking delegate <validator_address> 10000000ustaex \
  --from=<account_name> \
  --chain-id aesc-poc \
  --fees 2000uaex \
  -y
```

### Q5: 节点之间无法连接怎么办？

**A**: 检查以下几点：
1. 防火墙是否开放 26656 端口
2. persistent_peers 配置是否正确
3. Node ID 是否正确
4. 网络是否互通（ping 测试）

### Q6: 如何备份验证者密钥？

**A**: 备份以下文件：

```bash
# 备份验证者密钥
cp ~/.sei/config/priv_validator_key.json ~/backup/
cp ~/.sei/data/priv_validator_state.json ~/backup/

# 备份账户密钥
seid keys export validator > ~/backup/validator.key
```

---

## 注意事项

### ⚠️ 安全警告

1. **密码安全**：生产环境请使用强密码，不要使用 `12345678`
2. **密钥备份**：务必备份 `priv_validator_key.json` 和账户助记词
3. **防火墙**：生产环境应限制 RPC 端口（26657）的访问
4. **Oracle 惩罚**：当前配置已禁用 Oracle 惩罚（`min_valid_per_window=0`），生产环境需要配置 Price Feeder

### 💡 最佳实践

1. **节点命名**：使用有意义的节点名称，方便识别
2. **日志管理**：定期清理日志文件，避免磁盘占满
3. **监控**：建议使用 Prometheus + Grafana 监控节点状态
4. **备份**：定期备份 genesis.json 和验证者密钥
5. **测试**：在生产环境部署前，先在测试环境验证

---

## 总结

**方案 B** 提供了一种简单灵活的多节点部署方式：

- ✅ **validator0**：使用 `deploy.sh` 一键启动，无需手动配置
- ✅ **validator1-3**：通过 create-validator 动态加入
- ✅ **rpc**：专门的 RPC 节点，不参与共识
- ✅ **CHAIN_ID**：统一使用 `aesc-poc`（deploy.sh 的默认值）
- ✅ 模拟真实的验证者加入流程
- ✅ 适合测试环境和快速部署

**最终架构**：
- 4 个验证者节点（validator0 + validator1-3）
- 1 个 RPC 节点（只同步区块）
- 共 5 个节点

**与方案 A 的对比**：
- **方案 A**（`deploy_a.md`）：所有验证者在 genesis 中定义，需要收集 gentx，适合生产环境
- **方案 B**（本文档）：单节点启动 + 动态加入，简单快速，适合测试环境

**下一步**：
- 如果需要更复杂的配置，参考 `deploy_a.md`
- 如果需要批量测试，参考 `poc-deploy/tools/` 目录下的工具
- 如果遇到问题，查看故障排查部分或查看日志

根据你的需求选择合适的方案！🎉

