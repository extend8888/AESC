# 链标识规格增量

## MODIFIED Requirements

### Requirement: 链基础标识配置
系统必须（SHALL）使用 AESC 品牌标识作为链的基础配置，包括代币面额、地址前缀和链名称。

#### Scenario: 代币面额配置正确
- **GIVEN** 系统已初始化
- **WHEN** 查询链的基础代币配置
- **THEN** 基础代币单位应为 `uaex`
- **AND** 标准代币单位应为 `aex`
- **AND** 代币精度应为 6（1 aex = 10^6 uaex）

#### Scenario: 地址前缀配置正确
- **GIVEN** 系统已初始化
- **WHEN** 生成新的账户地址
- **THEN** 地址应以 `aesc1` 开头
- **AND** 验证者地址应以 `aescvaloper` 开头
- **AND** 共识地址应以 `aescvalcons` 开头

#### Scenario: Bech32 地址验证
- **GIVEN** 一个 `aesc1` 开头的地址
- **WHEN** 系统验证该地址
- **THEN** 验证应该通过
- **AND** 地址长度应符合 Bech32 规范

#### Scenario: 拒绝旧的 Sei 地址格式
- **GIVEN** 一个 `sei1` 开头的地址
- **WHEN** 系统验证该地址
- **THEN** 验证应该失败
- **AND** 返回地址前缀不匹配的错误

### Requirement: Genesis 配置
系统必须（SHALL）在 genesis 配置中使用正确的 AESC 标识符。

#### Scenario: Genesis 代币配置
- **GIVEN** 一个新的 genesis 配置文件
- **WHEN** 检查代币元数据配置
- **THEN** base denom 应为 `uaex`
- **AND** display denom 应为 `aex`
- **AND** 代币名称应为 "AESC Gas Token"
- **AND** 代币符号应为 "AEX"

#### Scenario: Genesis 账户地址
- **GIVEN** genesis 配置中的账户列表
- **WHEN** 检查所有账户地址
- **THEN** 所有地址都应以 `aesc1` 开头
- **AND** 不应存在 `sei1` 开头的地址

#### Scenario: Genesis 余额配置
- **GIVEN** genesis 配置中的余额列表
- **WHEN** 检查余额的代币面额
- **THEN** 所有余额都应使用 `uaex` 作为面额
- **AND** 不应存在 `usei` 面额的余额

## ADDED Requirements

### Requirement: 最小 Gas 价格配置
系统必须（SHALL）使用 `uaex` 作为最小 Gas 价格的面额单位。

#### Scenario: 默认最小 Gas 价格
- **GIVEN** 节点使用默认配置启动
- **WHEN** 查询最小 Gas 价格配置
- **THEN** 最小 Gas 价格应为 `0.02uaex`

#### Scenario: 自定义最小 Gas 价格
- **GIVEN** 节点配置了自定义最小 Gas 价格
- **WHEN** 配置值为 `0.05uaex`
- **THEN** 系统应接受该配置
- **AND** 交易的 Gas 价格必须不低于此值

#### Scenario: 拒绝错误的面额
- **GIVEN** 节点配置了最小 Gas 价格
- **WHEN** 配置值使用 `usei` 面额
- **THEN** 系统应拒绝该配置
- **OR** 在启动时显示警告

### Requirement: EVM 模块代币配置
系统必须（SHALL）在 EVM 模块中使用 `uaex` 作为基础代币。

#### Scenario: EVM 交易 Gas 费用
- **GIVEN** 用户发起一笔 EVM 交易
- **WHEN** 计算 Gas 费用
- **THEN** Gas 费用应以 `uaex` 计价
- **AND** 从用户账户扣除的代币应为 `uaex`

#### Scenario: EVM 余额查询
- **GIVEN** 用户通过 EVM RPC 查询余额
- **WHEN** 调用 eth_getBalance
- **THEN** 返回的余额应对应 `uaex` 数量
- **AND** 转换率应为 1 wei = 1 uaex

#### Scenario: EVM 代币转账
- **GIVEN** 用户通过 EVM 发起代币转账
- **WHEN** 转账 1 ETH（在 AESC 中对应 1 AEX）
- **THEN** 实际转账的代币应为 1000000 uaex
- **AND** 接收方余额增加 1000000 uaex

