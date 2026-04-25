# Feature Spec

## 基本信息

- Feature ID: SL-BE-AUTO-DEPLOY
- Feature 名称: main 分支合并后自动部署后端服务
- 负责人: AI agent
- 目标上线环境: production

## 背景

SteelLink-admin 已通过 GitHub Actions 在 PR 合并到 `main` 后触发生产部署。后端服务需要采用同类流程，让合并后的生产发布入口固定在 GitHub Actions，并由服务器侧部署脚本负责拉取代码、构建与重启服务。

## 范围

### In Scope

- 新增 `.github/workflows/deploy-prod.yml`。
- workflow 在 `main` 分支 push 后自动运行，并支持 `workflow_dispatch` 手动触发。
- workflow 通过 SSH 登录生产服务器，执行后端部署命令。
- 部署命令默认使用 `./deploy_backend.sh`，可通过 GitHub Actions Secret `PROD_DEPLOY_COMMAND` 覆盖。
- 补齐本 feature 的 spec、test cases、implementation plan 与 release checklist。

### Out of Scope

- 不新增或修改服务器上的部署脚本。
- 不改变 Go/Gin 业务代码、API、数据库结构、Docker 配置或运行时配置。
- 不在 GitHub Actions 内重复实现服务器部署脚本中的拉取、构建、迁移或重启逻辑。

## 接口与数据影响

- 受影响路由/handler: 无。
- 受影响服务: 无业务服务逻辑变更。
- 受影响 middleware: 无。
- 受影响表/model: 无。
- 配置或环境变量变化: 需要在 GitHub Actions Secrets 中配置 `PROD_HOST`、`PROD_USER`，并配置 `PROD_SSH_KEY` 或 `PROD_PASSWORD` 其中之一；如服务器脚本不是 `./deploy_backend.sh`，需配置 `PROD_DEPLOY_COMMAND`。
- Docker / 部署影响: PR 合并到 `main` 后触发生产部署，远程执行服务器上的后端部署脚本。

## 权限与合规影响

- 受影响角色: 无业务角色影响。
- 新增/调整权限: 需要 GitHub 仓库 production environment 或 Actions Secrets 的维护权限。
- 是否影响审核、风控、日志留痕: 不影响业务审核与风控；部署日志留存在 GitHub Actions run 记录中。

## 验收标准

1. `.github/workflows/deploy-prod.yml` 存在，并仅在 `main` 分支 push 或手动触发时运行生产部署。
2. workflow 使用 `appleboy/ssh-action@v1.2.0` 通过 GitHub Secrets 登录生产服务器。
3. workflow 默认执行 `./deploy_backend.sh`，并允许通过 `PROD_DEPLOY_COMMAND` 覆盖。
4. `./harness/review.sh docs/features/SL-BE-AUTO-DEPLOY` 通过。
5. `./harness/release.sh docs/features/SL-BE-AUTO-DEPLOY` 通过。
