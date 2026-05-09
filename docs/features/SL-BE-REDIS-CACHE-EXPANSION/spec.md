# Redis Cache Expansion Spec

## 基本信息

- Feature ID: SL-BE-REDIS-CACHE-EXPANSION
- Feature 名称: Redis 缓存与会话安全扩展
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

生产环境已有 Redis 服务。当前后端只稳定使用了短信验证码、短信频控和 AI 采购运行时配置，商品、服务资源、权限判断、市场统计和 AI 采购匹配仍主要依赖 MySQL 或外部模型实时计算。需要把可重建、短时有效、高频读取的数据放入 Redis，降低数据库压力并补齐刷新 token 重放防护。

## 范围

### In Scope

- 新增通用 Redis JSON 缓存 helper。
- 缓存用户鉴权快照与企业服务权限判断。
- 缓存公开商品列表、商品详情、品类列表。
- 缓存仓储、物流、资金服务公开列表与详情。
- 缓存市场大厅统计数据。
- 缓存 AI 采购解析、商品匹配、物流仓储匹配中间结果。
- 刷新 token 轮换后将旧 refresh token 加入 Redis 黑名单。

### Out of Scope

- 不把订单、支付、融资申请、履约事件、库存扣减结果迁移到 Redis。
- 不改变 AI 采购运行时配置的 Redis key 与现有 Admin 同步语义。
- 不新增外部 Redis 集群配置项。

## 接口与数据影响

- 受影响路由/handler: 商品、服务资源、市场大厅、AI 采购、鉴权中间件相关既有接口。
- 受影响服务: `ProductService`、`MarketService`、`ProcurementService`、`AuthService`、服务资源 handler。
- 受影响 middleware: `AuthRequired`、`RequireEnterpriseService`。
- 受影响表/model: 无新增表；MySQL 仍为主数据源。
- 配置或环境变量变化: 无新增必填配置，复用现有 `redis` 配置。
- Docker / 部署影响: 生产后端必须能连接现有 Redis。

## 权限与合规影响

- 受影响角色: 登录用户、企业用户、已开通销售/采购/物流/仓储/资金服务的企业。
- 新增/调整权限: 无新增权限，缓存现有权限判断结果。
- 是否影响审核、风控、日志留痕: 服务权限审核后会失效对应企业权限缓存，避免权限状态长期滞后。

## 验收标准

1. Redis 可用时，高频读接口能命中短 TTL 缓存；Redis 不可用时除原短信/AI配置强依赖路径外均回源 MySQL。
2. 商品、服务资源、权限变更后会删除相关缓存前缀。
3. refresh token 成功换发后，旧 refresh token 在有效期内不可重复使用。
4. `go test ./...` 与 `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 通过。
