# 变更提案：AESC 权益合约

## 状态：草案

## 依赖：aex-gas-token

## 1. 概述

实现 AESC 权益代币的智能合约系统，包括代币、金库、节点激励池和基础质押池。

## 2. 动机

AESC 链需要权益代币系统来支撑：
1. 节点激励机制 - 奖励节点参与者
2. 质押收益机制 - 奖励长期持有者
3. 金库管理 - 控制代币释放节奏

## 3. 变更内容

### 3.1 合约列表

| 合约 | 职责 | 标准 |
|------|------|------|
| `AESCToken.sol` | ERC-20 代币 | ERC-20 + UUPS |
| `Treasury.sol` | 金库管理 | UUPS |
| `NodePool.sol` | 节点激励池 | UUPS |
| `StakePool.sol` | 基础质押池 | UUPS |

### 3.2 代币参数

| 参数 | 值 |
|------|-----|
| 名称 | AESC |
| 标准 | ERC-20 |
| 精度 | 18 decimals |
| 总量 | 1,600,000,000 AESC（固定） |

### 3.3 金库分配

| 金库 | 占比 | 数量 |
|------|------|------|
| 节点激励池 | 25% | 400,000,000 |
| 质押池 | 15% | 240,000,000 |
| 其他 7 个金库 | 60% | 960,000,000 |

## 4. 影响范围

### 4.1 新增文件

- `contracts/src/aesc/AESCToken.sol`
- `contracts/src/aesc/Treasury.sol`
- `contracts/src/aesc/NodePool.sol`
- `contracts/src/aesc/StakePool.sol`
- `contracts/src/aesc/interfaces/`
- `contracts/test/aesc/`
- `contracts/scripts/deploy-aesc.js`

### 4.2 不影响

- 链代码
- 现有合约
- 现有预编译

## 5. 不可突破约束

1. AEX 与 AESC 强隔离
2. AESC 不作为 Gas 支付
3. 加速机制只影响分配权重，不改变释放总量
4. 波比 ≤ 40%

## 6. 风险

| 风险 | 缓解措施 |
|------|----------|
| 合约安全漏洞 | OpenZeppelin、审计、多签 |
| Gas 消耗高 | 算法优化、批量操作 |
| 升级风险 | UUPS、充分测试 |

## 7. 参考文档

- `tmp/AESC 链开发技术方案.md`
- `tmp/AESC 节点系统.md`

