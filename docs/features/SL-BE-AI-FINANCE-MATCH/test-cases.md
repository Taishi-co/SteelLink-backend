# AI Procurement Finance Product Matching Test Cases

## 测试范围

- Feature ID: `SL-BE-AI-FINANCE-MATCH`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | AI采购召回融资产品 | 存在已审核 `finance_services`，订单金额在额度范围内 | 调用 `match-finance` 且 `financing_required=true` | 返回真实金融产品，包含产品 ID、费率、期限、推荐原因 |
| TC-02 | 用户预选 AI 推荐产品 | TC-01 返回产品后选择一个 `finance_service_id` | 再次调用 `match-finance` 并传 `selected_finance_service_id` | 对应方案 `selected=true`、`ai_recommended=true`，订单备注写入推荐信息 |
| TC-03 | 融资产品增量同步 | pgvector 与 embedding 配置可用 | 调用内部 vector-sync `target=finance` | 仅同步水位后的融资产品到 `ai_finance_product_embeddings` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | pgvector 不可用 | vector enabled 但连接失败 | 调用 `match-finance` | 记录日志并回退 MySQL 金额范围匹配 |
| TC-EX-02 | 无已审核金融产品 | `finance_services` 无匹配记录 | 调用 `match-finance` | 返回空融资方案，审核项保持待匹配 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 非归属企业用户 | 调用他人的 AI采购 `match-finance` | 返回无权匹配该 AI采购单 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端验证 | `./harness/verify.sh` | Go 测试和构建通过 |
| TC-OPS-02 | 向量同步命令 | `cmd/aiprocurement-vector-sync --target=finance --schedule=manual` | 后端返回 finance sync succeeded |
