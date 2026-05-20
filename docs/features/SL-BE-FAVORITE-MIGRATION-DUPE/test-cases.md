# Test Cases

## 测试范围

- Feature ID: SL-BE-FAVORITE-MIGRATION-DUPE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 旧 schema 重复收藏清理 | `favorites` 表存在 `deleted_at` 列，且同一用户商品存在软删除记录与有效记录 | 执行收藏模块迁移 | 清理 SQL 优先保留有效记录，唯一索引可继续创建 |
| TC-02 | 当前 schema 重复收藏清理 | `favorites` 表不存在 `deleted_at` 列 | 执行收藏模块迁移 | 清理 SQL 不引用 `deleted_at`，保留最早记录 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 新库无收藏表 | `favorites` 表尚不存在 | 执行收藏模块迁移 | 跳过清理，继续 `AutoMigrate` 建表 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 登录用户 | 收藏接口权限 | 权限逻辑不变，仍由原有鉴权中间件控制 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 单元测试 | `go test ./internal/modules/favorite/infrastructure/persistence` | 清理 SQL 策略测试通过 |
| TC-OPS-02 | 仓库验证 | `./harness/verify.sh` | Go 测试、构建和 compose config 通过或明确环境前置条件 |
| TC-OPS-03 | Review gate | `./harness/review.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE` | review gate 通过 |
| TC-OPS-04 | Release gate | `./harness/release.sh docs/features/SL-BE-FAVORITE-MIGRATION-DUPE` | release gate 通过 |
