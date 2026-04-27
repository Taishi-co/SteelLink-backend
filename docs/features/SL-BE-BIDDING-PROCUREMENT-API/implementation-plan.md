# Implementation Plan

## 步骤

1. 从当前主干创建 `feat/SL-BE-BIDDING-PROCUREMENT-api`。
2. 新增 `market` 模块实体、DTO、service、handler、router。
3. 在 `cmd/api/main.go` 注册 market AutoMigrate 和公开路由。
4. 增加 `AUTO_MIGRATE_SCOPE=market|skip`，用于本地联调时只迁移本次新增表或跳过迁移。
5. 对空表写入基础业务样例数据，保证本地和测试环境可直接联调。
6. 与前端 `/bidding`、`/procurement`、`/procurement/:id` 联调。
7. 执行 `go test ./...`、`go build ./cmd/api` 和 harness gate。

## 风险与回滚

- 新增表为独立市场活动表，不修改现有交易表结构。
- 回滚本分支可移除接口注册；数据库新增表可保留或按发布回滚流程清理。
- 后续真实写接口上线前，需要补充企业身份、操作日志和状态机校验。
