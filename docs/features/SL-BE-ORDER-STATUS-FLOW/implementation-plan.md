# Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-ORDER-STATUS-FLOW`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/trade/domain/entity/order.go`
  - `internal/modules/trade/ui/dto/dto.go`
  - `internal/modules/trade/ui/handler/handler.go`
  - `internal/modules/trade/application/service/service.go`
  - `internal/modules/trade/infrastructure/router/router.go`
  - `internal/modules/trade/infrastructure/persistence/migrate.go`
- 文档文件:
  - `docs/features/SL-BE-ORDER-STATUS-FLOW/*`

## 实施步骤

1. 扩展订单状态枚举，加入一期闭环所需的中间状态。
2. 新增合同、支付回单、提货码模型并纳入 AutoMigrate。
3. 增加买家端合同提交、支付回单上传、提货码查看接口。
4. 增加平台端订单列表、详情、支付审核接口。
5. 增加提货码核销接口并驱动订单状态完结。
6. 将订单详情响应补齐为可返回合同、回单、提货码数据。

## 验证计划

- `go build ./...`
- 人工验证:
  - 创建订单
  - 提交合同
  - 上传支付回单
  - 平台审核支付
  - 获取并核销提货码

## 风险与回滚关注点

- 当前权限模型仍是简化版，核销主体未细分到仓库角色。
- 当前支付审核驳回会将订单退回 `pending_payment`，不支持多笔回单历史。
- 提货码为本地联调用轻量实现，不可直接视为生产安全方案。
