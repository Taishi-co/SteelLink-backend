# Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-MARKET-MIGRATION-ORDER`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/market/infrastructure/persistence/migrate.go`
  - `internal/modules/market/infrastructure/persistence/migrate_test.go`
- 文档文件:
  - `docs/features/SL-BE-MARKET-MIGRATION-ORDER/*`
- 配置文件:
  - 无
- 部署/脚本文件:
  - 无

## 实施步骤

1. 增加 market 迁移模型顺序测试，锁定 `ProcurementProject` 必须早于 `ProcurementLog`。
2. 将 market 自动迁移模型列表抽为 `migrationModels()`，使测试能覆盖真实迁移顺序。
3. 调整 `ProcurementProject` 在迁移列表中的位置，确保 `procurement_projects` 先于 `procurement_logs` 创建。
4. 运行目标包测试、`./harness/verify.sh`、review gate 和 release gate。

## 验证计划

- `go test ./internal/modules/market/infrastructure/persistence`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-MARKET-MIGRATION-ORDER`
- `./harness/release.sh docs/features/SL-BE-MARKET-MIGRATION-ORDER`
- 需要的人工验证:
  - 生产或 staging 后端启动后访问 `/api/v1/health`

## 风险与回滚关注点

- 风险: 仅调整迁移顺序，不修改业务字段和历史数据，主要风险是迁移模型列表遗漏。
- 回滚: 回滚本 PR 后，若生产库已创建 `procurement_projects`，旧代码仍可启动；全新环境会重新暴露同类迁移风险。
- 发布关注点: 启动日志不应再出现 `Failed to open the referenced table 'procurement_projects'`。
