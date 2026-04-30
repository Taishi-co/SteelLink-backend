# SL-GOV-LARK-PR-DOC Spec

## 背景

共享飞书 Wiki《交易平台PR记录》已经按 `Admin`、`Backend`、`Frontend` 分段维护 PR 记录，三段分别对应 `SteelLink-admin`、`SteelLink-backend`、`SteelLink-frontend` 三个仓库。

现有 PR 流程只要求提交代码、feature 文档和验证结果，缺少把 PRD/PR 记录同步到飞书的确定性步骤。

## 目标

- 在提交 GitHub PR 后，基于 `docs/features/<feature-id>/` 下的 PRD 产物自动创建飞书文档。
- 将新飞书文档链接写入共享 PR 记录的 `Backend` 段落。
- 允许通过环境变量覆盖共享 PR 记录文档和父 Wiki 节点，避免脚本绑定单一环境。
- 保持 `release.sh` 为本地验证入口，飞书写入作为显式 PR 后置动作执行。

## 非目标

- 不自动创建 GitHub PR。
- 不修改飞书文档权限模型。
- 不把飞书写入放进 CI 或 release gate 中强制执行。

## 验收标准

- `./harness/lark-pr-doc.sh --help` 能展示用法。
- `./harness/lark-pr-doc.sh docs/features/SL-GOV-LARK-PR-DOC --dry-run` 能展示创建飞书文档和插入共享 PR 记录的 dry-run 请求。
- `AGENTS.md`、`docs/ai/git-workflow.md`、release checklist 模板都包含飞书 PRD/PR 记录步骤。
