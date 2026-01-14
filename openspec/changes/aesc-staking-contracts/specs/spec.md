# AESC 权益合约规格

## 概述

AESC 权益代币通过智能合约实现，包括代币、金库、节点激励池和质押池。

---

## AESCToken.sol

### [ADDED] REQ-TOKEN-001: ERC-20 实现
使用 OpenZeppelin ERC20Upgradeable。

### [ADDED] REQ-TOKEN-002: UUPS 可升级
使用 UUPSUpgradeable 代理模式。

### [ADDED] REQ-TOKEN-003: 初始铸造
部署时铸造 1,600,000,000 AESC (18 decimals) 到 Treasury。

---

## Treasury.sol

### [ADDED] REQ-TREASURY-001: 金库管理
管理 9 个子金库，按比例分配 AESC。

### [ADDED] REQ-TREASURY-002: 释放控制
各金库按不同周期释放代币。

---

## NodePool.sol

### [ADDED] REQ-NODEPOOL-001: 节点注册
```solidity
function registerNode(address owner, uint8 level) external onlyBridge;
```
节点等级: 1=先锋(1.0×), 2=泰坦(1.4×), 3=创世(1.8×)

### [ADDED] REQ-NODEPOOL-002: 算力锁定
```solidity
function lockPower(uint256 nodeId, uint256 amount, uint256 days) external;
```
| 周期 | 系数 |
|------|------|
| ≥45天 | 1.15× |
| ≥90天 | 1.30× |
| ≥180天 | 1.45× |

### [ADDED] REQ-NODEPOOL-003: 多节点系数
当前固定 1.0×。

### [ADDED] REQ-NODEPOOL-004: 收益计算
```
权重 = 等级权重 × 多节点系数 × 锁定系数
收益 = (节点权重 ÷ 总权重) × 当期释放量
```

### [ADDED] REQ-NODEPOOL-005: Pending 释放
20 天线性释放。

### [ADDED] REQ-NODEPOOL-006: 提前解除
- 算力系数回落 1.0
- 扣除 30% 未解锁收益
- 30 天冷却期

---

## StakePool.sol

### [ADDED] REQ-STAKE-001: 质押
```solidity
function stake(uint256 amount, uint256 lockDays) external;
```

### [ADDED] REQ-STAKE-002: 时间权重
| 周期 | 系数 |
|------|------|
| ≥45天 | 1.10× |
| ≥90天 | 1.18× |
| ≥180天 | 1.25× |
| ≥360天 | 1.30× |

### [ADDED] REQ-STAKE-003: 数量权重
| 数量 | 系数 |
|------|------|
| <10k | 1.00 |
| 10k-50k | 1.05 |
| 50k-200k | 1.10 |
| ≥200k | 1.15 |

### [ADDED] REQ-STAKE-004: 收益计算
```
算力 = 数量 × 时间权重 × 数量权重
收益 = (个人算力 ÷ 总算力) × 当期释放量
```

### [ADDED] REQ-STAKE-005: Pending 释放
15 天线性释放。

### [ADDED] REQ-STAKE-006: 提前解除
- 扣除 15% 未解锁收益
- 时间权重清零
- 14 天冷却期

---

## 场景

### 场景: 用户质押 AESC

```
Given 用户持有 100,000 AESC
When 用户质押 100,000 AESC，锁定 90 天
Then 时间权重 = 1.18×
And 数量权重 = 1.10×
And 质押算力 = 100,000 × 1.18 × 1.10 = 129,800
```

### 场景: 节点收益计算

```
Given 用户持有创世节点，锁定 180 天
And 全网总算力 = 1,000,000
When 当期释放 100,000 AESC
Then 节点权重 = 1.8 × 1.0 × 1.45 = 2.61
And 用户收益 = (2.61 ÷ 1,000,000) × 100,000 = 0.261 AESC
```

