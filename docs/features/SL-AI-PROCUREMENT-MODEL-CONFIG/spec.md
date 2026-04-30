# SL-AI-PROCUREMENT-MODEL-CONFIG Spec

## 背景

AI 采购后端已具备规则解析、商品匹配、订单草稿、融资方案和审核提交结构，但采购指令解析仍是本地规则，缺少可配置的模型解析入口。管理端已经有 AI 采购配置页，但模型配置还不足以表达 MiniMax 这类 OpenAI-compatible provider 的生产参数。

## 目标

- 为 `SteelLink-backend` 增加 AI 采购模型解析配置。
- 支持 MiniMax/OpenAI-compatible Chat Completions 风格的结构化采购指令解析。
- 在未启用模型解析、未配置 API Key 或模型调用失败时，自动回退到现有规则解析。
- API Key 只通过环境变量读取，不写入配置文件或数据库。

## 范围

- 新增 `ai_procurement` 配置段和环境变量覆盖。
- `POST /api/v1/ai-procurements/parse` 与创建采购单流程优先使用模型解析，失败回退规则解析。
- `.env.example` 记录 MiniMax 推荐配置。

## 非范围

- 不在买家侧接口中暴露 API Key。
- 不直接对接管理端配置同步到买家侧后端。
- 不替代商品匹配、订单、融资和审核的现有确定性逻辑。

## 推荐 MiniMax 配置

- `AI_PROCUREMENT_MODEL_PARSING_ENABLED=true`
- `AI_PROCUREMENT_PROVIDER=minimax`
- `AI_PROCUREMENT_BASE_URL=https://api.minimax.io/v1/chat/completions`
- `AI_PROCUREMENT_API_KEY_ENV=MINIMAX_API_KEY`
- `MINIMAX_API_KEY=<真实密钥>`
- `AI_PROCUREMENT_MODEL=MiniMax-M2.7`
- `AI_PROCUREMENT_TEMPERATURE=0.2`
- `AI_PROCUREMENT_TOP_P=0.9`
- `AI_PROCUREMENT_MAX_TOKENS=1200`
- `AI_PROCUREMENT_TIMEOUT_SEC=20`
