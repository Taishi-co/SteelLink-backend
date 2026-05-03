# SL-AI-PROCUREMENT-MODEL-CONFIG Test Cases

## 配置

- 未设置 `AI_PROCUREMENT_MODEL_PARSING_ENABLED` 时，服务保持规则解析。
- 设置 `AI_PROCUREMENT_MODEL_PARSING_ENABLED=true` 但未设置 `MINIMAX_API_KEY` 时，服务回退规则解析。
- `.env.example` 不包含真实 API Key。

## API

- `POST /api/v1/ai-procurements/parse` 在模型不可用时仍返回结构化解析结果。
- 模型返回 JSON 代码块或纯 JSON 时，都应能解析出结构化字段。
- 模型响应缺字段时，应使用规则解析结果补齐默认字段。
- Base URL 为 `/anthropic/v1/messages` 时，使用 Anthropic Messages 请求体并从 `content[].text` 解析 JSON。
- Base URL 只填到 `/anthropic` 时，后端自动补全为 `/anthropic/v1/messages`。

## 回归

- AI 采购创建、列表、详情、提交审核接口不变。
- 商品、订单、市场、供求、用户路由不受影响。
