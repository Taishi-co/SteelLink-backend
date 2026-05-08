# SL-BE-CART-PRODUCT-FULFILLMENT

## 背景

买家购物车需要展示商品缩略图和可发货仓库。当前 `/cart/items` 只返回商品名、规格、价格、卖家和数量，无法支撑前端展示。

## 范围

- `internal/modules/cart/application_service.go`
- `internal/modules/cart/ui/dto/cart_dto.go`

## 需求

1. `/cart/items`、加入购物车和更新购物车返回的 `CartItemInfo` 包含商品首图 `image_url`。
2. 返回商品绑定仓库快照 `warehouse_info`，包含仓库名称、编码、地址和关联 ID。
3. 未绑定仓库或未上传图片时保持空值返回，不改变购物车增删改行为。

## 非目标

- 不新增数据库字段。
- 不改变商品仓库绑定规则。
- 不新增购物车仓库选择接口。
