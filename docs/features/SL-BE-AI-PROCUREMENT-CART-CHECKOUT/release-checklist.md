# AI Procurement Cart Checkout Release Checklist

## 发布基本信息

- Feature ID: SL-BE-AI-PROCUREMENT-CART-CHECKOUT
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更
- [ ] 已确认数据库影响
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-AI-PROCUREMENT-CART-CHECKOUT --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 合并前版本。
- 回滚步骤: Revert PR，确认 API 服务启动；如需清理新表，由 DBA 评估 `cart_items` 数据后处理。
- 回滚后验证: `go test ./...`，订单创建和取消 smoke test。
