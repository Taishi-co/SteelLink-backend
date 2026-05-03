# SL-AI-PROCUREMENT-MODEL-CONFIG Release Checklist

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-MODEL-CONFIG
- 发布负责人:
- 目标环境: backend

## 发布前检查

- [ ] `gofmt` 已执行
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 已确认生产环境通过 `MINIMAX_API_KEY` 注入真实密钥
- [ ] 已确认 `AI_PROCUREMENT_MODEL_PARSING_ENABLED` 默认关闭，支持灰度启用
- [ ] 已确认 Anthropic Messages endpoint 使用 `/anthropic/v1/messages`
- [ ] 已确认模型失败回退规则解析

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: 设置 `AI_PROCUREMENT_MODEL_PARSING_ENABLED=false` 或 revert 本次提交
- 回滚后验证: `POST /api/v1/ai-procurements/parse`
