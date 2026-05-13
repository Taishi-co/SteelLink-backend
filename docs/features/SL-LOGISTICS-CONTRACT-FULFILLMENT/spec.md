# SL-LOGISTICS-CONTRACT-FULFILLMENT

## 目标

在物流竞价选定报价后，增加物流合同签署环节，完善履约单据上传和物流费用结算流程，并将融资流程步骤在订单 timeline 中完整展示。

## 新增订单状态

- `OrderStatusPendingLogisticsContract` (27): 待物流合同签署

## 新增实体

### LogisticsContract（物流合同）
- 买方和物流方双方签署，签署顺序不限
- 双方签完后自动触发 Order 27 → 18（运输中）+ 创建 ShipmentTask
- `IsFinancingOrder` 字段区分融资/非融资路径

### FulfillmentDocument（履约单据）
- 9 种单据类型：提货单、出库单、铅封照片、GPS轨迹、验货结果、收货单、入库单、运单、送达照片
- 各角色按权限上传（物流/仓库/买方）

### LogisticsSettlement（物流结算）
- 买方确认收货时自动创建
- 状态：pending → paid → committed
- 运费、平台抽佣（3%）、实付金额

## 新增 API

| 端点 | 方法 | 角色 |
|------|------|------|
| `/orders/:id/logistics-contracts` | GET | 所有参与方 |
| `/orders/:id/logistics-contracts/buyer-sign` | POST | 买方 |
| `/orders/:id/logistics-contracts/logistics-sign` | POST | 物流方 |
| `/orders/:id/logistics-settlements` | GET | 所有参与方 |
| `/orders/:id/logistics-settlements/pay` | POST | 买方 |
| `/orders/:id/logistics-settlements/commission` | POST | 买方 |
| `/logistics/orders/:id/documents` | POST | 物流/仓库/买方 |
| `/logistics/orders/:id/documents` | GET | 所有参与方 |
| `/logistics/orders/:id/seal-confirm` | POST | 仓库 |
| `/logistics/orders/:id/inspection` | POST | 买方/仓库 |

## 融资流程优化

Timeline 展开为完整步骤链：
1. 风控审批通过
2. 资金方生成合同
3. 买方签署并支付手续费
4. 资金方放款
5. 货权质押记录

## 其他修复

- WarehouseConfirm 权限改为目的地仓确认
- ListByWarehouse 同时匹配提货仓和目的地仓
- ConfirmReceipt 允许资方确认收货（融资路径）
- 订单列表 N+1 性能修复（10s → 0.6s）
- 新增 fee_voucher 上传 purpose
