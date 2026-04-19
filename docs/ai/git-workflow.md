# Git Workflow

本文件定义 `SteelLink-backend` 的分支、PR、review 与合并规范。

说明：本规则假设当前目录已经是新的 Git 工作树。如果目录只是代码副本或 remote 尚未配置，先完成 `git init` / 重新 clone / `git remote add origin <new-repo>`，再执行下面流程。

## 主干规则

- `main` 是唯一受保护主干
- 禁止任何人或 agent 直接向 `main` 提交业务改动
- 所有需求必须从最新 `main` 拉出分支开发
- 所有代码必须通过 PR 合入 `main`
- PR 通过 review、CI 和必要的 CODEOWNERS 审核后，才允许合并

## 标准开发流程

1. `git fetch origin`
2. `git switch main`
3. `git pull --ff-only`
4. `git switch -c <type>/<feature-id>-<slug>`
5. 补齐 `docs/features/<feature-id>/`
6. 开发并执行 `./harness/verify.sh`
7. 发起指向 `main` 的 PR
8. 通过 `./harness/branch-guard.sh ci`、review gate、CI checks
9. 使用 `Squash merge` 合入 `main`
10. 如果仓库 remote 或默认分支发生变更，先同步更新 `harness/branch-guard.sh` 所依赖的环境变量或 CI 配置

## 分支命名规范

允许的前缀：

- `feat/`
- `fix/`
- `docs/`
- `chore/`
- `refactor/`
- `test/`
- `hotfix/`

## PR 规则

- base branch 必须是 `main`
- head branch 不能是 `main`
- PR 描述必须包含 `Feature ID`、`Feature 文档`、验证结果、风险与回滚说明
- 所有 review comments 必须处理或解释
- 合并前必须保证 CI checks 全绿

## 工具入口

- 本地分支检查：`./harness/branch-guard.sh local`
- CI PR 检查：`./harness/branch-guard.sh ci`
- 默认基线分支：`main`，必要时可通过环境变量 `DEFAULT_BRANCH` 覆盖
