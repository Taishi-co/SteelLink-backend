# Test Cases

## 测试范围

- Feature ID: SL-BE-NOTIFICATIONS
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 管理员发送指定企业站内信 | 管理员已登录，企业下存在有效用户 | POST `/api/v1/admin/notifications` | 返回创建数量，用户收件箱可见 |
| TC-02 | 用户读取站内信 | 用户已登录 | GET `/api/v1/notifications` | 只返回当前用户消息和未读数 |
| TC-03 | 订单流转自动通知 | 订单进入仓库/卖方/物流节点 | 执行业务动作 | 对应企业用户收到订单类站内信 |
| TC-04 | 商品审核通知 | 商品待审核 | 管理员更新商品状态为上架或拒绝 | 卖家企业用户收到审核类站内信 |
| TC-05 | WebSocket 实时通知 | 用户已登录并建立 WS 连接 | 触发一条站内信 | 当前连接收到 `notification.created` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 无接收人发送 | 管理员已登录 | 不传任何接收范围且不全量发送 | 返回 bad request |
| TC-EX-02 | 越权已读 | 用户 A 已登录 | 标记用户 B 的消息已读 | 返回 not found |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 查询站内信 | 401 |
| TC-AUTH-02 | 普通用户 | 调用管理员发送接口 | 403 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端测试 | `go test ./...` | 通过 |
| TC-OPS-02 | API 构建 | `CGO_ENABLED=0 go build -o /tmp/steellink-api-notifications ./cmd/api` | 通过 |
