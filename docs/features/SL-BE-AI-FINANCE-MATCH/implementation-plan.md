# AI Procurement Finance Product Matching Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-AI-FINANCE-MATCH`
- 开发负责人: Codex

## 变更范围

- 代码文件: `internal/modules/aiprocurement/**`, `internal/pkg/vectorstore/pgvector.go`, `cmd/aiprocurement-vector-sync/main.go`
- 文档文件: `docs/features/SL-BE-AI-FINANCE-MATCH/*`
- 配置文件: 复用既有 AI采购向量配置
- 部署/脚本文件: 复用 `cmd/aiprocurement-vector-sync`

## 实施步骤

1. 扩展 AI采购融资方案结构和 `match-finance` 请求体，支持真实融资产品 ID、AI 推荐和用户预选标记。
2. 基于 `finance_services` 实现融资产品召回: pgvector 优先、MySQL 金额范围兜底。
3. 将用户预选融资产品写回 `ai_procurements.finance_plans` 与关联订单备注。
4. 为 `ai_finance_product_embeddings` 增加 pgvector schema、upsert、search、watermark 同步能力，并启用 `--target=finance`。

## 验证计划

- `go test ./internal/modules/aiprocurement/application/service ./internal/pkg/vectorstore`
- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `CGO_ENABLED=0 go build -o /tmp/aiprocurement-vector-sync ./cmd/aiprocurement-vector-sync`
- 服务器 smoke: 调用生产内部 sync endpoint 的 `target=finance`，成功后加入 30 分钟 cron。

## 风险与回滚关注点

- pgvector 不可用时融资匹配仍可通过 MySQL `finance_services` 兜底。
- 回滚代码后，已写入订单备注的 `ai_finance_recommendation` 为附加 JSON 字段，不影响订单主流程。
