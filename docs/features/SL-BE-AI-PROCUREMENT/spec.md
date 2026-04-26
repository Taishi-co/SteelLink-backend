# SL-BE-AI-PROCUREMENT Spec

## 背景

买家在商城前端通过自然语言描述采购诉求后，需要后端提供稳定的 AI 采购工作流接口，支撑指令解析、商品匹配、订单草稿、融资方案和审核提交。

## 范围

- 新增 `ai_procurements` 持久化表。
- 新增买家侧受保护 API：
  - `POST /api/v1/ai-procurements/parse`
  - `POST /api/v1/ai-procurements`
  - `GET /api/v1/ai-procurements/mine`
  - `GET /api/v1/ai-procurements/:id`
  - `POST /api/v1/ai-procurements/:id/submit-review`
- 输出五阶段可视化状态：指令解析、商品匹配、订单生成、融资匹配、审核提交。
- 当前版本使用确定性规则解析和商品库查询，保留后续接入 LLM、订单中心和融资系统的结构字段。

## 非范围

- 不直接创建正式交易订单。
- 不直接发起真实融资申请。
- 不暴露平台管理端审核列表。
