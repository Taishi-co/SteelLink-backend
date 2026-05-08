# Platform Home Stats Implementation Plan

## Feature 信息

- Feature ID: SL-BE-PLATFORM-HOME-STATS
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `cmd/api/main.go`
  - `internal/pkg/handler/platform_stats.go`
- 文档文件:
  - `docs/features/SL-BE-PLATFORM-HOME-STATS/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新增平台首页统计 handler，使用 GORM 对企业、商品、订单做只读聚合。
2. 在 `/api/v1` 路由组注册公开接口 `GET /platform/stats`。
3. 执行 gofmt 与后端验证。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/verify.sh`
- 需要的人工验证: 部署后打开买家端首页，确认统计数字来自接口响应。

## 风险与回滚关注点

- 统计口径只读，不影响交易写路径。
- 回滚时移除新增 handler 和路由即可恢复。
