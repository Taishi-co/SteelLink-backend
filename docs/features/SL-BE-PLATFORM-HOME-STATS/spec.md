# Platform Home Stats Spec

## 基本信息

- Feature ID: SL-BE-PLATFORM-HOME-STATS
- Feature 名称: 平台首页真实统计接口
- 负责人: Codex
- 目标上线环境: production

## 背景

买家端首页当前展示“入驻商家、SKU商品、累计订单、交易金额”的静态演示数字。生产首页需要读取后端真实业务数据，避免页面数字与数据库实际状态不一致。

## 范围

### In Scope

- 新增公开只读统计接口 `GET /api/v1/platform/stats`。
- 统计已认证企业数、已上架 SKU 数、有效订单累计数和有效订单累计金额。
- 接口使用统一 `{ code, message, data }` 响应格式。

### Out of Scope

- 不新增数据库表或迁移。
- 不调整管理端数据概览。
- 不调整订单状态机或商品审核流程。

## 接口与数据影响

- 受影响路由/handler: `cmd/api/main.go`, `internal/pkg/handler/platform_stats.go`
- 受影响服务: 无独立 service，handler 内执行只读聚合查询。
- 受影响 middleware: 无，接口公开访问。
- 受影响表/model: `enterprises`, `products`, `orders`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 游客、买家、卖家均可读取首页聚合统计。
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不影响，仅聚合公开统计。

## 统计口径

- 入驻商家: `enterprises.status = approved`。
- SKU商品: `products.status = on_sale`。
- 累计订单: 排除草稿和已取消订单。
- 交易金额: 汇总排除草稿和已取消订单后的 `orders.total_amount`。

## 验收标准

1. `GET /api/v1/platform/stats` 无需登录即可返回四项统计。
2. 数据库查询失败时接口返回统一内部错误响应。
3. 不引入数据库结构、配置或权限变更。
