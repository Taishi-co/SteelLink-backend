# 多工作台会话切换后端支持

## 基本信息

- Feature ID: SL-MULTI-WORKBENCH-AUTH
- Feature 名称: 多服务能力企业跨工作台会话切换
- 负责人: Codex
- 目标上线环境: staging/prod

## 背景

平台设计允许一个企业申请多个服务能力。原 `switch-session` 只支持 `buyer` 和 `seller` 两类目标，物流、仓储、资金方工作台无法通过一次性切换凭证建立目标端会话，只能依赖共享 token，容易在目标工作台残留过期 token 时触发 `2000003 token expired`。

本需求扩展后端切换目标，使工作台准入以企业已审核服务能力为准。

## 范围

### In Scope

- `switch-session` 请求 DTO 支持 `buyer/seller/logistics/warehouse/finance`。
- `switch-session/consume` 请求 DTO 支持相同目标。
- `canAccessPortal` 按目标工作台映射到服务能力和企业角色兜底。

### Out of Scope

- 不修改 JWT claims 结构。
- 不修改 token 过期时间和刷新 token 策略。
- 不调整服务能力审核表结构。

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/auth/switch-session`
  - `POST /api/v1/auth/switch-session/consume`
- 受影响服务:
  - `internal/modules/user/application/service/auth_service.go`
- 受影响 middleware: 无
- 受影响表/model: 无
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 企业用户
- 新增/调整权限:
  - `buyer` 目标需要 `procurement` 能力或企业角色 `buyer`
  - `seller` 目标需要 `sales` 能力或企业角色 `seller`
  - `logistics` 目标需要 `logistics` 能力或企业角色 `logistics`
  - `warehouse` 目标需要 `warehouse` 能力或企业角色 `warehouse`
  - `finance` 目标需要 `finance` 能力或企业角色 `fund`
- 是否影响审核、风控、日志留痕: 不影响资金、风控或审核操作，仅影响登录会话切换准入。

## 验收标准

1. 已开通对应服务能力的企业可创建并消费目标工作台 `switch_token`。
2. 未开通对应服务能力且企业角色不匹配时，后端返回 forbidden。
3. 旧的 `buyer/seller` 切换行为保持兼容。
