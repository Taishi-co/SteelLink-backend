# Test Cases

## 测试范围

- Feature ID: `SL-BE-ORDER-STATUS-FLOW`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 创建订单 | 商品已审核通过且库存充足 | 调用 `POST /api/v1/orders` | 返回订单，状态为 `pending_contract` |
| TC-02 | 提交合同 | 订单处于 `pending_contract` | 调用 `POST /api/v1/orders/:id/contract/submit` | 返回订单，状态变为 `pending_payment` |
| TC-03 | 上传支付回单 | 订单处于 `pending_payment` | 调用 `POST /api/v1/orders/:id/payment-receipts` | 返回订单，状态变为 `pending_payment_audit` |
| TC-04 | 平台审核通过 | 订单处于 `pending_payment_audit` | 调用 `PUT /api/v1/admin/orders/:id/payment-audit`，状态为 `approved` | 返回订单，状态变为 `pending_pickup`，包含提货码 |
| TC-05 | 核销提货码 | 订单处于 `pending_pickup` | 调用 `POST /api/v1/orders/:id/pickup-code/verify` | 返回订单，状态变为 `completed` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 非法状态提交合同 | 订单不是 `pending_contract` | 调用合同提交流程 | 返回 400 |
| TC-EX-02 | 回单金额不匹配 | 订单处于 `pending_payment` | 上传与订单金额不一致的回单 | 返回 400 |
| TC-EX-03 | 错误提货码核销 | 订单处于 `pending_pickup` | 提交错误 `code` | 返回 400 |

## 接口与后端验证

| 编号 | 场景 | 验证命令 / 步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 主业务后端可构建 | `go build ./...` | 构建通过 |
| TC-BE-02 | 迁移可执行 | 启动服务触发 AutoMigrate | 新模型表成功创建 |
