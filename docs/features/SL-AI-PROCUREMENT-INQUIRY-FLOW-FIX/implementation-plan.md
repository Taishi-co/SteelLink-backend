# AI Procurement Inquiry Flow Fix Implementation Plan

## Feature 信息

- Feature ID: SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/aiprocurement/application/service/service.go`
  - `internal/modules/aiprocurement/application/service/service_test.go`
- 文档文件:
  - `docs/features/SL-AI-PROCUREMENT-INQUIRY-FLOW-FIX/*`
- 配置文件: 无
- 脚本文件: 无

## 实施步骤

1. 在 `toProcurementInfo` 中归一化 AI采购 JSON 列表字段。
2. 在 AI采购物流委托创建函数中增加数据库句柄空值防御。
3. 补充服务层单元测试覆盖响应归一和空 DB 防御。
4. 运行 AI采购服务层测试与 `cmd/api` 构建。

## 验证计划

- `go test ./internal/modules/aiprocurement/application/service`
- `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api`
- `./harness/branch-guard.sh local`
- 需要的人工验证: 联调环境用买家企业账号走 AI采购询价完整链路。

## 风险与回滚关注点

- 风险: 响应从 `null` 调整为 `[]`，前端应兼容且语义更稳定。
- 回滚: 回退本 feature 修改的服务层代码、测试和文档。
