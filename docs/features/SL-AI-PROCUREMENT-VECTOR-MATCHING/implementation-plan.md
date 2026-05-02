# SL-AI-PROCUREMENT-VECTOR-MATCHING Implementation Plan

## 当前变更

- 在 `ai_procurement.vector_search` 中补齐 pgvector/embedding/hybrid 参数。
- 在 `.env.example` 中补齐 pgvector 和 embedding 推荐环境变量。
- 新增 `internal/pkg/embedding`，通过 OpenAI-compatible embeddings API 生成采购查询向量。
- 新增 `internal/pkg/vectorstore`，通过 pgvector 搜索 `ai_product_embeddings` 商品向量表。
- 将 `matchProducts` 改为向量优先、MySQL 规则兜底；向量默认关闭，不改变当前生产行为。

## 已完成实施

1. 配置层支持 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED`、`PGVECTOR_DSN_ENV`、`PGVECTOR_DATABASE`、`PGVECTOR_USERNAME`、`PGVECTOR_PASSWORD_ENV`、`AI_EMBEDDING_*`、hybrid 权重和 score 阈值。
2. 商品匹配链路启用向量时执行：
   - 组装采购语义文本
   - 调用 embedding API
   - 搜索 pgvector 商品向量表
   - 回读 MySQL 商品详情并过滤上架状态、品类和可用库存
   - 按向量相似度、关键词、库存满足率做 hybrid 重排
3. pgvector 或 embedding 不可用、无命中、命中商品已下架时，自动回退原 MySQL 规则匹配。

## 后续实施步骤

1. 新增商品/供应商/融资产品向量同步 job：
   - 商品上下架/编辑后同步 `ai_product_embeddings`
   - 企业认证/资料更新后同步 `ai_supplier_embeddings`
   - 融资产品配置变更后同步 `ai_finance_product_embeddings`
2. 增加 pgvector 表初始化/迁移脚本，统一创建 schema 和向量索引。
3. 将供应商和融资产品也纳入 AI 采购结果，补齐供应商履约评分、金融产品约束匹配和审核项。
4. 管理端增加向量库状态页或状态卡片，展示集合数量、最近同步时间、失败原因和手动重建入口。

## 回滚

- 设置 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`。
- 保留 MySQL 规则匹配作为兜底路径。
