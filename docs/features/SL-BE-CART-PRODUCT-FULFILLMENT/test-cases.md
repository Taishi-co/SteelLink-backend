# Test Cases

| 用例 | 前置条件 | 操作 | 预期 |
| --- | --- | --- | --- |
| TC-01 图片返回 | 商品存在 `images` | 调用 `/cart/items` | 返回首张图片为 `image_url` |
| TC-02 仓库返回 | 商品存在 `warehouse_info` | 调用 `/cart/items` | 返回完整 `warehouse_info` |
| TC-03 空图片 | 商品未上传图片 | 调用 `/cart/items` | `image_url` 为空字符串且接口成功 |
| TC-04 空仓库 | 商品未绑定仓库 | 调用 `/cart/items` | 不返回 `warehouse_info` 或为空，接口成功 |
