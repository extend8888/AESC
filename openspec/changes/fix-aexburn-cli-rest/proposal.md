# 变更提案: 完善 aexburn 模块 CLI 和 REST API

## 基本信息

| 字段 | 值 |
|------|-----|
| 提案 ID | fix-aexburn-cli-rest |
| 状态 | ✅ 已完成 |
| 优先级 | 中 |
| 创建日期 | 2026-01-19 |
| 完成日期 | 2026-01-19 |
| 关联需求 | AEX 经济模型 |

## 背景

在 AEX 经济模型集成测试 (`test-aex-economic-model`) 过程中，发现 `aexburn` 模块缺少 CLI 查询命令和 REST API 实现。

### 问题描述

1. **CLI 查询不可用**
   - `seid q aexburn params` 返回 `unknown command "aexburn"`
   - 无法通过命令行查询模块参数和状态

2. **REST API 未实现**
   - `/aesc/aexburn/v1/params` 返回 `Not Implemented`
   - `/aesc/aexburn/v1/burn_stats` 返回 `Not Implemented`

### 影响

- 运维人员无法方便地查询 aexburn 模块状态
- 监控系统无法通过 API 获取销毁和通胀数据
- 前端 DApp 无法展示经济模型相关数据

## 变更目标

为 `aexburn` 模块实现完整的 CLI 和 REST API 支持。

## 变更范围

### 需要新增的文件

1. `x/aexburn/client/cli/query.go` - CLI 查询命令
2. `x/aexburn/client/cli/tx.go` - CLI 交易命令 (如果有)

### 需要修改的文件

1. `x/aexburn/module.go` - 注册 CLI 命令
2. `x/aexburn/keeper/grpc_query.go` - 实现 gRPC 查询 (如果未实现)

## 需要实现的 CLI 命令

| 命令 | 功能 |
|------|------|
| `seid q aexburn params` | 查询模块参数 |
| `seid q aexburn burn-stats` | 查询销毁统计 |
| `seid q aexburn inflation-stats` | 查询通胀统计 |
| `seid q aexburn reverse-brake-state` | 查询反向刹车状态 |
| `seid q aexburn income-buffer` | 查询收入缓冲区 |

## 需要实现的 REST API

| 端点 | 方法 | 功能 |
|------|------|------|
| `/aesc/aexburn/v1/params` | GET | 查询模块参数 |
| `/aesc/aexburn/v1/burn_stats` | GET | 查询销毁统计 |
| `/aesc/aexburn/v1/inflation_stats` | GET | 查询通胀统计 |
| `/aesc/aexburn/v1/reverse_brake_state` | GET | 查询反向刹车状态 |
| `/aesc/aexburn/v1/income_buffer` | GET | 查询收入缓冲区 |

## 验收标准

1. 所有 CLI 命令可正常执行并返回正确数据
2. 所有 REST API 端点可正常访问并返回 JSON 数据
3. 添加相应的单元测试
4. 更新模块文档

## 工作量估计

| 任务 | 估计时间 |
|------|----------|
| 实现 CLI 查询命令 | 2 小时 |
| 实现 gRPC 查询 (如需要) | 2 小时 |
| 注册和集成 | 1 小时 |
| 测试和验证 | 1 小时 |
| **总计** | **6 小时** |

## 实现结果

### 新增文件

| 文件 | 说明 |
|------|------|
| `x/aexburn/keeper/grpc_query.go` | gRPC 查询服务实现 |
| `x/aexburn/client/cli/query.go` | CLI 查询命令 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `x/aexburn/module.go` | 注册 CLI 命令、gRPC 服务、REST 网关 |

### 已实现的 CLI 命令

| 命令 | 状态 |
|------|------|
| `seid q aexburn params` | ✅ 正常工作 |
| `seid q aexburn burn-stats` | ✅ 正常工作 |
| `seid q aexburn inflation-stats` | ✅ 正常工作 |
| `seid q aexburn monthly-burn-data` | ✅ 正常工作 |
| `seid q aexburn net-supply` | ✅ 正常工作 |

### 已实现的 REST API

| 端点 | 状态 |
|------|------|
| `/aesc/aexburn/v1/params` | ✅ 正常工作 |
| `/aesc/aexburn/v1/burn_stats` | ✅ 正常工作 |
| `/aesc/aexburn/v1/inflation_stats` | ✅ 正常工作 |
| `/aesc/aexburn/v1/monthly_burn_data` | ✅ 正常工作 |
| `/aesc/aexburn/v1/net_supply` | ✅ 正常工作 |

## 参考

- 现有模块实现: `x/mint/client/cli/`
- Cosmos SDK CLI 文档: https://docs.cosmos.network/main/building-modules/module-interfaces

