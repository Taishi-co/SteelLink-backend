# 实现计划

## Phase 1: 数据模型与状态映射

- [ ] 扩展 `servicefulfillment` 物流状态枚举
- [ ] 在 `logistics_orders` 增加责任主体和关键节点时间字段
- [ ] 保持 AutoMigrate 兼容现有表结构

## Phase 2: 后端履约接口

- [ ] 调整物流单创建默认状态为 `task_created`
- [ ] 增强 `UpdateLogisticsOrderStatus`，按新状态维护责任主体与节点时间
- [ ] 增强 `GetOrderFulfillment` 摘要字段

## Phase 3: 兼容与校验

- [ ] 兼容旧状态到新文案/摘要的展示
- [ ] 自测物流状态推进和订单履约摘要查询
