# Dockerfile Backend Deployment Script

## 基本信息

- Feature ID: SL-BE-DOCKERFILE-DEPLOY
- Feature 名称: Dockerfile 后端部署脚本
- 负责人: Codex
- 目标上线环境: prod

## 背景

生产环境通过 `Dockerfile` 构建镜像并用 `docker run` 启动后端，不使用 `docker-compose`。上传接口依赖 `COS_SECRET_ID` 与 `COS_SECRET_KEY`，这些密钥不能写入 Dockerfile，必须在容器运行时通过 `.env` 注入。当前手写部署脚本缺少启动后健康检查，容器启动失败时只会在 nginx 侧表现为 502。

## 范围

### In Scope

- 新增仓库内标准 `deploy_backend.sh`。
- 使用 `docker run --env-file /opt/backend/.env` 注入数据库、Redis、JWT、COS 等运行时变量。
- 部署前校验关键环境变量存在。
- 部署后执行容器内 `/health` 检查，失败时打印容器日志。

### Out of Scope

- 修改 nginx 配置。
- 将密钥写入 Dockerfile 或提交到仓库。
- 引入 docker-compose 部署。

## 接口与数据影响

- 受影响路由/handler: 无。
- 受影响服务: 后端容器部署入口。
- 受影响 middleware: 无。
- 受影响表/model: 无。
- 配置或环境变量变化: 部署脚本要求 `.env` 包含 `DATABASE_*`、`REDIS_*`、`JWT_SECRET`、`COS_SECRET_ID`、`COS_SECRET_KEY`。
- Docker / 部署影响: 生产部署使用 `deploy_backend.sh` 构建 Dockerfile 镜像并启动 `backend` 容器。

## 权限与合规影响

- 受影响角色: 运维部署人员。
- 新增/调整权限: 无。
- 是否影响审核、风控、日志留痕: 不影响业务审计；部署失败时输出容器日志便于定位。

## 验收标准

1. 缺少 `.env` 或关键环境变量时部署脚本失败并提示缺失项。
2. 容器启动时通过 `--env-file` 注入 COS 密钥，上传接口不再因进程缺少密钥返回 `cos credentials are not configured`。
3. 容器启动失败时脚本打印最近后端日志，避免只看到 nginx 502。
