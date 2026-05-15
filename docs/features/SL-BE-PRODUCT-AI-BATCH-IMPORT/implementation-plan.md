# Product AI Batch Import Implementation Plan

## Feature 信息

- Feature ID: SL-BE-PRODUCT-AI-BATCH-IMPORT
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/product/application/service/product_import_*.go`
  - `internal/modules/product/ui/dto/product_dto.go`
  - `internal/modules/product/ui/handler/product_handler.go`
  - `internal/modules/product/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-PRODUCT-AI-BATCH-IMPORT/*`
- 配置文件: 增加 `AI_PROCUREMENT_VISION_PROVIDER`、`AI_PROCUREMENT_VISION_BASE_URL`、`AI_PROCUREMENT_VISION_API_KEY_ENV`。

## 实施步骤

1. 增加商品导入 DTO、解析接口和确认接口。
2. 实现 csv/tsv/txt/xlsx 规则解析和规格字段拆分。
3. 复用既有 AI 模型配置，增加文本识别和 MiniMax Token Plan 图片识别入口。
4. 接收 Admin 写入 Redis 的视觉识别配置，支持在 Admin 选择 Token Plan 端点和环境变量名。
5. 将确认后的草稿逐行创建商品，返回成功和失败行。
6. 增加解析单元测试并执行 release gate。

## 验证计划

- `go test ./internal/modules/product/...`
- `go test ./...`
- `./harness/release.sh docs/features/SL-BE-PRODUCT-AI-BATCH-IMPORT`
- 需要的人工验证: 卖家上传一份真实台账，确认编辑后批量创建草稿商品。

## 风险与回滚关注点

- 图片识别依赖 MiniMax Token Plan Key 和 `/v1/coding_plan/vlm` 配置，未启用时会明确失败。
- 默认创建草稿，避免未补齐图片和审核资料的商品直接进入商城。
- 回滚代码不会影响已创建商品。
