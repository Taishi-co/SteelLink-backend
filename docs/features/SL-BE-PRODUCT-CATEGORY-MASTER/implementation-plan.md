# SL-BE-PRODUCT-CATEGORY-MASTER Implementation Plan

1. 新增 `ProductCategory` GORM 实体、DTO、仓储和服务方法。
2. 启动迁移创建 `product_categories` 表，并从现有商品 `category` 去重种子化。
3. 在商品创建流程中确保品类主数据存在。
4. 注册公开读取接口和运营管理 CRUD 接口。
5. 运行 `go test ./...` 与 `go build ./cmd/api`。
