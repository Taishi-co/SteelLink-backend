# Seller Warehouse Binding Backend Implementation Plan

## Feature 信息

- Feature ID: SL-BE-SELLER-WAREHOUSE-BINDING
- 开发负责人: Codex

## 变更范围

- 代码文件: `internal/modules/serviceoffer/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-SELLER-WAREHOUSE-BINDING/*`
  - `docs/domain/permission-scope-audit.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 确认商品仓库绑定和仓储资源接口现状。
2. 将仓库资源自有 CRUD 权限从仅 `warehouse` 扩展为 `sales` 或 `warehouse`。
3. 保留商品绑定仓储资源必须审核通过的现有校验。
4. 更新权限口径和 feature 文档。
5. 运行 Go test/build 和 harness gate。

## 验证计划

- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-SELLER-WAREHOUSE-BINDING`
- `./harness/release.sh docs/features/SL-BE-SELLER-WAREHOUSE-BINDING`
- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- 需要的人工验证: 使用销售企业账号在卖家端创建仓库，待平台审核通过后绑定到商品。

## 风险与回滚关注点

- 销售企业创建的仓库资源会进入待审核，运营端审核前不能绑定为商品仓储资源。
- 如需回滚，恢复仓库资源路由中间件为仅 `RequireProviderService("warehouse")`。
