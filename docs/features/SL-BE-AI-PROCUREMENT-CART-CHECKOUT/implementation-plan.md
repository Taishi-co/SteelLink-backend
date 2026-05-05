# AI Procurement Cart Checkout Implementation Plan

## Feature 信息

- Feature ID: SL-BE-AI-PROCUREMENT-CART-CHECKOUT
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `cmd/api/main.go`
  - `internal/modules/cart/**`
  - `internal/modules/order/**`
  - `internal/modules/aiprocurement/**`
- 文档文件:
  - `docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT/*`
- 配置文件: 无。
- 部署/脚本文件: 无。

## 实施步骤

1. 新增购物车 entity/repository/service/handler/router，并注册迁移和路由。
2. 扩展订单创建 DTO 和服务逻辑，支持多商品订单、购物车结算输入和取消回填购物车。
3. 扩展 AI采购订单生成 DTO 和服务逻辑，支持 `selected_products` 拼单下单。
4. 执行 Go 格式化、测试、构建、review 和 release gate。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT`
- `./harness/release.sh docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT`
- 需要的人工验证: 使用真实企业账号验证 AI采购拼单下单、订单取消和购物车回填。

## 风险与回滚关注点

- 当前订单表只有单个卖家字段，跨卖家拼单被显式拒绝；后续若要支持跨卖家，需要父子订单模型。
- 回滚时需撤回后端代码并确认 `cart_items` 表不被旧版本依赖。
