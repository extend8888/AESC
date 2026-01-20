# 变更：为 AESC Chain 集成 x402 支付协议服务端

## 为什么

x402 是 Coinbase 开发的开放支付协议，使用 HTTP 402 状态码实现即时、自动的稳定币支付。通过在 AESC Chain 中集成 x402 服务端，可以：

1. **支持 API 付费访问**：让链上服务能够通过 x402 协议接受支付
2. **AI Agent 经济**：支持 AI 代理自主支付访问 API 和服务
3. **微支付场景**：支持按请求付费的微交易模式
4. **生态互操作**：与 Coinbase 及 x402 生态系统对接

## 变更内容

### 1. x402 服务端核心模块
- 实现 x402 协议的服务端 HTTP 中间件
- 支持 HTTP 402 响应和支付验证
- 支持 USDC 等稳定币支付（基于 AESC 的 EVM 兼容性）
- 支持 CAIP-2 网络标识符

### 2. 配置系统
- 添加 `[x402]` 配置段到 app.toml
- 支持通过配置启用/禁用 x402 服务
- 配置项：端口、钱包地址、Facilitator 地址等

### 3. 链启动集成
- **可选启动**：链启动时可选择是否启用 x402 服务
- 集成到 seid start 命令的服务启动流程
- 类似 EVM RPC 的启用模式

### 4. Facilitator 集成
- 支持连接 Coinbase CDP Facilitator
- 预留自建 Facilitator 的扩展点

## 影响

- **受影响的规格**：新增 `x402-server` 能力
- **受影响的代码**：
  - `services/x402/` - 新增 x402 服务实现
  - `cmd/seid/` - 启动命令集成
  - `config.yml` - 配置模板更新
  - `app/app.go` - 服务注册

## 技术选型

### 语言选择：Go

选择 Go 而非 TypeScript 的原因：
1. **一致性**：AESC Chain 核心代码库为 Go，保持技术栈一致
2. **集成便利**：直接调用 EVM keeper 进行链上操作
3. **性能**：Go 服务与链节点同进程运行，减少 RPC 开销
4. **官方支持**：Coinbase x402 提供 Go SDK

### 架构模式

参照 `evmrpc/` 的设计模式：
- Config 结构体定义配置项
- TOML 配置模板
- 可选启用的 HTTP 服务
- 与链生命周期集成

## 里程碑

1. **阶段一**：基础设施（配置、启动集成）
2. **阶段二**：x402 协议核心（402 响应、支付验证）
3. **阶段三**：Facilitator 集成
4. **阶段四**：测试与文档

## 参考资料

- [x402 协议文档](https://docs.cdp.coinbase.com/x402/welcome)
- [x402 GitHub 仓库](https://github.com/coinbase/x402)
- [x402 生态系统](https://www.x402.org/)

