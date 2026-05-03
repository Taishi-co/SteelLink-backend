# SL-AI-PROCUREMENT-RUNTIME-CONFIG Release Checklist

## 发布前检查

- [x] `gofmt` 已执行
- [x] `go test ./...` 已通过
- [x] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 已确认 Admin 和 Backend 连接同一个 Redis 实例与 DB
- [ ] 已确认 Redis key `ai_procurement_config` 已由 Admin 写入
- [ ] 已确认生产环境设置 `PGVECTOR_DSN`

## 发布后验证

- [ ] Admin 切换模型解析后，Backend 在 TTL 后使用新配置
- [ ] Admin 切换向量召回后，Backend 在 TTL 后使用新配置
- [ ] 删除 Redis key 后 Backend 不调用模型，回到规则解析

## 回滚

- 回滚目标版本: 合入前版本
- 回滚步骤: revert 本 feature
