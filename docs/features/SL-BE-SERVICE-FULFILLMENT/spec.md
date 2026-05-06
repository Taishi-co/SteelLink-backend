# Feature Spec

## 基本信息

- Feature ID: SL-BE-SERVICE-FULFILLMENT
- Feature 名称: 物流仓储履约后端适配
- 负责人: Codex
- 目标上线环境: 后端预览环境

## 背景

前端已经补齐物流与仓储业务闭环：卖家商品可绑定仓库，买家下单时根据仓库推荐物流商，无仓商品可转物流平台找车，订单详情需要展示仓库出库、物流接单、在途追踪、签收和履约保障。后端当前已有商品 `warehouse_info`、订单 `delivery_info`、服务资源上架接口，但缺少仓库/物流服务资源 CRUD、商品仓库绑定 API、出入库记录、物流委托、轨迹和保障数据能力。

## 范围

### In Scope

- 扩展 `serviceoffer` 仓储/物流服务资源，支持创建、查询、更新、删除/下线，满足仓库管理和物流商管理 CRUD。
- 新增物流仓储履约模块 `servicefulfillment`。
- 支持商品绑定平台登记仓库，并可查询某仓库绑定的商品。
- 支持仓储入库/出库记录创建、查询和状态更新。
- 支持基于仓库/目的地/重量/物流服务的订单价格估算。
- 支持客户物流发单、物流商接单、运输状态更新和轨迹节点追加。
- 支持订单级履约保障记录创建、查询和状态更新。
- 对接订单详情，提供订单物流仓储履约摘要接口。

### Out of Scope

- 不实现真实地图轨迹、司机 GPS 设备接入、保险公司外部理赔接口。
- 不改前端页面，本 feature 只补后端 API 与数据表。
- 不做复杂仓库多库位库存批次算法，当前记录到商品、捆包号、仓库和数量粒度。

## 接口与数据影响

- 受影响路由/handler:
  - `PUT /api/v1/warehouse-services/:id`
  - `DELETE /api/v1/warehouse-services/:id`
  - `PUT /api/v1/logistics-services/:id`
  - `DELETE /api/v1/logistics-services/:id`
  - `PUT /api/v1/products/:id/warehouse-binding`
  - `GET /api/v1/warehouses/:id/products`
  - `GET /api/v1/warehouse-movements`
  - `POST /api/v1/warehouse-movements`
  - `PUT /api/v1/warehouse-movements/:id/status`
  - `POST /api/v1/logistics/estimate`
  - `GET /api/v1/logistics/orders`
  - `POST /api/v1/logistics/orders`
  - `POST /api/v1/logistics/orders/:id/accept`
  - `PUT /api/v1/logistics/orders/:id/status`
  - `POST /api/v1/logistics/orders/:id/tracking-events`
  - `GET /api/v1/orders/:id/service-fulfillment`
  - `POST /api/v1/fulfillment-guarantees`
  - `PUT /api/v1/fulfillment-guarantees/:id/status`
- 受影响服务:
  - `internal/modules/serviceoffer`
  - 新增 `internal/modules/servicefulfillment`
- 受影响 middleware:
- 受影响 middleware: 复用 `AuthRequired`、`RequireEnterpriseService`
- 受影响表/model:
  - `warehouse_services`
  - `logistics_services`
  - 新增 `warehouse_inventory_movements`
  - 新增 `logistics_orders`
  - 新增 `logistics_tracking_events`
  - 新增 `fulfillment_guarantees`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买家、卖家、物流商、仓储商
- 新增/调整权限:
  - 仓储资源与入出库记录要求企业开通 `warehouse` 服务。
  - 物流资源、物流委托接单、轨迹要求企业开通 `logistics` 服务；买家发单可通过 `procurement` 服务。
  - 商品仓库绑定要求卖家企业拥有该商品。
- 是否影响审核、风控、日志留痕: 影响履约证据留存；出入库、物流轨迹和保障记录后续可作为风控、仲裁和融资审核依据。

## 验收标准

1. 仓储/物流服务资源支持创建、查询、更新和下线。
2. 卖家可将商品绑定到平台登记仓库，仓库可查询绑定商品。
3. 仓储方可创建入库/出库记录并更新作业状态。
4. 买家可创建物流委托并获得价格估算。
5. 物流商可接单、更新运输状态并追加轨迹节点。
6. 订单详情可查询物流仓储履约摘要和保障状态。
7. `./harness/verify.sh`、review gate、release gate 通过。
