# 物流全链路状态流转

## 背景

卖家端发起物流竞价后，无法查看竞价列表了解竞价情况，也无法结束竞价。物流竞价完成后，物流信息未能自动填写到订单流程中，导致物流状态与订单状态脱节，买卖家无法追踪物流进度。

## 需求

1. 卖家端可查看物流竞价列表、选择报价、关闭竞价
2. 物流确认接单后，司机/车辆/地址等信息自动同步到订单 delivery_info
3. 物流确认提货 → 买卖家同步看到"物流已提货"
4. 物流确认送达 → 买家可确认收货
5. 买家确认收货 → 卖家可对账确认
6. 三端(买家/卖家/物流)物流进度实时联动

## 状态流转

```
已支付(4) → 物流竞价中(15) → 待确认物流(16) → 运输中(17) → 已送达(7) → 已收货(18) → 已对账(19)
```

| 角色 | 操作 | 订单状态变化 | shipment_task状态 |
|------|------|-------------|------------------|
| 卖家 | 发起物流竞价 | → 15(物流竞价中) | - |
| 物流 | 提交报价 | 15 不变 | - |
| 卖家 | 选择报价 | → 16(待确认物流) | - |
| 物流 | 确认接单 | → 17(运输中) | pending_pickup |
| 物流 | 确认提货 | 17 不变 | picked_up |
| 物流 | 确认送达 | → 7(已送达) | delivered |
| 买家 | 确认收货 | → 18(已收货) | signed |
| 卖家 | 对账确认 | → 19(已对账) | signed |

## 新增API

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | /orders/:id/logistics-biddings/close | seller | 卖家关闭竞价 |
| GET | /orders/:id/logistics-quotes | buyer/seller | 查看报价列表 |
| POST | /orders/:id/logistics-quotes/select | seller | 卖家选择报价 |

## 核心逻辑变更

### ConfirmLogisticsQuote（物流确认接单）
- 自动填充 delivery_info：driver_name, plate_number, address, pickup_code
- 从竞价信息同步 destination_address 和 pickup_address
- 保留已有仓库信息
- 创建 ShipmentTask（status=pending_pickup）
- 通知买家"物流已确认接单"
- 通知物流方"报价已确认，请安排提货"

### UpdateShipmentProgress（物流进度回传）
- pickup：shipment_task → picked_up，通知买卖双方"物流已提货"
- deliver：shipment_task → delivered，订单 → 已送达(7)，通知买卖双方"物流已送达"
- signature：shipment_task → signed，通知"签收凭证已上传"（不再错误跳过买家确认收货步骤）

### ConfirmReceipt（买家确认收货）
- 订单 → 已收货(18)
- shipment_task → signed

### ReconcileOrder（卖家对账）
- 订单 → 已对账(19)
