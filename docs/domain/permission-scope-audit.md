# 权限口径审计

## 结论

当前权限应按两层理解：

- 交易能力：`procurement`、`sales`。代表企业作为买方/卖方参与交易。
- 服务商能力：`warehouse`、`logistics`、`finance`。代表企业作为平台服务提供方发布资源并执行服务商履约动作。

买卖家使用仓储、物流、金融服务时，不应被要求开通 `warehouse`、`logistics`、`finance`。这些服务商能力只用于服务提供方发布资源、报价、仓储确认、运输节点、融资审批、放款和质押释放等动作。

企业可以在本企业身份之外额外申请服务商能力，例如买方/卖方企业也可以申请 `warehouse`、`logistics` 或 `finance`。申请通过后，该企业才可以进入对应服务商工作台发布资源或执行服务商履约动作。

`sales/procurement` 仍然是需要审核的企业交易能力。企业认证提交时会同步提交企业身份对应的初始交易能力材料；企业审核通过后，才赋予对应交易能力。

## 已审计口径

订单主流程：

- 买方下单、取消、最终确认、支付、申请融资、确认融资方案、签署债务合同、确认收货：要求 `procurement`。
- 卖方确认条款、确认收款、发起物流竞价、对账归档：要求 `sales`。
- 买卖双方确认物流报价、签署交易合同、归档：要求 `procurement` 或 `sales`。
- 仓储方确认仓库条件：要求 `warehouse`。
- 物流方报价、提货、送达、签收回传：要求 `logistics`。
- 资金方风控、生成债务合同、放款、记录/释放质押：要求 `finance`。

服务资源：

- 公共查看已审核仓储/物流/金融服务：不要求服务商能力。
- 发布和编辑仓储资源：要求 `warehouse`。
- 发布和编辑物流资源：要求 `logistics`。
- 发布金融服务产品：要求 `finance`。

商品和交易：

- 商品发布/维护：要求 `sales`。
- 购物车、下单、AI 采购：要求 `procurement`。
- 卖方绑定已审核仓储资源，是销售履约配置，不要求卖方开通 `warehouse`。

工作台口径：

- buyer-app / seller-app 是交易工作台，可以查询和使用已审核仓储、物流、金融资源。
- warehouse-app / logistics-app / finance-app 是服务商工作台，资源发布和服务商履约动作必须具备对应服务商能力。

## 本次代码整理

为了避免后续继续混用“平台服务权限”这个模糊概念，后端路由已改用语义化中间件：

- `RequireTradeCapability(...)`：用于 `sales`、`procurement` 交易能力。
- `RequireProviderService(...)`：用于 `warehouse`、`logistics`、`finance` 服务商能力。
- `RequireFulfillmentParticipant()`：用于订单列表、详情、事件等所有履约参与方均可访问的读接口。

这些改动不改变数据库和权限判断结果，只把代码口径拆清楚。

权限表已拆分：

- `enterprise_trade_capability_permissions`：承载 `sales/procurement`。
- `enterprise_provider_service_permissions`：承载 `warehouse/logistics/finance`。
- `enterprise_service_permissions`：保留为历史迁移来源，不再作为运行时权限判断表。

对外接口仍保留 `service` 字段以兼容前端，同时新增 `capability_kind` 表示 `trade` 或 `provider`。

## 发现的剩余口径风险

- 前端主要入口已调整为“交易能力/服务商能力”文案，但仍需在后续页面迭代中持续避免把买卖家“使用服务”描述成“开通服务商能力”。
- `enterprise_service_applications` 仍保留统一申请表。后续如要进一步消除命名歧义，可再拆成交易能力申请表和服务商能力申请表，或重命名为 `enterprise_capability_applications`。

## 建议后续

- 管理端企业能力表后续可进一步分组展示交易能力和服务商能力。
- finance-app 需要接入真实金融服务商工作台，保持“服务商发布金融产品/审批融资”和“买方使用融资服务”的口径隔离。
