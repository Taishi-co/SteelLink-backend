# Implementation Plan

## 实现概述

本计划修复直采支付货权凭证问题，主要包含三个部分：
1. 修复订单项预加载错误
2. 增强直采支付货权凭证功能
3. 优化错误处理

## 实现步骤

### 1. 修复订单项预加载错误

**文件:** `internal/modules/order/application/service/order_service.go`

**改动:**
- 删除 `UploadPaymentVoucher` 方法中的错误预加载代码
- `Items` 字段是JSON类型，不是数据库关联关系，不需要预加载

**代码变更:**
```go
// 删除以下代码
if err := tx.Model(&order).Association("Items").Find(&order.Items); err != nil {
    return ecode.ErrDatabase
}
```

### 2. 增强直采支付货权凭证功能

**文件 1:** `internal/modules/order/ui/dto/order_dto.go`

**改动:**
- 在 `DirectPaymentRequest` 中添加 `TitleVoucherURL` 和 `TitleVoucherRemark` 字段

**代码变更:**
```go
type DirectPaymentRequest struct {
    VoucherURL        string   `json:"voucher_url"`
    TitleVoucherURL   string   `json:"title_voucher_url"`      // 新增
    TitleVoucherRemark string  `json:"title_voucher_remark"`   // 新增
    TransactionNo     string   `json:"transaction_no"`
    PayerAccount      string   `json:"payer_account"`
    PayeeAccount      string   `json:"payee_account"`
    CapturedAt        string   `json:"captured_at"`
    GPSLatitude       *float64 `json:"gps_latitude"`
    GPSLongitude      *float64 `json:"gps_longitude"`
}
```

**文件 2:** `internal/modules/order/application/service/order_service.go`

**改动:**
- 修改 `DirectPayment` 方法
- 优先使用请求中的货权凭证
- 如果请求中没有，才从商品获取

**代码变更:**
```go
// 优先使用请求中的货权凭证
titleVoucherURL = strings.TrimSpace(req.TitleVoucherURL)
titleVoucherRemark = strings.TrimSpace(req.TitleVoucherRemark)

// 如果请求中没有，尝试从商品获取
if titleVoucherURL == "" && len(order.Items) > 0 {
    product, prodErr := s.productRepo.FindByID(ctx, order.Items[0].ProductID)
    if prodErr == nil && product != nil {
        titleVoucherURL = strings.TrimSpace(product.TitleVoucherURL)
        titleVoucherRemark = strings.TrimSpace(product.TitleVoucherRemark)
    }
}
```

### 3. 优化错误处理

**文件:** `internal/modules/order/application/service/order_service.go`

**改动:**
- 增强 `UploadPaymentVoucher` 和 `DirectPayment` 方法的错误处理
- 使用 `fmt.Errorf` 包装具体的错误信息
- 在返回错误时，传递原始错误信息

**代码变更:**
```go
// 例如：
return fmt.Errorf("查询订单失败: %w", err)
return fmt.Errorf("保存订单失败: %w", err)
// ...
return nil, ecode.ErrDatabase.WithMessage(err.Error())
```

## 风险评估

| 风险项 | 风险等级 | 缓解措施 |
|--------|----------|----------|
| API兼容性 | 低 | 新增字段是可选的，不影响现有调用 |
| 数据兼容性 | 低 | 只新增JSON字段，不修改数据库结构 |
| 功能回归 | 中 | 通过test cases验证原有功能 |

## 回滚方案

如果发现问题，直接回滚这个commit即可：
```bash
git revert 750844f
```

## 部署检查清单

- [ ] 代码已通过 `go build` 验证
- [ ] 代码已通过 `go test` 验证（如有相关测试）
- [ ] 功能已在staging环境验证
- [ ] 前端已配合更新API调用
