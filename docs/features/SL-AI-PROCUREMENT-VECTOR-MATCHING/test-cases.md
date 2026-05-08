# SL-AI-PROCUREMENT-VECTOR-MATCHING Test Cases

## 配置

- 默认 `AI_PROCUREMENT_VECTOR_SEARCH_ENABLED=false`，现有匹配链路不变。
- 设置 pgvector 相关环境变量后，配置可被 `config.Init` 读取。
- `.env.example` 不包含 pgvector DSN/Postgres 密码和 embedding API Key 明文。

## 向量召回

- 启用向量且 pgvector 命中上架商品时，AI 采购优先返回向量召回商品。
- pgvector 不可用时，AI 采购解析仍回退 MySQL 规则匹配。
- pgvector 命中商品不存在、已下架或品类不匹配时，AI 采购仍回退 MySQL 规则匹配。
- `ai_product_embeddings` 只要求最小索引字段：`product_id`、`product_no`、`product_name`、`category`、`enterprise_id`、`embedding`。
- `ai_product_embeddings` 额外保存 `source_updated_at` 和 `synced_at`，用于增量水位和同步审计，不承载库存/价格等业务字段。
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
- `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/aiprocurement-vector-sync` 通过。

## 同步命令

- `POST /api/v1/internal/ai-procurements/vector-sync` 使用 `X-AI-Procurement-Sync-Token` 校验内部同步 token。
- 后端接口负责读取 MySQL、调用 embedding、写入 pgvector；脚本只调用后端接口。
- 每次同步只处理 `(products.updated_at, products.id)` 大于 `ai_product_embeddings_sync_state` 当前水位的商品，不做全量扫描。
- 单个商品成功写入 pgvector 后立即推进同步水位；中途失败时下次从失败点继续增量重试。
- 旧表第一次升级且没有同步状态时，后端初始化水位，避免 30 分钟定时任务自动全量重跑。
- `--schedule=manual` 调用一次后端商品向量同步接口后退出。
- `--schedule=minute|hour|day|30m` 按固定间隔重复同步，收到 SIGINT/SIGTERM 后退出。
- `--target=suppliers|finance` 当前为预留 no-op，不影响产品同步。
- `scripts/install-ai-vector-sync-cron.sh` 在服务器安装 `*/30 * * * *` cron，并通过 `flock` 防止并发同步。
