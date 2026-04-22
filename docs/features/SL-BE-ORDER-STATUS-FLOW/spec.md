# Feature Spec

## 基本信息

- Feature ID: `SL-BE-ORDER-STATUS-FLOW`
- Feature 名称: 订单合同、支付审核与提货码状态流转补齐
- 负责人: Codex
- 目标上线环境: 本地开发 / 测试联调环境

## 背景

当前主业务后端已经支持商品审核通过后的下单，但订单创建后只停留在基础状态，尚未补齐一期闭环里最关键的后续流转：

- 买家提交合同
- 买家上传支付回单
- 平台审核支付
- 平台生成提货码
- 提货码核销完成订单

这会导致前后端联调时只能验证“下单成功”，无法继续覆盖一期最重要的状态迁移逻辑。

## 范围

### In Scope

- 为订单补齐一期闭环的最小状态机。
- 新增订单合同、支付回单、提货码三类持久化模型。
- 新增买家侧合同提交、支付回单上传、提货码查看接口。
- 新增平台侧订单查询、支付审核接口。
- 新增提货码核销接口并驱动订单完结。

### Out of Scope

- 第三方电子签、OCR、GPS 水印。
- 生产级支付清结算。
- 仓库端独立角色与复杂权限模型。
- 多次支付、分批提货、异常仲裁。

## 接口与数据影响

- 受影响路由 / handler:
  - `/api/v1/orders`
  - `/api/v1/orders/:id/contract/submit`
  - `/api/v1/orders/:id/payment-receipts`
  - `/api/v1/orders/:id/pickup-code`
  - `/api/v1/orders/:id/pickup-code/verify`
  - `/api/v1/admin/orders`
  - `/api/v1/admin/orders/:id`
  - `/api/v1/admin/orders/:id/payment-audit`
- 受影响服务:
  - `internal/modules/trade/application/service`
- 受影响 middleware:
  - 无新增 middleware，继续复用现有鉴权
- 受影响表 / model:
  - `orders`
  - `order_items`
  - `order_contracts`
  - `payment_receipts`
  - `pickup_codes`
- 配置或环境变量变化:
  - 无
- Docker / 部署影响:
  - 仅新增 AutoMigrate 模型，无额外部署配置

## 权限与合规影响

- 受影响角色:
  - 买家
  - 卖家
  - 平台审核
- 新增 / 调整权限:
  - 买家可提交合同、支付回单、查看提货码
  - 平台可审核支付并查看订单详情
  - 买家 / 卖家可在当前简化模型下核销提货码
- 是否影响审核、风控、日志留痕:
  - 影响审核与状态流转，但暂未引入独立操作日志表

## 验收标准

1. 下单后订单初始状态为 `pending_contract`。
2. 提交合同后订单进入 `pending_payment`。
3. 上传支付回单后订单进入 `pending_payment_audit`。
4. 平台审核通过后订单进入 `pending_pickup` 并生成提货码。
5. 提货码核销成功后订单进入 `completed`。
