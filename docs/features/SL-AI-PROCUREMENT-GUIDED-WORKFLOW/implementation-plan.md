# Implementation Plan: AI采购渐进式后端工作流

## Feature 信息

- Feature ID: SL-AI-PROCUREMENT-GUIDED-WORKFLOW
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/aiprocurement/domain/entity/procurement.go`
  - `internal/modules/aiprocurement/domain/repository/repository.go`
  - `internal/modules/aiprocurement/infrastructure/persistence/procurement_repository.go`
  - `internal/modules/aiprocurement/infrastructure/router/router.go`
  - `internal/modules/aiprocurement/ui/dto/dto.go`
  - `internal/modules/aiprocurement/ui/handler/handler.go`
  - `internal/modules/aiprocurement/application/service/service.go`
  - `internal/modules/aiprocurement/application/service/vector_matcher.go`
- 文档文件:
  - `docs/features/SL-AI-PROCUREMENT-GUIDED-WORKFLOW/*`
- 配置文件: 无
- 脚本文件: 无

## 实施步骤

1. 扩展 AI采购 DTO 和 entity JSON 字段，支持候选商品详情、真实订单快照和历史保存标记。
2. 新增参数确认、商品召回、生成订单、融资匹配接口。
3. 商品匹配改为最多返回 10 个候选商品，向量不可用时使用 MySQL fallback。
4. 生成订单阶段调用真实订单服务创建订单，并记录 AI采购流程状态。
5. 提交审核支持 `save_history`，我的AI采购列表仅展示保存到历史的记录。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-backend-api ./cmd/api`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-AI-PROCUREMENT-GUIDED-WORKFLOW`
- `./harness/release.sh docs/features/SL-AI-PROCUREMENT-GUIDED-WORKFLOW`

## 风险与回滚关注点

- 风险: `generate-order` 会真实锁定商品库存，测试环境需要使用可用库存商品。
- 风险: 未保存历史的 AI采购记录仍在表内但不会出现在我的AI采购列表。
- 回滚: 移除新增路由，恢复 AI采购 entity/dto/service 到一次性流程，保留数据库新增字段不会影响旧逻辑。
