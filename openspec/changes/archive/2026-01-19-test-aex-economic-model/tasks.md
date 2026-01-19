# AEX 经济模型集成测试 - 任务清单

> **状态**：⏳ 进行中
> **分支**：`feature/test-aex-economic-model`

---

## 阶段 1：环境准备

- [ ] **T-1.1**: 编译链二进制
  - 执行 `make build` 编译 seid
  - 确认编译成功无错误

- [ ] **T-1.2**: 准备 Genesis 配置
  - 使用 `depoly-scripts/localnode/aesc_genesis_template.json` 作为基础
  - 配置初始账户和余额
  - 缩短 epoch 周期便于测试（如 1 分钟）

- [ ] **T-1.3**: 启动本地单节点
  - 执行启动脚本
  - 验证节点正常出块
  - 验证 RPC 端点可访问

- [ ] **T-1.4**: 准备测试账户
  - 创建至少 3 个测试账户
  - 确保每个账户有足够的 uaex 余额
  - 记录账户地址和私钥

---

## 阶段 2：基础验证测试

### 2.1 Gas 代币定位 (第1章)

- [ ] **T-2.1.1**: 执行 TC-1-01 - Gas 代币仅用于手续费
  - 命令：`seid tx bank send ...`
  - 记录结果到 test-results.md

- [ ] **T-2.1.2**: 执行 TC-1-02 - Gas 与 AESC 隔离
  - 命令：`seid q staking params`
  - 记录结果到 test-results.md

### 2.2 发行与供给 (第2章)

- [ ] **T-2.2.1**: 执行 TC-2-01 - 初始发行量验证
  - 命令：`seid q aexburn params`
  - 验证 InitialSupply = 500,000,000,000,000 uaex

- [ ] **T-2.2.2**: 执行 TC-2-02 - 年通胀上限验证
  - 命令：`seid q aexburn params`
  - 验证 MaxAnnualInflationRate = 0.03

- [ ] **T-2.2.3**: 执行 TC-2-03 - 通胀触发条件验证
  - 需要生成高负载触发通胀

- [ ] **T-2.2.4**: 执行 TC-2-04 - 净供给硬约束验证
  - 命令：`seid q aexburn params`
  - 验证 MaxNetSupplyRatePerYear = 0.05

---

## 阶段 3：销毁与动态调节测试

### 3.1 销毁机制 (第3章)

- [ ] **T-3.1.1**: 执行 TC-3-01 - 手续费销毁基本功能
  - 查询 BurnStats 前后变化
  - 计算实际销毁比例

- [ ] **T-3.1.2**: 执行 TC-3-02 - 动态销毁 (低使用率)
  - 保持低交易量
  - 验证销毁比例 ≈ 30%-40%

- [ ] **T-3.1.3**: 执行 TC-3-03 - 动态销毁 (高使用率)
  - 批量发送交易提高负载
  - 验证销毁比例调整

- [ ] **T-3.1.4**: 执行 TC-3-04 - 净供给反向刹车
  - 模拟连续净负场景
  - 验证自动下调机制

---

## 阶段 4：价格与可用性测试

### 4.1 Gas 价格 (第4章)

- [ ] **T-4.1.1**: 执行 TC-4-01 - 动态 Gas 价格
  - 使用 `cast` 或 RPC 查询 base fee
  - 观察拥堵/空闲时价格变化

- [ ] **T-4.1.2**: 执行 TC-4-02 - 目标使用成本
  - 记录简单转账的实际 Gas 消耗
  - 估算美元成本

---

## 阶段 5：验证者激励测试

### 5.1 验证者收入 (第5章)

- [ ] **T-5.1.1**: 执行 TC-5-01 - 验证者手续费分配
  - 查询验证者奖励变化
  - 验证分配比例正确

- [ ] **T-5.1.2**: 执行 TC-5-02 - 验证者通胀补贴
  - 触发通胀后检查分配

- [ ] **T-5.1.3**: 执行 TC-5-03 - 收入平滑机制 (可选)
  - 验证默认关闭状态

---

## 阶段 6：治理与边界测试

### 6.1 治理边界 (第6章)

- [ ] **T-6.1.1**: 执行 TC-6-01 - 可治理参数边界
  - 代码审查参数校验逻辑

- [ ] **T-6.1.2**: 执行 TC-6-02 - 不可治理约束
  - 验证无 uaex 质押机制

---

## 阶段 7：综合测试

- [ ] **T-7.1**: 执行 TC-E2E-01 - 完整生命周期测试
  - 运行多个 epoch
  - 记录全程数据

- [ ] **T-7.2**: 执行 TC-E2E-02 - EVM 兼容性测试
  - 使用 EVM 工具交互
  - 验证 Gas 消耗正确

---

## 阶段 8：结果汇总

- [ ] **T-8.1**: 整理测试结果
  - 更新 test-cases.md 中的实际结果
  - 填写 test-results.md

- [ ] **T-8.2**: 问题记录与跟进
  - 记录发现的问题
  - 创建修复任务（如需要）

- [ ] **T-8.3**: 更新提案状态
  - 所有测试通过后更新为 ✅ 已完成

---

## 命令速查表

```bash
# 编译
make build

# 启动节点 (根据实际脚本路径)
cd poc-deploy/localnode && ./start.sh

# 查询 aexburn 参数
seid q aexburn params

# 查询 aexburn 统计
seid q aexburn burn-stats

# 发送转账
seid tx bank send <from> <to> <amount>uaex --gas auto --gas-prices 0.02uaex

# 查询余额
seid q bank balances <address>

# EVM RPC 查询
curl -X POST http://localhost:8545 -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}'
```

