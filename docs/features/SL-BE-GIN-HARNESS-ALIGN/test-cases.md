# Test Cases

## 测试范围

- Feature ID: `SL-BE-GIN-HARNESS-ALIGN`
- 测试负责人: AI agent

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | Go 验证流程通过 | 本机已安装 Go | 执行 `./harness/verify.sh` | `go test`、`go build` 通过，脚本退出码为 0 |
| TC-02 | Review 门禁通过 | feature 文档四件套存在 | 执行 `./harness/review.sh docs/features/SL-BE-GIN-HARNESS-ALIGN` | review gate 通过 |
| TC-03 | Release 门禁通过 | Dockerfile、go.mod、docker-compose.yml、config 示例存在 | 执行 `./harness/release.sh docs/features/SL-BE-GIN-HARNESS-ALIGN` | release gate 通过 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 本地目录不是 Git 仓库 | 当前目录无 `.git` | 执行 `./harness/branch-guard.sh local` | 脚本明确提示先初始化或 clone 新仓库 |
| TC-EX-02 | 缺少 feature 文档目录 | 未传 feature 目录且存在多个 feature 目录 | 执行 `./harness/review.sh` | 脚本提示无法解析 feature 目录 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 开发者 | 查看新的工程流程文档 | 能够明确知道 Go/Gin 验证命令、Git 前置条件与 deploy 入口 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | Docker 编排文件可解析 | `docker compose config` | Compose 语法通过，无 Maven / JAR 依赖 |
| TC-OPS-02 | Docker 镜像构建入口正确 | 检查 `Dockerfile` | 使用 Go 多阶段构建并暴露 8080 |
