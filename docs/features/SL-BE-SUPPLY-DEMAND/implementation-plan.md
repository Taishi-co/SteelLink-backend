# Implementation Plan: 供求信息后端

## Feature 信息

- Feature ID: SL-BE-SUPPLY-DEMAND
- 开发负责人: AI agent

## 变更范围

- 代码文件:
  - `cmd/api/main.go`
  - `internal/modules/supplydemand/domain/entity/post.go`
  - `internal/modules/supplydemand/domain/repository/repository.go`
  - `internal/modules/supplydemand/infrastructure/persistence/post_repository.go`
  - `internal/modules/supplydemand/infrastructure/persistence/migrate.go`
  - `internal/modules/supplydemand/infrastructure/router/router.go`
  - `internal/modules/supplydemand/application/service/service.go`
  - `internal/modules/supplydemand/ui/dto/dto.go`
  - `internal/modules/supplydemand/ui/handler/handler.go`
  - `SteelLink-admin/backend/internal/handler/supply_demand.go`
  - `SteelLink-admin/frontend/src/pages/supply-demand/index.tsx`
- 文档文件:
  - `docs/features/SL-BE-SUPPLY-DEMAND/spec.md`
  - `docs/features/SL-BE-SUPPLY-DEMAND/test-cases.md`
  - `docs/features/SL-BE-SUPPLY-DEMAND/implementation-plan.md`
  - `docs/features/SL-BE-SUPPLY-DEMAND/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 补齐 feature 文档四件套。
2. 新增供求信息实体、仓储接口和 GORM 实现。
3. 新增 service，处理类型/状态校验、分页默认值、发布人权限和发布后待审核状态。
4. 新增 DTO 与 Gin handler。
5. 注册 `/api/v1/supply-demands` 路由和自动迁移。
6. 管理后台新增供求信息审核列表、审核通过/驳回接口和页面入口。
7. 执行 gofmt、Go/TS 测试、构建和 harness review/release。

## 验证计划

- `go test ./...`
- `cd ../SteelLink-admin/backend && go test ./...`
- `cd ../SteelLink-admin/frontend && pnpm typecheck`
- `cd ../SteelLink-frontend && pnpm --filter @steellink/buyer-app typecheck`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/review.sh docs/features/SL-BE-SUPPLY-DEMAND`
- `./harness/release.sh docs/features/SL-BE-SUPPLY-DEMAND`
- 需要的人工验证: 使用真实用户 token 调用发布、我的发布和关闭接口；使用管理后台账号审核通过/驳回并检查买家端公开可见性。

## 风险与回滚关注点

- 风险: 当前通过 GORM 自动迁移创建表，生产环境如要求手工 migration，需要在上线前补齐数据库变更脚本。
- 风险: 历史已发布状态数据仍会保持公开可见；如需重审历史数据，需要额外数据治理脚本。
- 回滚: 回滚代码并按环境数据库策略下线或保留 `supply_demand_posts` 表。
