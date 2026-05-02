# SL-AI-PROCUREMENT-VECTOR-MATCHING Test Cases

## 配置

- 默认 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`，现有匹配链路不变。
- 设置 pgvector 相关环境变量后，配置可被 `config.Init` 读取。
- `.env.example` 不包含 pgvector DSN/Postgres 密码和 embedding API Key 明文。

## 向量召回

- 启用向量且 pgvector 命中上架商品时，AI 采购优先返回向量召回商品。
- pgvector 不可用时，AI 采购解析仍回退 MySQL 规则匹配。
- pgvector 命中商品不存在、已下架或品类不匹配时，AI 采购仍回退 MySQL 规则匹配。
- 商品集合后续可按品类、状态、区域、库存做 scalar filter。
- 供应商集合后续可按认证状态和服务区域过滤。
- 融资产品集合后续可按金额范围和账期过滤。

## 重排

- 同等向量相似度下，库存满足率高的商品优先。
- 同等商品匹配下，后续应优先选择已认证和履约评分高的供应商。
- 融资需求存在时，后续应优先选择金额范围覆盖且材料要求少的融资产品。

## 回归

- 未启用向量时，`POST /api/v1/ai-procurements/parse` 输出与当前规则匹配兼容。
- 商品、供应商和融资产品同步失败不影响 AI 采购主接口可用性。
- `go test ./...` 通过。
- `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 通过。
