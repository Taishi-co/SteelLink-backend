# Redis Business Cache Expansion Test Cases

## 测试范围

- Feature ID: SL-BE-REDIS-BUSINESS-CACHE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | AI 商品匹配缓存 | Redis 可用，用户已开通 procurement | 同一用户连续调用 `match-products` | 第二次可命中按用户/企业/解析意图生成的缓存 |
| TC-02 | 工作台缓存 | Redis 可用，用户存在订单 | 连续调用 `/users/me/dashboard` | 返回结构一致，缓存 TTL 内减少重复聚合查询 |
| TC-03 | 企业服务 overview 缓存 | Redis 可用，企业已开通或申请服务 | 连续调用 `/enterprises/me/services` | 返回服务状态和材料要求一致 |
| TC-04 | 幂等写入 | Redis 可用，请求携带 `Idempotency-Key` | 对订单创建或 AI 生成订单重复提交同一个 key | 首次请求继续处理，重复请求返回 409 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | Redis 不可用 | Redis 客户端为空或连接失败 | 调用 dashboard、服务 overview、AI 匹配 | 接口回源数据库或继续正常执行 |
| TC-EX-02 | 无幂等 key | 客户端未升级 | 调用受保护写入接口 | 请求按原逻辑执行，不强制要求 header |
| TC-EX-03 | 服务状态变更 | Redis 有旧 overview 缓存 | 审核服务申请或更新权限状态 | 相关企业服务缓存被删除，后续查询返回新状态 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 调用企业服务或 dashboard | 仍由原鉴权中间件拒绝 |
| TC-AUTH-02 | 未开通 procurement 企业 | 调用 AI 匹配或生成订单 | 仍由服务能力权限拒绝 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 构建通过，配置解析无阻断 |
| TC-OPS-02 | 发布检查 | `./harness/release.sh docs/features/SL-BE-REDIS-BUSINESS-CACHE` | release gate 通过 |
