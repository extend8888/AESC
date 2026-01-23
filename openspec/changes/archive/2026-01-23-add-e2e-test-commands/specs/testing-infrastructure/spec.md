# testing-infrastructure Spec Delta

---

## ADDED Requirements

### Requirement: Tokenomics E2E 测试命令
系统必须（SHALL）在 Makefile 中提供 Tokenomics 经济模型的 E2E 测试命令。

#### Scenario: 运行 Tokenomics E2E 测试
- **GIVEN** 开发者在项目根目录
- **AND** localnode 已启动运行
- **WHEN** 执行 `make test-tokenomics-e2e`
- **THEN** 系统应运行所有 Tokenomics E2E 测试
- **AND** 测试应验证 AEX (uaex) Gas 代币功能
- **AND** 测试应验证 STAEX (ustaex) 质押代币功能
- **AND** 测试应验证 aexburn 销毁机制
- **AND** 测试结果应清晰显示 PASS/FAIL 状态

#### Scenario: Tokenomics 测试验证双代币体系
- **GIVEN** Tokenomics E2E 测试运行中
- **WHEN** 测试 TC-TK-01 和 TC-TK-02 执行
- **THEN** 应验证 `uaex` 用于 Gas 费用和转账
- **AND** 应验证 `ustaex` 用于质押和治理
- **AND** 应验证 `bond_denom` 配置为 `ustaex`

#### Scenario: Tokenomics 测试验证销毁机制
- **GIVEN** Tokenomics E2E 测试运行中
- **WHEN** 测试 TC-TK-03 和 TC-TK-04 执行
- **THEN** 应验证销毁比例在 30%-60% 范围
- **AND** 应验证高 Gas 使用率导致低销毁比例
- **AND** 应验证参数硬约束生效 (3%/5%/30-60%)

#### Scenario: Tokenomics 测试验证 Epoch Gas 采集
- **GIVEN** Tokenomics E2E 测试运行中
- **WHEN** 测试 TC-TK-05 执行
- **THEN** 应验证 EpochGasData 结构正确记录
- **AND** 应验证 gas_used 和 gas_limit 累计统计

#### Scenario: Tokenomics 测试通过 USDT 预编译合约验证
- **GIVEN** Tokenomics E2E 测试运行中
- **AND** EVM RPC 可访问
- **WHEN** 测试 TC-TK-06 至 TC-TK-09 执行
- **THEN** 应通过 USDT 预编译 (0x1010) 调用 ERC-20 方法
- **AND** 预编译返回值应与 REST API 查询结果一致

### Requirement: USDT 预编译合约 E2E 测试
系统必须（SHALL）通过 USDT 预编译合约调用验证 ERC-20 功能。

#### Scenario: USDT 预编译合约基础信息查询
- **GIVEN** USDT 预编译合约已注册
- **WHEN** 通过 EVM 调用 `name()`, `symbol()`, `decimals()`
- **THEN** `name()` 应返回 "Tether USD"
- **AND** `symbol()` 应返回 "USDT"
- **AND** `decimals()` 应返回 6

#### Scenario: USDT 预编译合约余额查询
- **GIVEN** 测试账户有 USDT 余额
- **WHEN** 通过 EVM 调用 USDT 预编译 `balanceOf(address)`
- **THEN** 返回值应等于账户的 USDT 余额
- **AND** 与 REST API bank/balances 查询结果一致

#### Scenario: USDT 预编译合约转账
- **GIVEN** 发送方有足够 USDT 余额
- **WHEN** 通过 EVM 调用 USDT 预编译 `transfer(to, amount)`
- **THEN** 交易应成功
- **AND** 发送方余额应减少 amount
- **AND** 接收方余额应增加 amount

#### Scenario: USDT 预编译合约授权转账
- **GIVEN** owner 有 USDT 余额
- **WHEN** owner 调用 `approve(spender, amount)` 授权
- **AND** spender 调用 `transferFrom(owner, to, amount)`
- **THEN** 转账应成功
- **AND** allowance 应相应减少

### Requirement: 共识 E2E 测试命令
系统必须（SHALL）在 Makefile 中提供多节点共识的 E2E 测试命令。

#### Scenario: 运行共识 E2E 测试
- **GIVEN** 开发者在项目根目录
- **AND** Docker 和 Docker Compose 已安装
- **WHEN** 执行 `make test-consensus-e2e`
- **THEN** 系统应启动 4 节点 Docker 集群
- **AND** 系统应运行所有共识 E2E 测试
- **AND** 测试结果应清晰显示 PASS/FAIL 状态
- **AND** 测试完成后应提供清理选项

#### Scenario: 共识测试验证集群启动
- **GIVEN** 共识 E2E 测试运行中
- **WHEN** 测试 TC-MC-01 执行
- **THEN** 应验证 4 个节点 RPC 端点可访问
- **AND** 应验证所有节点 chain-id 一致
- **AND** 应验证节点正常出块

#### Scenario: 共识测试验证交易同步
- **GIVEN** 共识 E2E 测试运行中
- **AND** 4 节点集群正常运行
- **WHEN** 测试 TC-MC-02 执行
- **THEN** 应在 node0 发送交易
- **AND** 应验证交易在 node1/2/3 可查询
- **AND** 交易同步延迟应 < 5 秒

#### Scenario: 共识测试验证状态一致性
- **GIVEN** 共识 E2E 测试运行中
- **WHEN** 测试 TC-MC-03 和 TC-MC-04 执行
- **THEN** 应验证所有节点区块高度差异 ≤ 1
- **AND** 应验证 aexburn 参数在所有节点一致
- **AND** 应验证 bank 余额在所有节点一致

### Requirement: E2E 测试脚本
系统必须（SHALL）提供可独立运行的 E2E 测试脚本。

#### Scenario: Tokenomics 测试脚本独立运行
- **GIVEN** localnode 已启动
- **WHEN** 直接执行 `./poc-deploy/localnode/scripts/run_tokenomics_e2e_tests.sh`
- **THEN** 脚本应检查环境依赖
- **AND** 脚本应执行所有 Tokenomics 测试用例
- **AND** 脚本应输出测试汇总 (通过/失败数量)
- **AND** 测试失败时脚本应返回非零退出码

#### Scenario: 共识测试脚本独立运行
- **GIVEN** 4 节点 Docker 集群已启动
- **WHEN** 直接执行 `./docker/localnode/scripts/run_consensus_e2e_tests.sh`
- **THEN** 脚本应检查所有节点可访问
- **AND** 脚本应执行所有共识测试用例
- **AND** 脚本应输出测试汇总 (通过/失败数量)
- **AND** 测试失败时脚本应返回非零退出码

