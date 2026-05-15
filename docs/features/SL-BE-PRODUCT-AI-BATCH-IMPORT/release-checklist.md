# Product AI Batch Import Release Checklist

## 发布基本信息

- Feature ID: SL-BE-PRODUCT-AI-BATCH-IMPORT
- 发布负责人: Codex
- 目标环境: backend API

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-PRODUCT-AI-BATCH-IMPORT` 已通过
- [ ] `go test ./...` 已通过
- [ ] 已确认 AI 模型配置是否在目标环境启用
- [ ] 已确认 MiniMax Token Plan Key 已写入 `AI_PROCUREMENT_VISION_API_KEY_ENV` 指向的环境变量
- [ ] 已确认 smoke test 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-BE-PRODUCT-AI-BATCH-IMPORT --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一版 backend API 镜像
- 回滚步骤: 回滚应用代码。
- 回滚后验证: 手工发布商品和商品列表可正常使用。
