# Feature Spec

## 基本信息

- Feature ID: `SL-BE-MARKET-MIGRATION-ORDER`
- Feature 名称: 集采项目迁移顺序修复
- 负责人: Codex
- 目标上线环境: prod

## 背景

生产后端容器启动时执行 market 自动迁移，`procurement_logs.project_id` 的外键引用 `procurement_projects.id`。当前迁移列表先迁移 `ProcurementLog`，后迁移 `ProcurementProject`，在全新或未建项目表的生产库上会报 `Error 1824 (HY000): Failed to open the referenced table 'procurement_projects'`，导致后端容器反复重启，前端 `/api/v1/*` 经 Nginx 返回 502。

本次修复目标是让 market 自动迁移按依赖顺序执行，确保父表 `procurement_projects` 在日志表尝试建立外键前已经存在。

## 范围

### In Scope

- 调整 market 模块自动迁移模型顺序。
- 增加迁移顺序回归测试，覆盖 `ProcurementProject` 必须早于 `ProcurementLog`。
- 补齐本次数据库迁移影响说明与发布验证记录。

### Out of Scope

- 不改动集采项目业务接口。
- 不新增或删除生产数据。
- 不调整 Nginx、Dockerfile、部署脚本和环境变量。

## 接口与数据影响

- 受影响路由/handler: 无直接接口变更。
- 受影响服务: `internal/modules/market/infrastructure/persistence.AutoMigrate`
- 受影响 middleware: 无
- 受影响表/model:
  - `procurement_projects`
  - `procurement_logs`
  - market 模块其他自动迁移模型仅保持原有迁移能力
- 配置或环境变量变化: 无
- Docker / 部署影响: 后端容器启动时自动迁移不再因缺少 `procurement_projects` 父表而退出。

## 权限与合规影响

- 受影响角色: 无
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不改变业务日志写入逻辑，仅修复日志表迁移顺序。

## 验收标准

1. `ProcurementProject` 在 market 迁移列表中早于 `ProcurementLog`。
2. 迁移顺序单元测试能防止后续重排再次引入同类问题。
3. `./harness/verify.sh` 通过，后端构建和测试均无阻断。
