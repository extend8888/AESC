# 测试基础设施规格增量

## ADDED Requirements

### Requirement: Makefile 测试命令
系统必须（SHALL）在 Makefile 中提供便捷的测试命令。

#### Scenario: 运行所有测试
- **GIVEN** 开发者在项目根目录
- **WHEN** 执行 `make test`
- **THEN** 系统应运行所有单元测试和集成测试
- **AND** 测试结果应清晰显示通过/失败状态
- **AND** 失败的测试应显示详细错误信息

#### Scenario: 只运行单元测试
- **GIVEN** 开发者需要快速验证代码修改
- **WHEN** 执行 `make test-unit`
- **THEN** 系统应只运行单元测试
- **AND** 不应运行集成测试
- **AND** 测试应在合理时间内完成（< 10 分钟）

#### Scenario: 只运行集成测试
- **GIVEN** 开发者需要验证端到端功能
- **WHEN** 执行 `make test-integration`
- **THEN** 系统应只运行集成测试
- **AND** 不应运行单元测试
- **AND** 测试应包含所有集成测试场景

#### Scenario: 测试命令错误处理
- **GIVEN** 某些测试失败
- **WHEN** 执行 `make test`
- **THEN** 命令应返回非零退出码
- **AND** 应显示失败测试的详细信息
- **AND** 不应继续执行后续步骤

### Requirement: 测试使用正确的标识符
系统的所有测试必须（SHALL）使用 AESC 品牌标识符。

#### Scenario: 测试地址格式
- **GIVEN** 测试代码需要生成测试地址
- **WHEN** 生成新的测试地址
- **THEN** 地址应以 `aesc1` 开头
- **AND** 不应使用 `sei1` 格式的地址

#### Scenario: 测试代币面额
- **GIVEN** 测试代码需要指定代币数量
- **WHEN** 创建测试交易或余额
- **THEN** 应使用 `uaex` 作为面额
- **AND** 不应使用 `usei` 面额

#### Scenario: 测试断言验证
- **GIVEN** 测试需要验证交易结果
- **WHEN** 检查交易中的代币面额
- **THEN** 断言应检查 `uaex` 面额
- **AND** 不应检查 `usei` 面额

#### Scenario: 集成测试配置
- **GIVEN** 集成测试需要启动测试节点
- **WHEN** 初始化测试节点配置
- **THEN** 配置应使用 `aesc` 地址前缀
- **AND** 配置应使用 `uaex` 代币面额
- **AND** Genesis 配置应使用正确的标识符

### Requirement: 测试工具和辅助函数
测试工具必须（SHALL）提供使用 AESC 标识符的辅助函数。

#### Scenario: 生成测试账户
- **GIVEN** 测试需要创建测试账户
- **WHEN** 调用账户生成辅助函数
- **THEN** 生成的账户地址应以 `aesc1` 开头
- **AND** 账户应有正确的 Bech32 格式

#### Scenario: 创建测试余额
- **GIVEN** 测试需要为账户设置初始余额
- **WHEN** 调用余额设置辅助函数
- **THEN** 余额应使用 `uaex` 面额
- **AND** 金额应正确转换（如 1 aex = 1000000 uaex）

#### Scenario: 验证测试交易
- **GIVEN** 测试需要验证交易执行结果
- **WHEN** 调用交易验证辅助函数
- **THEN** 验证应检查 `uaex` 面额的 Gas 费用
- **AND** 验证应检查 `aesc1` 格式的地址

## MODIFIED Requirements

### Requirement: 单元测试覆盖率
系统的单元测试必须（SHALL）覆盖所有核心功能，并使用正确的 AESC 标识符。

#### Scenario: 核心模块测试
- **GIVEN** 核心模块（x/evm, x/mint, x/oracle 等）
- **WHEN** 运行单元测试
- **THEN** 所有测试应通过
- **AND** 测试应使用 `uaex` 和 `aesc1` 标识符
- **AND** 不应有使用旧 Sei 标识符的测试

#### Scenario: 地址验证测试
- **GIVEN** 地址验证功能的测试
- **WHEN** 测试有效地址验证
- **THEN** 应测试 `aesc1` 格式的地址通过验证
- **AND** 应测试 `sei1` 格式的地址被拒绝

#### Scenario: 代币操作测试
- **GIVEN** 代币转账和余额查询的测试
- **WHEN** 执行测试
- **THEN** 所有代币操作应使用 `uaex` 面额
- **AND** 测试断言应验证正确的面额

### Requirement: 集成测试场景
系统的集成测试必须（SHALL）验证使用 AESC 标识符的端到端场景。

#### Scenario: 节点启动测试
- **GIVEN** 集成测试启动测试节点
- **WHEN** 节点初始化完成
- **THEN** 节点应使用 `aesc` 地址前缀
- **AND** 节点应使用 `uaex` 作为基础代币
- **AND** Genesis 配置应正确

#### Scenario: 交易执行测试
- **GIVEN** 集成测试执行交易
- **WHEN** 发送代币转账交易
- **THEN** 交易应使用 `aesc1` 格式的地址
- **AND** 交易应使用 `uaex` 面额
- **AND** Gas 费用应以 `uaex` 计价

#### Scenario: 跨模块交互测试
- **GIVEN** 集成测试验证模块间交互
- **WHEN** 测试 EVM 和 Cosmos 模块的交互
- **THEN** 两个模块应使用一致的标识符
- **AND** 代币转换应正确（EVM wei ↔ Cosmos uaex）

