# Release Checklist: 供求信息后端

## 发布基本信息

- Feature ID: SL-BE-SUPPLY-DEMAND
- 发布负责人: AI agent
- 目标环境: Backend API

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-SUPPLY-DEMAND` 已通过
- [ ] `go test ./...` 已通过
- [ ] `cd ../SteelLink-admin/backend && go test ./...` 已通过
- [ ] `cd ../SteelLink-admin/frontend && pnpm typecheck` 已通过
- [ ] `cd ../SteelLink-frontend && pnpm --filter @steellink/buyer-app typecheck` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更: 无
- [ ] 已确认数据库影响: `supply_demand_posts` 表新增审核字段，新发布状态默认待审核
- [ ] 已确认 smoke test 范围: 公开列表、详情、统计、发布、我的发布、关闭发布、管理后台审核通过/驳回

## 回滚信息

- 回滚目标版本: 上一稳定后端版本
- 回滚步骤: 回滚本 feature 代码；根据环境策略决定是否保留或删除 `supply_demand_posts` 表
- 回滚后验证: 执行 `go test ./...` 和 `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
