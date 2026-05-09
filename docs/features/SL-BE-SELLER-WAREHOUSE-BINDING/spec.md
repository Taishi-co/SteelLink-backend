# Seller Warehouse Binding Backend Feature Spec

## 基本信息

- Feature ID: SL-BE-SELLER-WAREHOUSE-BINDING
- Feature 名称: 卖家仓库管理与商品仓库绑定后端适配
- 负责人: Codex
- 目标上线环境: production

## 背景

商品、订单和 AI采购链路已经支持商品携带 `warehouse_service_id` 和 `warehouse_info`。卖家端也需要维护自有发货仓库并在发布商品时选择仓库，但当前仓储资源 CRUD 路由只允许 `warehouse` 服务商能力，普通销售企业无法在卖家后台维护仓库资源。

## 范围

### In Scope

- 允许具备 `sales` 交易能力的企业访问 `/warehouse-services` 的自有资源 CRUD 和 `/warehouse-services/mine`。
- 保留仓库资源创建/编辑后进入待审核状态。
- 保留商品绑定仓储资源必须审核通过的约束，避免未审核仓库进入订单和 AI采购物流估价。
- 同步权限口径文档，区分卖家维护自有发货仓库和仓储服务商经营性履约。

### Out of Scope

- 不新增仓库数据表。
- 不改变公共仓库资源列表默认只展示审核通过资源的逻辑。
- 不放开物流、金融资源的销售企业 CRUD 权限。

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/warehouse-services`
  - `GET /api/v1/warehouse-services/mine`
  - `PUT /api/v1/warehouse-services/:id`
  - `DELETE /api/v1/warehouse-services/:id`
- 受影响服务: serviceoffer 仓储资源路由。
- 受影响 middleware: 仓储资源写接口从仅 `RequireProviderService("warehouse")` 调整为 `RequireEnterpriseService("sales", "warehouse")`。
- 受影响表/model: 无结构变化，继续使用 `warehouse_services`。
- 配置或环境变量变化: 无。
- Docker / 部署影响: 无。

## 权限与合规影响

- 受影响角色: 销售企业、仓储服务商企业。
- 新增/调整权限: 仓库资源自有 CRUD 从仅 `warehouse` 扩展为 `sales` 或 `warehouse`；商品绑定已审核仓储资源仍不要求卖家开通 `warehouse`。
- 是否影响审核、风控、日志留痕: 仓库创建/编辑仍进入待审核，审核通过前不能绑定商品进入交易链路。

## 验收标准

1. 具备 `sales` 能力的卖家可创建、查询、编辑、停用自有仓库资源。
2. 新建或编辑仓库资源后状态为待审核。
3. 公共仓库资源列表仍默认只展示审核通过资源。
4. 商品绑定未审核仓库资源仍返回错误，避免未审核仓库进入交易履约。
