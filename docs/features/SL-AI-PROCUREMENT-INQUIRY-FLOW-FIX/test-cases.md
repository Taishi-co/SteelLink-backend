# AI Procurement Inquiry Flow Fix Test Cases

## 测试范围

- Feature ID: SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 历史记录响应归一 | AI采购记录中的匹配、融资或审核 JSON 为空 | 调用 `toProcurementInfo` 或查询详情接口 | 响应中的列表字段为 `[]`，不是 `null` |
| TC-02 | 完整链路响应 | 已生成真实订单的 AI采购记录 | 查询详情或提交审核 | 返回订单草稿、阶段和列表字段结构稳定 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 空数据库句柄 | 单元测试未初始化 MySQL | 调用 AI采购物流委托创建防御路径 | 不 panic，返回 nil |
| TC-EX-02 | 旧数据空列表 | 旧记录 `matched_products` 为空 | 前端打开历史记录 | 后端返回空数组，前端可渲染空态 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 买家企业 | 查询和提交自己的 AI采购 | 权限不变，可继续访问 |
| TC-AUTH-02 | 非所属企业 | 查询他人 AI采购 | 权限不变，仍返回无权访问 |

## 回归关注点

- 旧 `POST /ai-procurements` 创建记录不受影响。
- 真实订单创建和物流委托创建不改变业务状态机。
- 融资匹配和审核提交接口路径不变。
