# 品牌重塑验证清单

本文档提供详细的验证步骤，确保品牌重塑工作完整且正确。

## 自动化验证脚本

### 1. 搜索残留的 Sei 标识符

```bash
#!/bin/bash
# 文件名：check-sei-references.sh

echo "=== 检查代码库中的 Sei 标识符 ==="

echo -e "\n1. 检查 usei 引用（应该只在注释或第三方代码中）："
find . -type f -name "*.go" ! -path "*/vendor/*" ! -path "*/go-ethereum/*" \
  -exec grep -Hn "usei" {} \; | grep -v "// " | grep -v "/*" || echo "✓ 未发现问题"

echo -e "\n2. 检查 sei1 地址（应该只在注释中）："
find . -type f -name "*.go" ! -path "*/vendor/*" ! -path "*/go-ethereum/*" \
  -exec grep -Hn "sei1" {} \; | grep -v "// " | grep -v "/*" || echo "✓ 未发现问题"

echo -e "\n3. 检查配置文件中的 usei："
find . -type f \( -name "*.json" -o -name "*.toml" -o -name "*.yaml" \) \
  ! -path "*/vendor/*" -exec grep -Hn "usei" {} \; || echo "✓ 未发现问题"

echo -e "\n4. 检查脚本中的 usei："
find . -type f -name "*.sh" ! -path "*/vendor/*" \
  -exec grep -Hn "usei" {} \; || echo "✓ 未发现问题"

echo -e "\n5. 检查 Makefile 中的 sei 引用："
grep -Hn "sei" Makefile | grep -v "seid" | grep -v "# " || echo "✓ 未发现问题"

echo -e "\n=== 检查完成 ==="
```

### 2. 验证 AESC 标识符的正确使用

```bash
#!/bin/bash
# 文件名：verify-aesc-identifiers.sh

echo "=== 验证 AESC 标识符 ==="

echo -e "\n1. 验证 app/params/config.go："
if grep -q 'BaseCoinUnit.*=.*"uaex"' app/params/config.go && \
   grep -q 'HumanCoinUnit.*=.*"aex"' app/params/config.go && \
   grep -q 'Bech32PrefixAccAddr.*=.*"aesc"' app/params/config.go; then
    echo "✓ config.go 配置正确"
else
    echo "✗ config.go 配置有误"
    exit 1
fi

echo -e "\n2. 验证 x/evm/types/params.go："
if grep -q 'BaseDenom.*=.*"uaex"' x/evm/types/params.go 2>/dev/null || \
   grep -q 'BaseDenom.*=.*"uaex"' x/evm/keeper/params.go 2>/dev/null; then
    echo "✓ EVM 模块配置正确"
else
    echo "✗ EVM 模块配置有误"
    exit 1
fi

echo -e "\n3. 验证 Makefile 测试命令："
if grep -q "^test:" Makefile && \
   grep -q "^test-unit:" Makefile && \
   grep -q "^test-integration:" Makefile; then
    echo "✓ Makefile 测试命令已添加"
else
    echo "✗ Makefile 测试命令缺失"
    exit 1
fi

echo -e "\n=== 验证完成 ==="
```

### 3. 测试验证脚本

```bash
#!/bin/bash
# 文件名：run-tests.sh

echo "=== 运行测试验证 ==="

echo -e "\n1. 运行单元测试："
if make test-unit; then
    echo "✓ 单元测试通过"
else
    echo "✗ 单元测试失败"
    exit 1
fi

echo -e "\n2. 运行集成测试："
if make test-integration; then
    echo "✓ 集成测试通过"
else
    echo "✗ 集成测试失败"
    exit 1
fi

echo -e "\n=== 测试验证完成 ==="
```

## 手动验证清单

### 核心配置验证

- [ ] **app/params/config.go**
  - [ ] `BaseCoinUnit = "uaex"`
  - [ ] `HumanCoinUnit = "aex"`
  - [ ] `UaexExponent = 6`
  - [ ] `Bech32PrefixAccAddr = "aesc"`

- [ ] **x/evm/types/params.go 或 x/evm/keeper/params.go**
  - [ ] `BaseDenom = "uaex"`

- [ ] **cmd/seid/cmd/root.go**
  - [ ] `MinGasPrices = "0.02uaex"`

### Makefile 验证

- [ ] **测试命令存在**
  - [ ] `make test` 命令存在
  - [ ] `make test-unit` 命令存在
  - [ ] `make test-integration` 命令存在

- [ ] **测试命令功能**
  - [ ] `make test` 运行所有测试
  - [ ] `make test-unit` 只运行单元测试
  - [ ] `make test-integration` 只运行集成测试

### 代码文件验证

- [ ] **核心模块（x/evm/）**
  - [ ] 所有 `.go` 文件使用 `uaex`
  - [ ] 测试文件使用 `aesc1` 地址
  - [ ] 不存在 `usei` 引用（除注释）

- [ ] **其他模块（x/mint/, x/oracle/ 等）**
  - [ ] 代币面额使用 `uaex`
  - [ ] 测试地址使用 `aesc1` 格式

- [ ] **应用层（app/）**
  - [ ] 测试文件更新
  - [ ] 示例代码更新

### 配置文件验证

- [ ] **Genesis 配置**
  - [ ] `poc-deploy/localnode/exmaple_genesis.json`
    - [ ] `denom` 字段为 `uaex`
    - [ ] 地址以 `aesc1` 开头
    - [ ] `denom_metadata` 正确配置

- [ ] **部署脚本**
  - [ ] `poc-deploy/localnode/scripts/` 中的脚本
  - [ ] `depoly-scripts/` 中的脚本
  - [ ] 环境变量使用 `uaex`

- [ ] **Docker 配置**
  - [ ] `docker/localnode/` 配置
  - [ ] `docker/rpcnode/` 配置
  - [ ] docker-compose.yml 环境变量

### 测试文件验证

- [ ] **单元测试**
  - [ ] `x/evm/*_test.go` 使用正确标识符
  - [ ] `x/mint/*_test.go` 使用正确标识符
  - [ ] `x/oracle/*_test.go` 使用正确标识符
  - [ ] `app/*_test.go` 使用正确标识符

- [ ] **集成测试**
  - [ ] `integration_test/` 下所有测试
  - [ ] 测试配置使用 `aesc` 前缀
  - [ ] 测试金额使用 `uaex` 面额

- [ ] **测试工具**
  - [ ] `testutil/` 辅助函数更新
  - [ ] 地址生成函数使用 `aesc` 前缀
  - [ ] 余额设置函数使用 `uaex` 面额

### 功能测试验证

- [ ] **本地节点启动**
  - [ ] 可以成功启动单节点
  - [ ] 节点使用正确的配置
  - [ ] 日志中显示正确的标识符

- [ ] **地址生成**
  - [ ] 可以生成 `aesc1` 格式的地址
  - [ ] 地址验证功能正常
  - [ ] 拒绝 `sei1` 格式的地址

- [ ] **代币操作**
  - [ ] 可以使用 `uaex` 进行转账
  - [ ] 余额查询返回 `uaex` 面额
  - [ ] Gas 费用以 `uaex` 计价

- [ ] **EVM 功能**
  - [ ] EVM RPC 端点正常工作
  - [ ] eth_getBalance 返回正确余额
  - [ ] EVM 交易使用 `uaex` 支付 Gas

- [ ] **多节点集群**
  - [ ] 可以启动 4 节点集群
  - [ ] 节点间共识正常
  - [ ] 跨节点交易正常

### 文档验证

- [ ] **技术文档迁移**
  - [ ] `tmp/AESC 链开发技术方案.md` 已整合到 design.md
  - [ ] design.md 中的 Sei 引用已更新
  - [ ] 添加了文档来源说明

- [ ] **README 更新**
  - [ ] 示例地址使用 `aesc1` 格式
  - [ ] 示例金额使用 `uaex` 面额
  - [ ] 不包含过时的 Sei 信息

- [ ] **部署文档**
  - [ ] `poc-deploy/README.md` 更新
  - [ ] 示例命令正确
  - [ ] 配置示例正确

### 清理验证

- [ ] **搜索验证**
  - [ ] 搜索 `usei` 只返回必要的引用
  - [ ] 搜索 `sei1` 只返回必要的引用
  - [ ] 搜索 `"sei"` 只返回 `seid` 和必要的引用

- [ ] **第三方代码**
  - [ ] vendor/ 目录未修改
  - [ ] go-ethereum/ 目录未修改
  - [ ] sei-cosmos/ 只修改了测试和配置

## 回归测试清单

### 基础功能测试

```bash
# 1. 构建二进制
make install

# 2. 初始化节点
seid init test-node --chain-id aesc-testnet-1

# 3. 创建测试账户
seid keys add test-account

# 4. 验证地址格式
# 地址应以 aesc1 开头

# 5. 启动节点
seid start

# 6. 查询余额
seid query bank balances $(seid keys show test-account -a)

# 7. 发送交易
seid tx bank send <from> <to> 1000000uaex --fees 1000uaex

# 8. 验证交易
seid query tx <tx-hash>
```

### 集成测试

```bash
# 1. 启动本地测试网络
cd poc-deploy/localnode
./scripts/start.sh

# 2. 等待节点就绪
sleep 10

# 3. 执行测试交易
# （使用预配置的测试账户）

# 4. 验证结果
# 检查余额、交易记录等

# 5. 停止网络
./scripts/stop.sh
```

### Docker 集群测试

```bash
# 1. 启动 4 节点集群
make docker-cluster-start

# 2. 等待集群就绪
sleep 30

# 3. 检查节点状态
docker ps

# 4. 执行跨节点交易测试

# 5. 停止集群
make docker-cluster-stop
```

## 性能基准验证

- [ ] **启动时间**
  - [ ] 单节点启动时间 < 10 秒
  - [ ] 4 节点集群启动时间 < 30 秒

- [ ] **测试执行时间**
  - [ ] 单元测试 < 10 分钟
  - [ ] 集成测试 < 30 分钟

- [ ] **交易性能**
  - [ ] 交易确认时间 < 1 秒
  - [ ] TPS 满足基本要求

## 最终检查

- [ ] 所有自动化验证脚本通过
- [ ] 所有手动验证项目完成
- [ ] 所有测试通过
- [ ] 文档更新完成
- [ ] 没有遗留的 Sei 标识符（除必要的注释）
- [ ] 功能测试全部通过
- [ ] 性能基准满足要求

## 批准标准

提案可以批准实施的条件：
1. ✅ 所有验证清单项目完成
2. ✅ 至少一位技术负责人审查通过
3. ✅ 没有阻塞性问题
4. ✅ 回滚方案已准备
5. ✅ 文档完整且准确

