# Implementation Plan

## Feature 信息

- Feature ID: SL-BE-SELLER-PRODUCT-CRUD
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/product/application/service/product_service.go`
  - `internal/modules/product/domain/repository/product_repository.go`
  - `internal/modules/product/infrastructure/persistence/product_repository.go`
  - `internal/modules/product/infrastructure/router/router.go`
  - `internal/modules/product/ui/handler/product_handler.go`
- 文档文件:
  - `docs/features/SL-BE-SELLER-PRODUCT-CRUD/*`
- 配置文件: 无
- 脚本文件: 无

## 实施步骤

1. Repository 增加软删除方法。
2. Product service 增加 Update/Delete，校验企业归属、主图和锁定库存。
3. Handler 增加编辑和删除入口。
4. Router 在 sales capability 分组下注册 `PUT /products/:id` 和 `DELETE /products/:id`。

## 验证计划

- `go test ./internal/modules/product/...`
- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- 需要的人工验证: 使用卖家账号调用编辑、删除接口并走查 seller-app。

## 风险与回滚关注点

- 编辑已上架商品会重新进入待审核，短期内会影响买家端可见性。
- 删除为 GORM 软删除；如误删需要数据库层恢复 `deleted_at`。
