# Implementation Plan

## Feature ID: SL-BE-FINANCE-FIX

## 实施步骤

### Step 1: AuditFinance/AuditWarehouse/AuditLogistic 审核后刷新缓存

- 文件: `internal/modules/serviceoffer/ui/handler/handler.go`
- 改动: `auditOffer` 成功后调用 `invalidateServiceOfferCache(ctx)`
- 影响范围: 仓储/物流/融资三种服务的审核接口

### Step 2: DisburseFinancing 添加 DeductStock

- 文件: `internal/modules/order/application/service/order_service.go`
- 改动: DisburseFinancing 事务中遍历 order.Items 调用 DeductStock
- 与 UploadPaymentVoucher 中 DeductStock 逻辑一致

### Step 3: SubmitRiskResult 风控拒绝区分 UnlockStock/ReturnStock

- 文件: `internal/modules/order/application/service/order_service.go`
- 改动: 参照 cancelOrderWithUnlock 的逻辑，区分 LockedQty>=qty 和已扣减情况

### Step 4: CreateFinancingRequest 校验 SupportedScenarios

- 文件: `internal/modules/order/application/service/order_service.go`
- 改动: 查询 FinanceService 后校验 SupportedScenarios 是否与订单品类/用途匹配

### Step 5: RecordPledge 后订单状态流转

- 文件: `internal/modules/order/application/service/order_service.go`
- 改动: RecordPledge 成功后将订单状态改为 PendingDelivery

### Step 6: 验证

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steelink-api ./cmd/api`

## 风险与回滚

- 缓存刷新：审核后立即生效，无回滚风险
- DeductStock：已在 UploadPaymentVoucher 中验证，逻辑成熟
- RecordPledge 状态流转：需确认 PendingDelivery 在状态机映射中存在 → 已确认存在
- 回滚策略：git revert 即可
