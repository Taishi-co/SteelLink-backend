# Feature Spec

## 基本信息

- Feature ID: SL-BE-FINANCE-FIX
- Feature 名称: 托盘垫资流程 bug 修复
- 负责人: backend
- 目标上线环境: staging → prod

## 背景

买家在订单支付环节选择申请融资后，新上架的融资商品不会在列表中显示。经 review 发现以下问题：

1. AuditFinance 审核通过后未刷新缓存，导致买家端继续命中旧缓存看到空列表
2. DisburseFinancing 缺少 DeductStock，融资路径库存不扣减
3. SubmitRiskResult 风控拒绝时只 UnlockStock 不区分已扣减情况
4. CreateFinancingRequest 不校验融资产品的 SupportedScenarios
5. RecordPledge 后订单状态未流转，缺少进入物流履约的明确信号

## 范围

### In Scope

- AuditFinance / AuditWarehouse / AuditLogistics 审核后刷新缓存
- DisburseFinancing 添加 DeductStock
- SubmitRiskResult 风控拒绝区分 UnlockStock / ReturnStock
- CreateFinancingRequest 校验 SupportedScenarios
- RecordPledge 后订单状态流转到 PendingDelivery

### Out of Scope

- 融资审核通知机制（需要 notification 模块扩展，另开 feature）
- 资金方定向打款至卖方凭证追踪（需要新增字段，另开 feature）
- 前端改动

## 接口与数据影响

- 受影响路由/handler:
  - PUT /warehouse-services/:id/audit → 缓存刷新
  - PUT /logistics-services/:id/audit → 缓存刷新
  - PUT /finance-services/:id/audit → 缓存刷新
  - POST /orders/:id/financing-requests/disburse → 新增 DeductStock
  - POST /orders/:id/financing-requests/risk-result → 区分 UnlockStock/ReturnStock
  - POST /orders/:id/financing-requests → 校验 SupportedScenarios
  - POST /orders/:id/pledges → 订单状态流转
- 受影响服务: OrderService, Handler(serviceoffer)
- 受影响 middleware: 无
- 受影响表/model: finance_services(缓存), order_financing_requests, orders
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 资金方、运营端、买家
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 审核后缓存刷新确保数据一致性

## 验收标准

1. 运营端审核融资/仓储/物流服务后，买家端立即能查询到最新已审核数据
2. 融资放款后库存正确扣减（LockedQty 减少、Quantity 减少）
3. 风控拒绝时：库存锁定则 UnlockStock，已扣减则 ReturnStock
4. 融资申请时校验 SupportedScenarios，不匹配则拒绝
5. RecordPledge 后订单状态流转到 PendingDelivery，可继续物流履约
6. go test ./... 和 go build 通过
