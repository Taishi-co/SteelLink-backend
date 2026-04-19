# SteelLink Backend Claude Guide

本文件保留给 Claude Code 使用，但不再单独维护规则内容。统一规则以仓库内治理文档为准。

## 必读顺序

1. `AGENTS.md`
2. `docs/ai/entrypoint.md`
3. `docs/ai/project-map.md`
4. `docs/ai/engineering-rules.md`
5. `docs/ai/git-workflow.md`
6. 如处理具体需求，再阅读 `docs/features/<feature-id>/`

## 执行要求

- 不要跳过设计文档直接编码
- 不要绕过 `harness/` 直接拼接验证、发布、部署命令
- 不要直接在 `main` 上开发或发起从非规范分支创建的 PR
- 涉及权限、审核、风控时必须核对 `docs/domain/role-permission-matrix.md`

## 标准命令

```bash
./harness/bootstrap.sh
./harness/branch-guard.sh local
./harness/verify.sh
./harness/review.sh docs/features/<feature-id>
./harness/release.sh docs/features/<feature-id>
./harness/deploy.sh staging
```
