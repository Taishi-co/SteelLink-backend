# AI Procurement Inquiry Flow Fix Spec

## 基本信息

- Feature ID: SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX
- Feature 名称: AI采购询价链路通畅性修复
- 负责人: Codex
- 目标上线环境: backend api

## 背景

AI采购询价链路依赖后端返回稳定的阶段数据、商品匹配、订单草稿、融资方案和审核项。历史记录或旧数据中 JSON 列表可能为空或 null，前端打开时需要稳定收到空列表语义；物流委托创建也应在无数据库句柄的局部测试场景下安全返回，避免 panic 干扰链路验证。

## 范围

### In Scope

- `toProcurementInfo` 输出时将空 JSON 列表归一为空数组。
- `order_draft.orders`、`finance_plans`、`audit_items`、`matched_products` 为空时保持可序列化的空列表语义。
- AI采购物流委托创建增加数据库句柄防御。
- 补充服务层单元测试覆盖空列表归一和空数据库防御。

### Out of Scope

- 不新增数据库字段。
- 不改变 AI采购接口路径和权限要求。
- 不改变真实订单状态机。

## 接口影响

- 受影响 API:
  - `GET /api/v1/ai-procurements/mine`
  - `GET /api/v1/ai-procurements/:id`
  - `POST /api/v1/ai-procurements/generate-order`
  - `POST /api/v1/ai-procurements/:id/match-finance`
  - `POST /api/v1/ai-procurements/:id/submit-review`
- 请求/响应变化: 不新增字段；空列表从 `null` 归一为 `[]`。
- 环境变量变化: 无

## 数据影响

- 受影响表: `ai_procurements`
- 迁移要求: 无
- 回滚策略: 回退服务层响应归一化和测试。

## 权限影响

- 受影响角色: 买家企业用户
- 新增/调整权限: 无

## 验收标准

1. AI采购历史记录中空 JSON 列表返回为空数组。
2. 空数据库句柄下调用物流委托创建防御逻辑不会 panic。
3. 现有 AI采购服务层单元测试通过。
4. `cmd/api` 构建通过。
