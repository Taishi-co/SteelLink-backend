# Backend AI Entrypoint

本目录是 SteelLink 后端给 AI agent 的统一上手入口，目标是让新 agent 在 5 分钟内获得足够上下文，并进入同一套 Go + Gin 开发工作流。

## 阅读顺序

1. `AGENTS.md`
2. `docs/ai/project-map.md`
3. `docs/ai/engineering-rules.md`
4. `docs/ai/git-workflow.md`
5. `docs/ai/review-rules.md`
6. `docs/ai/release-rules.md`
7. 具体需求对应的 `docs/features/<feature-id>/`

## 本项目的单一事实源

- 项目结构与模块边界：`docs/ai/project-map.md`
- 工程规范与命令入口：`docs/ai/engineering-rules.md`
- Review 阻断条件：`docs/ai/review-rules.md`
- 上线与部署规则：`docs/ai/release-rules.md`
- 分支与 PR 规则：`docs/ai/git-workflow.md`
- 角色权限与合规边界：`docs/domain/role-permission-matrix.md`

## 新需求执行协议

收到需求后不要立即写代码，先确认以下四件事：

1. 是否已经存在 feature 文档目录
2. 是否已有 `spec.md`
3. 是否已有 `test-cases.md`
4. 当前需求是否会影响权限、接口、数据库、配置或部署
5. 当前工作分支是否符合 Git 合并规范

若缺少 feature 文档，请基于 `docs/features/_template/` 新建一个目录后再开始开发。
若当前目录还不是 Git worktree，先补齐新的仓库初始化或重新 clone，再运行 `./harness/branch-guard.sh local`。

## 标准命令

```bash
./harness/bootstrap.sh
./harness/branch-guard.sh local
./harness/verify.sh
./harness/review.sh docs/features/<feature-id>
./harness/release.sh docs/features/<feature-id>
./harness/deploy.sh staging
```

## 默认原则

- 优先复用现有 Gin router、handler、middleware、database、pkg 边界
- 先收敛影响面，再实现代码
- 先让验证路径可执行，再宣称功能完成
- 所有部署动作必须可由脚本重放，不能依赖个人记忆
- 不要假设旧仓库 remote、旧 Java 包路径或旧部署方式仍然有效，统一以当前目录实际内容为准
