# Seller Warehouse Binding Backend Release Checklist

## 发布基本信息

- Feature ID: SL-BE-SELLER-WAREHOUSE-BINDING
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-SELLER-WAREHOUSE-BINDING` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更
- [ ] 已确认数据库影响
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-SELLER-WAREHOUSE-BINDING --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 合并前主干版本
- 回滚步骤: revert 本 feature PR
- 回滚后验证: 仓库资源写接口重新仅允许 `warehouse` 服务商能力访问

## 发布后验证

- [ ] 销售企业可访问自有仓库资源 CRUD
- [ ] 未审核仓库不能绑定商品
- [ ] 已审核仓库可被商品绑定并进入订单/AI采购仓库快照
