# 物流履约责任流拆分

## 背景

当前交易订单主链路已经能表达高层状态，但物流履约仍主要依赖 `orders` 内的 `shipment_task` 和少量状态联动，难以承载重工业交付中的责任交接、仓储协同和多节点留痕。

现有 `servicefulfillment` 模块已具备独立的 `logistics_orders`、`logistics_tracking_events`、`warehouse_inventory_movements` 与 `fulfillment_guarantees` 基础能力，适合作为独立物流履约子系统继续演进。

## 目标

1. 保持订单主状态机不新增、不重命名、不拆分细粒度物流节点。
2. 将物流履约细状态迁移到 `servicefulfillment.logistics_orders`。
3. 让物流履约能表达仓储、物流、买方之间的责任转移。
4. 买家、卖家都能通过独立履约摘要观测进度，而不依赖订单状态细节。

## 非目标

- 不重构 `orders` 主状态机。
- 不替换现有 `shipment_task` 相关接口，先保持兼容。
- 不在本轮引入新的独立工作台应用或新的管理端主流程。

## 目标流程

物流履约状态改为以下细粒度流程：

1. `task_created` 创建运输任务
2. `pickup_scheduled` 提货预约
3. `arrival_signed_in` 到场签到
4. `outbound_confirmed` 出库确认
5. `loading_confirmed` 装货确认
6. `in_transit` 运输中
7. `arrival_confirmed` 到货确认
8. `receipt_signed` 验收签收
9. `inbound_completed` 入库完成
10. `completed` 履约完成

责任主体按节点映射：

| 状态 | 当前责任主体 |
| --- | --- |
| `task_created` | logistics |
| `pickup_scheduled` | warehouse |
| `arrival_signed_in` | warehouse |
| `outbound_confirmed` | warehouse |
| `loading_confirmed` | logistics |
| `in_transit` | logistics |
| `arrival_confirmed` | buyer |
| `receipt_signed` | buyer |
| `inbound_completed` | warehouse |
| `completed` | buyer |

## 设计约束

1. `orders.status` 只保留高层交易结果态，物流细节点不再映射为新增订单状态。
2. `logistics_orders.status` 成为物流履约唯一细粒度状态。
3. `tracking_events` 继续保留，但应围绕上述责任节点生成更明确的事件语义。
4. `GET /api/v1/orders/:id/service-fulfillment` 作为买卖家观测入口，返回更丰富的物流履约摘要。
5. 物流服务商仍通过 `POST /api/v1/logistics/orders/:id/accept`、`PUT /api/v1/logistics/orders/:id/status`、`POST /api/v1/logistics/orders/:id/tracking-events` 推进履约，接口优先兼容。

## 接口与数据变更

### 数据模型

扩展 `logistics_orders`：

- `status` 改为支持新的细粒度状态集合
- 新增 `current_responsible_party`
- 新增 `handover_note`
- 新增 `pickup_appointment_at`
- 新增 `arrival_signed_in_at`
- 新增 `outbound_confirmed_at`
- 新增 `loading_confirmed_at`
- 新增 `arrival_confirmed_at`
- 新增 `receipt_signed_at`
- 新增 `inbound_completed_at`

### 摘要接口

`GET /api/v1/orders/:id/service-fulfillment` 返回：

- 物流单当前细状态
- 当前责任主体
- 关键节点时间
- 轨迹事件
- 仓储动作与物流动作的组合摘要

### 状态推进

`PUT /api/v1/logistics/orders/:id/status` 支持上述细状态推进，并根据状态自动：

- 维护责任主体
- 记录关键时间字段
- 在需要时补写轨迹事件

## 风险与兼容

1. 买卖家现有页面仍读取 `shipment_task`，本轮需同时补充 `servicefulfillment` 摘要展示，避免旧页面失效。
2. AI 采购与服务市场入口会写入 `logistics_orders`，新状态必须兼容其默认创建逻辑。
3. 旧数据中的 `pending/accepted/delivered/completed` 仍需被前端兼容展示。
