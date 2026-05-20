# Implementation Plan

## Feature 信息

- Feature ID: SL-BE-NOTIFICATIONS
- 开发负责人: Codex

## 变更范围

- 代码文件: `internal/modules/notification/**`、`cmd/api/main.go`、`internal/modules/order/application/service/order_service.go`、`internal/modules/product/application/service/product_service.go`
- 文档文件: `docs/features/SL-BE-NOTIFICATIONS/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新增 notification domain/service/handler/router/migration。
2. 在 API 启动时注册迁移和路由。
3. 在商品审核和订单关键状态流转后写入通知。
4. 执行 Go 测试和 API 构建验证。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api-notifications ./cmd/api`
- 需要的人工验证: 登录用户查看 `/api/v1/notifications`，管理员发送 `/api/v1/admin/notifications`。

## 风险与回滚关注点

- 回滚代码后需保留或人工清理 `notifications` 表数据。
- 当前未引入 MQ，站内信写入随业务事务同步完成。
