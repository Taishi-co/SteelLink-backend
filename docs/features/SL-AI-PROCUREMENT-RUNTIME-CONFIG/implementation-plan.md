# SL-AI-PROCUREMENT-RUNTIME-CONFIG Implementation Plan

## 实施内容

1. 在 `config.AIProcurementConfig` 增加 `runtime_config.redis_key`。
2. 增加 Admin 配置 JSON 映射和 Redis 缓存刷新逻辑。
3. `ProcurementService.Parse` 前从 Redis 刷新运行时配置。
4. `matchProducts` 动态创建 vector matcher，确保 Admin 开关在缓存刷新后生效。
5. 更新 `config.yaml`、`config.example.yaml` 和 `.env.example`。

## 数据与权限影响

- 无新增业务表。
- 业务后端直接读取 Redis key `ai_procurement_config`。
- Redis miss、Redis 异常或 JSON 无效时，使用代码内置默认值，不读取本地 AI env/config 开关。
- 生产环境需要 Admin 和 Backend 连接同一个 Redis 实例与 DB。

## 回滚

- 回滚代码版本。
- Redis key 缺失时默认关闭模型解析和向量召回，保持规则解析可用。
