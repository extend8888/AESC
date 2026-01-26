## ADDED Requirements

### Requirement: x402 服务端配置

系统必须（SHALL）提供可配置的 x402-relayer 独立服务，支持通过配置文件启用或禁用。

#### Scenario: 默认禁用
- **WHEN** x402-relayer 服务未配置或未启动
- **THEN** x402 支付功能不可用
- **AND** 不占用任何端口

#### Scenario: 启用 x402 服务
- **WHEN** 配置 `x402-relayer.enabled = true` 并启动服务
- **THEN** x402-relayer 在配置的端口启动
- **AND** 记录启动日志

#### Scenario: 启动时域分隔符校验
- **WHEN** x402-relayer 启动
- **THEN** 调用链上 Token 合约的 `DOMAIN_SEPARATOR()` 方法
- **AND** 与本地计算的域分隔符比对
- **IF** 不一致
- **THEN** 拒绝启动并记录错误日志

### Requirement: HTTP 402 响应

系统必须（SHALL）能够返回符合 x402 协议的 HTTP 402 Payment Required 响应。

#### Scenario: 返回支付要求
- **WHEN** 客户端请求付费资源且未提供支付凭证
- **THEN** 返回 HTTP 402 状态码
- **AND** 响应头包含 `X-PAYMENT-REQUIRED` 字段
- **AND** 字段内容为 JSON 格式的支付指令

#### Scenario: 支付指令格式
- **WHEN** 生成 402 响应
- **THEN** 支付指令包含接收地址、金额、代币合约、网络 ID
- **AND** 网络 ID 使用 CAIP-2 格式 (如 `eip155:71603`)

### Requirement: 支付验证

系统必须（SHALL）能够验证客户端提交的 x402 支付签名。

#### Scenario: 有效支付
- **WHEN** 客户端提交有效的 EIP-712 签名支付
- **AND** 支付金额和接收地址正确
- **THEN** 请求被放行处理
- **AND** 返回请求的资源

#### Scenario: 无效签名
- **WHEN** 客户端提交无效的支付签名
- **THEN** 返回 HTTP 402 状态码
- **AND** 包含错误描述

### Requirement: 本地支付结算

系统必须（SHALL）使用本地 verifier/settler 模块验证和结算支付，通过调用链上 EIP-3009 合约完成转账。

#### Scenario: 支付结算
- **WHEN** 收到有效的支付签名
- **THEN** 调用链上 Token 合约的 `transferWithAuthorization` 方法
- **AND** 等待交易确认后放行请求

#### Scenario: EVM RPC 不可用
- **WHEN** EVM RPC 服务不可达或超时
- **THEN** 返回 HTTP 503 Service Unavailable
- **AND** 记录错误日志

### Requirement: 健康检查

系统必须（SHALL）提供健康检查端点。

#### Scenario: 健康检查成功
- **WHEN** 请求 `/health` 端点
- **AND** x402-relayer 服务正常运行
- **THEN** 返回 HTTP 200 OK

