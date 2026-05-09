# AI Procurement Finance Product Matching Release Checklist

## 发布基本信息

- Feature ID: `SL-BE-AI-FINANCE-MATCH`
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-AI-FINANCE-MATCH` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/aiprocurement-vector-sync` 已通过
- [ ] 已确认配置与环境变量变更: 复用 pgvector/embedding/vector sync token 配置
- [ ] 已确认数据库影响: `ai_finance_product_embeddings`、同步状态表、`orders.remark.ai_finance_recommendation`
- [ ] 已确认 smoke test 范围: AI采购 `match-finance`、订单详情融资申请、vector sync `target=finance`
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-AI-FINANCE-MATCH --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 合并前 `main`
- 回滚步骤: revert PR，保留 pgvector 附加表不影响主业务
- 回滚后验证: AI采购可跳过融资或走旧订单融资入口
