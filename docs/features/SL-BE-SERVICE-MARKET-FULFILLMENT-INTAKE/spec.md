# SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE Spec

## 基本信息

- Feature ID: SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE
- Feature 名称: 服务市场仓储托管需求入口
- 负责人: Codex
- 目标上线环境: prod

## 背景

服务市场不仅展示仓储和物流资源，还要支持用户基于已有货物发起真实履约需求。物流委托、履约保障、服务资源上架已具备后端接口；缺口是用户有货但没有存放地时，向平台仓库提交仓储托管需求。

## 范围

### In Scope

- 新增买家侧仓储托管请求接口，用已审核仓库创建待处理入库作业记录。
- 仓储出入库记录增加 `request_enterprise_id`，区分委托方和仓储方。
- 仓储出入库列表支持按 `role=requester|warehouse` 查询。

### Out of Scope

- 不新增独立仓储租赁合同表。
- 不实现仓储报价确认流，当前先用待处理入库作业承载需求。
- 不调整物流委托、履约保障已有表结构。

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/warehouse-storage-requests`
  - `GET /api/v1/warehouse-storage-requests`
  - `GET /api/v1/warehouse-movements?role=requester|warehouse`
- 受影响服务: `internal/modules/servicefulfillment`
- 受影响 middleware: 新接口需要登录并具备 `procurement` 或 `warehouse` 服务能力
- 受影响表/model: `warehouse_inventory_movements` 新增 `request_enterprise_id`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买家企业用户、仓储服务商
- 新增/调整权限: 新增买家侧仓储托管请求；仓储方仍通过现有作业记录处理
- 是否影响审核、风控、日志留痕: 需求以作业记录留痕，状态由仓储方后续处理

## 验收标准

1. 用户可以对已审核仓库提交仓储托管请求，返回待处理入库作业记录。
2. 记录保存委托企业 ID 和仓储企业 ID。
3. 仓储方仍可通过出入库记录接口处理状态。
