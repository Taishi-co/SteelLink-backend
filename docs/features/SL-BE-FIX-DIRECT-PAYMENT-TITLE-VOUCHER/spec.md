# Feature Spec

## 基本信息

- Feature ID: SL-BE-FIX-DIRECT-PAYMENT-TITLE-VOUCHER
- Feature 名称: 修复直采支付货权凭证问题
- 负责人: AI Agent
- 目标上线环境: Staging / Prod

## 背景

直采支付功能存在两个问题：
1. 上传支付凭证时出现"预加载订单项失败: unsupported relations: Items"错误
2. 货权凭证只能在商品上架时认证，无法在支付时上传
3. 错误信息不够详细，难以调试

## 范围

### In Scope

- 修复订单项预加载错误
- 增强直采支付货权凭证功能，支持在支付时上传
- 优化错误处理，提供更详细的错误信息

### Out of Scope

- 不修改其他支付方式的逻辑
- 不修改商品上架时的货权凭证认证流程

## 接口与数据影响

- 受影响路由/handler:
  - `POST /api/v1/orders/:id/payments/direct`
- 受影响服务:
  - `OrderService.UploadPaymentVoucher`
  - `OrderService.DirectPayment`
- 受影响 middleware: 无
- 受影响表/model:
  - `Order` model (JSON字段无变更)
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买家
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 否

## 验收标准

1. 直采支付时不再出现"预加载订单项失败"错误
2. 可以在直采支付时通过`title_voucher_url`和`title_voucher_remark`字段上传货权凭证
3. 如果商品已有货权凭证，会优先使用商品的凭证
4. 错误信息更加详细，包含具体的错误原因
5. 原有的支付凭证上传功能不受影响
