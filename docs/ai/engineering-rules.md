# Engineering Rules

## 开发原则

- 设计先于编码：没有 `spec.md` 与 `test-cases.md` 的需求，不应直接开始开发
- 小步提交：每次变更聚焦一个 feature 或一个子问题
- 影响可追溯：涉及权限、接口、数据库、部署时，必须写入 feature 文档
- 统一入口：验证、评审、发布、部署统一走 `harness/`
- 分支规范：所有开发分支必须从最新 `main` 拉出，并通过 PR 合入 `main`
- 仓库切换时先校验当前目录是否为正确的 Git worktree，不要沿用旧 remote 或旧仓库假设

## 代码约束

### 通用规范

- 优先复用现有 Gin router、handler、middleware、database、pkg 分层
- 权限改动必须同步核对 `docs/domain/role-permission-matrix.md`
- 涉及审核、风控、统计口径的改动，必须显式说明影响范围和回滚路径
- 涉及数据库改动时，必须说明受影响表、迁移顺序和回滚策略
- 新增脚本、配置、流程变更时优先更新文档再改 CI

### Go代码规范

- 使用 `go fmt` 格式化代码
- 遵循 Go 命名规范
- 错误处理使用 `pkg/ecode` 定义的错误码
- API响应使用 `pkg/response` 统一格式
- 日志使用 `pkg/logger` 的结构化日志

## 命令约束

```bash
./harness/bootstrap.sh
./harness/branch-guard.sh local
./harness/verify.sh
./harness/review.sh docs/features/<feature-id>
./harness/release.sh docs/features/<feature-id>
./harness/deploy.sh <staging|prod>
```

## 通过标准

1. `go test ./...` 测试通过
2. `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 构建成功
3. 如环境安装了 Docker Compose，`docker compose config` 可解析
4. feature 文档四件套完整
5. 变更说明覆盖接口、数据、权限、回滚关注点
6. 分支名、PR base branch 与 merge 方式符合 `docs/ai/git-workflow.md`

## 禁止事项

- 禁止跳过 `verify` 直接发起 review
- 禁止直接向 `main` push 业务改动
- 禁止只在对话中描述测试结果而不运行命令
- 禁止手工拼装与 CI 不一致的部署步骤
- 禁止继续使用旧的 Maven / JAR / `pom.xml` 假设解释当前仓库
