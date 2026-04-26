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
- 提供登录后发布供求信息、查看我的发布、更新发布状态接口。
- 支持四类信息类型: `supply`、`demand`、`warehouse`、`logistics`。

### Out of Scope

- 不实现 AI 推荐、撮合算法、融资审批或审核工作流。
- 不新增后台运营审核接口。
- 不新增手写 SQL migration 文件。

## 接口与数据影响

- 受影响路由/handler:
  - `GET /api/v1/supply-demands`
  - `GET /api/v1/supply-demands/:id`
  - `GET /api/v1/supply-demands/stats`
  - `GET /api/v1/supply-demands/mine`
  - `POST /api/v1/supply-demands`
  - `PUT /api/v1/supply-demands/:id/status`
- 受影响服务: `internal/modules/supplydemand/application/service`
- 受影响 middleware: 复用 `AuthRequired` 与 `OptionalAuth`
- 受影响表/model: 新增 `supply_demand_posts`
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 游客可查看已发布列表、详情和统计；登录用户可发布、查看自己的发布、更新自己的发布状态。
- 新增/调整权限: 无新增权限码。
- 是否影响审核、风控、日志留痕: 暂不涉及审核和风控；状态更新仅允许发布人操作。

## 验收标准

1. 服务启动时自动迁移 `supply_demand_posts` 表。
2. 游客可查看已发布供求信息列表、详情和统计。
3. 登录用户可发布四类供求信息，并可查询自己的发布。
4. 登录用户只能更新自己发布的信息状态。
5. `go test ./...` 与 `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 通过。
