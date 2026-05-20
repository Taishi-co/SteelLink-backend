# Release Checklist

## 发布基本信息

- Feature ID: `SL-BE-WORKBENCH-FLOW-BASELINE`
- 发布负责人: Codex
- 目标环境: 本地开发 / 测试联调环境

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-WORKBENCH-FLOW-BASELINE` 已通过
- [ ] `go test ./internal/modules/order/...` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] 已确认无数据库迁移、配置、环境变量变更
- [ ] 已确认旧前端 `buyer-sign` 路径与新路径兼容
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 发布前版本
- 回滚步骤: 回退订单路由兼容路径和 feature 文档
- 回滚后验证: 重新执行订单模块测试与后端 verify
