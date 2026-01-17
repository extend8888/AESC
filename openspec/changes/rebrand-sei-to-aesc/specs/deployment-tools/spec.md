# 部署工具规格增量

## MODIFIED Requirements

### Requirement: 本地节点部署配置
本地节点部署工具必须（SHALL）使用 AESC 品牌标识符。

#### Scenario: 单节点快速启动
- **GIVEN** 开发者使用 poc-deploy 启动本地节点
- **WHEN** 执行启动脚本
- **THEN** 节点应使用 `aesc` 地址前缀
- **AND** 节点应使用 `uaex` 作为基础代币
- **AND** Genesis 配置应包含正确的代币元数据

#### Scenario: 多节点集群启动
- **GIVEN** 开发者启动 4 节点测试集群
- **WHEN** 使用 docker-compose 启动集群
- **THEN** 所有节点应使用一致的 AESC 标识符
- **AND** 节点间通信应正常
- **AND** 共识应正常工作

#### Scenario: Genesis 文件生成
- **GIVEN** 部署脚本生成 genesis 文件
- **WHEN** 检查生成的 genesis.json
- **THEN** 代币面额应为 `uaex`
- **AND** 账户地址应以 `aesc1` 开头
- **AND** 验证者地址应以 `aescvaloper` 开头

#### Scenario: 配置文件模板
- **GIVEN** 部署工具使用配置文件模板
- **WHEN** 生成节点配置
- **THEN** 最小 Gas 价格应为 `0.02uaex`
- **AND** 不应包含 `usei` 引用

### Requirement: Docker 容器配置
Docker 部署配置必须（SHALL）使用正确的 AESC 标识符。

#### Scenario: Docker 镜像构建
- **GIVEN** 使用 Dockerfile 构建节点镜像
- **WHEN** 构建完成
- **THEN** 镜像应包含正确配置的二进制
- **AND** 默认配置应使用 AESC 标识符

#### Scenario: Docker Compose 环境变量
- **GIVEN** docker-compose.yml 配置文件
- **WHEN** 检查环境变量
- **THEN** 不应包含 `usei` 相关的环境变量
- **AND** 应使用 `uaex` 相关的配置

#### Scenario: 容器网络配置
- **GIVEN** Docker 容器网络配置
- **WHEN** 容器启动
- **THEN** 容器应能正常通信
- **AND** 使用正确的链 ID 和标识符

### Requirement: 部署脚本
部署脚本必须（SHALL）使用 AESC 标识符进行节点初始化和配置。

#### Scenario: 节点初始化脚本
- **GIVEN** 节点初始化脚本
- **WHEN** 执行初始化
- **THEN** 脚本应使用 `aesc` 作为地址前缀
- **AND** 脚本应使用 `uaex` 作为代币面额
- **AND** 生成的配置文件应正确

#### Scenario: 账户创建脚本
- **GIVEN** 测试账户创建脚本
- **WHEN** 创建新账户
- **THEN** 账户地址应以 `aesc1` 开头
- **AND** 账户应有正确的初始余额（以 uaex 计）

#### Scenario: 验证者配置脚本
- **GIVEN** 验证者配置脚本
- **WHEN** 配置验证者
- **THEN** 验证者地址应以 `aescvaloper` 开头
- **AND** 质押代币应使用 `uaex` 面额

## ADDED Requirements

### Requirement: 快速启动工具
系统必须（SHALL）提供使用 AESC 标识符的快速启动工具。

#### Scenario: 一键启动本地节点
- **GIVEN** 开发者需要快速启动测试环境
- **WHEN** 执行快速启动命令
- **THEN** 应自动生成正确的配置
- **AND** 应创建测试账户（aesc1 格式）
- **AND** 应启动节点并等待就绪

#### Scenario: 预配置测试账户
- **GIVEN** 快速启动工具创建测试账户
- **WHEN** 节点启动完成
- **THEN** 应有至少 2 个测试账户
- **AND** 每个账户应有初始余额（如 1000000000 uaex）
- **AND** 账户私钥应保存在已知位置

#### Scenario: 清理和重置
- **GIVEN** 开发者需要重置测试环境
- **WHEN** 执行清理命令
- **THEN** 应删除所有链数据
- **AND** 应删除生成的配置文件
- **AND** 下次启动应使用新的配置

### Requirement: 配置验证工具
系统必须（SHALL）提供配置验证工具，确保使用正确的标识符。

#### Scenario: Genesis 配置验证
- **GIVEN** 一个 genesis.json 文件
- **WHEN** 运行验证工具
- **THEN** 工具应检查代币面额是否为 `uaex`
- **AND** 工具应检查地址前缀是否为 `aesc`
- **AND** 发现错误应给出清晰的错误信息

#### Scenario: 节点配置验证
- **GIVEN** 节点配置文件（app.toml, config.toml）
- **WHEN** 运行验证工具
- **THEN** 工具应检查最小 Gas 价格配置
- **AND** 工具应检查是否有遗留的 Sei 标识符
- **AND** 应提供修复建议

## MODIFIED Requirements

### Requirement: 部署文档
部署文档必须（SHALL）使用 AESC 标识符和正确的示例。

#### Scenario: README 文档
- **GIVEN** poc-deploy/README.md 文档
- **WHEN** 开发者阅读文档
- **THEN** 所有示例地址应使用 `aesc1` 格式
- **AND** 所有代币数量应使用 `uaex` 面额
- **AND** 不应包含 Sei 相关的过时信息

#### Scenario: 部署指南
- **GIVEN** 部署指南文档
- **WHEN** 按照指南操作
- **THEN** 所有命令应使用正确的标识符
- **AND** 示例配置应正确
- **AND** 故障排除部分应更新

#### Scenario: 配置示例
- **GIVEN** 示例配置文件
- **WHEN** 开发者复制使用
- **THEN** 配置应立即可用
- **AND** 不需要手动替换标识符
- **AND** 注释应清晰说明各项配置

