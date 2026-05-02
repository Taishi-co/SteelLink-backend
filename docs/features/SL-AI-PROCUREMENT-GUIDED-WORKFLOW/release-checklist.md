# Release Checklist: AI采购渐进式后端工作流

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-GUIDED-WORKFLOW
- 发布负责人: Codex
- 目标环境: backend api

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-AI-PROCUREMENT-GUIDED-WORKFLOW` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api` 已通过
- [ ] 已确认 `ai_procurements.history_saved` 自动迁移策略
- [ ] 已确认前端同步接入新增接口
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-AI-PROCUREMENT-GUIDED-WORKFLOW --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录

## 回滚信息

- 回滚目标版本: 上一版 AI采购后端接口
- 回滚步骤: 回滚本 feature 涉及的 AI采购 service、handler、router、dto、entity 和 repository 变更
- 回滚后验证: `go test ./...`、`CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`

## 发布后验证

- [ ] 新增四个渐进式接口可访问
- [ ] 商品召回返回候选商品详情
- [ ] 生成订单接口创建真实订单
- [ ] 审核提交接口按 `save_history` 控制历史列表展示
