# Test Cases: AI Procurement Order Grouping And Logistics Visibility Fix

| ID | Scenario | Steps | Expected |
| --- | --- | --- | --- |
| TC-01 | Same seller and same warehouse | Select two AI procurement products from the same seller and warehouse, each with one ton, then generate order | One trade order is created with two items |
| TC-02 | Same seller but different warehouse | Select two AI procurement products from the same seller but different warehouses | Two trade orders are created |
| TC-03 | Matched logistics visibility | Generate AI procurement order with selected `logistics_service_id` | A `logistics_orders` row is created for every generated trade order |
| TC-04 | Negotiated logistics | Generate AI procurement order with `logistics_match_status=negotiation` or service name `协商解决` | Trade order is created, no carrier-specific `logistics_orders` row is created |
| TC-05 | Regression build | Run `go test ./internal/modules/aiprocurement/application/service` and `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` | Commands pass |
