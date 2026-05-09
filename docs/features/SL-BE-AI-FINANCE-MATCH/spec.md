# AI Procurement Finance Product Matching Spec

## 基本信息

- Feature ID: `SL-BE-AI-FINANCE-MATCH`
- Feature 名称: AI采购融资产品匹配与 pgvector 增量同步
- 负责人: Codex
- 目标上线环境: production

## 背景

AI采购的融资动作只能发生在真实订单生成、买卖双方合同签署并进入支付/融资节点之后。AI采购阶段不能替用户发起融资申请，但可以基于订单金额、采购用途和平台已审核金融产品，先推荐若干融资产品供用户预选；订单进入真实融资申请时，预选产品应被带回并标注为 AI 推荐。

## 范围

### In Scope

- `match-finance` 从已审核 `finance_services` 生成融资产品推荐，不再返回硬编码融资方案。
- `match-finance` 支持 `selected_finance_service_id`，保存用户在 AI采购阶段预选的融资产品。
- 将预选融资产品写入关联订单 `remark.ai_finance_recommendation`，供订单支付/融资节点展示。
- 将 pgvector 同步目标 `finance` 从 no-op 改为真实融资产品增量同步。

### Out of Scope

- 不自动创建 `order_financing_requests`。
- 不调整资金方审批、债务合同、放款和质押流程。
- 不新增权限角色。

## 接口与数据影响

- 受影响路由/handler: `POST /api/v1/ai-procurements/:id/match-finance`, `POST /api/v1/internal/ai-procurements/vector-sync`
- 受影响服务: `internal/modules/aiprocurement/application/service`
- 受影响 middleware: 无
- 受影响表/model: 读取 `finance_services`；更新 `ai_procurements.finance_plans`；更新 `orders.remark.ai_finance_recommendation`；新增/维护 pgvector 表 `ai_finance_product_embeddings` 及其同步状态表
- 配置或环境变量变化: 复用既有 `AI_FINANCE_PRODUCT_VECTOR_COLLECTION`、pgvector 连接与 embedding 配置
- Docker / 部署影响: `cmd/aiprocurement-vector-sync --target=finance --schedule=30m` 可作为服务器定时任务

## 权限与合规影响

- 受影响角色: 买家企业用户、资金方服务配置方
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不改变融资审批；仅提前记录 AI 推荐来源，真实申请仍走订单融资接口校验金额和产品状态。

## 验收标准

1. AI采购融资匹配返回真实已审核金融产品，并包含 `finance_service_id`、金额、费率、期限、材料和 AI 推荐标记。
2. 用户选择融资产品后，AI采购记录中的该产品标记为 `selected=true`，关联订单 `remark.ai_finance_recommendation` 写入产品信息。
3. `cmd/aiprocurement-vector-sync --target=finance` 通过后端内部接口增量同步 MySQL `finance_services` 到 pgvector。
