# Feature Spec

## 基本信息

- Feature ID: `SL-BE-GIN-HARNESS-ALIGN`
- Feature 名称: Gin 重写后的 harness 与 AI 文档对齐
- 负责人: AI agent
- 目标上线环境: staging

## 背景

仓库代码已经迁移到 Go + Gin，但 `harness/`、`AGENTS.md`、`docs/ai/` 和 feature 模板中仍残留 Maven / JAR / Admin Backend 的旧约定。继续沿用旧文档会让后续开发、review、release 与 deploy 流程失真。

## 范围

### In Scope

- 将 `harness` 脚本切换到 Go/Gin 项目的验证与发布入口
- 补全 `docs/ai/*.md`，使其描述与当前目录结构、验证命令、Git 前置条件一致
- 调整 feature 模板，让新需求文档以 Gin handler / middleware / model / config / deploy 为中心
- 补齐本次迁移所需的 feature 文档四件套

### Out of Scope

- 业务接口重写
- 数据库 schema 设计
- CI 平台或云部署平台的具体流水线改造

## 接口与数据影响

- 受影响路由/handler: 无业务接口改动
- 受影响服务: `harness/*.sh`
- 受影响 middleware: 无
- 受影响表/model: 无
- 配置或环境变量变化: `DEFAULT_BRANCH`、`ENV_FILE`、`DEPLOY_EXECUTE`
- Docker / 部署影响: `Dockerfile` 切换为 Go 构建，多阶段镜像与容器健康检查对齐 Gin 服务

## 权限与合规影响

- 受影响角色: 开发者、reviewer、release owner
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 仅影响工程流程，不改业务权限

## 验收标准

1. `./harness/verify.sh` 可以在当前 Go 仓库中执行通过，不再依赖 Maven。
2. `docs/ai/`、`AGENTS.md` 与 feature 模板不再引用旧的 Java / JAR 约定。
3. `./harness/review.sh` 与 `./harness/release.sh` 可基于本次 feature 文档通过基础门禁。
