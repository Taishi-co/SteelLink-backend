# SL-AI-PROCUREMENT-RUNTIME-CONFIG Spec

## 背景

AI采购模型解析和向量召回已经支持 env/config 配置，但运营希望在 Admin 平台通过开关调整，不再频繁改业务后端 `.env` 并重启。

## 目标

- SteelLink-backend 支持从 Redis 读取 AI采购运行时配置。
- SteelLink-admin 保存配置到 MySQL 后同步写入 Redis。
- Admin 配置覆盖模型解析开关、模型参数、向量召回开关、pgvector 和 embedding 元数据。
- 真实密钥仍只从业务后端环境变量读取。
- Redis miss、Redis 异常或配置 JSON 无效时，业务后端使用代码内置默认值并继续规则兜底，不再拉 Admin HTTP 或使用本地 AI env/config 开关。

## 范围

- 新增 `ai_procurement.runtime_config.redis_key` 配置。
- 新增 `AI_PROCUREMENT_RUNTIME_CONFIG_REDIS_KEY` 环境变量。
- `Parse` 前按缓存 TTL 从 Redis 刷新 Admin 写入的配置。
- 向量 matcher 每次匹配时读取最新配置，避免服务启动后开关不生效。
- `Parse` 输出模型解析链路日志，标明模型解析跳过、开始、成功、失败回退和规则兜底原因，便于生产确认是否实际调用大模型。
- `match-products` 输出商品匹配链路日志，标明向量匹配是否启用、向量召回成功/失败/空结果、规则查询阶段和参考方案兜底原因。

## 非范围

- 不从 Admin 同步真实 API Key、pgvector DSN 或 Postgres 密码。
- 不实现配置变更审计，审计仍在 Admin 后端记录。

## 生产配置

- Admin 和业务后端使用同一个 Redis 实例与 DB。
- Redis key 默认为 `ai_procurement_config`，可用 `AI_PROCUREMENT_RUNTIME_CONFIG_REDIS_KEY` 覆盖。
- Redis miss 时默认不开启模型解析和向量召回。
