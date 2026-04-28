# Dockerfile Backend Deployment Script Implementation Plan

## Feature 信息

- Feature ID: SL-BE-DOCKERFILE-DEPLOY
- 开发负责人: Codex

## 变更范围

- 代码文件: 无。
- 文档文件:
  - `docs/features/SL-BE-DOCKERFILE-DEPLOY/*`
- 配置文件: 无。
- 部署/脚本文件:
  - `deploy_backend.sh`

## 实施步骤

1. 新增 Dockerfile 部署脚本，基于 `/opt/backend` 拉取 `main`、构建镜像、启动 `backend` 容器。
2. 通过 `--env-file /opt/backend/.env` 注入运行时配置，并校验数据库、Redis、JWT、COS 必填变量。
3. 部署后在容器内请求 `/health`，失败时打印最近容器日志。

## 验证计划

- `bash -n deploy_backend.sh`
- `./harness/branch-guard.sh local`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-DOCKERFILE-DEPLOY`
- `./harness/release.sh docs/features/SL-BE-DOCKERFILE-DEPLOY`
- 人工验证: 在生产服务器确认 `.env` 使用已轮换 COS 密钥后执行 `./deploy_backend.sh`。

## 风险与回滚关注点

- 如果 MySQL/Redis 部署在宿主机，`.env` 里的 `DATABASE_HOST` / `REDIS_HOST` 不能使用 `localhost`，应改为 `host.docker.internal` 或实际内网 IP。
- 如果健康检查失败，脚本会保留失败容器并输出日志；修复 `.env` 后重新执行脚本即可替换容器。
- 回滚时切回上一版代码并重新执行脚本。
