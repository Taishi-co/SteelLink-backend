# SL-BE-PRODUCT-CATEGORY-MASTER Spec

## 背景

买家端 `/mall` 品类导航当前从商品列表前 12 条数据反推，商品列表筛选和卖家创建商品页又各自维护硬编码品类，导致线上新增品类无法稳定出现在筛选中。

## 目标

- 新增商品品类主数据，作为商城导航、商品筛选、卖家商品创建的统一来源。
- 品类支持业务名、展示名、导航图片、排序、启停状态。
- 商品创建时允许使用新业务品类，并自动补齐默认品类主数据。
- 保持现有商品 `category` 字段兼容，避免大规模历史商品迁移。

## 非目标

- 本次不迁移历史订单或 AI 采购记录中的品类字段。
- 本次不实现多级子品类树。
- 本次不改变商品筛选参数语义，仍使用 `category` 字符串过滤。

## 接口

- `GET /api/v1/product-categories`
  - 公开读取启用品类，返回 `name/display_name/image_url/icon/sort/status`。
- `POST /api/v1/product-categories`
  - 平台运营或超级管理员创建品类。
- `PUT /api/v1/product-categories/:id`
  - 平台运营或超级管理员修改展示名、图片、排序、状态。
- `DELETE /api/v1/product-categories/:id`
  - 平台运营或超级管理员软删除品类。

## 数据约定

- `name` 是业务品类值，商品表继续保存这个值。
- `display_name` 是前台展示名，管理员可修改。
- `image_url` 是商城导航图片。
- 商品创建时如 `name` 不存在，后端创建一条启用品类，`display_name=name`，图片为空。
