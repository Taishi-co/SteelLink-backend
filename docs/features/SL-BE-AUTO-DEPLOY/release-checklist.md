# Release Checklist

## 发布基本信息

- Feature ID: SL-BE-AUTO-DEPLOY
- 发布负责人: AI agent
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-AUTO-DEPLOY` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已在 GitHub Actions Secrets 配置 `PROD_HOST`、`PROD_USER`，并配置 `PROD_SSH_KEY` 或 `PROD_PASSWORD` 其中之一
- [ ] 如服务器部署脚本不在 SSH 登录后的默认目录下，已配置 `PROD_DEPLOY_COMMAND`
- [ ] 已确认远程部署脚本会拉取最新 `main` 并完成构建/重启
- [ ] 已确认 smoke test 范围

## 回滚信息

- 回滚目标版本: 合并前 `main` 版本。
- 回滚步骤: Revert 新增的 `.github/workflows/deploy-prod.yml` 及 `docs/features/SL-BE-AUTO-DEPLOY` 文档，或在 GitHub Actions 中临时禁用 `Deploy Production` workflow。
- 回滚后验证: 确认 `main` push 不再触发后端自动部署，按原人工流程执行一次后端部署 smoke test。
