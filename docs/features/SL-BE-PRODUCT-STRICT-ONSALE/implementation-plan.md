# SL-BE-PRODUCT-STRICT-ONSALE Implementation Plan

## Feature 信息

- Feature ID: SL-BE-PRODUCT-STRICT-ONSALE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/product/infrastructure/persistence/product_repository.go`
- 文档文件:
  - `docs/features/SL-BE-PRODUCT-STRICT-ONSALE/spec.md`
  - `docs/features/SL-BE-PRODUCT-STRICT-ONSALE/implementation-plan.md`
  - `docs/features/SL-BE-PRODUCT-STRICT-ONSALE/test-cases.md`
  - `docs/features/SL-BE-PRODUCT-STRICT-ONSALE/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 移除商品仓储列表查询中 `approved_at IS NOT NULL` 的可售兜底。
2. 将所有状态查询统一为 `status = ?` 精确匹配。
3. 保持 handler 的公开列表默认 `status=2` 行为不变。
4. 补齐 feature 文档并运行 harness 验证。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-PRODUCT-STRICT-ONSALE`
- `./harness/release.sh docs/features/SL-BE-PRODUCT-STRICT-ONSALE`
- 需要的人工验证: 生产部署后请求公开商品列表，确认异常 `status=1` 商品不再展示。

## 风险与回滚关注点

- 历史依赖 `approved_at` 兜底展示的脏数据会从前台消失，需要通过数据清洗明确改成 `status=2` 才可展示。
- 回滚时恢复原查询逻辑并重新部署 backend。
