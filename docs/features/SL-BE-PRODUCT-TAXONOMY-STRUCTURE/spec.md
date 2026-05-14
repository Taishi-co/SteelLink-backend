# Product Taxonomy Structure Spec

## 基本信息

- Feature ID: SL-BE-PRODUCT-TAXONOMY-STRUCTURE
- Feature 名称: 商品品种牌号规格结构化
- 负责人: Codex
- 目标上线环境: backend API

## 背景

卖家商品当前以文本品类、文本钢厂和 `spec` JSON 为主，无法稳定支持“品种、牌号、规格、产地、存放地”的行业化检索。商家后续上传台账或截图进行 AI 批量识别时，也需要先有稳定的标准字段承接识别结果。

## 范围

### In Scope

- 保留现有商品文本快照字段，同时新增品种、牌号、钢厂主数据 ID。
- 将厚度/直径/高度、宽度/壁厚、长度类型和值保存为可索引商品字段。
- 商品列表支持按品种 ID、牌号 ID、钢厂 ID、牌号文本、规格数值范围查询。
- 商品创建和修改时自动补齐品种、牌号、钢厂主数据。
- `bundle_no` 从全局唯一改为可重复，迁移时移除旧唯一索引并保留普通索引。

### Out of Scope

- 不在本 feature 内实现完整 AI 文件解析、OCR、多模态识别和批量确认工作流。
- 不新增运营后台的完整主数据维护页面。
- 不改变订单、购物车、仓储履约的历史快照读取方式。

## 接口与数据影响

- 受影响路由/handler:
  - `GET /api/v1/product-grades`
  - `GET /api/v1/product-steel-mills`
  - `POST /api/v1/products`
  - `PUT /api/v1/products/:id`
  - `GET /api/v1/products`
- 受影响服务:
  - `internal/modules/product/application/service`
  - `internal/modules/product/infrastructure/persistence`
- 受影响 middleware: 无
- 受影响表/model:
  - `products`
  - `product_categories`
  - `product_grades`
  - `product_steel_mills`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无新增依赖；需要随服务启动执行 AutoMigrate。

## 权限与合规影响

- 受影响角色:
  - 买家/未登录用户读取商品列表时可使用新筛选字段。
  - 卖家发布和修改商品时可提交结构化字段。
  - 运营角色继续负责商品审核。
- 新增/调整权限: 无新增权限点。
- 是否影响审核、风控、日志留痕: 商品关键字段变更仍会进入待审核状态。

## 验收标准

1. 卖家创建商品时可以提交品种 ID、牌号 ID、钢厂 ID 和拆分规格字段，返回结果包含这些字段。
2. 未提交 ID 但提交文本品种、牌号、钢厂时，后端自动补齐主数据并保存 ID。
3. 商品列表可以通过牌号、钢厂 ID、规格范围过滤。
4. 数据库不再阻止重复 `bundle_no`。
5. 旧的 `category`、`sub_category`、`steel_mill`、`spec.material`、`spec.spec` 字段继续返回，历史前端不被破坏。
