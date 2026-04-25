# Implementation Plan

## Feature 信息

- Feature ID: SL-BE-AUTO-DEPLOY
- 开发负责人: AI agent

## 变更范围

- 代码文件: 无业务代码变更。
- 文档文件: `docs/features/SL-BE-AUTO-DEPLOY/*`
- 配置文件: `.github/workflows/deploy-prod.yml`
- 部署/脚本文件: 复用服务器已有部署脚本，仓库内不新增部署脚本。

## 实施步骤

1. 参照 SteelLink-admin 生产部署 workflow，新增后端仓库专用 `.github/workflows/deploy-prod.yml`。
2. 将触发条件限定为 `main` 分支 push 与 `workflow_dispatch`。
3. 通过 GitHub Secrets 注入 SSH 连接信息，支持密钥或密码认证。
4. 默认执行 `./deploy_backend.sh`，并允许 `PROD_DEPLOY_COMMAND` 覆盖实际服务器命令。
5. 补齐 feature 文档，并执行 review/release gate。

## 验证计划

- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-AUTO-DEPLOY`
- `./harness/release.sh docs/features/SL-BE-AUTO-DEPLOY`
- 需要的人工验证: 在 GitHub 仓库配置 production environment 与 Secrets 后，手动触发一次 workflow，确认远程脚本路径与生产部署结果。

## 风险与回滚关注点

- 如果服务器部署脚本名称或工作目录不是 `./deploy_backend.sh`，必须配置 `PROD_DEPLOY_COMMAND`。
- 如果 workflow 合并后误触发，可临时禁用 workflow 或回滚 `.github/workflows/deploy-prod.yml`。
- 服务器部署脚本仍是生产发布的真实执行点，脚本必须具备幂等性与失败返回码。
