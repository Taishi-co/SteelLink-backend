# 多工作台会话切换后端发布检查

## 发布基本信息

- Feature ID: SL-MULTI-WORKBENCH-AUTH
- 发布负责人: Codex
- 目标环境: staging/prod

## 发布前检查

- [x] `./harness/release.sh docs/features/SL-MULTI-WORKBENCH-AUTH` 已通过
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [x] 已确认配置与环境变量无变更
- [x] 已确认数据库无变更
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-MULTI-WORKBENCH-AUTH --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 当前生产版本
- 回滚步骤: 回滚本 PR，恢复 switch-session 仅支持 `buyer/seller`
- 回滚后验证: 买家/卖家中心切换仍可用；服务商工作台回退共享 token 兜底
