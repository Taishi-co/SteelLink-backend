# SL-AI-PROCUREMENT-RUNTIME-CONFIG Spec

## 背景

AI采购模型解析和向量召回已经支持 env/config 配置，但运营希望在 Admin 平台通过开关调整，不再频繁改业务后端 `.env` 并重启。

## 目标

- SteelLink-backend 支持从 SteelLink-admin 拉取 AI采购运行时配置。
- Admin 配置覆盖模型解析开关、模型参数、向量召回开关、Milvus 和 embedding 元数据。
- 真实密钥仍只从业务后端环境变量读取。
- 拉取失败时保留本地 env/config 配置并继续规则兜底。

## 范围

- 新增 `ai_procurement.runtime_config` 配置段。
- 新增 `AI_PROCUREMENT_RUNTIME_CONFIG_*` 环境变量。
- `Parse` 前按缓存 TTL 刷新 Admin 配置。
- 向量 matcher 每次匹配时读取最新配置，避免服务启动后开关不生效。

## 非范围

- 不从 Admin 同步真实 API Key 或 Milvus 密码。
- 不实现 Admin 推送配置到业务后端。
- 不实现配置变更审计，审计仍在 Admin 后端记录。

## 生产配置

- `AI_PROCUREMENT_RUNTIME_CONFIG_ENABLED=true`
- `AI_PROCUREMENT_RUNTIME_CONFIG_ENDPOINT=https://<admin-host>/api/v1/internal/ai-procurement/runtime-config`
- `AI_PROCUREMENT_RUNTIME_CONFIG_TOKEN_ENV=AI_PROCUREMENT_CONFIG_SYNC_TOKEN`
- `AI_PROCUREMENT_CONFIG_SYNC_TOKEN=<shared-internal-token>`
