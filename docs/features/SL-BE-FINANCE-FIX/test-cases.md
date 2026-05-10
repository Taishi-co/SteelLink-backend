# Test Cases

## 测试范围

- Feature ID: SL-BE-FINANCE-FIX
- 测试负责人: backend

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 审核融资商品后买家可见 | 融资商品 status=pending | 运营端审核通过 → 买家端查询融资列表 | 列表包含该商品，无需等待缓存过期 |
| TC-02 | 审核仓储服务后可见 | 仓储服务 status=pending | 运营端审核通过 → 买家端查询仓储列表 | 列表包含该服务 |
| TC-03 | 审核物流服务后可见 | 物流服务 status=pending | 运营端审核通过 → 买家端查询物流列表 | 列表包含该服务 |
| TC-04 | 融资放款扣减库存 | 订单融资状态=buyer_signed_fee_paid | 资金方放款 | 订单商品 LockedQty-qty, Quantity-qty |
| TC-05 | 融资申请校验场景 | 融资产品 supported_scenarios=["钢材采购"] | 买家申请融资，订单不含匹配场景 | 返回错误 |
| TC-06 | RecordPledge 后状态流转 | 融资状态=disbursed | 资金方记录货权质押 | 订单状态变为 PendingDelivery(5) |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 风控拒绝-库存已锁定 | 订单库存 LockedQty>=qty | 资金方风控拒绝 | UnlockStock，LockedQty 减少 |
| TC-EX-02 | 风控拒绝-库存已扣减 | 订单库存 LockedQty<qty | 资金方风控拒绝 | ReturnStock，Quantity+qty, AvailableQty+qty |
| TC-EX-03 | 融资申请场景不匹配 | 融资产品场景不包含订单用途 | 买家申请融资 | 返回参数错误 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 运营端 | 审核融资商品 | 审核成功，缓存刷新 |
| TC-AUTH-02 | 非运营端 | 尝试审核融资商品 | 403 Forbidden |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 构建通过 | `CGO_ENABLED=0 go build -o /tmp/steelink-api ./cmd/api` | 构建成功 |
| TC-OPS-02 | 测试通过 | `go test ./...` | 全部通过 |
