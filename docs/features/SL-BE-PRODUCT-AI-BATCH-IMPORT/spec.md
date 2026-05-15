# Product AI Batch Import Spec

## 基本信息

- Feature ID: SL-BE-PRODUCT-AI-BATCH-IMPORT
- Feature 名称: 商品 AI 批量导入
- 负责人: Codex
- 目标上线环境: backend API

## 背景

卖家希望上传商品台账、表格或截图，由系统识别出品种、牌号、规格、产地、存放地等字段，先生成可确认草稿，再批量添加商品。第一阶段已经补齐商品结构化字段，本阶段承接导入识别和确认落库。

## 范围

### In Scope

- 新增卖家商品批量导入解析接口，支持 csv、tsv、txt、xlsx 和图片文件。
- 新增流式解析接口，AI 图片识别期间通过 SSE 返回进度，降低 30 秒请求超时风险。
- 表格类文件提供规则解析兜底；启用既有 AI 模型配置时优先走大模型解析。
- 图片类文件优先通过 MiniMax Token Plan VLM 解析，识别失败时返回明确错误。
- 新增批量确认接口，将商家确认后的草稿批量创建为商品。
- 默认保存为草稿；商家选择直接提交审核时执行更严格字段校验。

### Out of Scope

- 不新增持久化导入会话表；草稿由前端确认页持有。
- 不实现 PDF OCR 和旧版 xls 解析。
- 不绕过现有商品审核流程。

## 接口与数据影响

- 新增路由:
  - `POST /api/v1/products/batch-import/parse`
  - `POST /api/v1/products/batch-import/parse-stream`
  - `POST /api/v1/products/batch-import/confirm`
- 受影响服务:
  - `internal/modules/product/application/service`
  - `internal/modules/product/ui/handler`
- 受影响表/model:
  - `products`
- 配置或环境变量变化:
  - 复用 `AI_PROCUREMENT_MODEL_PARSING_ENABLED`、`AI_PROCUREMENT_BASE_URL`、`AI_PROCUREMENT_MODEL`、`AI_PROCUREMENT_API_KEY_ENV` 等既有模型配置。
  - 新增图片识别配置: `AI_PROCUREMENT_VISION_PROVIDER`、`AI_PROCUREMENT_VISION_BASE_URL`、`AI_PROCUREMENT_VISION_API_KEY_ENV`、`AI_PROCUREMENT_VISION_TIMEOUT_SEC`。
  - 默认图片识别 provider 为 `minimax_token_plan`，默认端点为 `https://api.minimaxi.com/v1/coding_plan/vlm`。

## 权限与合规影响

- 受影响角色: 具备 `sales` 能力的卖家。
- 新增/调整权限: 无新增权限点，复用商品发布权限。
- 审核影响: 批量确认默认创建草稿；直接提交审核仍进入待审核状态。

## 验收标准

1. 卖家上传 csv/xlsx 台账后，后端返回可编辑商品草稿。
2. 返回草稿包含品种、牌号、规格、产地、存放地、库存、价格和识别置信度。
3. 商家确认后可批量创建商品，默认状态为草稿。
4. 选择直接提交审核时，缺少主图、库存、单价等关键字段的行返回逐行错误。
5. 启用 AI 模型配置和 MiniMax Token Plan Key 后，截图文件可以通过 MiniMax VLM 解析。
6. 前端调用流式解析接口时，后端在识别期间持续返回阶段事件，最终返回同结构的商品草稿。
