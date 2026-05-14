# Product Taxonomy Structure Implementation Plan

## Feature 信息

- Feature ID: SL-BE-PRODUCT-TAXONOMY-STRUCTURE
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/product/domain/entity/*`
  - `internal/modules/product/domain/repository/*`
  - `internal/modules/product/infrastructure/persistence/*`
  - `internal/modules/product/application/service/product_service.go`
  - `internal/modules/product/ui/dto/product_dto.go`
  - `internal/modules/product/ui/handler/product_handler.go`
  - `internal/modules/product/infrastructure/router/router.go`
- 文档文件:
  - `docs/features/SL-BE-PRODUCT-TAXONOMY-STRUCTURE/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新增牌号、钢厂主数据实体、仓储和列表接口。
2. 扩展商品实体、DTO、创建/修改/列表映射，保留旧文本快照。
3. 调整迁移逻辑，移除 `bundle_no` 唯一索引并创建普通索引。
4. 增加商品列表结构化规格过滤。
5. 执行 gofmt、测试和 release gate。

## 验证计划

- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/release.sh docs/features/SL-BE-PRODUCT-TAXONOMY-STRUCTURE`
- 需要的人工验证: 使用卖家账号创建相同捆包号商品，并在买家商品列表验证规格筛选。

## 风险与回滚关注点

- `bundle_no` 唯一索引移除后，外部系统不能再依赖捆包号单独定位商品，必须使用 `product_id` 或 `enterprise_id + bundle_no`。
- AutoMigrate 会新增字段和主数据表，回滚代码前需确认旧服务是否忽略新增列。
- 结构化字段作为增量兼容字段上线，旧前端仍可提交旧字段。
