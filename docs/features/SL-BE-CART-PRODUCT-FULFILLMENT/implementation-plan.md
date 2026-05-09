# Implementation Plan

1. 扩展购物车 DTO，新增 `image_url` 和 `warehouse_info`。
2. 在购物车应用服务从商品实体读取首张图片和 `WarehouseInfo` 快照。
3. 保持空图片、空仓库兼容。
4. 执行 `gofmt` 和后端构建验证。
