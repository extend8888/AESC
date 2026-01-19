# 任务清单: 完善 aexburn 模块 CLI 和 REST API

**状态**: ✅ 已完成 (2026-01-19)

## 任务列表

### 阶段 1: 分析现状

- [x] T-1.1: 检查现有 aexburn 模块结构
- [x] T-1.2: 确认 proto 文件中定义的查询方法
- [x] T-1.3: 检查 grpc_query.go 实现状态

### 阶段 2: 实现 CLI 命令

- [x] T-2.1: 创建 `x/aexburn/client/cli/query.go`
  - 实现 `GetQueryCmd()` 函数
  - 实现 `CmdQueryParams()` - 查询参数
  - 实现 `CmdQueryBurnStats()` - 查询销毁统计
  - 实现 `CmdQueryInflationStats()` - 查询通胀统计
  - 实现 `CmdQueryMonthlyBurnData()` - 查询月度数据
  - 实现 `CmdQueryNetSupply()` - 查询净供给

- [-] T-2.2: 创建 `x/aexburn/client/cli/tx.go` (如有交易命令)
  - 跳过：当前模块无交易命令

### 阶段 3: 实现 gRPC 查询

- [x] T-3.1: 检查/更新 `x/aexburn/types/query.proto`
  - 确认所有查询方法已定义
  - proto 文件已存在，无需修改

- [x] T-3.2: 实现 `x/aexburn/keeper/grpc_query.go`
  - 实现 `Params()` 方法
  - 实现 `BurnStats()` 方法
  - 实现 `InflationStats()` 方法
  - 实现 `MonthlyBurnData()` 方法
  - 实现 `NetSupply()` 方法

### 阶段 4: 注册和集成

- [x] T-4.1: 更新 `x/aexburn/module.go`
  - 在 `GetQueryCmd()` 中返回 CLI 命令
  - 在 `RegisterServices()` 中注册 gRPC 查询服务
  - 在 `RegisterGRPCGatewayRoutes()` 中注册 REST 网关

- [x] T-4.2: 更新 `app/app.go` (如需要)
  - 无需修改，模块已正确注册

### 阶段 5: 测试和验证

- [-] T-5.1: 添加 CLI 命令测试 (可选后续)
- [x] T-5.2: 添加 gRPC 查询测试
  - ✅ 创建 `x/aexburn/keeper/grpc_query_test.go`
  - ✅ TestGRPCQueryParams - 通过
  - ✅ TestGRPCQueryBurnStats - 通过
  - ✅ TestGRPCQueryInflationStats - 通过
  - ✅ TestGRPCQueryMonthlyBurnData - 通过
  - ✅ TestGRPCQueryNetSupply - 通过
- [x] T-5.3: 手动验证
  - ✅ `seid q aexburn params` - 正常工作
  - ✅ `seid q aexburn burn-stats` - 正常工作
  - ✅ `seid q aexburn inflation-stats` - 正常工作
  - ✅ `seid q aexburn monthly-burn-data` - 正常工作
  - ✅ `seid q aexburn net-supply` - 正常工作
  - ✅ REST API 端点 - 正常工作

### 阶段 6: 文档

- [-] T-6.1: 更新模块 README (可选后续)
- [-] T-6.2: 记录 API 端点 (可选后续)

## 依赖关系

```
T-1.x → T-2.x, T-3.x
T-2.x, T-3.x → T-4.x
T-4.x → T-5.x
T-5.x → T-6.x
```

## 验收检查点

1. `seid q aexburn params` 返回正确的参数 JSON
2. `seid q aexburn burn-stats` 返回销毁统计数据
3. `curl http://localhost:1317/aesc/aexburn/v1/params` 返回参数 JSON
4. 所有测试通过

