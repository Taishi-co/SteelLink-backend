# Redis Business Cache Expansion Spec

## 基本信息

- Feature ID: SL-BE-REDIS-BUSINESS-CACHE
- Feature 名称: Redis 业务链路缓存与幂等控制
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

平台已完成 Redis 基础缓存、限流、平台统计和向量同步锁。下一步需要把 Redis 用到更高价值的业务链路，降低重复匹配、重复聚合查询和重复提交风险。

## 范围

### In Scope

- AI 采购商品匹配缓存按用户、企业、解析意图和向量配置生成缓存 key。
- 用户首页/工作台聚合数据缓存 60 秒。
- 企业服务开通 overview 缓存 5 分钟，并在申请、审核和状态调整后失效。
- 订单创建、AI 生成订单、AI 创建采购、AI 提交审核、支付凭证、直接付款和融资申请支持 `Idempotency-Key` 防重复提交。

### Out of Scope

- 不缓存订单详情和金融风控决策。
- 不新增数据库表或迁移。
- 不改变现有接口响应结构。

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/ai-procurements/match-products`
  - `POST /api/v1/ai-procurements/generate-order`
  - `POST /api/v1/ai-procurements`
  - `POST /api/v1/ai-procurements/:id/submit-review`
  - `GET /api/v1/users/me/dashboard`
  - `GET /api/v1/enterprises/me/services`
  - `POST /api/v1/orders`
  - `POST /api/v1/orders/:id/payment-voucher`
  - `POST /api/v1/orders/:id/payments/direct`
  - `POST /api/v1/orders/:id/financing-requests`
- 受影响服务:
  - `ProcurementService.MatchProducts`
  - `UserService.GetDashboard`
  - `EnterprisePermissionService.GetOverview`
- 受影响 middleware:
  - 新增 Redis 幂等中间件 `RedisIdempotency`
  - CORS 允许 `Idempotency-Key`
- 受影响表/model: 无
- 配置或环境变量变化: 无
- Docker / 部署影响: 无，沿用现有 Redis 连接

## 权限与合规影响

- 受影响角色: 买家企业、企业服务申请方、平台运营、服务商。
- 新增/调整权限: 无。
- 是否影响审核、风控、日志留痕: 不改变审核和风控决策，只减少重复查询和重复提交。

## 验收标准

1. AI 商品匹配缓存不会跨用户/企业复用，向量配置变化后 key 自动变化。
2. `/users/me/dashboard` 和 `/enterprises/me/services` 在 Redis 可用时命中缓存，Redis 不可用时回源数据库。
3. 带相同 `Idempotency-Key` 的重复写入请求在 5 分钟内不会重复落库。
4. 服务申请、审核和权限状态调整后会清理服务开通状态缓存。
