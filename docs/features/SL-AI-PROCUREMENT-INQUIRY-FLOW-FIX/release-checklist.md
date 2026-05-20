# AI Procurement Inquiry Flow Fix Release Checklist

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX
- 发布负责人: Codex
- 目标环境: backend api

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `go test ./internal/modules/aiprocurement/application/service` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 已确认环境变量、CI 或跨仓依赖
- [ ] 已确认 smoke test 范围: AI采购解析、商品匹配、物流匹配、订单生成、融资匹配、审核提交、历史详情
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 修复前 AI采购服务层
- 回滚步骤: 回退 `internal/modules/aiprocurement/application/service/service.go`、`service_test.go` 和本 feature 文档
- 回滚后验证: AI采购服务层测试和 `cmd/api` 构建

## 发布后验证

- [ ] AI采购历史详情响应列表字段为 `[]`
- [ ] 联调环境完整链路可通过
