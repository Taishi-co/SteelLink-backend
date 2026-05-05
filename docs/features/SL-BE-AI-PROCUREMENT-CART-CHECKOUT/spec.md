# AI Procurement Cart Checkout Spec

## 基本信息

- Feature ID: SL-BE-AI-PROCUREMENT-CART-CHECKOUT
- Feature 名称: AI采购复用购物车式下单与取消回填
- 负责人: Codex
- 目标上线环境: production

## 背景

AI采购订单生成原本只接受单个召回商品并直接创建订单。业务希望订单生成走和购物车结算一致的库存锁定语义，并支持用户在商品匹配阶段按多个商品拆分采购数量。若订单取消，AI采购锁定的商品应回到购物车，便于客户继续调整和结算。

## 范围

### In Scope

- 新增后端购物车持久化模块与 `/api/v1/cart/*` 接口，支撑前端已有购物车封装。
- 扩展订单创建请求，支持 `items` 多商品下单和 `restore_cart_on_cancel` 取消回填标记。
- AI采购 `generate-order` 支持 `selected_products`，生成同一卖家下的多商品订单。
- 取消订单时，带回填标记的订单项会重新写入买家购物车。

### Out of Scope

- 跨卖家拼单父子订单模型。
- 购物车独立 checkout 页面重构。
- 支付、合同、物流状态机调整。

## 接口与数据影响

- 受影响路由/handler:
  - 新增 `/api/v1/cart/items` `GET/POST/PUT/DELETE`
  - 新增 `/api/v1/cart/count`
  - 变更 `/api/v1/orders`
  - 变更 `/api/v1/ai-procurements/generate-order`
- 受影响服务:
  - `OrderService.Create`
  - `OrderService.Cancel`
  - `ProcurementService.GenerateOrder`
- 受影响 middleware: 无新增，沿用企业账号鉴权。
- 受影响表/model:
  - 新增 `cart_items`
  - `orders.remark` 增加 `source` 与 `restore_cart_on_cancel` 运行时标记。
- 配置或环境变量变化: 无。
- Docker / 部署影响: 默认 AutoMigrate 会创建 `cart_items` 表。

## 权限与合规影响

- 受影响角色: enterprise。
- 新增/调整权限: 无新增权限码，沿用登录企业用户。
- 是否影响审核、风控、日志留痕: 影响订单取消后的购物车回填行为，不改变审核提交权限。

## 验收标准

1. AI采购可提交多个同卖家 `selected_products` 并生成真实订单。
2. 多商品订单取消时，订单项重新回填到当前买家购物车。
3. 不同卖家的拼单请求返回明确错误，不生成结构不一致的订单。
