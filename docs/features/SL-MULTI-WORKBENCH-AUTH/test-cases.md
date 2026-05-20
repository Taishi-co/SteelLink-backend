# 多工作台会话切换后端测试用例

## 测试范围

- Feature ID: SL-MULTI-WORKBENCH-AUTH
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 开通物流能力后创建切换凭证 | 企业服务能力包含 `logistics` | 调用 `/auth/switch-session`，target 为 `logistics` | 返回 `switch_token` |
| TC-02 | 开通仓储能力后消费切换凭证 | 企业服务能力包含 `warehouse` | 创建并消费 target 为 `warehouse` 的 `switch_token` | 返回新的登录响应 |
| TC-03 | 企业角色为资金方 | 企业 `enterprise_role=fund` | 创建 target 为 `finance` 的切换凭证 | 允许创建凭证 |
| TC-04 | 兼容买卖方切换 | 企业开通 `procurement` 或 `sales` | 继续使用 target 为 `buyer/seller` | 行为与原逻辑一致 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未开通目标服务能力 | 企业无 `warehouse` 能力且角色不是仓储方 | 创建 target 为 `warehouse` 的切换凭证 | 返回 forbidden |
| TC-EX-02 | 不支持的目标值 | 请求 target 不在枚举内 | 调用 switch-session | DTO 校验失败 |
| TC-EX-03 | 消费目标不匹配 | 创建 `logistics` 凭证后用 `finance` 消费 | 调用 consume | 返回“切换目标不匹配” |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | `enterprise_role=seller` 且有 `logistics` 能力 | 创建物流工作台凭证 | 允许 |
| TC-AUTH-02 | `enterprise_role=buyer` 且无 `finance` 能力 | 创建资金方凭证 | 拒绝 |
| TC-AUTH-03 | `enterprise_role=fund` | 创建资金方凭证 | 允许 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端测试 | `go test ./internal/modules/user/... ./pkg/auth ./internal/pkg/middleware` | 通过 |
| TC-OPS-02 | 全量验证 | `./harness/verify.sh` | 通过 |

## 回归关注点

- JWT claims 中 `service_capabilities` 的生成不变。
- Redis 中一次性切换凭证 TTL 与一次性消费行为不变。
