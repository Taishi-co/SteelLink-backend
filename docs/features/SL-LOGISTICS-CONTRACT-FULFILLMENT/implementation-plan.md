# Implementation Plan

## Phase 1: 物流合同签署（后端）
1. 新增 OrderStatusPendingLogisticsContract (27) + transition map
2. 新增 LogisticsContract 实体 + migration
3. 修改 ConfirmLogisticsQuote：选定报价后创建合同，Order → 27
4. 新增合同签署 handler（buyer-sign / logistics-sign）
5. 双方签完自动 Order 27 → 18 + 创建 ShipmentTask + 同步 LogisticsOrder

## Phase 2: 履约单据（后端）
1. 新增 FulfillmentDocument 实体 + migration
2. 新增单据上传/查询 handler
3. 新增铅封确认、验货结果端点
4. 路由注册到 servicefulfillment 模块

## Phase 3: 物流结算（后端）
1. 新增 LogisticsSettlement 实体 + migration
2. 买方确认收货时自动创建结算单
3. 新增支付/抽佣 handler
4. 路由注册

## Phase 4: 融资流程优化
1. Timeline 融资步骤展开（基于 order_events 判断）
2. WarehouseConfirm 权限改为目的地仓
3. ListByWarehouse 同时匹配提货仓和目的地仓
4. ConfirmReceipt 允许资方确认收货
5. 新增 fee_voucher 上传 purpose

## Phase 5: 前端集成
1. buyer-app: 融资流程 UI 修正（status 23/24 映射、按钮触发）
2. finance-app: 合同与放款 tab 融资进度列
3. buyer-app: 卖家跳转 switch_token 免登录

## 关键文件
- `internal/modules/order/domain/entity/order.go`
- `internal/modules/order/application/service/order_service.go`
- `internal/modules/order/ui/handler/order_handler.go`
- `internal/modules/order/infrastructure/router/router.go`
- `internal/modules/servicefulfillment/domain/entity/entity.go`
- `internal/modules/servicefulfillment/ui/handler/handler.go`
- `internal/pkg/storage/cos.go`
- `internal/pkg/storage/validate.go`
