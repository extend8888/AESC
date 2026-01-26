# x402-relayer 问题修复方案

> 本文档列出已发现的问题及修复方案，供 Review 后实施。

---

## 问题概览

| 优先级 | 问题 | 影响文件 |
|--------|------|----------|
| **高** | 规格要求外部 Facilitator，实现为本地 verifier/settler | spec.md, proposal.md, tasks.md, design.md |
| **中** | X-PAYMENT-REQUIRED 头使用 base64 编码 | payment.go |
| **中** | USDT 地址硬编码在 verifier/balance/settler | verifier.go, balance.go, settler.go, config.go |
| **中** | SignedTxHash 切片可能 panic | relay.go |
| **中** | EVM RPC 不可用时未返回 503 | middleware/payment.go, handler/relay.go |
| **低** | 文档/e2e 使用过时 Chain ID 713715 | README.md, e2e/*.go |
| **低** | 规格描述与独立二进制架构不一致 | spec.md |

---

## 修复方案

### 1. 【高】更新规格：移除外部 Facilitator 要求

**现状**：
- `spec.md` 要求调用 Coinbase CDP Facilitator 并在不可用时返回 503
- 实现为本地 verifier/settler，无外部依赖

**决策**：不对接 Coinbase CDP Facilitator，使用本地实现

**修复**：
1. 更新 `openspec/changes/add-x402-server/specs/x402-server/spec.md`:
   - 移除 "Facilitator 集成" 需求
   - 将 "Facilitator 不可用返回 503" 改为 "EVM RPC 不可用返回 503"

2. 更新 `openspec/changes/add-x402-server/proposal.md`:
   - 移除 "阶段三: Facilitator 集成"
   - 明确说明采用本地 verifier/settler 架构

3. 同步更新其他文档：
   - `tasks.md`：更新相关任务描述
   - `design.md`：更新架构描述
   - `x402-technical-analysis.md`：更新技术分析

---

### 2. 【中】修复 X-PAYMENT-REQUIRED 头编码

**现状**：`payment.go:154` 使用 base64 编码
```go
w.Header().Set(PaymentRequiredHeader, base64.StdEncoding.EncodeToString(reqJSON))
```

**规格要求**：`spec.md:25` 要求 JSON 格式
> **AND** 字段内容为 JSON 格式的支付指令

**修复**：
更新 `x402-relayer/middleware/payment.go`:
```go
// 直接使用 JSON，不再 base64 编码
w.Header().Set(PaymentRequiredHeader, string(reqJSON))
```

**注意**：需同步更新客户端解析逻辑和文档

---

### 3. 【中】使用配置的 Token 地址替代硬编码 + 启动时域分隔符校验

**现状**：
- `config.go` 提供 `USDTPrecompile` 配置项
- `verifier.go:15`、`balance.go:37`、`settler.go:48` 硬编码 `0x...1010`

**需求**：支持带 EIP-3009 的 ERC20 合约

**修复**：

#### 3.1 配置项变更

```go
// config.go
type Config struct {
    // ...existing fields...

    // TokenContract is the ERC20 contract address (must support EIP-3009)
    TokenContract string `mapstructure:"token_contract"`

    // TokenName is the EIP-712 domain name (default: "Tether USD")
    TokenName string `mapstructure:"token_name"`

    // TokenVersion is the EIP-712 domain version (default: "1")
    TokenVersion string `mapstructure:"token_version"`
}

// DefaultConfig - 保持 USDT 兼容
func DefaultConfig() *Config {
    return &Config{
        // ...
        TokenContract: DefaultUSDTPrecompile,  // 0x...1010
        TokenName:     "Tether USD",           // USDT 默认值
        TokenVersion:  "1",                    // USDT 默认值
    }
}
```

#### 3.2 配置别名（向后兼容）

```go
// ReadConfig - 使用完整 key 注册别名
func ReadConfig(v *viper.Viper) (*Config, error) {
    cfg := DefaultConfig()

    // 注册别名：新字段 → 旧字段
    v.RegisterAlias("x402-relayer.token_contract", "x402-relayer.usdt_precompile")

    if v.IsSet("x402-relayer") {
        if err := v.UnmarshalKey("x402-relayer", cfg); err != nil {
            return nil, err
        }
    }

    // 显式兜底：如果 TokenContract 为空，使用旧字段
    if cfg.TokenContract == "" {
        if old := v.GetString("x402-relayer.usdt_precompile"); old != "" {
            cfg.TokenContract = old
        }
    }

    return cfg, nil
}
```

#### 3.3 启动时域分隔符校验（关键）

```go
// server.go - 启动时校验
func (s *Server) validateDomainSeparator(ctx context.Context) error {
    // 1. 调用链上合约的 DOMAIN_SEPARATOR() 方法
    onChainDS, err := s.balanceChecker.GetDomainSeparator(ctx)
    if err != nil {
        return fmt.Errorf("contract does not implement DOMAIN_SEPARATOR(): %w", err)
    }

    // 2. 本地计算域分隔符
    localDS := s.verifier.ComputeDomainSeparator()

    // 3. 比对
    if onChainDS != localDS {
        return fmt.Errorf(
            "DOMAIN_SEPARATOR mismatch: on-chain=%s, local=%s. "+
            "Check token_name/token_version config",
            hex.EncodeToString(onChainDS[:]),
            hex.EncodeToString(localDS[:]),
        )
    }

    return nil
}

// Start - 启动前校验
func (s *Server) Start() error {
    // 校验域分隔符
    if err := s.validateDomainSeparator(context.Background()); err != nil {
        return fmt.Errorf("domain separator validation failed: %w", err)
    }

    // ...继续启动...
}
```

#### 3.4 BalanceChecker 添加 DOMAIN_SEPARATOR 查询

```go
// balance.go
func (bc *BalanceChecker) GetDomainSeparator(ctx context.Context) ([32]byte, error) {
    // 调用合约的 DOMAIN_SEPARATOR() view 方法
    data, err := bc.usdtABI.Pack("DOMAIN_SEPARATOR")
    if err != nil {
        return [32]byte{}, err
    }

    result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
        To:   &bc.tokenAddress,
        Data: data,
    }, nil)
    if err != nil {
        return [32]byte{}, err
    }

    var ds [32]byte
    copy(ds[:], result)
    return ds, nil
}
```

#### 3.5 更新各模块构造函数

```go
// verifier.go
func NewVerifier(chainID *big.Int, tokenAddr, tokenName, tokenVersion string) *Verifier

// balance.go
func NewBalanceChecker(rpcURL, tokenAddr string) (*BalanceChecker, error)

// settler.go
func NewSettler(rpcURL, privateKeyHex string, chainID *big.Int, tokenAddr string) (*Settler, error)
```

---

### 4. 【中】修复 SignedTxHash 切片越界

**现状**：`relay.go:77`
```go
SignedTxHash: req.SignedTx[:66], // First 66 chars as identifier
```
若 `req.SignedTx` 长度 < 66 会 panic

**修复**：
```go
signedTxHash := req.SignedTx
if len(signedTxHash) > 66 {
    signedTxHash = signedTxHash[:66]
}
// ...
SignedTxHash: signedTxHash,
```

---

### 5. 【中】EVM RPC 不可用时返回 503

**现状**：EVM RPC 错误会返回 402 或 500

**修复**：

#### 5.1 定义 RPC 错误类型

```go
// errors.go
type RPCUnavailableError struct {
    Err error
}

func (e *RPCUnavailableError) Error() string {
    return fmt.Sprintf("EVM RPC unavailable: %v", e.Err)
}

func IsRPCUnavailable(err error) bool {
    var rpcErr *RPCUnavailableError
    return errors.As(err, &rpcErr)
}
```

#### 5.2 更新 middleware/payment.go

```go
// payment.go - 余额检查失败时
balance, err := pm.balanceChecker.GetBalance(ctx, fromAddr)
if err != nil {
    if IsRPCUnavailable(err) {
        http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
        return
    }
    // 其他错误...
}
```

#### 5.3 更新 handler/relay.go

```go
// relay.go - 交易广播失败时
receipt, err := h.settler.Settle(ctx, payment.Payload)
if err != nil {
    if IsRPCUnavailable(err) {
        h.writeError(w, http.StatusServiceUnavailable, "EVM RPC unavailable")
        return
    }
    // 其他错误...
}
```

#### 5.4 包装 RPC 错误

```go
// balance.go, settler.go - 连接/超时错误包装为 RPCUnavailableError
client, err := ethclient.Dial(rpcURL)
if err != nil {
    return nil, &RPCUnavailableError{Err: err}
}
```

---

### 6. 【低】更新 Chain ID 到 AESC

**现状**：文档和 e2e 使用 `eip155:713715` / `ChainID=713715`

**AESC 当前配置**：`aesc-poc` → `71603`

**修复**：
1. `x402-relayer/README.md`:
   - 第 20 行: `eip155:713715` → `eip155:71603`
   - 第 94 行: `eip155:713715` → `eip155:71603`

2. `x402-relayer/e2e/helpers_test.go`:
   - 第 27 行: `ChainID = 713715` → `ChainID = 71603`

3. `x402-relayer/e2e/test_config.toml`:
   - 第 5 行: `eip155:713715` → `eip155:71603`

---

### 7. 【低】更新规格：明确独立二进制架构

**现状**：
- `spec.md:8` 描述 "链节点启动时未配置 x402" 暗示与节点集成
- 实际实现是独立二进制 `x402-relayer`

**修复**：
更新 `spec.md` 中的场景描述：
```markdown
#### Scenario: 默认禁用
- **WHEN** x402-relayer 服务未配置或未启动
- **THEN** x402 支付功能不可用
```

---

## 配置项变更汇总

| 旧名称 | 新名称 | 默认值 | 说明 |
|--------|--------|--------|------|
| `usdt_precompile` | `token_contract` | `0x...1010` | 支持任意 EIP-3009 ERC20 合约 |
| - | `token_name` | `"Tether USD"` | EIP-712 domain name |
| - | `token_version` | `"1"` | EIP-712 domain version |

**向后兼容**：
- `usdt_precompile` 作为 `token_contract` 的别名
- 默认值保持 USDT 兼容

---

## 任务清单

### 规格与文档更新
- [x] 1.1 更新 `specs/x402-server/spec.md`：移除外部 Facilitator 要求
- [x] 1.2 更新 `proposal.md`：修改架构描述
- [x] 1.3 更新 `tasks.md`：同步任务描述
- [x] 1.4 更新 `design.md`：同步架构描述
- [x] 1.5 更新 `x402-technical-analysis.md`：同步技术分析

### 代码修复
- [x] 2.1 修复 `middleware/payment.go`：X-PAYMENT-REQUIRED 改用 JSON
- [x] 2.2 更新 `config/config.go`：添加 token_name/token_version，配置别名
- [x] 2.3 更新 `facilitator/verifier.go`：使用配置参数
- [x] 2.4 更新 `facilitator/balance.go`：使用配置参数，添加 DOMAIN_SEPARATOR 查询
- [x] 2.5 更新 `facilitator/settler.go`：使用配置参数
- [x] 2.6 更新 `server.go`：启动时域分隔符校验
- [x] 2.7 修复 `handler/relay.go`：防止切片越界
- [x] 2.8 添加 503 错误映射：RPC 不可用时返回 503

### Chain ID 更新
- [x] 3.1 更新 `README.md`：Chain ID 改为 71603
- [x] 3.2 更新 `e2e/helpers_test.go`：Chain ID 改为 71603
- [x] 3.3 更新 `e2e/test_config.toml`：Chain ID 改为 71603

### 测试验证
- [x] 4.1 编写单元测试
  - `middleware/payment_test.go` - 16 测试通过 (IsRPCUnavailableError)
  - `config/config_test.go` - 21 测试通过 (配置读取、验证、兼容性)
- [x] 4.2 运行 Scenario 1 E2E 测试 (USDT Precompile) - 全部通过
- [x] 4.3 运行 Scenario 2 E2E 测试 (Custom EIP-3009 ERC20) - 全部通过

---

## 关键设计决策

### D1: 域分隔符校验策略

**决策**：启动时调用链上 `DOMAIN_SEPARATOR()` 方法进行校验

**原因**：
- 很多 ERC20 不提供 `version()` 方法
- 即使提供 `name()`/`version()`，域构造方式可能不同
- 直接比对 `DOMAIN_SEPARATOR()` 是最可靠的方式
- 如果合约不实现 `DOMAIN_SEPARATOR()`，拒绝启动

### D2: 配置兼容性策略

**决策**：使用 viper 别名 + 显式兜底

**实现**：
1. `v.RegisterAlias("x402-relayer.token_contract", "x402-relayer.usdt_precompile")`
2. `ReadConfig` 中显式检查：`TokenContract` 为空则使用 `usdt_precompile`
3. 默认值保持 USDT 兼容

---

## 测试方案

### 测试架构概述

x402-relayer 的测试分为三个层次：

```
┌─────────────────────────────────────────────────────────────┐
│                      E2E 测试 (集成)                         │
│  - 启动本地 seid 节点                                        │
│  - 启动 x402-relayer 服务                                    │
│  - 测试完整的支付→验证→结算→广播流程                          │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                     单元测试 (模块)                          │
│  - IsRPCUnavailableError 错误检测                           │
│  - 配置读取和兼容性                                          │
│  - EIP-712 签名验证                                          │
│  - 域分隔符计算                                              │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                    编译验证 (基础)                           │
│  - go build ./x402-relayer/...                              │
└─────────────────────────────────────────────────────────────┘
```

### 测试场景分析

#### 场景 1：使用 USDT 预编译合约（推荐）

**架构**：
```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   客户端      │────▶│  x402-relayer    │────▶│  seid 节点           │
│              │     │  (HTTP API)      │     │  (USDT 预编译)       │
└──────────────┘     └──────────────────┘     └─────────────────────┘
                                                      │
                                              ┌───────▼───────┐
                                              │ 预编译合约     │
                                              │ 0x...1010     │
                                              │ EIP-3009 ✓    │
                                              └───────────────┘
```

**优点**：
- USDT 预编译已实现完整的 EIP-3009
- 不需要额外的合约部署工具链
- 与生产环境一致

**测试流程**：
1. 启动本地 seid 节点
2. 给测试账户铸造 USDT（通过 bank 模块）
3. 启动 x402-relayer
4. 执行 E2E 测试

#### 场景 2：部署自定义 EIP-3009 ERC20（扩展）

**架构**：
```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   客户端      │────▶│  x402-relayer    │────▶│  seid 节点           │
│              │     │  (HTTP API)      │     │  (EVM)              │
└──────────────┘     └──────────────────┘     └─────────────────────┘
                                                      │
                                              ┌───────▼───────┐
                                              │ 自定义 ERC20   │
                                              │ 0x...动态地址  │
                                              │ EIP-3009 ✓    │
                                              └───────────────┘
```

**需要**：
- 编写 EIP-3009 兼容的 Solidity 合约
- 使用 Hardhat/Foundry 部署
- 配置 x402-relayer 使用新合约地址

**适用场景**：
- 验证 x402-relayer 对非 USDT 预编译的通用性
- 测试 `token_contract`/`token_name`/`token_version` 配置

### 单元测试清单

| 测试文件 | 测试内容 | 优先级 |
|----------|----------|--------|
| `middleware/payment_test.go` | `IsRPCUnavailableError` 错误检测 | 高 |
| `config/config_test.go` | 配置读取、别名兼容性、默认值 | 高 |
| `facilitator/verifier_test.go` | 域分隔符计算、签名验证 | 中 |
| `facilitator/balance_test.go` | 余额查询（mock RPC） | 中 |

### E2E 测试清单

| 测试用例 | 描述 | 状态 |
|----------|------|------|
| `TestHealthEndpoint` | 健康检查 | ✅ 已有 |
| `TestPaymentRequirementsEndpoint` | 支付要求查询 | ✅ 已有 |
| `TestRelayWithoutPayment` | 无支付头返回 402 | ✅ 已有 |
| `TestFullRelayWithPayment` | 完整支付流程 | ✅ 已有 |
| `TestDomainSeparatorValidation` | 域分隔符校验 | ⏳ 新增 |
| `TestRPCUnavailable503` | RPC 不可用返回 503 | ⏳ 新增 |

### 执行步骤

#### 步骤 1：编写单元测试

```bash
# 创建测试文件
x402-relayer/middleware/payment_test.go
x402-relayer/config/config_test.go
```

#### 步骤 2：修复 E2E 测试配置

```bash
# 更新 Chain ID
x402-relayer/e2e/run_test.sh  # 713715 → 71603
x402-relayer/e2e/e2e_test.go  # 更新 network 断言
```

#### 步骤 3：运行测试

```bash
# 单元测试
go test ./x402-relayer/... -v

# E2E 测试（需要本地节点）
cd x402-relayer/e2e && ./run_test.sh
```

---

**请 Review 后确认是否可以开始实施。**

