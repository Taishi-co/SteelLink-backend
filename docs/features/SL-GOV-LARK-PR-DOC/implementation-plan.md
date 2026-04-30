# SL-GOV-LARK-PR-DOC Implementation Plan

## 变更范围

- 新增 `harness/lark-pr-doc.sh`
- 更新 `AGENTS.md`
- 更新 `docs/ai/git-workflow.md`
- 更新 `docs/features/_template/release-checklist.md`

## 实施步骤

1. 新增飞书 PRD/PR 记录同步脚本。
2. 通过仓库名将 `SteelLink-backend` 映射到共享 PR 记录的 `Backend` 段落。
3. 脚本读取 feature PRD 产物，生成临时 Markdown，并调用 `lark-cli docs +create` 创建飞书文档。
4. 脚本解析新文档 URL，并调用 `lark-cli docs +update` 写入共享 PR 记录。
5. 将该步骤写入 PR 工作流和 release checklist 模板。

## 风险与缓解

- 飞书 token 过期：脚本依赖 `lark-cli doctor` 可验证授权状态；失败时输出明确错误。
- 共享 PR 记录结构变化：可通过 `LARK_PR_INDEX_DOC` 和 `LARK_PR_WIKI_NODE` 覆盖目标；如果段落标题变化，需要同步脚本映射。
- 外部写操作不可重复：提供 `--dry-run` 用于验证命令和权限，不把写操作并入 `release.sh`。
