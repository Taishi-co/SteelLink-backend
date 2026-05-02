# SL-AI-PROCUREMENT-VECTOR-MATCHING Release Checklist

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-VECTOR-MATCHING
- 发布负责人:
- 目标环境: backend

## 发布前检查

- [x] `gofmt` 已执行
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [x] 默认 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`
- [x] `.env.example` 不包含真实 Milvus 密码或 embedding API Key
- [ ] 已确认 Milvus 集合命名、维度和 embedding 模型一致
- [ ] 已确认 `ai_product_embeddings` 中存在 `product_id` 字段并能关联 MySQL `products.id`

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: 设置 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false` 或 revert 本次向量配置改动
- 回滚后验证: `POST /api/v1/ai-procurements/parse`
