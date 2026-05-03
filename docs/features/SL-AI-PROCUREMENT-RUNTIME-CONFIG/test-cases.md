# SL-AI-PROCUREMENT-RUNTIME-CONFIG Test Cases

## 配置同步

- Redis 中 `ai_procurement_config` 的 `model.parsingEnabled=true` 时，模型解析开关生效。
- Redis 中 `ai_procurement_config` 的 `matching.vectorEnabled=true` 时，向量召回开关生效。
- Redis miss、Redis 不可用或 JSON 无效时，AI采购接口不失败，使用代码内置默认值并继续规则兜底。
- Redis miss 时即使本地 AI env/config 打开模型解析，业务后端也不调用模型。
- Redis miss 或 `matching.vectorEnabled=false` 时，内部 vector-sync 接口拒绝同步，避免脚本绕过运行时开关写入 pgvector。

## 安全

- 业务后端只读取 Admin 保存的环境变量名，不读取 Admin 明文密钥。
- MiniMax API Key 仍来自业务后端 `MINIMAX_API_KEY`。
- pgvector DSN 仍来自业务后端 `PGVECTOR_DSN` 或 Admin 配置中的 connection env 对应变量。
- Postgres 密码仍来自业务后端 `PGVECTOR_PASSWORD` 或 Admin 配置中的 password env 对应变量。

## 回归

- `POST /api/v1/ai-procurements/parse` 未登录仍按鉴权规则处理。
- 模型解析失败仍回退规则解析。
- 向量召回失败仍回退 MySQL 规则匹配。
- 模型解析开启且配置完整时，日志包含 `AI procurement model parser started` 与 `AI procurement model parser succeeded`。
- 模型解析关闭、密钥缺失或模型失败时，日志包含跳过或失败原因，并记录 `AI procurement parse fallback to rule parser`。
