# 上线检查

- [ ] 已确认仅扩展 `servicefulfillment.logistics_orders`，不改订单主状态枚举
- [ ] 已确认 `GET /orders/:id/service-fulfillment` 返回结构变更对前端兼容
- [ ] 已执行 `go test ./...`
- [ ] 已执行 `CGO_ENABLED=0 go build -o /tmp/steellink-backend ./cmd/api`
- [ ] 已记录仍需人工关注的旧物流状态兼容情况
