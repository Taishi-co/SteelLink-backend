# Test Cases

## 测试范围

- Feature ID: `SL-BE-WORKBENCH-FLOW-BASELINE`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 旧前端路径兼容 | 订单处于待签署物流合同状态，当前企业为买家或卖家 | 调用 `POST /api/v1/orders/:id/logistics-contracts/buyer-sign` | 命中发起方签署逻辑，不返回 404 |
| TC-02 | 新路径继续可用 | 同上 | 调用 `POST /api/v1/orders/:id/logistics-contracts/initiator-sign` | 行为保持不变 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未登录访问 | 无 token | 调用 `buyer-sign` 路径 | 返回认证失败 |
| TC-EX-02 | 状态不满足 | 订单不在待签署物流合同状态 | 调用签署路径 | 返回业务错误，状态不被推进 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 非买家/卖家服务能力账号 | 调用 `buyer-sign` 路径 | 被 `RequireTradeCapability("procurement", "sales")` 拦截 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 订单模块验证 | `go test ./internal/modules/order/...` | 测试通过 |
| TC-BE-02 | 全量后端验证 | `./harness/verify.sh` | 后端测试/构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 发布 gate | `./harness/release.sh docs/features/SL-BE-WORKBENCH-FLOW-BASELINE` | 必需文档、测试、构建通过 |
