# Test Cases: AI采购渐进式后端工作流

## 测试范围

- Feature ID: SL-AI-PROCUREMENT-GUIDED-WORKFLOW
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 指令解析 | 已登录买家企业 | 调用 `POST /ai-procurements/parse` | 返回 parsed 和指令解析阶段，不直接生成订单 |
| TC-02 | 确认参数 | 已登录买家企业 | 调用 `POST /ai-procurements/confirm-instruction` | 返回修正后的 parsed，流程进入商品匹配 |
| TC-03 | 商品召回 | 已确认 parsed | 调用 `POST /ai-procurements/match-products` | 返回最多 10 个候选商品，包含商品详情字段 |
| TC-04 | 生成订单 | 已选择商品并填写订单表单 | 调用 `POST /ai-procurements/generate-order` | 创建真实订单，返回 AI采购流程记录和订单草稿 |
| TC-05 | 需要融资 | 订单已生成且融资需求为 true | 调用 `POST /ai-procurements/:id/match-finance` | 返回融资方案并保持融资匹配阶段 |
| TC-06 | 不需要融资 | 订单已生成且融资需求为 false | 调用 `POST /ai-procurements/:id/match-finance` | 返回空融资方案并进入审核提交阶段 |
| TC-07 | 提交并保存历史 | 订单已生成 | 调用 `POST /ai-procurements/:id/submit-review` 且 `save_history=true` | 状态更新为审核提交，记录出现在我的AI采购 |
| TC-08 | 显式品类解析 | 已开启大模型解析 | 输入 `我想买100吨的钢，品类为热轧板卷` 调用 `POST /ai-procurements/parse` | 返回 `product_name=钢`、`category=热轧板卷`、`quantity=100`、`unit=吨` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未登录访问 | 无登录态 | 调用任一 AI采购业务接口 | 返回未登录 |
| TC-EX-02 | 未选商品生成订单 | 已确认 parsed | 调用 `generate-order` 但 product_id 为空 | 返回业务错误 |
| TC-EX-03 | 未生成订单提交审核 | 仅有旧 AI采购草稿 | 调用 `submit-review` | 返回请先生成真实订单 |
| TC-EX-04 | 不保存历史 | 订单已生成 | `submit-review` 传 `save_history=false` | 提交成功但不会出现在我的AI采购列表 |
| TC-EX-05 | 模型布尔字段抖动 | 大模型返回 `financing_required` 为字符串 | 解析模型响应 | 不因 `"true"`、`"false"`、`"是"`、`"否"` 类型抖动整体回落规则解析 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 买家企业 | 完整 AI采购流程 | 可执行 |
| TC-AUTH-02 | 非企业用户 | 调用新增接口 | 被企业用户鉴权拦截 |

## 回归关注点

- 旧 `POST /ai-procurements` 仍可创建 AI采购记录。
- 真实订单创建仍复用订单模块库存锁定规则。
- 向量召回不可用时商品匹配回落 MySQL 查询。
