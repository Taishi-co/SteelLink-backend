# Release Rules

## 目标

确保“准备上线”和“部署执行”是标准化、可复现、可回滚的同一套流程。

## 准备上线阶段必须完成

1. 执行 `./harness/release.sh docs/features/<feature-id>`
2. 完成 `release-checklist.md`
3. 确认环境变量、数据库、Docker 变更
4. 明确 smoke test 范围与回滚点
5. 确认变更已通过 PR 合入 `main`

## 部署原则

- staging 与 production 统一走 `./harness/deploy.sh`
- 不允许直接在服务器上手工拼装 docker 命令替代脚本
- 部署命令应与 CI 中调用的命令保持一致

## 上线前最少检查项

- 测试与打包通过
- Docker 构建配置存在且可用
- 数据库、Redis、JWT 等依赖来源明确
- Gin 服务端口、路由前缀、配置文件与 mode 切换方式明确

## 上线后最少验证项

- 健康检查或 Swagger 页可访问
- 管理员登录接口可响应
- 仪表盘接口可返回数据
- 日志输出与容器状态正常
