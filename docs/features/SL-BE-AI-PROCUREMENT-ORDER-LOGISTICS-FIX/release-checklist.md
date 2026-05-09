# Release Checklist: AI Procurement Order Grouping And Logistics Visibility Fix

- [x] `go test ./internal/modules/aiprocurement/application/service` passes.
- [x] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` passes.
- [ ] Confirm AI procurement same-seller/same-warehouse selections generate one trade order.
- [ ] Confirm logistics workspace shows one logistics fulfillment row per generated order with matched logistics service.
- [ ] Confirm negotiated logistics remains a non-carrier-specific flow.
