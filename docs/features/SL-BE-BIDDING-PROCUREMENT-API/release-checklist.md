# Release Checklist

## 发布前

- [ ] `go test ./...`
- [ ] `CGO_ENABLED=0 go build ./cmd/api`
- [ ] `./harness/review.sh docs/features/SL-BE-BIDDING-PROCUREMENT-API`
- [ ] `./harness/release.sh docs/features/SL-BE-BIDDING-PROCUREMENT-API`

## 联调检查

- [ ] `GET /api/v1/biddings/stats` 可返回统计数据。
- [ ] `GET /api/v1/biddings` 可返回竞价列表。
- [ ] `GET /api/v1/procurements/stats` 可返回统计数据。
- [ ] `GET /api/v1/procurements` 可返回集采列表。
- [ ] `GET /api/v1/procurements/:id` 可返回活动详情和日志。
- [ ] 本地远端库联调可使用 `AUTO_MIGRATE_SCOPE=market` 避免重复迁移历史表。

## 回滚

- 回滚本分支提交。
- 如需清理数据，按 DBA 流程删除 `bidding_activities`、`procurement_activities`、`procurement_logs`。
