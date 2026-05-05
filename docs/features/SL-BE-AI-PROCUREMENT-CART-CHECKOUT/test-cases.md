# AI Procurement Cart Checkout Test Cases

## 测试范围

- Feature ID: SL-BE-AI-PROCUREMENT-CART-CHECKOUT
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | AI采购单商品下单 | 企业用户已认证，商品库存充足 | 调用 `generate-order` 传 `selected_product` | 生成订单并锁定库存 |
| TC-02 | AI采购拼单下单 | 两个同卖家商品库存充足 | 调用 `generate-order` 传 `selected_products` | 生成包含两个订单项的订单 |
| TC-03 | 取消回填购物车 | TC-02 已生成订单 | 调用 `/orders/:id/cancel` | 库存解锁，两个订单项回填购物车 |
| TC-04 | 购物车接口列表 | 企业用户已登录 | 调用 `/cart/items` | 返回当前用户购物车明细 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未选择商品 | 企业用户已登录 | 调用 `generate-order` 且不传商品 | 返回参数错误 |
| TC-EX-02 | 跨卖家拼单 | 两个商品来自不同卖家 | 调用 `generate-order` 传两个商品 | 返回不支持跨卖家拼单 |
| TC-EX-03 | 库存不足 | 商品可用库存小于下单量 | 调用 `generate-order` | 返回库存不足，不生成订单 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 访问 `/cart/items` | 返回未登录 |
| TC-AUTH-02 | 非企业用户 | 调用购物车或AI采购下单 | 被企业账号鉴权拦截 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | Go 测试 | `go test ./...` | 通过 |
| TC-OPS-02 | API 构建 | `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` | 通过 |
| TC-OPS-03 | Release gate | `./harness/release.sh docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT` | 通过 |
