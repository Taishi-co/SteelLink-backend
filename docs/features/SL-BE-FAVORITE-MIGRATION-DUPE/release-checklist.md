# Release Checklist

## 发布基本信息

- Feature ID: SL-BE-FAVORITE-MIGRATION-DUPE
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [x] `./harness/release.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE` 已通过
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [x] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [x] 已确认配置与环境变量变更: 无变更
- [x] 已确认数据库影响: `favorites` 迁移前会清理重复 `user_id/product_id` 行
- [x] 已确认 smoke test 范围: `/health` 与 `/api/v1/health`
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: PR 合入前版本
- 回滚步骤: revert 本 PR；如后端启动仍被重复数据阻断，先按生产数据备份/重复行检查结果处理 `favorites` 重复数据
- 回滚后验证: 后端容器健康，`curl http://127.0.0.1:3000/health` 与公网 `/api/v1/health` 返回 `200 OK`
