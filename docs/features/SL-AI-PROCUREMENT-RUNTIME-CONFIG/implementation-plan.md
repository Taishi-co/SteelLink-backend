# SL-AI-PROCUREMENT-RUNTIME-CONFIG Implementation Plan

## 实施内容

1. 在 `config.AIProcurementConfig` 增加 `runtime_config`。
2. 增加 Admin runtime config JSON 映射和缓存刷新逻辑。
3. `ProcurementService.Parse` 前刷新运行时配置。
4. `matchProducts` 动态创建 vector matcher，确保 Admin 开关在缓存刷新后生效。
5. 更新 `config.yaml`、`config.example.yaml` 和 `.env.example`。

## 数据与权限影响

- 无新增业务表。
- 新增出站 HTTP 请求到 Admin 内部配置接口。
- 生产环境需要 Admin 和 Backend 配置同一个 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN`。

## 回滚

- 设置 `AI_PROCUREMENT_RUNTIME_CONFIG_ENABLED=false`。
- 保留本地 env/config 作为兜底配置。
