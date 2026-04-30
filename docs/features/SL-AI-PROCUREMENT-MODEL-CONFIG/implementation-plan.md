# SL-AI-PROCUREMENT-MODEL-CONFIG Implementation Plan

## 变更范围

- `config/config.go`
- `config/config.yaml`
- `config/config.example.yaml`
- `.env.example`
- `internal/modules/aiprocurement/application/service/`

## 实施步骤

1. 增加 AI 采购模型解析配置结构、默认值和环境变量覆盖。
2. 增加 OpenAI-compatible Chat Completions 调用器。
3. 将采购指令解析流程改为“模型优先，规则兜底”。
4. 补充 `.env.example` 和 feature 文档。
5. 执行 `gofmt`、`go test ./...`、`CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api`。

## 风险与回滚

- 风险：模型接口超时或返回非 JSON。缓解：服务回退现有规则解析。
- 风险：密钥误提交。缓解：只保存 API Key 环境变量名，真实密钥只放运行环境。
- 回滚：关闭 `AI_PROCUREMENT_MODEL_PARSING_ENABLED` 即可回到规则解析；必要时 revert 本次改动。
