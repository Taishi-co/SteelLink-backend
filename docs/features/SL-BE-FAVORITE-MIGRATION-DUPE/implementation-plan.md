# Implementation Plan

## Feature 信息

- Feature ID: SL-BE-FAVORITE-MIGRATION-DUPE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/favorite/infrastructure/persistence/migrate.go`
  - `internal/modules/favorite/infrastructure/persistence/migrate_test.go`
- 文档文件:
  - `docs/features/SL-BE-FAVORITE-MIGRATION-DUPE/spec.md`
  - `docs/features/SL-BE-FAVORITE-MIGRATION-DUPE/test-cases.md`
  - `docs/features/SL-BE-FAVORITE-MIGRATION-DUPE/implementation-plan.md`
  - `docs/features/SL-BE-FAVORITE-MIGRATION-DUPE/release-checklist.md`
- 配置文件: 无。
- 部署/脚本文件: 无。

## 实施步骤

1. 在收藏模块迁移入口读取当前 `favorites` 表与 `deleted_at` 列状态。
2. 若表已存在，执行基于 `ROW_NUMBER()` 的重复数据清理，按 `user_id/product_id` 分组保留一条记录。
3. 对旧 schema 优先保留未软删除记录；对当前 schema 保留最早记录。
4. 清理完成后执行原有 `AutoMigrate(&entity.Favorite{})`。
5. 增加单元测试覆盖旧 schema 与当前 schema 的 SQL 策略。
6. 执行 verify、review、release gate 后创建 PR。

## 验证计划

- `go test ./internal/modules/favorite/infrastructure/persistence`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE`
- `./harness/release.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE`
- 需要的人工验证: 生产发布后确认后端容器 `healthy`，`/api/v1/health` 返回 `200 OK`。

## 风险与回滚关注点

- 风险: 清理逻辑会物理删除重复收藏行。重复行本身无法满足新的唯一约束，保留策略必须明确。
- 回滚: 回退本 PR 后，如生产仍存在重复收藏行，需要手动删除重复记录或临时跳过收藏迁移。
- 数据恢复: 如需恢复被清理的收藏数据，应从发布前数据库备份中按 `favorites.id` 定点恢复。
