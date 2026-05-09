# Test Cases

| ID | 场景 | 前置条件 | 操作 | 预期 |
| --- | --- | --- | --- | --- |
| TC-01 | 仓储方查看参与订单详情 | 订单 `delivery_info.warehouse_enterprise_id` 等于当前企业 ID | `GET /api/v1/orders/:id` | 返回订单详情 |
| TC-02 | 非参与企业查看详情 | 当前企业不是买方、卖方、仓储方、资金方 | `GET /api/v1/orders/:id` | 返回 403 无权查看 |
| TC-03 | 买卖资方原有权限 | 当前企业是买方、卖方或资金方 | `GET /api/v1/orders/:id` | 保持可查看 |

## 自动验证

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/verify.sh`
