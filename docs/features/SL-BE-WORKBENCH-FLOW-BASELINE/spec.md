# Feature Spec

## 基本信息

- Feature ID: `SL-BE-WORKBENCH-FLOW-BASELINE`
- Feature 名称: 工作台流程接口兼容基线
- 负责人: Codex
- 目标上线环境: 本地开发 / 测试联调环境

## 背景

五个工作台围绕同一订单、仓储、物流、融资流程协作。梳理前后端调用后发现，买家工作台仍存在物流合同签署接口的旧路径调用：前端提供 `buyer-sign` 调用，而后端只注册 `initiator-sign`。这会导致待签署物流合同状态下，买方/卖方发起方签署动作可能出现 404。

本轮后端只补兼容入口，不改变订单状态机和 handler 实现。

## 范围

### In Scope

- 在订单路由下补充 `/orders/:id/logistics-contracts/buyer-sign` 兼容路径。
- 兼容路径复用现有 `InitiatorSignLogisticsContract` handler 和买家/卖家贸易能力校验。
- 补齐 feature 文档，记录后续仍需处理的接口缺口。

### Out of Scope

- 不修改物流合同签署业务逻辑。
- 不调整订单状态枚举或迁移数据。
- 不新增收藏、admin 通知等跨模块接口。
- 不改变权限 middleware 的判断口径。

## 接口与数据影响

- 受影响路由 / handler:
  - `POST /api/v1/orders/:id/logistics-contracts/buyer-sign`
  - `POST /api/v1/orders/:id/logistics-contracts/initiator-sign`
- 受影响服务:
  - `internal/modules/order`
- 受影响 middleware:
  - 复用 `RequireTradeCapability("procurement", "sales")`
- 受影响表 / model: 无
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买家、卖家
- 新增 / 调整权限: 无新增权限，兼容路由沿用原发起方签署权限
- 是否影响审核、风控、日志留痕: 不影响

## 验收标准

1. `buyer-sign` 与 `initiator-sign` 两个路径都能命中同一个物流合同发起方签署 handler。
2. 未认证用户仍被 `AuthRequired` 拦截。
3. 非买家/卖家贸易能力账号仍被 capability middleware 拦截。
4. 订单模块测试和构建通过。
