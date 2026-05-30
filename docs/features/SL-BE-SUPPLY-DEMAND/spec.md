# Feature Spec: 供求信息后端

## 基本信息

- Feature ID: SL-BE-SUPPLY-DEMAND
- Feature 名称: 供求信息基础 API
- 负责人: AI agent
- 目标上线环境: Backend API

## 背景

买家端已经将原信息撮合页调整为“供求信息”，页面需要承载供应、求购、仓储服务和物流服务信息。后端需要提供基础数据能力，支撑前端从静态演示逐步切换到真实列表、详情、统计和用户发布数据。

## 范围

### In Scope

- 新增 `supplydemand` Go/Gin 模块。
- 新增供求信息 GORM model 和自动迁移。
- 提供公开列表、详情和统计接口。
- 提供登录且个人实名认证后发布供求信息、查看我的发布、关闭自己的发布接口。
- 发布后默认进入待审核，需管理后台审核通过后才进入公开商城/供求信息列表。
- 支持四类信息类型: `supply`、`demand`、`warehouse`、`logistics`。

### Out of Scope

- 不实现 AI 推荐、撮合算法或融资审批。
- 不新增手写 SQL migration 文件。

## 接口与数据影响

- 受影响路由/handler:
  - `GET /api/v1/supply-demands`
  - `GET /api/v1/supply-demands/:id`
  - `GET /api/v1/supply-demands/stats`
  - `GET /api/v1/supply-demands/mine`
  - `POST /api/v1/supply-demands`
  - `PUT /api/v1/supply-demands/:id/status`
- 管理后台审核接口:
  - `GET /api/v1/supply-demands`
  - `PUT /api/v1/supply-demands/:id/audit`
- 受影响服务: `internal/modules/supplydemand/application/service`
- 受影响 middleware: 复用 `AuthRequired` 与 `OptionalAuth`
- 受影响表/model: `supply_demand_posts` 新增审核状态、审核备注、审核人和审核时间字段
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 游客可查看已发布列表、详情和统计；登录用户可发布、查看自己的发布、更新自己的发布状态。
- 新增/调整权限: 管理后台新增 `supply-demand:audit` 菜单权限。
- 是否影响审核、风控、日志留痕: 供求信息审核通过后公开展示，驳回需填写原因；管理后台记录审核操作日志。

## 验收标准

1. 服务启动时自动迁移 `supply_demand_posts` 表。
2. 游客可查看已发布供求信息列表、详情和统计。
3. 登录且个人实名认证用户可发布四类供求信息，并可查询自己的发布。
4. 新发布信息默认为待审核，不出现在公开列表/详情/统计中。
5. 管理员审核通过后公开可见，审核驳回后继续不可见并回显驳回原因。
6. 登录用户只能关闭自己发布的信息，不能绕过审核改成公开状态。
7. `go test ./...` 与 `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 通过。
