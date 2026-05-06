# Release Checklist

## 发布基本信息

- Feature ID: SL-BE-SERVICE-FULFILLMENT
- 发布负责人: Codex
- 目标环境: 后端预览环境

## 发布前检查

- [x] `./harness/release.sh docs/features/SL-BE-SERVICE-FULFILLMENT` 已通过
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [x] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [x] 已确认配置与环境变量变更: 无
- [x] 已确认数据库影响: 新增 `warehouse_inventory_movements`、`logistics_orders`、`logistics_tracking_events`、`fulfillment_guarantees`，扩展使用既有 `warehouse_services`、`logistics_services`、`products.warehouse_info`
- [x] 已确认 smoke test 范围
- [x] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-SERVICE-FULFILLMENT --pr-url https://github.com/Taishi-co/SteelLink-backend/pull/48 --pr-title "feat: add logistics warehouse fulfillment backend"` 并写入共享飞书 PR 记录
- 飞书 PRD/PR 记录: https://www.feishu.cn/wiki/XqRXwzfGZiWiRRkKOZ9ckfBLnEb

## 回滚信息

- 回滚目标版本: 合并前版本
- 回滚步骤: 回滚 `serviceoffer` CRUD 扩展、`servicefulfillment` 模块、`cmd/api` 路由/迁移接入和 feature 文档
- 回滚后验证: `./harness/verify.sh`
