# SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE Release Checklist

## 发布基本信息

- Feature ID: SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE
- 发布负责人: Codex
- 目标环境: prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更
- [ ] 已确认数据库影响
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 合并前后端稳定版本
- 回滚步骤: 回滚本 PR；新增列可保留
- 回滚后验证: 服务市场物流委托和资源上架仍使用既有接口
