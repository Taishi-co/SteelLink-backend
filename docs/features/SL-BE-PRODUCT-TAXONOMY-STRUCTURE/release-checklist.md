# Product Taxonomy Structure Release Checklist

## 发布基本信息

- Feature ID: SL-BE-PRODUCT-TAXONOMY-STRUCTURE
- 发布负责人: Codex
- 目标环境: backend API

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-PRODUCT-TAXONOMY-STRUCTURE` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更
- [ ] 已确认数据库影响
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-PRODUCT-TAXONOMY-STRUCTURE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一版 backend API 镜像
- 回滚步骤: 回滚应用代码；新增字段和主数据表可保留，旧代码会忽略。
- 回滚后验证: 商品列表、商品详情、下单、卖家商品管理可正常使用。
