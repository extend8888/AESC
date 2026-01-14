# 任务清单：AESC 权益合约

## 第一阶段：核心合约

### 基础设施

- [ ] **TASK-101**: 创建合约目录结构
  - `contracts/src/aesc/`
  - `contracts/src/aesc/interfaces/`
  - `contracts/test/aesc/`

- [ ] **TASK-102**: 配置 Hardhat
  - 添加 OpenZeppelin 依赖
  - 配置 UUPS 代理

### AESCToken.sol

- [ ] **TASK-110**: 实现 AESCToken 合约
  - ERC-20 标准（OpenZeppelin）
  - UUPS 可升级
  - 初始铸造 16 亿到 Treasury

### Treasury.sol

- [ ] **TASK-120**: 实现 Treasury 合约
  - 9 个子金库管理
  - 释放控制逻辑
  - 权限管理（多签/时间锁）

### NodePool.sol

- [ ] **TASK-130**: 实现 NodePool 基础功能
  - 节点注册接口（供跨链桥调用）
  - 节点等级权重（1.0×/1.4×/1.8×）
  - 算力锁定系数（1.15×/1.30×/1.45×）

- [ ] **TASK-131**: 实现 NodePool 收益分配
  - 收益计算公式
  - Pending 周期（20 天线性释放）

- [ ] **TASK-132**: 实现 NodePool 惩罚机制
  - 提前解除惩罚（扣除 30%）
  - 冷却期（30 天）

### StakePool.sol

- [ ] **TASK-140**: 实现 StakePool 基础功能
  - 质押/解质押
  - 时间权重系数（1.10×~1.30×）
  - 数量权重系数（1.00×~1.15×）

- [ ] **TASK-141**: 实现 StakePool 收益分配
  - 收益计算公式
  - Pending 周期（15 天线性释放）

- [ ] **TASK-142**: 实现 StakePool 惩罚机制
  - 提前解除惩罚（扣除 15%）
  - 冷却期（14 天）

### 部署

- [ ] **TASK-150**: 编写部署脚本
  - 代理部署
  - 初始化配置
  - 金库分配

---

## 第二阶段：Growth Pool

- [ ] **TASK-201**: NodePool Growth Pool 扩展
- [ ] **TASK-202**: 裂变系数逻辑（+5%/+3%/+1%）
- [ ] **TASK-203**: Growth Pool 二级封顶（裂变≤20%，生态≤20%）

---

## 测试

- [ ] **TASK-301**: 单元测试 - AESCToken
- [ ] **TASK-302**: 单元测试 - Treasury
- [ ] **TASK-303**: 单元测试 - NodePool
- [ ] **TASK-304**: 单元测试 - StakePool
- [ ] **TASK-305**: 集成测试 - 完整流程
- [ ] **TASK-306**: Gas 消耗测试

