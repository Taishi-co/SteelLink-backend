# Redis Ops Cache Test Cases

## 测试范围

- Feature ID: SL-BE-REDIS-OPS-CACHE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 登录接口限流 | Redis 可用 | 同一 IP 在窗口内重复调用登录接口 | 未超过阈值时正常进入登录逻辑 |
| TC-02 | AI 采购限流 | Redis 可用，用户已登录且开通采购服务 | 调用 parse/match 接口 | 未超过阈值时正常返回 |
| TC-03 | 首页统计缓存 | Redis 可用且 MySQL 有统计数据 | 连续调用 `/platform/stats` | 返回结构一致，缓存 TTL 内不需要重复聚合查询 |
| TC-04 | 向量同步互斥锁 | Redis 可用 | 同一 target 并发调用 vector-sync | 一个执行同步，其它返回 running 状态 |
| TC-05 | 查询向量同步状态 | 已执行过 vector-sync | 调用 `GET /internal/ai-procurements/vector-sync/status?target=products` | 返回最近状态、时间、耗时和同步数量 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 超过限流阈值 | Redis 可用 | 同一身份在窗口内超过阈值 | 返回 HTTP 429，并带 `Retry-After` |
| TC-EX-02 | Redis 不可用 | Redis 连接失败 | 调用限流接口 | 限流降级放行，不阻塞主流程 |
| TC-EX-03 | 无同步状态 | 从未同步过指定 target | 查询状态接口 | 返回 `unknown` 状态 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未带内部 token 调用方 | 查询 vector-sync 状态 | 返回 unauthorized |
| TC-AUTH-02 | 企业采购用户 | 调用 AI 采购 parse/match | 先通过原有 auth/capability，再执行 Redis 限流 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 测试、构建和 docker compose config 通过 |
| TC-OPS-02 | 无新增必填配置 | 检查配置文件 diff | 复用现有 Redis 和向量同步 token 配置 |
