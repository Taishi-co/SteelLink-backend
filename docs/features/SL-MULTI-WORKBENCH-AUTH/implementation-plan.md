# 多工作台会话切换后端实施计划

## Feature 信息

- Feature ID: SL-MULTI-WORKBENCH-AUTH
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/user/ui/dto/auth_dto.go`
  - `internal/modules/user/application/service/auth_service.go`
- 文档文件:
  - `docs/features/SL-MULTI-WORKBENCH-AUTH/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展 switch-session DTO target 枚举，覆盖五类工作台。
2. 将 `canAccessPortal` 改为目标工作台到服务能力、企业角色兜底的明确映射。
3. 保持原有 Redis 一次性切换凭证结构和消费流程不变。
4. 执行后端用户模块测试和 harness 验证。

## 验证计划

- `./harness/verify.sh`
- `go test ./internal/modules/user/... ./pkg/auth ./internal/pkg/middleware`
- 需要的人工验证:
  - 使用已开通多服务能力企业分别创建 `logistics/warehouse/finance` 切换凭证。
  - 使用未授权企业请求目标工作台切换凭证并确认被拒绝。

## 风险与回滚关注点

- 前端需同步发送新 target 并消费目标端 `switch_token`。
- 回滚后端会导致新 target 被 DTO 拒绝，前端将依赖共享 token 兜底。
