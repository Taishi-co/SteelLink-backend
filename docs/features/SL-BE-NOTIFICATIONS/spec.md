# Feature Spec

## 基本信息

- Feature ID: SL-BE-NOTIFICATIONS
- Feature 名称: 站内信后端能力
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

平台需要向用户发送站内消息，包括管理员通知、订单流转提醒和审核结果提醒。

## 范围

### In Scope

- 新增 `notifications` 表，存储收件人、类型、标题、内容、站内链接、已读状态和业务上下文。
- 新增用户站内信接口：列表、未读数、单条已读、全部已读。
- 新增 WebSocket 实时订阅接口，在线用户可实时收到新站内信。
- 新增管理员发送站内信接口，支持按用户 ID、企业 ID、用户角色或全量发送。
- 在商品审核、仓库确认、卖方确认、物流报价确认和物流进度回传时写入站内信。

### Out of Scope

- 不引入 Kafka/RabbitMQ 等外部消息队列。
- 不做短信、邮件、App Push 分发。

## 接口与数据影响

- 受影响路由/handler: `/api/v1/notifications`、`/api/v1/notifications/ws`、`/api/v1/admin/notifications`
- 受影响服务: product/order/notification
- 受影响 middleware: 复用现有 `AuthRequired`、`RequireUserType`
- 受影响表/model: 新增 `notifications`
- 配置或环境变量变化: 无
- Docker / 部署影响: 首次启动默认 AutoMigrate 会创建新表

## 权限与合规影响

- 受影响角色: 登录用户可查自己的站内信；`operator` / `super_admin` 可发送管理员通知。
- 新增/调整权限: 无独立 RBAC 权限码，复用现有用户类型校验。
- 是否影响审核、风控、日志留痕: 商品审核和订单履约节点会增加用户通知留痕，不改变原业务状态机。

## 验收标准

1. 管理员可以向指定用户、企业、角色或全量用户发送站内信。
2. 用户只能查看、订阅和标记自己的站内信。
3. 商品审核通过/拒绝、关键订单流转会自动生成站内信。
