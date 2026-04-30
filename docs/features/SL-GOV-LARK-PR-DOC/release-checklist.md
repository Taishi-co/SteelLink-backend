# SL-GOV-LARK-PR-DOC Release Checklist

## 发布基本信息

- Feature ID: SL-GOV-LARK-PR-DOC
- 发布负责人:
- 目标环境: repository governance

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-GOV-LARK-PR-DOC` 已通过
- [ ] `./harness/lark-pr-doc.sh docs/features/SL-GOV-LARK-PR-DOC --dry-run` 已通过
- [ ] 已确认 `lark-cli doctor` 授权可用
- [ ] 已确认共享 PR 记录中存在 `Backend` 段落

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: revert 本次治理脚本与文档改动
- 回滚后验证: `./harness/verify.sh`
