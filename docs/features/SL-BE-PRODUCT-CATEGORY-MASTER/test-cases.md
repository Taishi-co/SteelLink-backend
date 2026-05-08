# SL-BE-PRODUCT-CATEGORY-MASTER Test Cases

## API

- `GET /api/v1/product-categories` 返回启用品类，按 `sort, id` 排序。
- 已有商品分类但没有品类主数据时，迁移后自动补齐品类记录。
- 创建商品传入新 `category` 时，商品创建成功并补齐品类记录。
- 操作员或超级管理员可以创建、更新、删除品类。
- 非授权用户不能修改品类。

## 兼容性

- `GET /api/v1/products?category=冷硬` 仍按商品 `category/sub_category` 过滤。
- 老商品没有 `product_categories` 记录时不影响商品列表。
