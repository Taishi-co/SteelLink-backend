# Redis Cache Expansion Implementation Plan

## Feature 信息

- Feature ID: SL-BE-REDIS-CACHE-EXPANSION
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/pkg/cache/cache.go`
  - `internal/pkg/middleware/auth.go`
  - `internal/pkg/middleware/rbac.go`
  - `internal/modules/user/application/service/auth_service.go`
  - `internal/modules/user/application/service/enterprise_permission_service.go`
  - `internal/modules/product/application/service/product_service.go`
  - `internal/modules/serviceoffer/ui/handler/handler.go`
  - `internal/modules/market/application/service/service.go`
  - `internal/modules/aiprocurement/application/service/service.go`
- 文档文件:
  - `docs/features/SL-BE-REDIS-CACHE-EXPANSION/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新增 Redis JSON 缓存 helper、hash key 与按 pattern 删除能力。
2. 接入鉴权快照、企业服务权限缓存与 refresh token 黑名单。
3. 接入商品、服务资源、市场统计和 AI 采购中间结果短缓存。
4. 在写路径和审核路径删除受影响缓存。
5. 执行 Go 测试、构建与 harness 验证。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- `./harness/verify.sh`
- 如涉及镜像或容器：`docker compose config`
- 需要的人工验证: 生产环境确认业务后端、Admin 后端连接同一个 Redis 实例和 DB。

## 风险与回滚关注点

- Redis 短缓存可能带来秒级数据滞后；写路径会主动删除核心前缀，仍保留 TTL 兜底。
- refresh token 黑名单依赖 Redis；Redis 不可用时维持既有 JWT 行为，不阻塞登录刷新。
- 回滚时删除本 feature 代码即可，MySQL 数据结构不需要回滚。
