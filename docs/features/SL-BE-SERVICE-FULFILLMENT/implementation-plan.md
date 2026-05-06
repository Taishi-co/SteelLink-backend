# Implementation Plan

## Feature 信息

- Feature ID: SL-BE-SERVICE-FULFILLMENT
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `cmd/api/main.go`
  - `internal/modules/serviceoffer/*`
  - `internal/modules/servicefulfillment/*`
- 文档文件:
  - `docs/features/SL-BE-SERVICE-FULFILLMENT/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展 `serviceoffer` handler/router/dto，补齐仓储和物流资源更新、下线能力。
2. 新增 `servicefulfillment` entity/dto/handler/router/migrate。
3. 实现商品仓库绑定和仓库商品查询。
4. 实现出入库记录、物流估价、物流委托、接单、运输轨迹、保障记录接口。
5. 在 `cmd/api/main.go` 接入迁移与路由。
6. 执行 Go 格式化、构建、测试、harness review/release。

## 验证计划

- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-BE-SERVICE-FULFILLMENT`
- `./harness/release.sh docs/features/SL-BE-SERVICE-FULFILLMENT`
- 如涉及镜像或容器：`docker compose config`
- 需要的人工验证: 使用真实账号走仓储商、物流商、卖家、买家 API smoke test

## 风险与回滚关注点

- 新增表会随默认 AutoMigrate 创建；回滚代码前需确认线上是否已有业务数据。
- 当前估价按物流服务 `price_rule` 的数字或默认单价进行简化估算，后续可替换为线路价格模型。
- 当前商品仓库绑定仍写入既有 `products.warehouse_info` JSON，未引入单独商品仓库关系表。
