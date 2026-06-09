# Release Checklist

## 发布基本信息

- Feature ID: `SL-BE-MARKET-MIGRATION-ORDER`
- 发布负责人: Codex
- 目标环境: prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-MARKET-MIGRATION-ORDER` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [x] 已确认配置与环境变量变更: 无
- [x] 已确认数据库影响: 仅调整 market 自动迁移顺序，不删除数据
- [ ] 已确认 smoke test 范围: `/api/v1/health`
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-MARKET-MIGRATION-ORDER --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 当前生产后端上一稳定镜像或回滚本 PR。
- 回滚步骤: 停止新后端容器，恢复上一镜像；如已建 `procurement_projects` 空表，保留该表不影响旧逻辑。
- 回滚后验证: 访问 `/api/v1/health`，确认后端容器不再重启。
