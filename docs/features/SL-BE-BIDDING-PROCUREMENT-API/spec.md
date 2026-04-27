# Feature Spec

## 基本信息

- Feature ID: SL-BE-BIDDING-PROCUREMENT-API
- Feature 名称: 竞价交易与集采大厅后端接口
- 负责人: Codex
- 目标上线环境: SteelLink-backend

## 背景

买家端竞价交易和集采大厅需要从后端读取真实接口数据，不能继续由前端页面维护静态展示数据。本次新增市场活动接口，覆盖竞价大厅统计、热门竞价、竞价成功、集采大厅统计、集采活动列表和活动详情。

## 范围

### In Scope

- 新增市场活动模块 `internal/modules/market`。
- 新增竞价活动表 `bidding_activities`。
- 新增集采活动表 `procurement_activities` 与活动日志表 `procurement_logs`。
- 新增公开接口：
  - `GET /api/v1/biddings/stats`
  - `GET /api/v1/biddings`
  - `GET /api/v1/biddings/:id`
  - `GET /api/v1/procurements/stats`
  - `GET /api/v1/procurements`
  - `GET /api/v1/procurements/:id`
- 空表初始化时写入基础业务样例数据，供前后端联调和演示环境使用。

### Out of Scope

- 不实现真实竞价报价提交。
- 不实现集采报名、审核、签约、履约写接口。
- 不新增权限控制。

## 接口与数据影响

- 受影响路由/handler: 新增 `market/infrastructure/router` 和 `market/ui/handler`。
- 受影响服务: 新增 `market/application/service`。
- 受影响 middleware: 无。
- 受影响表/model: 新增 `bidding_activities`、`procurement_activities`、`procurement_logs`。
- 配置或环境变量变化: 新增本地联调用 `AUTO_MIGRATE_SCOPE=market|skip`，默认不设置时仍执行原全量迁移。
- Docker / 部署影响: 启动时通过 GORM AutoMigrate 建表。

## 权限与合规影响

- 受影响角色: 公开买家端访问。
- 新增/调整权限: 无。
- 是否影响审核、风控、日志留痕: 当前只读接口不影响；后续写接口需要补操作日志和权限校验。

## 验收标准

1. 后端启动后自动创建竞价和集采相关表。
2. 竞价统计和列表接口返回标准 `{ code, message, data }`。
3. 集采统计、列表和详情接口返回标准 `{ code, message, data }`。
4. 前端可通过这些接口渲染竞价交易与集采大厅，不依赖页面静态业务数据。
