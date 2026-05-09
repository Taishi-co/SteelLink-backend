# Redis Ops Cache Release Checklist

## 发布基本信息

- Feature ID: SL-BE-REDIS-OPS-CACHE
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-REDIS-OPS-CACHE` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认 Redis 配置可用
- [ ] 已确认 `AI_PROCUREMENT_VECTOR_SYNC_TOKEN` 仍按现有方式配置
- [ ] 已确认数据库无 schema 变更
- [ ] 已确认 smoke test 范围: 登录、短信、上传、AI 采购 parse/match、平台统计、vector-sync/status
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-REDIS-OPS-CACHE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一个稳定后端镜像
- 回滚步骤: 回滚本 feature 代码并重新发布后端
- 回滚后验证: 登录、短信、上传、平台首页统计和 AI 采购基础流程正常
