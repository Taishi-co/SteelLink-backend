# Test Cases

## 测试范围

- Feature ID: `SL-BE-MARKET-MIGRATION-ORDER`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 迁移模型顺序正确 | market persistence 包可测试 | 运行 `go test ./internal/modules/market/infrastructure/persistence` | `ProcurementProject` 早于 `ProcurementLog`，测试通过 |
| TC-02 | 后端完整验证 | 当前分支包含修复 | 运行 `./harness/verify.sh` | `go test ./...`、后端构建和 compose config 无阻断 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 迁移顺序被错误重排 | 将 `ProcurementLog` 放回 `ProcurementProject` 前 | 运行目标包测试 | 测试失败并指出项目表迁移顺序晚于日志表 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 任意角色 | 本次仅调整数据库迁移顺序 | 不新增权限，不改变接口鉴权结果 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 构建通过，配置解析无阻断 |
| TC-OPS-02 | Docker Compose 可解析 | `docker compose config` 或 harness optional compose check | compose 配置可解析 |
| TC-OPS-03 | 生产故障回归 | 部署修复后访问 `/api/v1/health` | 后端健康检查返回 200 |
