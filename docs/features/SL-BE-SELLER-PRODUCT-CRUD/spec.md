# Feature Spec

## 基本信息

- Feature ID: SL-BE-SELLER-PRODUCT-CRUD
- Feature 名称: 卖家商品编辑与删除接口
- 负责人: Codex
- 目标上线环境: Backend 生产环境

## 背景

卖家后台商品管理需要支持完整 CRUD。当前后端只开放商品创建和展示资料补充接口，缺少卖家编辑核心商品字段与删除误建商品的能力。

## 范围

### In Scope

- 新增卖家自有商品编辑接口。
- 新增卖家自有商品软删除接口。
- 编辑商品后重置为待审核，等待平台重新审核。
- 删除商品前校验企业归属和锁定库存。

### Out of Scope

- 不调整运营端审核接口。
- 不新增批量删除或批量上下架。
- 不改变买家端商品公开查询规则。

## 交互与页面影响

- 受影响页面: seller-app `/seller/products`
- 新增页面: 无
- 路由变化: 新增 `PUT /api/v1/products/:id` 与 `DELETE /api/v1/products/:id`
- 菜单变化: 无

## 数据与接口影响

- 受影响 API:
  - `PUT /api/v1/products/:id`
  - `DELETE /api/v1/products/:id`
- 请求/响应变化:
  - `PUT /products/:id` 复用商品创建请求体，返回更新后的商品详情。
  - `DELETE /products/:id` 成功返回删除成功消息。
- 环境变量变化: 无

## 权限影响

- 受影响角色: 具备 sales trade capability 的企业用户。
- 新增/调整权限: 只能编辑或删除所属企业商品；有锁定库存的商品不能删除。

## 验收标准

1. 卖家可编辑所属企业商品，提交后商品状态为待审核。
2. 卖家不能编辑其他企业商品。
3. 卖家可删除无锁定库存的所属企业商品。
4. 有锁定库存商品删除时返回业务错误。
