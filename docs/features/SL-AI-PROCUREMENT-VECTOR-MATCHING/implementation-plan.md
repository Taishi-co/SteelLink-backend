# SL-AI-PROCUREMENT-VECTOR-MATCHING Implementation Plan

## 当前变更

- 在 `ai_procurement.vector_search` 中补齐 Milvus/embedding/hybrid 参数。
- 在 `.env.example` 中补齐 Milvus 和 embedding 推荐环境变量。
- 将向量召回作为默认关闭能力，不改变当前生产行为。

## 后续实施步骤

1. 新增 `internal/pkg/vectorstore`，封装 Milvus 连接、collection 初始化、upsert、search。
2. 新增 `internal/pkg/embedding`，封装 MiniMax embedding 调用。
3. 新增 `internal/modules/aiprocurement/application/matching`，把商品、供应商、融资产品召回从 service 中拆出。
4. 新增同步 job：
   - 商品上下架/编辑后同步 `ai_product_embeddings`
   - 企业认证/资料更新后同步 `ai_supplier_embeddings`
   - 融资产品配置变更后同步 `ai_finance_product_embeddings`
5. `matchProducts` 改为：
   - vector enabled -> hybrid retriever
   - vector error/no hit -> current MySQL retriever
6. 管理端增加向量库状态页或状态卡片。

## 回滚

- 设置 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`。
- 保留 MySQL 规则匹配作为兜底路径。
