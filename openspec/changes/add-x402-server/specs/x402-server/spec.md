## ADDED Requirements

### Requirement: x402 服务端配置

系统必须（SHALL）提供可配置的 x402 服务端，支持通过配置文件启用或禁用。

#### Scenario: 默认禁用
- **WHEN** 链节点启动时未配置 x402
- **THEN** x402 服务不启动
- **AND** 不占用任何端口

#### Scenario: 启用 x402 服务
- **WHEN** 配置 `x402.enabled = true`
- **THEN** x402 服务在配置的端口启动
- **AND** 记录启动日志

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
- **AND** 网络 ID 使用 CAIP-2 格式 (如 `eip155:CHAIN_ID`)

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

### Requirement: Facilitator 集成

系统必须（SHALL）支持通过 Coinbase CDP Facilitator 验证和结算支付。

#### Scenario: 调用 Facilitator 验证
- **WHEN** 收到支付签名
- **THEN** 调用配置的 Facilitator URL 进行验证
- **AND** 等待验证结果后再放行请求

#### Scenario: Facilitator 不可用
- **WHEN** Facilitator 服务不可达
- **THEN** 返回 HTTP 503 Service Unavailable
- **AND** 记录错误日志

### Requirement: 健康检查

系统必须（SHALL）提供健康检查端点。

#### Scenario: 健康检查成功
- **WHEN** 请求 `/health` 端点
- **AND** x402 服务正常运行
- **THEN** 返回 HTTP 200 OK

