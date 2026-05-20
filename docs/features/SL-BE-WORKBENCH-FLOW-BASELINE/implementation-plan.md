# Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-WORKBENCH-FLOW-BASELINE`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/order/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-WORKBENCH-FLOW-BASELINE/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 对照前端订单服务调用与后端订单路由，确认兼容路径缺口。
2. 在现有物流合同签署路由旁注册 `buyer-sign` 兼容路径。
3. 复用原 handler 和权限 middleware，避免新增分叉逻辑。
4. 执行订单模块测试和后端 verify。

## 验证计划

- `go test ./internal/modules/order/...`
- `./harness/verify.sh`
- 需要的人工验证: 使用联调 token 调用 `buyer-sign` 路径确认不再 404。

## 风险与回滚关注点

- 风险: 兼容路径名称为 buyer-sign，但权限仍允许买家或卖家，因为业务含义是物流合同发起方签署。
- 回滚: 移除新增路由行即可，不涉及数据迁移。
