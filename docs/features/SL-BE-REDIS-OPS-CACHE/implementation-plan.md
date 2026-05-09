# Redis Ops Cache Implementation Plan

## Feature 信息

- Feature ID: SL-BE-REDIS-OPS-CACHE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/pkg/cache/cache.go`
  - `internal/pkg/middleware/rate_limit.go`
  - `cmd/api/main.go`
  - `internal/modules/user/infrastructure/router/router.go`
  - `internal/modules/aiprocurement/infrastructure/router/router.go`
  - `internal/modules/aiprocurement/ui/handler/handler.go`
  - `internal/modules/aiprocurement/ui/dto/dto.go`
  - `internal/modules/aiprocurement/application/service/vector_sync.go`
  - `internal/pkg/handler/platform_stats.go`
- 文档文件:
  - `docs/features/SL-BE-REDIS-OPS-CACHE/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展 Redis cache helper，支持 `INCR + EXPIRE`、TTL、SETNX 锁和安全解锁。
2. 新增 Redis 限流中间件，并接入登录、短信、上传、AI 采购模型/匹配接口。
3. 为平台首页统计增加短 TTL JSON 缓存。
4. 为 AI vector-sync 增加 target 级 Redis 锁、运行状态写入和状态查询接口。
5. 补齐 feature 文档并执行 harness 验证。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-REDIS-OPS-CACHE`
- `./harness/release.sh docs/features/SL-BE-REDIS-OPS-CACHE`
- 需要的人工验证: 生产 Redis 连接正常；内部同步状态接口使用有效 `AI_PROCUREMENT_VECTOR_SYNC_TOKEN` 调用。

## 风险与回滚关注点

- 限流阈值为代码常量，若生产流量高于预期可回滚或后续改为配置化。
- Redis 不可用时限流降级放行，避免误伤核心登录/上传流程。
- 向量同步锁 Redis 不可用时降级为直接执行，仍保留现有脚本 `flock` 防并发能力。
