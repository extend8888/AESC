# POC Single Node Deployment

这是一个简化的单节点验证人部署方案，基于 `docker/localnode` 但移除了 Docker 和多节点支持。

## 目录结构

```
poc-deploy/localnode/
├── config/           # 配置文件
│   ├── app.toml
│   └── config.toml
├── scripts/          # 部署脚本
│   ├── deploy.sh           # 主部署脚本
│   ├── deploy-debug.sh     # Debug 模式部署
│   ├── step0_build.sh      # 构建 seid
│   ├── step1_configure_init.sh  # 初始化节点
│   ├── step2_genesis.sh    # 准备 genesis
│   ├── add_validator_to_genesis.sh  # 添加 validators 到 genesis
│   ├── step3_config_override.sh # 配置覆盖
│   ├── step4_start_sei.sh  # 启动节点
│   ├── stop.sh             # 停止节点
│   ├── clean.sh            # 清理数据
│   ├── logs.sh             # 查看日志
│   ├── test.sh             # 测试部署
│   └── verify_genesis.sh   # 验证 genesis 文件
└── README.md         # 本文件
```

## 快速开始

### 1. 完整部署（从头开始）

```bash
# 赋予执行权限
chmod +x poc-deploy/localnode/scripts/*.sh

# 运行部署
./poc-deploy/localnode/scripts/deploy.sh

# 如果遇到问题，使用 debug 模式
./poc-deploy/localnode/scripts/deploy-debug.sh
```

这将执行以下步骤：
- **Step 0**: 构建 seid 二进制文件
- **Step 1**: 初始化节点和创建验证人账户
  - 初始化链配置（chain-id: aesc-poc）
  - 创建 validator 账户（余额: 10 UAEX）
  - 生成 gentx（质押: 10 UAEX，voting power: 10）
- **Step 2**: 准备 genesis 文件
  - 配置 genesis 参数（staking、oracle、gov 等）
  - 添加账户到 genesis
  - **手动添加 validators 到顶层 `.validators` 数组**（使用 bc 避免整数溢出）
  - 收集 gentxs 到 `.app_state.genutil.gen_txs`
- **Step 3**: 应用配置覆盖
- **Step 4**: 启动节点

### 2. 查看日志

```bash
./poc-deploy/localnode/scripts/logs.sh
```

或直接查看日志文件：
```bash
tail -f build/generated/logs/seid.log
```

### 3. 停止节点

```bash
./poc-deploy/localnode/scripts/stop.sh
```

### 4. 清理所有数据

```bash
./poc-deploy/localnode/scripts/clean.sh
```

### 5. 验证 Genesis 文件

```bash
# 验证当前的 genesis.json
./poc-deploy/localnode/scripts/verify_genesis.sh

# 或指定文件路径
./poc-deploy/localnode/scripts/verify_genesis.sh build/generated/genesis.json
```

这将：
- 停止运行中的节点
- 删除 `build/generated` 目录
- 删除 `~/.sei` 目录

## 配置说明

### 关键参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| **Chain ID** | `aesc-poc` | 链标识符 |
| **Moniker** | `aesc-node-poc` | 节点名称 |
| **初始余额** | `10000000uaex` | 10 UAEX（与 docker/localnode 一致）|
| **质押金额** | `10000000uaex` | 10 UAEX |
| **Voting Power** | `10` | delegation / 1,000,000 |
| **测试账户** | `5` | 自动创建的测试账户数量 |

### 环境变量

可以通过环境变量自定义部署：

```bash
# 设置链 ID（默认: aesc-poc）
export CHAIN_ID=my-chain

# 设置节点名称（默认: aesc-node-poc）
export MONIKER=my-node

# 设置测试账户数量（默认: 5）
export NUM_ACCOUNTS=10

# 启用 mock balances（默认: false）
export MOCK_BALANCES=true

# 运行部署
./poc-deploy/localnode/scripts/deploy.sh
```

## 账户信息

部署完成后，会创建以下账户：

1. **validator** - 验证人账户
   - 密码: `12345678`
   - 初始余额: `10000000uaex` (10 UAEX) + `10000000uusdc` + `10000000uatom`
   - 质押金额: `10000000uaex` (10 UAEX)
   - Voting Power: `10`

2. **admin** - 管理员账户
   - 密码: `12345678`
   - 初始余额: `1000000000000000000000uaex` + `1000000000000000000000uusdc` + `1000000000000000000000uatom`

3. **测试账户** - 由 `populate_genesis_accounts.py` 创建
   - 数量: 由 `NUM_ACCOUNTS` 环境变量控制（默认 5 个）
   - 每个账户余额: `1000000000000000000000uaex` + uusdc + uatom

## 查看账户

```bash
# 列出所有账户
seid keys list

# 查看特定账户
printf "12345678\n" | seid keys show validator
printf "12345678\n" | seid keys show admin
```

## 与节点交互

```bash
# 查询账户余额
seid query bank balances $(seid keys show validator -a)

# 查询验证人信息
seid query staking validators

# 发送交易
seid tx bank send validator <recipient-address> 1000uaex --chain-id aesc-poc --fees 1000uaex
```

## 生成的文件

部署过程会在 `build/generated/` 目录下生成以下文件：

```
build/generated/
├── genesis.json              # Genesis 文件
├── gentx/                    # Genesis 交易
├── exported_keys/            # 导出的验证人密钥
├── genesis_accounts.txt      # Genesis 账户列表
├── logs/
│   └── seid.log             # 节点日志
├── node_data/
│   └── snapshots/           # 快照目录
└── seid.pid                 # 进程 PID
```

## 故障排查

### 使用 Debug 模式

```bash
./poc-deploy/localnode/scripts/deploy-debug.sh
```

Debug 模式会显示：
- 每个执行的命令
- 所有环境变量
- 每一步的验证结果
- Genesis 文件的详细信息

### 验证 Genesis 文件

```bash
./poc-deploy/localnode/scripts/verify_genesis.sh
```

这会检查：
- Validators 数组结构
- Gen_txs 数组内容
- PubKey、Power、Address 是否匹配

### 完全重置

```bash
./poc-deploy/localnode/scripts/clean.sh
./poc-deploy/localnode/scripts/deploy.sh
```

## 关键技术修复

### 1. Bash 整数溢出问题

**问题**: Bash 的 `$((...))` 运算符无法处理超过 64 位的大整数，导致 voting power 计算错误。

**示例**:
```bash
# 错误的计算（Bash 溢出）
DELEGATION=100000000000000000000000
POWER=$(($DELEGATION / 1000000))  # 结果: 200376420520 (错误!)

# 正确的计算（使用 bc）
POWER=$(echo "$DELEGATION / 1000000" | bc)  # 结果: 100000000000000000 (正确)
```

**修复**: 在 `add_validator_to_genesis.sh` 中使用 `bc` 命令进行大数字除法。

### 2. Genesis Validators 结构

**问题**: `genesisValidators[0] != req.Validators[0]` 错误。

**原因**:
- `.validators` 数组需要在 `seid collect-gentxs` **之前**手动添加
- 必须包含正确的 `power`、`pub_key` 字段
- `seid collect-gentxs` 会自动补充 `address` 和 `name`

**解决方案**:
1. 使用 `add_validator_to_genesis.sh` 在 collect-gentxs 之前添加 validators
2. 只添加 `power` 和 `pub_key`，让 collect-gentxs 补充其他字段
3. 使用 `bc` 正确计算 power 值

### 3. 质押金额配置

**问题**: 初始使用过大的质押金额导致 power 值溢出。

**修复**: 改为与 `docker/localnode` 一致的配置：
- 初始余额: `10000000uaex` (10 UAEX)
- 质押金额: `10000000uaex` (10 UAEX)
- Voting Power: `10`

## 与 docker/localnode 的区别

1. **单节点** - 只运行一个验证人节点
2. **无 Docker** - 直接在本地运行
3. **无 Price Feeder** - 移除了 step6_start_price_feeder.sh
4. **简化配置** - 不需要处理多节点的 persistent peers
5. **使用 bc** - 避免大数字计算溢出

## 注意事项

- 所有账户密码默认为 `12345678`（仅用于测试）
- 这是一个 POC 环境，不适合生产使用
- 节点数据存储在 `~/.sei` 目录
- 使用 `clean.sh` 会删除所有数据，请谨慎操作

