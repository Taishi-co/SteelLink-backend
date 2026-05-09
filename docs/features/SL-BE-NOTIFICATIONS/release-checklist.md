# Release Checklist

## 发布基本信息

- Feature ID: SL-BE-NOTIFICATIONS
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o /tmp/steellink-api-notifications ./cmd/api` 已通过
- [ ] `./harness/release.sh docs/features/SL-BE-NOTIFICATIONS` 已通过
- [ ] 已确认生产启动会执行 `notifications` 表迁移
- [ ] 已完成管理员发送和用户收件箱 smoke test
- [ ] 已完成 WebSocket 订阅 smoke test

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: 回滚本 feature 代码；如需清理数据，备份后删除 `notifications` 表
- 回滚后验证: API 启动、订单和商品审核流程正常
