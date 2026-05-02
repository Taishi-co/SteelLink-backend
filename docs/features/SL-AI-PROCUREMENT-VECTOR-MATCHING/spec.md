# SL-AI-PROCUREMENT-VECTOR-MATCHING Spec

## 当前实现

### 意图识别

AI 采购解析入口在 `ProcurementService.Parse`：

1. 优先调用模型解析，把自然语言采购指令转成结构化 JSON。
2. 模型解析未启用、缺少 API Key 或调用失败时，回退规则解析。
3. 结构化结果包含商品名称、品类、数量、单位、用途、融资意图、关键词和置信度。

规则解析能力包括：

- 数量/单位：正则提取 `100吨`、`20套` 这类表达。
- 商品名：优先提取 `采购...` 后面的商品片段，再匹配常见关键词。
- 品类：按商品名和原始 prompt 中的关键词推断。
- 融资意图：识别 `融资`、`账期` 等词。

### 商品匹配

当前商品匹配是 MySQL 规则召回：

1. 用解析出的 `product_name` 作为关键词、`category` 作为品类过滤。
2. 只查上架商品，按库存倒序。
3. 没有关键词命中时，用品类再查一次。
4. 召回后按名称包含关系给 82/95 的简单分数。
5. 仍无结果时生成参考采购方案。

## 问题

- `LIKE` 无法识别“HRB400E 螺纹钢”和“抗震三级钢筋”这类语义等价。
- 供应商能力、交付区域、历史履约、金融产品适配都没有进入统一召回。
- 召回与排序耦合在业务服务里，后续扩展物流/仓储/金融会越来越硬编码。

## 目标

- 引入 pgvector 作为 AI 采购向量召回层。
- 商品、供应商、融资产品分别建立向量表。
- AI 采购匹配采用 hybrid 策略：关键词/结构化过滤 + 向量召回 + 业务重排。
- 当前版本先补齐配置与设计面，保留 MySQL 规则匹配作为默认兜底。

## pgvector 表设计

### `ai_product_embeddings`

- `entity_id`: 商品 ID
- `entity_type`: `product`
- `enterprise_id`: 供应商企业 ID
- `status`: 商品状态
- `category`: 一级品类
- `sub_category`: 二级品类
- `province`, `city`: 仓库区域
- `price`: 单价
- `available_qty`: 可用库存
- `text`: 用于 embedding 的商品文本
- `embedding`: 向量
- `updated_at`: 同步时间

商品 embedding 文本建议：

```text
商品名 + 品类 + 规格材质 + 钢厂 + 仓库城市 + 交付条件 + 价格区间 + 库存描述
```

### `ai_supplier_embeddings`

- `entity_id`: 企业 ID
- `entity_type`: `supplier`
- `status`: 认证状态
- `main_categories`: 主营品类
- `service_regions`: 服务区域
- `delivery_score`: 履约评分
- `text`: 企业能力文本
- `embedding`: 向量
- `updated_at`: 同步时间

### `ai_finance_product_embeddings`

- `entity_id`: 金融产品 ID
- `entity_type`: `finance_product`
- `provider`: 金融机构
- `min_amount`, `max_amount`: 金额范围
- `term_days`: 账期/期限
- `required_docs`: 必需材料
- `text`: 产品适配文本
- `embedding`: 向量
- `updated_at`: 同步时间

## 召回流程

1. 意图识别输出结构化字段。
2. 生成 query embedding，文本包含商品、用途、数量、地域、融资诉求。
3. pgvector 搜索商品向量表，使用状态、品类、区域、库存等 scalar filter。
4. pgvector 搜索供应商向量表，召回可履约供应商。
5. 如 `financing_required=true`，搜索融资产品集合。
6. 对召回结果做重排：
   - 向量相似度
   - 品类/规格硬匹配
   - 库存满足率
   - 价格容忍度
   - 供应商认证/履约优先级
   - 融资金额和资料适配
7. 如果 pgvector 或 embedding 服务不可用，回退当前 MySQL 规则匹配。

## 配置

后端新增 `ai_procurement.vector_search` 配置段，默认关闭：

- `enabled`
- `provider`
- `endpoint`
- `database`
- `product_collection`，pgvector 模式下表示商品向量表名
- `supplier_collection`
- `finance_product_collection`，pgvector 模式下表示融资产品向量表名
- `embedding_provider`
- `embedding_base_url`
- `embedding_api_key_env`
- `embedding_model`
- `dimension`
- `top_k`
- `min_score`
- `hybrid_keyword_weight`
- `hybrid_vector_weight`

## 分阶段交付

### P0: 配置与设计

- 补齐后端配置、环境变量和管理端配置项。
- 默认关闭向量召回。

### P1: pgvector 基础设施

- 增加 pgvector client。
- 增加向量表初始化脚本。
- 增加商品/供应商/融资产品 embedding 同步 job。

### P2: 查询链路接入

- `matchProducts` 先走 vector/hybrid 召回。
- 失败时回退 MySQL。
- 输出匹配原因包含向量相似度和业务重排理由。

### P3: 运营闭环

- 管理端展示向量库状态、集合数量、最近同步时间。
- 支持手动重建索引、单商品重算 embedding。
