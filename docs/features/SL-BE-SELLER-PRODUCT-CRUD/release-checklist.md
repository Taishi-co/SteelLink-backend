# Release Checklist

## 发布基本信息

- Feature ID: SL-BE-SELLER-PRODUCT-CRUD
- 发布负责人: Codex
- 目标环境: Backend 生产环境

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-SELLER-PRODUCT-CRUD` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api` 已通过
- [ ] 已确认环境变量、CI 或跨仓依赖
- [ ] 已确认 smoke test 范围: `PUT /products/:id`、`DELETE /products/:id`、卖家商品列表
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-SELLER-PRODUCT-CRUD --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 发布前版本
- 回滚步骤: 回滚商品编辑/删除 service、handler、router 和 repository 变更
- 回滚后验证: 商品创建、商品列表、商品展示资料补充接口仍可使用

## 发布后验证

- [ ] 卖家可编辑自有商品
- [ ] 卖家可删除无锁定库存的自有商品
- [ ] 运营端审核接口仍可正常更新商品状态
