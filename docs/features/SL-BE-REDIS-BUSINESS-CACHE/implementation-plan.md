# Redis Business Cache Expansion Implementation Plan

## Feature 信息

- Feature ID: SL-BE-REDIS-BUSINESS-CACHE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/pkg/cache/cache.go`
  - `internal/pkg/middleware/idempotency.go`
  - `internal/pkg/middleware/cors.go`
  - `internal/modules/aiprocurement/application/service/service.go`
  - `internal/modules/aiprocurement/ui/handler/handler.go`
  - `internal/modules/aiprocurement/infrastructure/router/router.go`
  - `internal/modules/user/application/service/user_service.go`
  - `internal/modules/user/application/service/enterprise_permission_service.go`
  - `internal/modules/order/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-REDIS-BUSINESS-CACHE/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展 Redis TTL 常量，新增幂等中间件并允许 `Idempotency-Key` 跨域头。
2. 调整 AI 商品匹配缓存 key，加入用户、企业和向量配置维度。
3. 为用户 dashboard 和企业服务 overview 增加短 TTL 缓存及服务权限失效。
4. 在高风险写入路由挂载 Redis 幂等保护。
5. 执行 Go 测试、构建和 harness review/release。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-REDIS-BUSINESS-CACHE`
- `./harness/release.sh docs/features/SL-BE-REDIS-BUSINESS-CACHE`
- 需要的人工验证:
  - 前端提交订单或 AI 生成订单时带 `Idempotency-Key`，重复点击应收到 409。
  - 服务开通状态变更后刷新 `/enterprises/me/services` 验证状态更新。

## 风险与回滚关注点

- 幂等中间件只在请求带 `Idempotency-Key` 时生效，不影响旧客户端。
- 缓存均为短 TTL，Redis 故障时按现有 helper 行为回源或放行。
- 回滚本 PR 后业务恢复为数据库直查和无幂等保护。
