# Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-GIN-HARNESS-ALIGN`
- 开发负责人: AI agent

## 变更范围

- 代码文件: `Dockerfile`
- 文档文件: `AGENTS.md`、`docs/ai/*.md`、`docs/features/_template/*`
- 配置文件: `docker-compose.yml`
- 部署/脚本文件: `harness/*.sh`

## 实施步骤

1. 读取现有 Go/Gin 项目结构与 `harness` 脚本，确认哪些地方仍依赖 Maven / JAR。
2. 将 `bootstrap`、`verify`、`review`、`release`、`deploy`、`resolve-feature-dir` 改为适配 Go 项目的工作流，并补充 Git 仓库切换时的提示。
3. 更新 `AGENTS.md`、`docs/ai/*.md` 与 feature 模板，确保项目地图、质量门槛、发布规则、Git 前置条件都与当前仓库一致。
4. 将 `Dockerfile` 对齐当前 Go 版本，并去掉 `docker-compose.yml` 的过时字段，避免 release / deploy 文档继续引用旧栈。

## 验证计划

- `./harness/bootstrap.sh`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-GIN-HARNESS-ALIGN`
- `./harness/release.sh docs/features/SL-BE-GIN-HARNESS-ALIGN`
- 如涉及镜像或容器：`docker compose config`
- 需要的人工验证: Git remote 初始化后再次执行 `./harness/branch-guard.sh local`

## 风险与回滚关注点

- 当前目录不是 Git worktree，无法在本次变更中验证新的 branch guard happy path。
- `docker-compose.yml` 依赖的 `scripts/sql/init.sql` 当前仓库中仍未落地，真实启动 MySQL 时需要补齐初始化脚本或移除挂载。
- 若需要回滚，只需恢复 `harness/`、`docs/`、`Dockerfile`、`docker-compose.yml` 到迁移前版本。
