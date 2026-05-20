# Feature Spec

## 基本信息

- Feature ID: SL-BE-FAVORITE-MIGRATION-DUPE
- Feature 名称: 收藏表唯一索引迁移重复数据清理
- 负责人: Codex
- 目标上线环境: production

## 背景

生产后端容器在启动迁移收藏表时反复退出，nginx 对 `/api/v1/*` 返回 `502 Bad Gateway`。日志显示 MySQL 连接成功，但 `favorites` 表创建 `idx_favorites_user_product` 唯一索引失败，原因是历史数据中存在同一 `user_id` 与 `product_id` 的重复收藏记录。

当前代码中的收藏实体已不再使用 `deleted_at` 软删除字段，但生产表保留了旧字段与旧数据。迁移必须在创建唯一索引前清理这些历史重复行，避免服务启动被数据约束阻断。

## 范围

### In Scope

- 在收藏模块 `AutoMigrate` 前执行幂等重复数据清理。
- 对旧表存在 `deleted_at` 字段的场景，优先保留未软删除记录；否则保留最早记录。
- 增加单元测试覆盖清理 SQL 的分组与保留策略。

### Out of Scope

- 不调整收藏接口协议。
- 不修改商品、订单、用户等业务表。
- 不改变生产部署脚本或 nginx 配置。

## 接口与数据影响

- 受影响路由/handler: 无路由变更；间接影响 `/api/v1/favorites/*` 启动可用性。
- 受影响服务: 收藏模块迁移。
- 受影响 middleware: 无。
- 受影响表/model: `favorites`。
- 配置或环境变量变化: 无。
- Docker / 部署影响: 后端容器启动迁移阶段会先清理重复收藏行，再创建唯一索引。

## 权限与合规影响

- 受影响角色: 所有使用收藏能力的登录用户。
- 新增/调整权限: 无。
- 是否影响审核、风控、日志留痕: 不影响审核、风控或审计权限；仅清理无法同时保留的重复收藏数据。

## 验收标准

1. 当 `favorites` 中存在重复 `user_id/product_id` 时，迁移前清理重复行，`AutoMigrate` 不因唯一索引创建失败而中断。
2. 旧表存在 `deleted_at` 时，清理优先保留 `deleted_at IS NULL` 的有效记录。
3. 当前 schema 不存在 `deleted_at` 时，清理 SQL 不引用该字段。
4. `./harness/verify.sh`、`./harness/review.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE` 和 `./harness/release.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE` 通过。
