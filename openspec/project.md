# 项目概览

## 项目目标
AESC Chain（基于 Sei Protocol）是一条高性能、并行化的 EVM 兼容 L1 区块链。它融合了 Solana 的高性能与以太坊 EVM 的兼容性，具备 400ms 区块最终确认、EVM 和 CosmWasm 的乐观并行执行，以及通过 IBC 实现的 Cosmos 生态无缝互操作性。

## 技术栈
- **主要语言**：Go 1.24+
- **区块链框架**：Cosmos SDK v0.45.10
- **共识机制**：Tendermint（sei-tendermint 分支）
- **EVM 实现**：go-ethereum 分支（支持并行化）
- **智能合约**：CosmWasm (wasmd)、Solidity (EVM)
- **数据库**：RocksDB、Pebble、BadgerDB、SeiDB
- **构建工具**：Make、Docker
- **测试框架**：Go testing、Hardhat（用于 Solidity 合约）
- **消息协议**：Protocol Buffers

## 项目规范

### 代码风格
- 遵循 Go 标准规范（gofmt、golint）
- 使用描述性的变量和函数命名
- 保持函数职责单一，便于测试
- 优先使用组合而非继承
- 适当使用表驱动测试

### 架构模式
- **模块化 SDK 设计**：自定义模块位于 `x/` 目录（epoch、evm、mint、oracle、tokenfactory）
- **预编译合约**：`precompiles/` 目录中的 EVM 预编译合约，用于桥接 Cosmos 功能
- **分支依赖**：cosmos-sdk、tendermint、iavl、wasmd 的自定义分支位于 `sei-*` 目录
- **RPC 层**：`evmrpc/` 中的自定义 EVM RPC 实现
- **并行执行**：`aclmapping/` 中的乐观并发控制（OCC）与 ACL 映射

### 测试策略
- 单元测试与源文件同目录（`*_test.go`）
- 集成测试位于 `integration_test/` 目录
- OCC 专项测试位于 `occ_tests/`
- 负载测试工具位于 `loadtest/`
- 合约测试使用 Hardhat，位于 `contracts/`

### Git 工作流
- 主开发分支：`main`
- 为变更创建功能分支
- 合并前运行测试
- 尽量使用规范化提交信息

## 领域知识
- **双涡轮共识（Twin Turbo Consensus）**：实现 400ms 最终确认
- **乐观并行化**：无需开发者额外工作即可实现 EVM/CosmWasm 并行执行
- **SeiDB**：支持并行写入的高性能存储层
- **可互操作 EVM**：完全 EVM 兼容，同时可访问 Cosmos 功能（IBC、费用授权、多签等）
- **预编译合约**：EVM 与 Cosmos SDK 模块之间的桥梁（质押、治理、预言机等）
- **Gas 经济模型**：自定义 Gas 代币（AEX），基于节点的经济模型

## 重要约束
- 必须保持与现有 EVM 合约的向后兼容性
- 必须保留 Cosmos SDK 模块接口以支持 IBC 兼容
- 性能关键路径必须支持并行执行
- 状态变更必须在所有验证者之间保持确定性
- 硬件要求：最低 64GB RAM、1TB NVME SSD、16 核 CPU

## 外部依赖
- **Cosmos 生态**：IBC 协议、Cosmos SDK 模块
- **以太坊工具链**：标准 EVM 工具（Hardhat、ethers.js 等）
- **云服务商**：AWS S3（状态快照）、Azure Blob Storage
- **DNS/CDN**：Cloudflare、AWS Route53
- **预言机服务**：`oracle/` 目录中的价格预言机
