## MODIFIED Requirements

### Requirement: 本地节点部署配置
本地开发环境必须（MUST）正确配置双代币架构参数。

#### Scenario: 单节点快速启动
- **GIVEN** 开发者执行本地节点启动命令
- **WHEN** 启动完成
- **THEN** 节点应运行在本地端口（默认 26657）
- **AND** 所有核心模块应正常加载
- **AND** 质押代币配置为 `ustaex`

#### Scenario: 多节点集群启动
- **GIVEN** 开发者需要测试多节点场景
- **WHEN** 启动多节点集群（默认3个节点）
- **THEN** 所有节点应正常启动并互相连接
- **AND** 共识应正常运行
- **AND** 质押代币配置为 `ustaex`

#### Scenario: Genesis 文件生成
- **GIVEN** 开发者初始化新链
- **WHEN** 生成 genesis 文件
- **THEN** genesis 应包含正确的 chain-id（aesc-local-*）
- **AND** 应包含初始账户配置
- **AND** `staking.params.bond_denom` 应为 `ustaex`
- **AND** Gas 代币 `mint.params.mint_denom` 应为 `uaex`

#### Scenario: 配置文件模板
- **GIVEN** 开发者需要自定义配置
- **WHEN** 查看配置模板
- **THEN** 应提供完整的配置项说明
- **AND** 质押相关配置使用 `ustaex`
- **AND** Gas 相关配置使用 `uaex`

