# SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE Implementation Plan

## Feature 信息

- Feature ID: SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/servicefulfillment/domain/entity/entity.go`
  - `internal/modules/servicefulfillment/ui/dto/dto.go`
  - `internal/modules/servicefulfillment/ui/handler/handler.go`
  - `internal/modules/servicefulfillment/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 在出入库记录模型中增加委托企业字段。
2. 增加仓储托管请求 DTO、创建接口和我的请求查询接口。
3. 扩展出入库记录列表按委托方/仓储方角色过滤。
4. 更新 feature 文档并执行后端验证。

## 验证计划

- `gofmt`
- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/verify.sh`

## 风险与回滚关注点

- `request_enterprise_id` 是新增 nullable 语义字段，旧记录为空不影响既有仓储作业。
- 回滚本 PR 后需确认已创建的托管请求不再依赖新接口处理。
