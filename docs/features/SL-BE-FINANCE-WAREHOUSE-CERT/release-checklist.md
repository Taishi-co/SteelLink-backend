# SL-BE-FINANCE-WAREHOUSE-CERT Release Checklist

## 发布基本信息

- Feature ID: SL-BE-FINANCE-WAREHOUSE-CERT
- 发布负责人: Codex
- 目标环境: staging

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-FINANCE-WAREHOUSE-CERT` 已通过
- [ ] `go test ./...` 与 `go build` 已通过
- [ ] 已确认监管仓授权数据已存在或由前端入口创建
- [ ] GitHub PR 创建后已执行飞书 PRD/PR 记录同步

## 回滚信息

- 回滚目标版本: 上一稳定版本
- 回滚步骤: 回滚本 feature 代码。
- 回滚后验证: AI 采购融资匹配恢复为按金额范围匹配。
