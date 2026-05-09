# SL-BE-PRODUCT-STRICT-ONSALE Release Checklist

## 发布基本信息

- Feature ID: SL-BE-PRODUCT-STRICT-ONSALE
- 发布负责人: Codex
- 目标环境: Backend production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-PRODUCT-STRICT-ONSALE` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更: 无
- [ ] 已确认数据库影响: 无 schema 变更；异常历史数据需另行清洗
- [ ] 已确认 smoke test 范围: `GET /api/v1/products`、`GET /api/v1/products?status=1`、`GET /api/v1/products?status=2`
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-PRODUCT-STRICT-ONSALE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一个稳定 backend 版本
- 回滚步骤: 回滚本 PR 对应提交并重新部署 backend
- 回滚后验证: 公开商品列表恢复到回滚版本行为
