# Product AI Batch Import Test Cases

## 测试范围

- Feature ID: SL-BE-PRODUCT-AI-BATCH-IMPORT
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | CSV 台账解析 | 卖家已登录 | 上传包含品种、牌号、规格、产地、存放地的 csv | 返回商品草稿列表 |
| TC-02 | XLSX 台账解析 | 卖家已登录 | 上传第一张 sheet 包含商品台账的 xlsx | 返回商品草稿列表 |
| TC-03 | 确认创建草稿 | 已有解析草稿 | POST `/products/batch-import/confirm`，`submit_for_review=false` | 批量创建商品，状态为草稿 |
| TC-04 | 直接提交审核 | 草稿补齐主图、库存、价格、存放地 | POST confirm，`submit_for_review=true` | 商品状态为待审核 |
| TC-05 | 流式截图识别 | 已配置 MiniMax Token Plan Key | POST `/products/batch-import/parse-stream` 上传截图 | 先收到进度事件，最终收到 done 事件和商品草稿 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 不支持旧版 xls | 上传 `.xls` 文件 | 调用 parse | 返回不支持旧版 xls |
| TC-EX-02 | 图片识别未启用模型 | 未配置 AI 模型或 MiniMax Token Plan Key | 上传截图 | 返回需要启用支持图片识别的 AI 模型 |
| TC-EX-03 | 行字段缺失 | 草稿缺少品种或牌号 | 调用 confirm | 返回逐行错误，不影响其他有效行 |
| TC-EX-04 | MiniMax VLM 返回错误 | Token Plan Key 无效或额度不足 | 上传截图 | 返回 MiniMax 图像识别错误 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录 | 调用 parse/confirm | 返回未登录 |
| TC-AUTH-02 | 无 sales 能力用户 | 调用 parse/confirm | 被商品发布权限拦截 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端测试 | `go test ./internal/modules/product/...` | 商品模块测试通过 |
| TC-OPS-02 | 后端全量测试 | `go test ./...` | 全量测试通过 |
| TC-OPS-03 | 发布门禁 | `./harness/release.sh docs/features/SL-BE-PRODUCT-AI-BATCH-IMPORT` | release gate 通过 |
| TC-OPS-04 | MiniMax Token Plan 配置 | 设置 `AI_PROCUREMENT_MODEL_PARSING_ENABLED=true`、`AI_PROCUREMENT_VISION_PROVIDER=minimax_token_plan`、`AI_PROCUREMENT_VISION_BASE_URL=https://api.minimaxi.com/v1/coding_plan/vlm`、`AI_PROCUREMENT_VISION_API_KEY_ENV=MINIMAX_TOKEN_PLAN_API_KEY`、`AI_PROCUREMENT_VISION_TIMEOUT_SEC=120` | 截图解析走 MiniMax VLM，长耗时识别不再被 20 秒默认文本模型超时截断 |
