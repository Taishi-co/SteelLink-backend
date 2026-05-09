# Redis Cache Expansion Release Checklist

## 发布基本信息

- Feature ID: SL-BE-REDIS-CACHE-EXPANSION
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-REDIS-CACHE-EXPANSION` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认业务后端 Redis 连接配置
- [ ] 已确认 Admin 和业务后端 AI 采购配置仍使用同一个 Redis 实例与 DB
- [ ] 已确认数据库无 schema 变更
- [ ] 已确认 smoke test 范围：登录刷新、商品列表、服务资源列表、市场统计、AI 采购匹配
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-REDIS-CACHE-EXPANSION --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一个稳定后端镜像
- 回滚步骤: 回滚本 feature 代码并重新发布后端
- 回滚后验证: 登录、商品列表、服务资源列表、AI 采购基础流程正常
