# Sei Chain 多节点部署指南

## 架构概述

本文档描述如何在 **4 台独立服务器**上部署 Sei Chain 测试网络（4 个验证者节点）。

### 网络拓扑

```
┌─────────────────────────────────────────────────────────────┐
│                      Sei Chain 测试网络                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Validator 0  │  │ Validator 1  │  │ Validator 2  │     │
│  │ 192.168.1.11 │  │ 192.168.1.12 │  │ 192.168.1.13 │     │
│  │ (协调节点)    │  │              │  │              │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                 │                 │              │
│         └─────────────────┼─────────────────┘              │
│                           │                                │
│                  ┌────────┴────────┐                       │
│                  │  Validator 3    │                       │
│                  │  192.168.1.14   │                       │
│                  └─────────────────┘                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 服务器规划

| 服务器 | IP 地址 | 角色 | 说明 |
|--------|---------|------|------|
| Server 0 | 192.168.1.11 | Validator 0 | 协调节点，负责生成 genesis |
| Server 1 | 192.168.1.12 | Validator 1 | 验证者节点 |
| Server 2 | 192.168.1.13 | Validator 2 | 验证者节点 |
| Server 3 | 192.168.1.14 | Validator 3 | 验证者节点 |

**注意**：请将示例 IP 地址替换为实际的服务器 IP。

### 端口规划

| 端口 | 用途 | 开放范围 |
|------|------|---------|
| 26656 | P2P 通信 | 所有节点互通 |
| 26657 | RPC | 可选开放（调试用） |
| 9090 | gRPC | 可选开放（调试用） |

---

## 前置准备

### 1. 软件依赖

所有服务器需要安装：

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装基础工具
sudo apt install -y build-essential git curl wget jq bc snapd

# 安装 Go 1.24.9 (使用 snap)
sudo snap install go --classic --channel=1.24/stable

# 配置环境变量
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc

# 创建 GOBIN 目录
mkdir -p $GOBIN

# 验证安装
go version
```

### 2. 克隆代码

所有服务器执行：

```bash
cd ~
git clone https://github.com/sei-protocol/sei-chain.git
cd sei-chain
git checkout main  # 或指定版本
```

---

## 部署步骤

### 步骤概览

```
阶段 1: 所有节点编译程序
  ↓
阶段 2: 所有节点初始化配置
  ↓
阶段 3: 所有节点准备 genesis 参数
  ↓
阶段 4: 收集 gentx 到协调节点 (手动)
  ↓
阶段 5: 协调节点生成 genesis.json (手动)
  ↓
阶段 6: 分发 genesis.json 到所有节点 (手动)
  ↓
阶段 7: 所有节点应用配置
  ↓
阶段 8: 所有节点启动
```

---

### 阶段 1：编译程序（所有节点）

在**所有 4 台服务器**上执行：

```bash
cd ~/sei-chain

# 设置环境变量
export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # 每台服务器使用不同的名称：validator0, validator1, validator2, validator3

# 执行编译脚本
./poc-deploy/localnode/multi-node-scripts/step0_build.sh
```

---

### 阶段 2：初始化节点（所有节点）

在**每台服务器**上执行，注意替换 `VALIDATOR_NAME`：

**Server 0 (192.168.1.11) - 协调节点**：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"

./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
```

执行完成后，**记录输出的 Node ID**，例如：
```
Node ID: aaa111222333444555666777888999000
```

**Server 1 (192.168.1.12)**：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator1"

./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
```

记录 Node ID。

**Server 2 (192.168.1.13)**：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator2"

./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
```

记录 Node ID。

**Server 3 (192.168.1.14)**：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator3"

./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
```

记录 Node ID。

---

### 阶段 3：准备 Genesis 参数（所有节点）

在**所有 4 台服务器**上执行：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # 替换为对应的验证者名称

./poc-deploy/localnode/multi-node-scripts/step2_genesis.sh
```

---

### 阶段 4：收集 gentx 文件（手动操作）

**目标**：将 validator1、validator2、validator3 的 gentx 文件复制到 validator0 的 `build/generated/gentx/` 目录。

#### 方式 1：使用 scp（推荐）

在 **Server 0 (validator0)** 上执行：

```bash
cd ~/sei-chain

# 从 validator1 复制
scp root@192.168.1.12:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/

# 从 validator2 复制
scp root@192.168.1.13:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/

# 从 validator3 复制
scp root@192.168.1.14:~/sei-chain/build/generated/gentx/*.json build/generated/gentx/

# 验证文件
ls -l build/generated/gentx/
```

应该看到 4 个 gentx 文件（包括 validator0 自己的）。

#### 方式 2：手动复制

在 **Server 1-3** 上执行：

```bash
# 查看 gentx 内容
cat ~/sei-chain/build/generated/gentx/*.json
```

复制输出内容，然后在 **Server 0** 上创建文件：

```bash
cd ~/sei-chain/build/generated/gentx/

# 创建文件并粘贴内容
vim gentx-validator1.json  # 粘贴 validator1 的 gentx
vim gentx-validator2.json  # 粘贴 validator2 的 gentx
vim gentx-validator3.json  # 粘贴 validator3 的 gentx
```

---

### 阶段 5：生成 Genesis 文件（协调节点，手动操作）

**仅在 Server 0 (validator0)** 上执行：

```bash
cd ~/sei-chain

./poc-deploy/localnode/multi-node-scripts/step3_add_validator_to_genesis.sh
```

执行完成后，会输出：

```
Genesis file: build/generated/genesis.json
Genesis hash: abc123def456...
```

**记录 Genesis Hash**，用于验证其他节点的 genesis 文件是否一致。

---

### 阶段 6：分发 Genesis 文件（手动操作）

**目标**：将 validator0 的 `build/generated/genesis.json` 复制到其他所有节点的 `build/generated/` 目录。

#### 方式 1：使用 scp（推荐）

从 **Server 0 (validator0)** 执行：

```bash
cd ~/sei-chain

# 分发到 validator1
scp build/generated/genesis.json root@192.168.1.12:~/sei-chain/build/generated/genesis.json

# 分发到 validator2
scp build/generated/genesis.json root@192.168.1.13:~/sei-chain/build/generated/genesis.json

# 分发到 validator3
scp build/generated/genesis.json root@192.168.1.14:~/sei-chain/build/generated/genesis.json
```

#### 方式 2：手动复制

在 **Server 0** 上查看 genesis.json：

```bash
cat ~/sei-chain/build/generated/genesis.json
```

在 **Server 1-3** 上创建文件：

```bash
cd ~/sei-chain
mkdir -p build/generated

# 创建文件并粘贴内容
vim build/generated/genesis.json
```

#### 验证 Genesis Hash

在**所有节点**上执行：

```bash
sha256sum ~/sei-chain/build/generated/genesis.json
```

确保所有节点的 hash 值一致！

---

### 阶段 7：应用配置（所有节点）

在**所有 4 台服务器**上执行：

```bash
cd ~/sei-chain

export VALIDATOR_NAME="validator0"  # 替换为对应的验证者名称

./poc-deploy/localnode/multi-node-scripts/step4_config_override.sh
```

执行完成后，脚本会提示配置 `persistent_peers`。

#### 配置 Persistent Peers

根据阶段 2 记录的 Node ID，编辑 `~/.sei/config/config.toml`：

**Server 0 (validator0)** 配置：

```bash
# 连接到 validator1, validator2, validator3
PEERS="<validator1_node_id>@192.168.1.12:26656,<validator2_node_id>@192.168.1.13:26656,<validator3_node_id>@192.168.1.14:26656"

sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.sei/config/config.toml
```

**Server 1 (validator1)** 配置：

```bash
# 连接到 validator0, validator2, validator3
PEERS="<validator0_node_id>@192.168.1.11:26656,<validator2_node_id>@192.168.1.13:26656,<validator3_node_id>@192.168.1.14:26656"

sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.sei/config/config.toml
```

**Server 2 (validator2)** 配置：

```bash
# 连接到 validator0, validator1, validator3
PEERS="<validator0_node_id>@192.168.1.11:26656,<validator1_node_id>@192.168.1.12:26656,<validator3_node_id>@192.168.1.14:26656"

sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.sei/config/config.toml
```

**Server 3 (validator3)** 配置：

```bash
# 连接到 validator0, validator1, validator2
PEERS="<validator0_node_id>@192.168.1.11:26656,<validator1_node_id>@192.168.1.12:26656,<validator2_node_id>@192.168.1.13:26656"

sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.sei/config/config.toml
```

**注意**：将 `<validatorX_node_id>` 替换为实际的 Node ID。

---

### 阶段 8：启动节点（所有节点）

在**所有 4 台服务器**上执行：

```bash
cd ~/sei-chain

export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # 替换为对应的验证者名称

./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

**等待共识**：至少需要 3 个验证者节点在线才能开始出块（4个验证者的 2/3+1）。

---

## 验证和测试

### 1. 检查节点状态

在任意节点上执行：

```bash
# 查看节点状态
curl http://localhost:26657/status | jq

# 查看连接的 peers（应该显示 3 个 peers）
curl http://localhost:26657/net_info | jq '.result.n_peers'

# 查看最新区块高度
curl http://localhost:26657/block | jq '.result.block.header.height'

# 查看详细的 peers 信息
curl http://localhost:26657/net_info | jq '.result.peers[] | {moniker, remote_ip}'
```

### 2. 验证共识

```bash
# 查看验证者集合（应该显示 4 个验证者）
seid query staking validators --output json | jq '.validators[] | {moniker, status, tokens}'

# 查看当前区块高度
seid status | jq '.SyncInfo.latest_block_height'

# 查看是否在同步
seid status | jq '.SyncInfo.catching_up'
```

### 3. 测试交易

```bash
# 查看账户余额
seid query bank balances $(seid keys show validator0 -a)

# 发送测试交易
seid tx bank send validator0 <接收地址> 1000usei \
  --chain-id sei-testnet \
  --fees 2000usei \
  --gas 200000 \
  -y
```

---

## 故障排查

### 常见问题

#### 1. 节点无法连接到 peers

**症状**：`curl http://localhost:26657/net_info` 显示 peers 数量为 0

**解决方法**：

```bash
# 1. 检查 persistent_peers 配置
grep persistent_peers ~/.sei/config/config.toml

# 2. 检查防火墙（确保 26656 端口开放）
sudo ufw status
sudo ufw allow 26656/tcp

# 3. 手动测试连接
telnet 192.168.1.11 26656

# 4. 检查节点是否在运行
ps aux | grep seid

# 5. 查看日志中的连接错误
grep -i "peer" build/generated/logs/seid.log | tail -20
```

#### 2. 链无法出块

**症状**：区块高度停留在 0 或很小的数字

**原因**：验证者节点数量不足（需要至少 3/4 = 3 个节点在线）

**解决方法**：

```bash
# 1. 确保至少 3 个验证者节点在线
# 在每个节点上检查进程
ps aux | grep seid

# 2. 检查每个节点的日志
tail -f build/generated/logs/seid.log

# 3. 查看验证者状态
seid query staking validators --output json | jq '.validators[] | {moniker, status}'

# 4. 检查是否有错误日志
grep -i "error\|fail" build/generated/logs/seid.log | tail -20
```

#### 3. Genesis 文件不匹配

**症状**：节点启动失败，提示 genesis hash 不匹配

**解决方法**：

```bash
# 1. 在所有节点上验证 genesis hash
sha256sum build/generated/genesis.json

# 2. 如果不一致，重新从 validator0 复制
scp root@192.168.1.11:~/sei-chain/build/generated/genesis.json build/generated/genesis.json

# 3. 重新应用配置
./poc-deploy/localnode/multi-node-scripts/step4_config_override.sh

# 4. 重启节点
pkill seid
./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

#### 4. 节点启动失败

**症状**：执行 step5_start_sei.sh 后进程立即退出

**解决方法**：

```bash
# 1. 查看详细日志
cat build/generated/logs/seid.log

# 2. 检查端口是否被占用
lsof -i :26656
lsof -i :26657

# 3. 检查 genesis 文件是否存在
ls -l ~/.sei/config/genesis.json

# 4. 验证 genesis 文件格式
seid validate-genesis

# 5. 清理并重新初始化
rm -rf ~/.sei
./poc-deploy/localnode/multi-node-scripts/step1_configure_init.sh
# 然后重新执行后续步骤
```

### 查看日志

```bash
# 实时查看日志
tail -f build/generated/logs/seid.log

# 搜索错误
grep -i error build/generated/logs/seid.log

# 搜索警告
grep -i warn build/generated/logs/seid.log

# 查看最近 100 行
tail -n 100 build/generated/logs/seid.log

# 查看连接相关日志
grep -i "peer\|connection" build/generated/logs/seid.log | tail -20
```

---

## 维护和监控

### 停止节点

```bash
# 方式 1：使用 PID 文件
kill $(cat build/generated/seid.pid)

# 方式 2：查找进程并停止
ps aux | grep seid
pkill seid

# 方式 3：强制停止（不推荐）
pkill -9 seid
```

### 重启节点

```bash
cd ~/sei-chain

# 停止节点
kill $(cat build/generated/seid.pid) 2>/dev/null || pkill seid

# 等待几秒
sleep 5

# 重新启动
export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # 替换为对应的验证者名称
./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

### 备份关键文件

```bash
# 创建备份目录
mkdir -p ~/sei-backup

# 备份验证者密钥（非常重要！）
cp ~/.sei/config/priv_validator_key.json ~/sei-backup/

# 备份节点密钥
cp ~/.sei/config/node_key.json ~/sei-backup/

# 备份 genesis
cp ~/.sei/config/genesis.json ~/sei-backup/

# 备份账户密钥
cp -r ~/.sei/keyring-test ~/sei-backup/

# 打包备份
tar -czf ~/sei-backup-$(date +%Y%m%d).tar.gz ~/sei-backup/
```

### 监控脚本

创建 `~/sei-chain/monitor.sh`：

```bash
#!/bin/bash

LOG_FILE="build/generated/logs/monitor.log"

while true; do
  HEIGHT=$(curl -s http://localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height // "N/A"')
  PEERS=$(curl -s http://localhost:26657/net_info 2>/dev/null | jq -r '.result.n_peers // "N/A"')
  CATCHING_UP=$(curl -s http://localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.catching_up // "N/A"')

  TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
  LOG_LINE="$TIMESTAMP - Height: $HEIGHT, Peers: $PEERS, Syncing: $CATCHING_UP"

  echo "$LOG_LINE"
  echo "$LOG_LINE" >> "$LOG_FILE"

  sleep 10
done
```

运行监控：

```bash
cd ~/sei-chain
chmod +x monitor.sh
./monitor.sh
```

---

## 快速参考

### 环境变量

```bash
# 在每台服务器上设置对应的环境变量
export CHAIN_ID="sei-testnet"
export VALIDATOR_NAME="validator0"  # validator0, validator1, validator2, validator3
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

### 常用命令

```bash
# 查看节点状态
curl http://localhost:26657/status | jq

# 查看 peers 数量
curl http://localhost:26657/net_info | jq '.result.n_peers'

# 查看区块高度
curl http://localhost:26657/block | jq '.result.block.header.height'

# 查看验证者列表
seid query staking validators --output json | jq '.validators[] | {moniker, status}'

# 查看账户余额
seid query bank balances $(seid keys show $VALIDATOR_NAME -a)

# 查看日志
tail -f build/generated/logs/seid.log

# 停止节点
kill $(cat build/generated/seid.pid)

# 重启节点
./poc-deploy/localnode/multi-node-scripts/step5_start_sei.sh
```

### 脚本位置

所有多节点部署脚本位于：`poc-deploy/localnode/multi-node-scripts/`

- `step0_build.sh` - 编译程序
- `step1_configure_init.sh` - 初始化节点
- `step2_genesis.sh` - 准备 genesis 参数
- `step3_add_validator_to_genesis.sh` - 添加验证者到 genesis（仅协调节点）
- `step4_config_override.sh` - 应用配置
- `step5_start_sei.sh` - 启动节点

---

## 总结

本文档描述了如何在 4 台独立服务器上部署 Sei Chain 测试网络。关键步骤包括：

1. ✅ 所有节点编译程序（step0）
2. ✅ 所有节点初始化配置（step1）
3. ✅ 所有节点准备 genesis 参数（step2）
4. ✅ 收集 gentx 到协调节点（手动）
5. ✅ 协调节点生成 genesis.json（step3）
6. ✅ 分发 genesis.json 到所有节点（手动）
7. ✅ 所有节点应用配置并设置 persistent_peers（step4）
8. ✅ 所有节点启动（step5）

**重要提示**：
- 至少需要 3 个验证者节点在线才能开始出块
- 确保所有节点的 genesis hash 一致
- 正确配置 persistent_peers
- 备份验证者密钥文件

如有问题，请参考故障排查部分或查看日志文件。

