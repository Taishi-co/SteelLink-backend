# 交易履约流程设计

## 定位

交易履约是平台主链路，围绕商品、订单、货权、合同、资金、仓储、物流形成闭环。

核心业务原则：

- 货权质押是平台核心能力，货权流转、出库、签收、放款必须有明确凭证和事件留痕。
- 仓储方、物流方、资金方是平台合作履约方，对其确认、承运、放款等动作承担对应责任。
- 平台不直接沉淀客户资金，资金动作以支付确认、融资请求、资金方审批和放款结果同步为准。
- 企业权限仍按 5 类平台服务控制：销售、采购、物流、仓储、金融。

## 角色与权限边界

| 角色 | 平台服务 | 主要职责 |
| --- | --- | --- |
| 买方 | `procurement` | 询价、下单、确认仓库、签署合同、支付/申请托盘、确认收货 |
| 卖方 | `sales` | 商品上架、订单确认、签署合同、发起物流竞价、对账确认 |
| 仓储方 | `warehouse` | 维护仓库资源、核验仓库条件、出库交接、货权凭证协同 |
| 物流方 | `logistics` | 维护车辆/司机资质、参与竞价、执行提货运输、回传签收 |
| 资金方 | `finance` | 授信审批、债务合同、手续费确认、放款同步 |
| 平台运营 | 后台权限 | 商品审核、资源审核、异常处理、流程审计 |

## 主流程阶段

### 阶段 1：商品上架与审核

卖方上架商品，商品可以是现货或期货，必须包含仓库地址或交付仓库信息。

平台审核重点：

- 商品基础信息完整性。
- 企业销售服务权限是否开通。
- 商品关联仓库是否可识别。
- 如商品绑定合作仓储方，仓储方权限和仓库资源是否有效。

结果：

- 审核通过后商品进入可售状态。
- 驳回后卖方可修改重提。

### 阶段 2.1：智能询价核算

买方输入：

- 是否自配车队。
- 收货仓库或收货地址。
- 订单明细。
- 期望交付条件。

询价系统输出：

- 可组合订单方案。
- 托盘价和直采价对比。
- 预计物流成本。
- 可选仓库/物流匹配结果。

买方确认意向后进入下单流程。

### 阶段 2.2：下单、仓库确认与合同签署

买方选择商品发起订单后，平台推送商品关联仓库和仓储方信息。

关键动作：

- 买方确认仓储方/仓库是否满足需求。
- 仓储方核验仓库条件并反馈。
- 平台把订单和仓库确认信息推送给卖方。
- 卖方确认价格、数量、交付条款。
- 买方最终确认订单。
- 买卖双方线上签署电子合同，合同应包含物流、仓储、货权和违约责任条款。

### 阶段 2.3：货款结算或托盘垫资

支持两种支付路径：

- 直采支付：买方向卖方支付对应货款。
- 托盘支付：买方选择资金方和账期，平台同步风控数据，资金方审批授信，买方确认后签署债务合同并支付手续费，资金方向卖方放款。

托盘模式必须记录：

- 资金方。
- 授信金额。
- 账期。
- 定价/费率。
- 风控审批结果。
- 债务合同。
- 手续费支付结果。
- 放款结果。

### 阶段 3：物流竞价与运输交付

卖方发起物流车队竞价需求。

物流方提交司机、车辆、承运资质参与竞价。平台做合规筛选后，生成候选车队列表并推送给买卖双方。

锁定物流资源的条件：

- 买方确认。
- 卖方确认。
- 两方确认结果一致。

锁定后平台向物流方下发运输任务和指定提货仓库。物流方到仓库提货出库，运输至买方目的地。买方现场验货、实物签收并在线确认收货。

### 阶段 4：对账归档

平台同步收货完成状态给卖方。卖方完成收款对账确认后，订单进入归档。

归档数据包括：

- 商品和订单快照。
- 仓库确认记录。
- 电子合同。
- 支付或融资记录。
- 物流竞价和运输记录。
- 出库、签收、对账事件。

## 订单状态机

建议主订单状态如下：

| 状态 | 含义 | 主要操作 |
| --- | --- | --- |
| `draft` | 草稿/意向单 | 买方确认询价方案 |
| `pending_warehouse_confirm` | 待仓库确认 | 买方选择仓库，仓储方核验 |
| `pending_seller_confirm` | 待卖方确认 | 卖方确认价格、数量、交付条款 |
| `pending_buyer_confirm` | 待买方最终确认 | 买方确认订单最终版本 |
| `pending_contract` | 待签署合同 | 买卖双方签署电子合同 |
| `pending_payment` | 待支付/待融资选择 | 买方直采支付或申请托盘 |
| `financing_review` | 托盘审批中 | 资金方风控审批 |
| `pending_financing_contract` | 待融资合同 | 买方签署债务合同、支付手续费 |
| `paid` | 已支付/已放款 | 可进入物流履约 |
| `logistics_bidding` | 物流竞价中 | 物流方报价，平台筛选 |
| `pending_logistics_confirm` | 待确认物流 | 买卖双方确认候选车队 |
| `shipping` | 运输中 | 物流方提货运输 |
| `delivered` | 已送达待确认 | 买方验货签收 |
| `received` | 已收货 | 平台同步卖方 |
| `reconciled` | 已对账 | 卖方确认收款对账 |
| `archived` | 已归档 | 流程完成 |
| `cancelled` | 已取消 | 归档取消原因 |
| `exception` | 异常 | 平台运营介入 |

## 子状态与附属对象

主订单状态不能承载所有细节，建议拆出附属对象：

| 对象 | 作用 |
| --- | --- |
| `order_warehouse_confirmations` | 仓库条件确认、仓储方反馈、买方确认结果 |
| `order_contracts` | 电子合同模板、买卖双方签署状态、签署文件 |
| `order_payments` | 直采支付确认、支付凭证、结算状态 |
| `financing_requests` | 托盘申请、资金方、账期、授信、审批、放款 |
| `logistics_bidding_demands` | 物流竞价需求、提货仓库、目的地、时效要求 |
| `logistics_bid_quotes` | 物流方报价、司机车辆资质、候选状态 |
| `shipment_tasks` | 提货、出库、运输、送达、签收 |
| `order_events` | 全流程操作留痕 |

## 核心数据模型草案

### 商品扩展

`products` 后续需要支持：

- `goods_type`: `spot` / `futures`
- `warehouse_id`
- `warehouse_enterprise_id`
- `warehouse_address_snapshot`
- `review_status`
- `review_reason`

### 订单主表

建议新增或扩展 `orders`：

- `order_no`
- `buyer_enterprise_id`
- `seller_enterprise_id`
- `status`
- `pricing_mode`: `direct` / `pallet`
- `warehouse_id`
- `warehouse_enterprise_id`
- `logistics_enterprise_id`
- `funding_enterprise_id`
- `total_amount`
- `freight_amount`
- `financing_amount`
- `contract_status`
- `payment_status`
- `shipment_status`

### 事件留痕

`order_events` 最少字段：

- `order_id`
- `event_type`
- `actor_user_id`
- `actor_enterprise_id`
- `actor_service`
- `from_status`
- `to_status`
- `payload`
- `created_at`

所有关键状态变更必须写事件。

## 接口草案

### 买方接口

- `POST /api/v1/inquiries/calculate`
- `POST /api/v1/orders`
- `POST /api/v1/orders/:id/warehouse-confirm`
- `POST /api/v1/orders/:id/buyer-confirm`
- `POST /api/v1/orders/:id/contracts/sign`
- `POST /api/v1/orders/:id/payments/direct`
- `POST /api/v1/orders/:id/financing-requests`
- `POST /api/v1/orders/:id/receipts/confirm`

### 卖方接口

- `POST /api/v1/products`
- `POST /api/v1/orders/:id/seller-confirm`
- `POST /api/v1/orders/:id/contracts/sign`
- `POST /api/v1/orders/:id/logistics-biddings`
- `POST /api/v1/orders/:id/reconcile`

### 仓储方接口

- `POST /api/v1/orders/:id/warehouse-check`
- `POST /api/v1/shipments/:id/outbound-confirm`

### 物流方接口

- `POST /api/v1/logistics-biddings/:id/quotes`
- `POST /api/v1/shipments/:id/pickup`
- `POST /api/v1/shipments/:id/deliver`
- `POST /api/v1/shipments/:id/signature`

### 资金方接口

- `POST /api/v1/financing-requests/:id/risk-result`
- `POST /api/v1/financing-requests/:id/contracts`
- `POST /api/v1/financing-requests/:id/disbursements`

### 管理端接口

- `GET /api/v1/admin/trade-orders`
- `GET /api/v1/admin/trade-orders/:id/events`
- `POST /api/v1/admin/products/:id/audit`
- `POST /api/v1/admin/logistics-bids/:id/audit`
- `POST /api/v1/admin/trade-orders/:id/exception`

## 后台审核与风控点

必须审核：

- 商品上架审核。
- 仓储资源审核。
- 物流车辆/司机资质审核。
- 企业服务权限审核。

需要运营可查看但不代替资金方决策：

- 托盘融资风控结果。
- 授信金额。
- 放款状态。
- 债务合同状态。

异常介入场景：

- 仓库条件不匹配。
- 买卖双方物流确认不一致。
- 资金方拒绝授信。
- 已放款但仓库不出库。
- 物流提货失败。
- 买方签收争议。

## MVP 实施建议

第一阶段：订单状态机和事件留痕。

- 建立主订单状态流转。
- 所有状态变更写入 `order_events`。
- 先支持直采支付占位，不接真实资金。

第二阶段：仓库确认和合同签署。

- 商品关联仓库。
- 买方确认仓库。
- 仓储方核验。
- 买卖双方合同签署状态。

第三阶段：物流竞价和运输任务。

- 卖方发起物流竞价。
- 物流方报价。
- 买卖双方确认同一候选车队。
- 运输任务、出库、签收。

第四阶段：托盘融资。

- 托盘申请。
- 资金方审批回传。
- 债务合同和手续费。
- 放款结果同步。

## 与现有模块的关系

- 用户体系：复用企业服务权限控制。
- 商品模块：扩展商品类型、仓库绑定、审核信息。
- 竞价模块：现有采购竞价继续保留；物流竞价建议独立建模，不和采购竞价混用。
- 订单模块：作为履约主状态机承载交易闭环。
- 服务资源模块：仓储/物流资源上架和审核结果作为履约准入数据。
