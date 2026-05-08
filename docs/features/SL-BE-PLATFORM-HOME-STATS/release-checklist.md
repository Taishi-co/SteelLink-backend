# Platform Home Stats Release Checklist

## 发布基本信息

- Feature ID: SL-BE-PLATFORM-HOME-STATS
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-PLATFORM-HOME-STATS` 已通过
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [x] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [x] 已确认配置与环境变量变更: 无
- [x] 已确认数据库影响: 只读查询现有表，无迁移
- [x] 已确认 smoke test 范围: `GET /api/v1/platform/stats`、买家端首页统计展示
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-PLATFORM-HOME-STATS --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一稳定后端版本
- 回滚步骤: 回滚本 feature 新增 handler 和路由
- 回滚后验证: `GET /api/v1/health` 正常，首页回退到前端降级显示
