# Test Cases

## 测试范围

- Feature ID: SL-BE-SELLER-PRODUCT-CRUD
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 编辑自有商品 | 登录卖家且商品属于当前企业 | 调用 `PUT /api/v1/products/:id` | 返回更新后的商品，状态为待审核 |
| TC-02 | 删除自有商品 | 登录卖家且商品无锁定库存 | 调用 `DELETE /api/v1/products/:id` | 返回删除成功，后续列表不再展示该商品 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 缺少主图 | 编辑请求 `images` 为空 | 调用 `PUT /products/:id` | 返回请至少保留 1 张商品主图 |
| TC-EX-02 | 锁定库存删除 | 商品 `locked_qty > 0` | 调用 `DELETE /products/:id` | 返回商品存在锁定库存，不能删除 |
| TC-EX-03 | 商品不存在 | 使用不存在的商品 ID | 调用编辑或删除接口 | 返回商品不存在 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 编辑或删除商品 | 返回未登录 |
| TC-AUTH-02 | 非商品所属企业卖家 | 编辑或删除商品 | 返回无权操作 |

## 回归关注点

- `GET /products` 默认仍只返回买家可见商品。
- `PUT /products/:id/status` 仍仅运营/超管可调用。
