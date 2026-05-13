# Feature Spec: 合同违约诉讼流程

## 基本信息

- Feature ID: SL-BE-CONTRACT-BREACH-LITIGATION
- Feature 名称: 合同不履约的诉讼/异常处理流程
- 负责人: backend
- 目标上线环境: staging / prod

## 背景

根据产品流程图，当一方拒绝履行合同时，需进入异常处理流程：
1. 守约方上传违约证据
2. 平台冻结订单，管理端审批介入
3. 保留证据，联系双方，开展协调
4. 如协调失败，进入诉讼流程
5. 诉讼结果归档

当前系统只有 `OrderStatusException` 和 `OrderStatusFrozen` 状态，缺少违约记录、证据管理、协调流程和诉讼流程的实体和接口。

## 范围

### In Scope

- 违约记录实体（BreachRecord）和诉讼记录实体（LitigationRecord）
- 违约证据上传接口
- 订单冻结/解冻接口（管理端）
- 协调流程接口（管理端发起协调、记录协调结果）
- 诉讼流程接口（发起诉讼、更新诉讼状态、记录判决结果）
- 订单状态机增加违约相关状态流转

### Out of Scope

- 真实法律文书生成
- 外部法院系统对接
- 自动违约判定

## 接口与数据影响

- 受影响路由/handler:
  - 新增 breach 相关路由在 order module
  - 新增 litigation 相关路由在 order module
  - 新增 admin 管理端异常处理路由
- 受影响服务: order_service 扩展
- 受影响 middleware: 无变更
- 受影响表/model: 新增 breach_records、litigation_records
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买方(sales/procurement)、平台运营(operator/super_admin)
- 新增/调整权限:
  - 买方/卖方: 上传违约证据
  - 平台运营: 冻结/解冻订单、发起协调、记录协调结果
  - 买方/卖方: 发起诉讼
  - 平台运营: 更新诉讼状态、记录判决
- 审核留痕: 所有操作必须写入 order_events

## 验收标准

1. 守约方可对已签合同订单上传违约证据，订单自动进入 exception 状态
2. 平台运营可冻结/解冻异常订单
3. 平台运营可发起协调并记录协调结果
4. 买卖方可发起诉讼，平台运营可更新诉讼状态
5. 诉讼判决结果归档后订单可进入 cancelled 或恢复履约
6. 所有操作留痕在 order_events
