# Feature Spec: AI采购渐进式后端工作流

## 基本信息

- Feature ID: SL-AI-PROCUREMENT-GUIDED-WORKFLOW
- Feature 名称: AI采购渐进式后端工作流
- 负责人: Codex
- 目标上线环境: backend api

## 背景

AI采购前端需要从一次性展示结果升级为客户逐步确认的业务流程。用户应先确认大模型解析参数，再召回商品并人工选择商品，随后生成真实订单、按需匹配融资，最后提交审核并决定是否保存到历史AI采购。

## 范围

### In Scope

- 拆分 AI采购接口，支持参数确认、商品召回、真实订单生成、融资匹配和带保存选项的审核提交。
- 指令解析支持大模型 JSON 容错和“品类为 X”显式句式归一化，避免模型字段类型抖动或商品名/品类边界不清导致回落到规则解析。
- 商品匹配返回最多 10 个候选商品，优先复用向量召回，失败时回落 MySQL 规则查询。
- 生成订单阶段复用真实订单服务创建订单，并把订单摘要写回 AI采购流程记录。
- 融资匹配按 `financing_required` 开关执行；不需要融资时流程跳过融资阶段。
- 历史AI采购只展示用户选择保存的记录。

### Out of Scope

- 不新增 Milvus/pgvector SDK 能力；继续复用现有向量匹配封装和 SQL fallback。
- 不改真实订单主流程的状态机。
- 不实现融资产品真实授信申请。

## 接口影响

- 保留:
  - `POST /api/v1/ai-procurements/parse`
  - `POST /api/v1/ai-procurements`
  - `GET /api/v1/ai-procurements/mine`
  - `GET /api/v1/ai-procurements/:id`
  - `POST /api/v1/ai-procurements/:id/submit-review`
- 新增:
  - `POST /api/v1/ai-procurements/confirm-instruction`
  - `POST /api/v1/ai-procurements/match-products`
  - `POST /api/v1/ai-procurements/generate-order`
  - `POST /api/v1/ai-procurements/:id/match-finance`
- 调整:
  - `POST /api/v1/ai-procurements/:id/submit-review` 请求体支持 `save_history`。

## 数据影响

- `ai_procurements` 新增 `history_saved` 字段，默认 true。
- `matched_products` JSON 增加商品详情字段，便于前端展示详情并选择商品。
- `order_draft` JSON 增加真实订单 ID/单号、订单表单字段和订单详情快照。

## 权限影响

- 受影响角色: 买家企业用户
- 新增权限: 无。新增接口继续复用企业用户 JWT 鉴权。

## 验收标准

1. `parse` 只返回可编辑的解析结果和当前步骤，不再直接推动完整流程。
2. 用户确认参数后可调用 `match-products` 获取最多 10 个候选商品。
3. 用户选择商品并填写订单信息后，`generate-order` 创建真实订单并生成 AI采购流程记录。
4. `match-finance` 可按融资意图生成或跳过融资方案。
5. `submit-review` 可携带 `save_history`，并影响我的AI采购历史列表。
