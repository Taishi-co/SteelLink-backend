# Redis Business Cache Expansion Release Checklist

## 发布基本信息

- Feature ID: SL-BE-REDIS-BUSINESS-CACHE
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-REDIS-BUSINESS-CACHE` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更: 无新增配置
- [ ] 已确认数据库影响: 无 schema 变更
- [ ] 已确认 smoke test 范围: AI 匹配、dashboard、企业服务 overview、幂等写入接口
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-REDIS-BUSINESS-CACHE --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 合入本 PR 前的 `main`
- 回滚步骤: revert 本 PR 或回滚到上一镜像
- 回滚后验证:
  - `/api/v1/users/me/dashboard`
  - `/api/v1/enterprises/me/services`
  - `/api/v1/ai-procurements/match-products`
  - `/api/v1/orders`
