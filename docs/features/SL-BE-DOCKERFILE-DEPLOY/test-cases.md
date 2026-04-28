# Dockerfile Backend Deployment Script Test Cases

## 测试范围

- Feature ID: SL-BE-DOCKERFILE-DEPLOY
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | Dockerfile 部署 | `/opt/backend/.env` 包含必填变量，Docker 可用 | 执行 `./deploy_backend.sh` | 构建镜像、启动容器、健康检查通过 |
| TC-02 | COS 环境注入 | `.env` 包含 `COS_SECRET_ID` 与 `COS_SECRET_KEY` | 启动容器后调用上传接口 | 后端进程可读取 COS 密钥，不返回缺密钥错误 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 缺少 `.env` | `/opt/backend/.env` 不存在 | 执行 `./deploy_backend.sh` | 脚本失败并提示缺少 env file |
| TC-EX-02 | 缺少 COS 密钥 | `.env` 缺少 `COS_SECRET_KEY` | 执行 `./deploy_backend.sh` | 脚本失败并提示缺少 `COS_SECRET_KEY` |
| TC-EX-03 | 容器启动失败 | 数据库或 Redis 配置不可达 | 执行 `./deploy_backend.sh` | 健康检查失败并输出 `docker logs backend` |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 部署人员 | 执行生产部署脚本 | 只需 Docker 与仓库目录权限，不需要把密钥写入镜像 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 测试、构建、配置检查通过 |
| TC-OPS-02 | 脚本语法检查 | `bash -n deploy_backend.sh` | Shell 语法通过 |
