# 任务清单：AEX Gas 代币（链层完整实现）

> **说明**：AEX 是链层原生 Gas 代币，所有功能通过修改链代码实现。
> 本任务清单包含 AEX 系统的完整功能实现。

---

## AEX-P0：基础配置 ✅ 已完成

### 链代码修改

- [x] **AEX-001**: 修改 `app/params/config.go`
  - 将 `HumanCoinUnit` 从 "sei" 改为 "aex"
  - 将 `BaseCoinUnit` 从 "uaex" 改为 "uaex"
  - 将 `UaexExponent` 重命名为 `UaexExponent`
  - 将 `Bech32PrefixAccAddr` 从 "sei" 改为 "aesc"

- [x] **AEX-002**: 修改 `x/evm/keeper/params.go`
  - 将 `BaseDenom` 常量从 "uaex" 改为 "uaex"

- [x] **AEX-003**: 修改 `cmd/seid/cmd/root.go`
  - 将 `MinGasPrices` 从 "0.02uaex" 改为 "0.02uaex"

- [x] **AEX-004**: 全局搜索替换（30+ 文件）
  - 所有 "uaex" → "uaex"
  - 所有 "aesc1" 地址 → "aesc1" 地址

- [x] **AEX-005**: Genesis 配置模板
  - `depoly-scripts/localnode/aesc_genesis_template.json`
  - `depoly-scripts/localnode/GENESIS_CONFIG.md`

- [x] **AEX-006**: poc-deploy 脚本适配
  - 所有脚本、配置、文档已更新

- [x] **AEX-007**: 编译验证通过
- [x] **AEX-008**: 核心测试通过

---

## AEX-P1：通胀与供给控制 ✅ 已完成

### 通胀机制

- [x] **AEX-101**: 分析现有 `x/mint` 模块结构
  - 理解 Sei 链的 mint 模块实现
  - 确定扩展点和修改方案

- [x] **AEX-102**: 实现 AEX 通胀机制
  - 年通胀上限 3%（协议级硬约束）
  - 通胀触发条件：交易量、区块稳定性、Gas 使用率
  - 使用 `x/epoch` 的 `AfterEpochEnd` hook 触发

- [x] **AEX-103**: 实现净供给硬约束
  - 任意连续 12 个月净增发 ≤ 初始量 5%
  - 新增状态存储记录 12 个月滚动窗口数据

- [x] **AEX-104**: 单元测试
  - 通胀计算逻辑测试
  - 净供给约束测试

---

## AEX-P2：手续费销毁与动态调节 ✅ 已完成

### 手续费销毁

- [x] **AEX-201**: 分析现有手续费处理逻辑
  - 理解 `x/evm/ante` 中的 fee 处理
  - 理解 fee 分配给验证者的机制

- [x] **AEX-202**: 实现手续费销毁机制
  - 修改 `x/evm/ante` 中的 fee 处理逻辑
  - 销毁部分 fee（发送到 burn 模块账户）
  - 剩余部分分配给验证者

- [x] **AEX-203**: 实现动态销毁比例
  - 销毁比例区间：30% - 60%
  - 基于 Gas 使用率自动调节
  - 使用率偏低 → 30-40%
  - 使用率正常 → ≈50%
  - 使用率过高 → 下调销毁，更多给验证者

- [x] **AEX-204**: 实现净供给反向刹车
  - 连续 3 个统计周期净供给为负
  - 自动下调销毁比例
  - 与 `x/mint` 模块联动

- [x] **AEX-205**: 状态存储与历史数据
  - 记录每个周期的销毁量
  - 记录每个周期的通胀量
  - 计算净供给变化

- [x] **AEX-206**: 单元测试
  - 销毁逻辑测试
  - 动态调节测试
  - 反向刹车测试

---

## AEX-P3：动态 Gas 价格 ✅ 已评估

- [x] **AEX-301**: 评估现有 EIP-1559 实现
  - ✅ Sei 链已完整实现 EIP-1559 风格动态 base fee 机制
  - 核心实现：`x/evm/keeper/fee.go` (`AdjustDynamicBaseFeePerGas`)
  - 支持目标 gas 使用量、动态上下调整、最低/最高限制

- [x] **AEX-302**: 参数评估
  - ✅ 当前参数满足目标单笔成本 $0.01 - $0.05
  - MinimumFeePerGas: 1 gwei (简单转账 < $0.001)
  - MaximumFeePerGas: 1000 gwei (拥堵时简单转账 ~$0.02)
  - **结论：无需调整**

- [ ] **AEX-303**: 参数调优与测试（可选）
  - 本地网络测试
  - 负载测试验证
  - 可根据实际运行情况通过治理调整参数

---

## AEX-P4：辅助功能 ✅ 已完成

- [x] **AEX-401**: Gas Credit 兜底机制 - **不需要实现**
  - ✅ 可通过 ERC-4337 账户抽象 / Paymaster 代付机制实现
  - 用户无 Gas 时，第三方可代付交易费用
  - 无需单独开发链层 Gas Credit 系统

- [x] **AEX-402**: 验证者收入平滑机制 ✅ 已实现
  - 高活跃周期（>70%）：部分收入（10%）进入缓冲池
  - 低活跃周期（<30%）：缓冲池释放补贴（5%）
  - 默认关闭，可通过治理启用
  - 实现位置：`x/aexburn/keeper/income_smoother.go`
  - 集成到 `BurnFees` 方法中，零侵入 distribution 模块

---

## 本地测试

- [ ] **AEX-901**: 本地节点测试
  - 启动本地测试网
  - 验证 AEX 作为 Gas 正常工作
  - 验证地址前缀为 "aesc"

- [ ] **AEX-902**: 通胀机制测试
  - 验证通胀触发条件
  - 验证年上限约束

- [ ] **AEX-903**: 销毁机制测试
  - 验证手续费销毁
  - 验证动态调节

