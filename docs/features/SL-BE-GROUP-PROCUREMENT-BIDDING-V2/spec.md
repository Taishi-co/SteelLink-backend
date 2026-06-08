# Feature Spec

## 基本信息

- Feature ID: `SL-BE-GROUP-PROCUREMENT-BIDDING-V2`
- Feature 名称: 集采竞价 V2 后端闭环
- 负责人: Codex
- 目标上线环境: staging

## 背景

当前集采后端是“平台选商品 -> 单卖家确认价格 -> 买家参团 -> 管理端统一成单”的轻量团购模型。该模型不支持多个供货商围绕同一集采需求报名和报价，无法满足集采平台内“卖家参与竞价”的核心业务。

本次改造将集采升级为项目制竞价闭环，但控制在可落地的第一期范围内：保留买家申请、平台归集、订单履约能力，补齐供应商报名、报价、平台定标和基于中标报价成单的主链路。

## 范围

### In Scope

- 新增“集采项目”层，允许一个项目关联多个买家集采申请。
- 新增供应商报名、邀请、审核和报价数据结构。
- 支持管理端创建集采竞价项目、邀请供应商、查看报价、手工定标。
- 支持卖家查看可参与项目、报名项目、提交/更新一轮报价。
- 支持买家查看项目详情、项目进度、已归集需求和中标结果摘要。
- 支持基于单供应商中标结果生成集采订单。
- 保持旧 `procurement_activities` 公开列表接口可兼容运行，并为旧数据保留读路径。

### Out of Scope

- 多轮报价与暗标解密。
- 多供应商分量中标与按供应商拆多主单。
- 电子合同签署、金融自动放款、售后评价、归档报告。
- 牵头方独立工作台与 AI 自动评标。

## 接口与数据影响

- 受影响路由/handler:
  - `GET/POST /api/v1/procurements/projects`
  - `GET /api/v1/procurements/projects/:id`
  - `POST /api/v1/procurements/projects/:id/enrollments`
  - `POST /api/v1/procurements/projects/:id/quotes`
  - `GET /api/v1/procurements/projects/supplier-open`
  - `GET /api/v1/procurements/projects/seller/mine`
  - `GET /api/v1/procurements/projects/:id/quotes/mine`
  - 兼容调整现有 `procurements/requests*`
- 受影响服务:
  - `internal/modules/market/application/service`
  - `internal/modules/market/ui/handler`
  - `internal/modules/market/infrastructure/router`
- 受影响 middleware:
  - 复用 `AuthRequired` 与 `RequireTradeCapability(procurement|sales)`
- 受影响表/model:
  - 新增 `procurement_projects`
  - 新增 `procurement_project_requests`
  - 新增 `procurement_supplier_enrollments`
  - 新增 `procurement_supplier_quotes`
  - 新增 `procurement_project_awards`
  - 扩展 `procurement_logs`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无新增部署组件；需完成数据库自动迁移

## 权限与合规影响

- 受影响角色:
  - 买家企业：发起集采申请、查看项目、查看归集进度
  - 卖家企业：报名项目、提交报价
  - 管理员：创建项目、邀请供应商、审核报名、定标、成单
- 新增/调整权限:
  - 买家仍要求 `procurement`
  - 卖家仍要求 `sales`
  - 管理端复用现有后台角色权限
- 是否影响审核、风控、日志留痕:
  - 是，项目创建、邀请、报名审核、报价提交、定标、成单必须写日志

## 验收标准

1. 平台可将多个买家集采申请归并为一个竞价项目，并邀请多个供应商参与。
2. 卖家可在项目截止前报名并提交/更新报价，平台可查看报价列表并手工定标。
3. 单供应商中标后，管理端可基于中标报价生成真实集采订单，买家和卖家均能看到正确项目状态。
