# SL-BE-PRODUCT-STRICT-ONSALE Spec

## 基本信息

- Feature ID: SL-BE-PRODUCT-STRICT-ONSALE
- Feature 名称: 商品可售池严格按上架状态过滤
- 负责人: Codex
- 目标上线环境: Backend production

## 背景

前台商城商品列表默认查询已上架商品，但仓储层对 `status=2` 的查询保留了历史兜底逻辑：只要 `approved_at IS NOT NULL` 且状态不是下架/拒绝，也会被视为可售。SQL 脚本导入数据时如果写入 `status=1` 且带有 `approved_at`，这些待审核商品会进入商城展示。

## 范围

### In Scope

- 商品列表按请求状态精确过滤。
- 前台默认可售池只返回 `products.status = 2`。
- 保留显式 `enterprise_id` 或状态查询能力，但状态语义不再混用。

### Out of Scope

- 不自动修改生产库商品状态。
- 不删除商品记录。
- 不调整管理端商品审核接口。

## 接口与数据影响

- 受影响路由/handler: `GET /api/v1/products`
- 受影响服务: ProductService 列表查询
- 受影响 middleware: 无
- 受影响表/model: `products.status`
- 配置或环境变量变化: 无
- Docker / 部署影响: 需要部署 backend

## 权限与合规影响

- 受影响角色: 访客、买家、卖家、运营
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 影响商品审核后的展示边界，不改变审核写入与日志。

## 验收标准

1. 未传 `enterprise_id/status` 的公开商品列表只返回 `status=2` 商品。
2. `status=1` 且 `approved_at IS NOT NULL` 的异常数据不会进入公开可售列表。
3. 显式传 `status=1` 时仍可查询待审核商品，供管理或排查场景使用。
