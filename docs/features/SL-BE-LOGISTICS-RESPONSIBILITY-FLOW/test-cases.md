# 测试用例

| ID | 场景 | 前置 | 步骤 | 预期 |
| --- | --- | --- | --- | --- |
| TC-01 | 创建物流履约单默认状态 | 存在物流服务资源 | 创建物流委托 | `logistics_orders.status=task_created`，责任主体为 `logistics` |
| TC-02 | 提货预约 | 物流单处于 `task_created` | 更新状态为 `pickup_scheduled` | 状态推进成功，责任主体切换为 `warehouse` |
| TC-03 | 仓储签到 | 物流单处于 `pickup_scheduled` | 更新状态为 `arrival_signed_in` | 记录签到时间 |
| TC-04 | 出库确认 | 物流单处于 `arrival_signed_in` | 更新状态为 `outbound_confirmed` | 责任主体仍为 `warehouse` |
| TC-05 | 装货确认 | 物流单处于 `outbound_confirmed` | 更新状态为 `loading_confirmed` | 责任主体切换为 `logistics` |
| TC-06 | 在途运输 | 物流单处于 `loading_confirmed` | 更新状态为 `in_transit` | 轨迹摘要显示运输中 |
| TC-07 | 到货确认 | 物流单处于 `in_transit` | 更新状态为 `arrival_confirmed` | 责任主体切换为 `buyer` |
| TC-08 | 验收签收 | 物流单处于 `arrival_confirmed` | 更新状态为 `receipt_signed` | 返回签收时间，买卖家摘要可见 |
| TC-09 | 入库完成 | 物流单处于 `receipt_signed` | 更新状态为 `inbound_completed` | 责任主体切换为 `warehouse` |
| TC-10 | 履约摘要 | 订单已关联物流单 | 查询 `/orders/:id/service-fulfillment` | 返回细状态、责任主体、关键节点时间和轨迹 |
| TC-11 | 兼容旧状态 | 库存中存在旧物流状态 | 查询履约摘要 | 接口仍可返回并被前端兼容 |

## 异常用例

| ID | 场景 | 前置 | 步骤 | 预期 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 无权限推进物流状态 | 非承运方企业 | 调用状态更新接口 | 返回 403 |
| TC-EX-02 | 无效状态推进 | 当前状态与目标状态不合法 | 调用状态更新接口 | 返回 400 |
