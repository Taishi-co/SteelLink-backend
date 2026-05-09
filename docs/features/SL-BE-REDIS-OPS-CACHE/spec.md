# Redis Ops Cache Spec

## 基本信息

- Feature ID: SL-BE-REDIS-OPS-CACHE
- Feature 名称: Redis 限流、平台统计缓存与向量同步锁
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

第一阶段 Redis 改造已覆盖鉴权快照、服务权限、商品、服务资源、市场统计和 AI 采购中间结果。第二阶段继续把 Redis 用在更偏运维和风控的场景：高频接口限流、首页平台统计短缓存、AI 向量同步互斥锁和最近同步状态，避免接口滥用、重复同步和首页聚合查询反复打到 MySQL。

## 范围

### In Scope

- 新增 Redis 原子计数、TTL、字符串读取、分布式锁辅助函数。
- 新增 Gin Redis 限流中间件。
- 对登录、refresh、短信发送、短信登录、重置密码、上传、AI 采购解析和匹配接口加 Redis 限流。
- 对 `GET /api/v1/platform/stats` 加 60 秒短缓存。
- 对内部 AI 向量同步接口加 Redis SETNX 锁，避免并发同步。
- 新增内部 `GET /api/v1/internal/ai-procurements/vector-sync/status` 查询最近同步状态。

### Out of Scope

- 不把订单、库存、支付、融资、合同和履约事件迁移到 Redis。
- 不新增数据库表。
- 不改变现有 AI 向量同步 token 校验方式。
- 不做 Admin 前端页面展示同步状态。

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/sms/send`
  - `POST /api/v1/auth/sms/login`
  - `POST /api/v1/auth/password/reset`
  - `POST /api/v1/uploads`
  - `POST /api/v1/ai-procurements/parse`
  - `POST /api/v1/ai-procurements/match-products`
  - `POST /api/v1/ai-procurements/match-logistics-warehouse`
  - `GET /api/v1/platform/stats`
  - `POST /api/v1/internal/ai-procurements/vector-sync`
  - `GET /api/v1/internal/ai-procurements/vector-sync/status`
- 受影响服务: AI 采购向量同步服务。
- 受影响 middleware: 新增 Redis 限流中间件。
- 受影响表/model: 无新增表。
- 配置或环境变量变化: 无新增必填配置，复用现有 Redis 与 `AI_PROCUREMENT_VECTOR_SYNC_TOKEN`。
- Docker / 部署影响: 后端容器需要连接 Redis；Redis 不可用时限流降级放行，向量同步锁降级为单进程直接执行。

## 权限与合规影响

- 受影响角色: 未登录用户、登录用户、企业采购用户、内部向量同步调用方。
- 新增/调整权限: 内部同步状态接口复用现有向量同步 token。
- 是否影响审核、风控、日志留痕: 限流会对高频请求返回 HTTP 429；向量同步状态在 Redis 保留 24 小时。

## 验收标准

1. Redis 可用时，受保护接口超过窗口阈值返回 429 和 `Retry-After`。
2. `GET /api/v1/platform/stats` 首次回源 MySQL，后续 60 秒内可从 Redis 返回同等结构。
3. 并发调用向量同步时，同一 target 只能有一个同步任务执行，其它请求返回 running 状态。
4. `GET /api/v1/internal/ai-procurements/vector-sync/status` 可返回最近一次同步状态或 unknown。
5. `go test ./...`、构建、review gate、release gate 通过。
