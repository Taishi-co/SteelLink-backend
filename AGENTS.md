# AGENTS.md

本文件为 Codex、Claude Code 以及新接入 agent 在 `SteelLink-backend` 中工作的统一入口。

当前仓库已经切到 Go + Gin 重写方向，所有流程、脚本和文档都应以现有 `go.mod`、`cmd/api`、`internal/pkg`、`pkg` 为准，不再沿用旧的 Maven / JAR 假设。

## 1. 启动协议

进入仓库后按以下顺序建立上下文：

1. 阅读 `AGENTS.md`
2. 阅读 `docs/ai/entrypoint.md`
3. 阅读 `docs/ai/project-map.md`
4. 阅读 `docs/ai/engineering-rules.md`
5. 阅读 `docs/ai/git-workflow.md`
6. 如处理具体需求，再阅读对应 `docs/features/<feature-id>/`

如果以上文档与代码不一致，应先更新文档或在结果中明确指出差异，不要自行假设。

## 2. 标准工作流

所有变更必须按以下阶段推进：

1. 设计阶段：先补齐 `spec.md` 与 `test-cases.md`
2. 开发阶段：按 `implementation-plan.md` 实施
3. Review 阶段：执行 `./harness/review.sh`
4. 准备上线阶段：执行 `./harness/release.sh`
5. 部署阶段：统一通过 `./harness/deploy.sh <staging|prod>`

禁止直接在 `main` 上开发或绕过 PR 合并。

## 3. 必用命令入口

- `./harness/bootstrap.sh`
- `./harness/branch-guard.sh local`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/<feature-id>`
- `./harness/release.sh docs/features/<feature-id>`
- `./harness/deploy.sh staging`
- `./harness/deploy.sh prod`

## 4. 质量底线

- 必须通过 `go test ./...`
- 必须通过 `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api`
- 如本机存在 Docker Compose，必须确认 `docker compose config` 可解析
- 涉及角色、权限、审批、风控时，必须核对 `docs/domain/role-permission-matrix.md`
- 涉及路由、handler、middleware、model、数据库、环境变量、Docker 或部署改动时，必须在 feature 文档中写明影响范围
- 提交 PR 前必须确认当前分支来自最新 `main`
- 分支命名必须符合 `feat|fix|docs|chore|refactor|test|hotfix/<feature-id>-<slug>`
- 如果当前目录尚未初始化为 Git worktree，先补齐新的仓库 remote/clone，再执行 `branch-guard` 与 PR 流程

## 5. 输出要求

提交结果时至少说明：

- 当前对应的 feature 文档路径
- 已执行的 harness 命令
- 是否完成 test/package 验证
- 是否存在遗留风险、人工验证点或部署前依赖
