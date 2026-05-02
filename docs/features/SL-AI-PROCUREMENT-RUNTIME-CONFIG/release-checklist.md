# SL-AI-PROCUREMENT-RUNTIME-CONFIG Release Checklist

## 发布前检查

- [x] `gofmt` 已执行
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 已确认生产环境设置 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN`
- [ ] 已确认 Admin runtime config endpoint 可从 Backend 服务器访问
- [ ] 已确认生产环境设置 `PGVECTOR_DSN`

## 发布后验证

- [ ] Admin 切换模型解析后，Backend 在 TTL 后使用新配置
- [ ] Admin 切换向量召回后，Backend 在 TTL 后使用新配置
- [ ] 关闭 runtime sync 后 Backend 回到本地 env/config

## 回滚

- 回滚目标版本: 合入前版本
- 回滚步骤: 设置 `AI_PROCUREMENT_RUNTIME_CONFIG_ENABLED=false` 或 revert 本 feature
