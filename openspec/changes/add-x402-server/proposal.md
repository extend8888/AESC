# 变更：为 AESC Chain 集成 x402 支付协议服务端

## 为什么

x402 是 Coinbase 开发的开放支付协议，使用 HTTP 402 状态码实现即时、自动的稳定币支付。通过在 AESC Chain 中集成 x402 服务端，可以：

1. **支持 API 付费访问**：让链上服务能够通过 x402 协议接受支付
2. **AI Agent 经济**：支持 AI 代理自主支付访问 API 和服务
3. **微支付场景**：支持按请求付费的微交易模式
4. **生态互操作**：与 x402 生态系统对接

## 变更内容

### 1. x402-relayer 独立服务
- 实现 x402 协议的服务端 HTTP 中间件
- 支持 HTTP 402 响应和支付验证
- 支持 USDT 等带 EIP-3009 的 ERC20 代币支付
- 支持 CAIP-2 网络标识符（如 `eip155:71603`）

### 2. 配置系统
- 独立 TOML 配置文件 `config.toml`
- 配置项：端口、钱包地址、Token 合约地址、EVM RPC 等
- 支持配置 Token 的 EIP-712 域参数（name/version）
- 启动时校验链上 `DOMAIN_SEPARATOR()` 一致性

### 3. 独立二进制架构
- **独立部署**：`x402-relayer` 作为独立 Go 模块运行
- 通过 EVM RPC 与链节点通信
- 不集成到 seid 进程，便于独立扩展和运维

### 4. 本地 Verifier/Settler 架构
- **本地验证**：使用本地 EIP-712 签名验证
- **本地结算**：直接调用链上 `transferWithAuthorization` 方法
- **不依赖外部 Facilitator**：无需对接 Coinbase CDP Facilitator

## 影响

- **受影响的规格**：新增 `x402-server` 能力
- **受影响的代码**：
  - `x402-relayer/` - 独立 Go workspace 模块
  - `precompiles/usdt/` - USDT 预编译合约（EIP-3009 支持）

## 技术选型

### 语言选择：Go

选择 Go 而非 TypeScript 的原因：
1. **一致性**：AESC Chain 核心代码库为 Go，保持技术栈一致
2. **性能**：Go 原生并发支持，适合高吞吐场景
3. **官方支持**：Coinbase x402 提供 Go SDK

### 架构模式

独立服务模式：
- 独立 Go workspace 模块 (`github.com/sei-protocol/x402-relayer`)
- TOML 配置文件
- 通过 EVM RPC 与链节点通信
- 本地 verifier/settler 模块处理支付

## 里程碑

1. **阶段一**：基础设施（配置、类型定义）✅
2. **阶段二**：Facilitator 模块（验证、余额检查、结算）✅
3. **阶段三**：Relayer 模块（交易广播、Gas 估算）✅
4. **阶段四**：HTTP 服务（处理器、中间件）✅
5. **阶段五**：集成测试 ✅
6. **阶段六**：文档与部署 ✅

## 参考资料

- [x402 协议文档](https://docs.cdp.coinbase.com/x402/welcome)
- [x402 GitHub 仓库](https://github.com/coinbase/x402)
- [x402 生态系统](https://www.x402.org/)
- [EIP-3009: Transfer With Authorization](https://eips.ethereum.org/EIPS/eip-3009)

