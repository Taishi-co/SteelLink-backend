# SL-AI-PROCUREMENT-RUNTIME-CONFIG Test Cases

## 配置同步

- 未开启 runtime sync 时，AI采购继续使用本地 env/config。
- 开启 runtime sync 且 Admin 返回 `model.parsingEnabled=true` 时，模型解析开关生效。
- 开启 runtime sync 且 Admin 返回 `matching.vectorEnabled=true` 时，向量召回开关生效。
- Admin 接口不可用、返回非 2xx 或 token 错误时，AI采购接口不失败，继续使用本地配置。

## 安全

- 业务后端只读取 Admin 保存的环境变量名，不读取 Admin 明文密钥。
- MiniMax API Key 仍来自业务后端 `MINIMAX_API_KEY`。
- Milvus 密码仍来自业务后端 `MILVUS_PASSWORD` 或 Admin 配置中的 password env 对应变量。

## 回归

- `POST /api/v1/ai-procurements/parse` 未登录仍按鉴权规则处理。
- 模型解析失败仍回退规则解析。
- 向量召回失败仍回退 MySQL 规则匹配。
