# SL-AI-PROCUREMENT-VECTOR-MATCHING Test Cases

## 配置

- 默认 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`，现有匹配链路不变。
- 设置 Milvus 相关环境变量后，配置可被 `config.Init` 读取。
- `.env.example` 不包含 Milvus 密码和 embedding API Key 明文。

## 向量召回

- Milvus 不可用时，AI 采购解析仍回退 MySQL 规则匹配。
- 商品集合可按品类、状态、区域、库存做 scalar filter。
- 供应商集合可按认证状态和服务区域过滤。
- 融资产品集合可按金额范围和账期过滤。

## 重排

- 同等向量相似度下，库存满足率高的商品优先。
- 同等商品匹配下，已认证和履约评分高的供应商优先。
- 融资需求存在时，金额范围覆盖且材料要求少的融资产品优先。

## 回归

- 未启用向量时，`POST /api/v1/ai-procurements/parse` 输出与当前规则匹配兼容。
- 商品、供应商和融资产品同步失败不影响 AI 采购主接口可用性。
