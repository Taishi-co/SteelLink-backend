package service

import (
	"context"
	"fmt"
	"time"

	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	orderRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/repository"
	orderPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/order/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo      orderRepository.OrderRepository
	productRepo    productRepository.ProductRepository
	enterpriseRepo userRepository.EnterpriseRepository
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo:      orderPersistence.NewOrderRepository(),
		productRepo:    productPersistence.NewProductRepository(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
	}
}

func (s *OrderService) Create(ctx context.Context, req *dto.CreateOrderRequest, userID, buyerEnterpriseID uint64) (*dto.OrderInfo, error) {
	if buyerEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能下单")
	}

	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}
	if !product.CanOrder(req.Quantity) {
		return nil, ecode.ErrBadRequest.WithMessage("商品不可下单或库存不足")
	}

	buyerEnterprise, err := s.enterpriseRepo.FindByID(ctx, buyerEnterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if buyerEnterprise == nil || !buyerEnterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("买家企业不存在或未通过审核")
	}

	sellerEnterprise, err := s.enterpriseRepo.FindByID(ctx, product.EnterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if sellerEnterprise == nil {
		return nil, ecode.ErrBadRequest.WithMessage("卖家企业不存在")
	}

	itemTotal := req.Quantity * product.UnitPrice
	itemTaxAmount := 0.0
	if !product.TaxIncluded {
		itemTaxAmount = itemTotal * (product.TaxRate / 100)
	}

	order := &orderEntity.Order{
		BuyerID:            buyerEnterprise.ID,
		BuyerName:          buyerEnterprise.Name,
		SellerID:           sellerEnterprise.ID,
		SellerEnterpriseID: sellerEnterprise.ID,
		SellerName:         sellerEnterprise.Name,
		Items: orderEntity.OrderItems{
			{
				ProductID:   product.ID,
				ProductNo:   product.ProductNo,
				ProductName: product.Name,
				SteelMill:   product.SteelMill,
				Category:    product.SubCategory,
				BundleNo:    product.BundleNo,
				Quantity:    req.Quantity,
				UnitPrice:   product.UnitPrice,
				TotalAmount: itemTotal,
				TaxRate:     product.TaxRate,
				TaxAmount:   itemTaxAmount,
			},
		},
		Currency:  product.Currency,
		Status:    orderEntity.OrderStatusPendingPayment,
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if !product.LockStock(req.Quantity) {
			return ecode.ErrBadRequest.WithMessage("库存不足")
		}
		product.UpdatedBy = userID
		if err := tx.Save(product).Error; err != nil {
			return err
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase.WithMessage("创建订单失败")
	}

	return toOrderInfo(order), nil
}

func (s *OrderService) GetByID(ctx context.Context, id, enterpriseID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if err := s.reconcileOrderState(ctx, order); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase
	}
	if enterpriseID == 0 || (order.BuyerID != enterpriseID && order.SellerID != enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权查看该订单")
	}

	return toOrderInfo(order), nil
}

func (s *OrderService) ListByEnterprise(ctx context.Context, enterpriseID uint64) ([]*dto.OrderInfo, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看订单")
	}
	orders, err := s.orderRepo.ListByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		result = append(result, toOrderInfo(order))
	}
	return result, nil
}

func (s *OrderService) Cancel(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if err := s.reconcileOrderState(ctx, order); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权取消该订单")
	}
	if effectiveOrderStatus(order) != orderEntity.OrderStatusPendingPayment {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可取消")
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		return nil, ecode.ErrBadRequest.WithMessage("已上传支付凭证，无法直接取消订单")
	}

	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			product, findErr := s.productRepo.FindByID(ctx, item.ProductID)
			if findErr != nil {
				return findErr
			}
			if product == nil {
				return ecode.ErrBadRequest.WithMessage("订单商品不存在，无法取消")
			}
			if !product.UnlockStock(item.Quantity) {
				return ecode.ErrBadRequest.WithMessage("订单锁定库存异常，无法取消")
			}
			product.UpdatedBy = operatorID
			if saveErr := tx.Save(product).Error; saveErr != nil {
				return saveErr
			}
		}

		order.Status = orderEntity.OrderStatusCancelled
		order.UpdatedBy = operatorID
		if saveErr := tx.Save(order).Error; saveErr != nil {
			return saveErr
		}
		return nil
	}); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase.WithMessage("取消订单失败")
	}

	return toOrderInfo(order), nil
}

func (s *OrderService) ListByBuyer(ctx context.Context, buyerID uint64) ([]*dto.OrderInfo, error) {
	if buyerID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看订单")
	}
	orders, err := s.orderRepo.ListByBuyer(ctx, buyerID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		result = append(result, toOrderInfo(order))
	}
	return result, nil
}

func (s *OrderService) ListBySeller(ctx context.Context, sellerID uint64) ([]*dto.OrderInfo, error) {
	if sellerID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看订单")
	}
	orders, err := s.orderRepo.ListBySeller(ctx, sellerID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		result = append(result, toOrderInfo(order))
	}
	return result, nil
}

func (s *OrderService) UploadPaymentVoucher(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.UploadPaymentVoucherRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if err := s.reconcileOrderState(ctx, order); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权上传支付凭证")
	}
	if effectiveOrderStatus(order) != orderEntity.OrderStatusPendingPayment {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可上传支付凭证")
	}

	order.PaymentInfo = &orderEntity.PaymentInfo{
		PaymentMethod: "bank_transfer",
		PaymentTime:   time.Now(),
		TransactionNo: req.TransactionNo,
		PayeeAccount:  req.PayeeAccount,
		PayerAccount:  req.PayerAccount,
		VoucherURL:    req.VoucherURL,
	}
	order.UpdatedBy = operatorID

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("上传支付凭证失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) ConfirmPayment(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if err := s.reconcileOrderState(ctx, order); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase
	}
	if order.SellerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认收款")
	}
	currentStatus := effectiveOrderStatus(order)
	if currentStatus != orderEntity.OrderStatusPendingPayment && currentStatus != orderEntity.OrderStatusPaid {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认收款")
	}
	if order.PaymentInfo == nil || order.PaymentInfo.VoucherURL == "" {
		return nil, ecode.ErrBadRequest.WithMessage("买家尚未上传支付凭证")
	}

	order.Status = orderEntity.OrderStatusPendingDelivery
	order.UpdatedBy = operatorID
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认收款失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) SetDeliveryInfo(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.SetDeliveryInfoRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权录入提货信息")
	}
	if effectiveOrderStatus(order) != orderEntity.OrderStatusPendingDelivery {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可录入提货信息")
	}

	pickupCode := ""
	if order.DeliveryInfo != nil && order.DeliveryInfo.PickupCode != "" {
		pickupCode = order.DeliveryInfo.PickupCode
	} else {
		pickupCode = generatePickupCode()
	}

	order.DeliveryInfo = &orderEntity.DeliveryInfo{
		DriverName:    req.DriverName,
		DriverIdCard:  req.DriverIDCard,
		DriverPhone:   req.DriverPhone,
		PlateNumber:   req.PlateNumber,
		PickupCode:    pickupCode,
		WarehouseID:   req.WarehouseID,
		WarehouseName: req.WarehouseName,
		Address:       req.Address,
		EstimatedTime: req.EstimatedTime,
	}
	order.UpdatedBy = operatorID

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("保存提货信息失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) ConfirmOutbound(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.SellerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认出库")
	}
	if effectiveOrderStatus(order) != orderEntity.OrderStatusPendingDelivery {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认出库")
	}
	if order.DeliveryInfo == nil || order.DeliveryInfo.PickupCode == "" {
		return nil, ecode.ErrBadRequest.WithMessage("买家尚未录入提货信息")
	}

	order.Status = orderEntity.OrderStatusDelivering
	order.UpdatedBy = operatorID
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认出库失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) ConfirmReceipt(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.ConfirmReceiptRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认收货")
	}
	if effectiveOrderStatus(order) != orderEntity.OrderStatusDelivering {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认收货")
	}

	now := time.Now()
	order.ReceiptVoucher = req.ReceiptVoucher
	order.ReceiptAt = &now
	order.Status = orderEntity.OrderStatusCompleted
	order.UpdatedBy = operatorID
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认收货失败")
	}
	return toOrderInfo(order), nil
}

func toOrderInfo(order *orderEntity.Order) *dto.OrderInfo {
	if order == nil {
		return nil
	}
	items := make([]dto.OrderItemInfo, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.OrderItemInfo{
			ProductID:   item.ProductID,
			ProductNo:   item.ProductNo,
			ProductName: item.ProductName,
			SteelMill:   item.SteelMill,
			Category:    item.Category,
			BundleNo:    item.BundleNo,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalAmount: item.TotalAmount,
			TaxRate:     item.TaxRate,
			TaxAmount:   item.TaxAmount,
		})
	}
	var paymentInfo *dto.PaymentInfo
	if order.PaymentInfo != nil {
		paymentInfo = &dto.PaymentInfo{
			PaymentMethod: order.PaymentInfo.PaymentMethod,
			PaymentTime:   order.PaymentInfo.PaymentTime,
			TransactionNo: order.PaymentInfo.TransactionNo,
			PayeeAccount:  order.PaymentInfo.PayeeAccount,
			PayerAccount:  order.PaymentInfo.PayerAccount,
			VoucherURL:    order.PaymentInfo.VoucherURL,
		}
	}
	var deliveryInfo *dto.DeliveryInfo
	if order.DeliveryInfo != nil {
		deliveryInfo = &dto.DeliveryInfo{
			DriverName:    order.DeliveryInfo.DriverName,
			DriverPhone:   order.DeliveryInfo.DriverPhone,
			DriverIDCard:  order.DeliveryInfo.DriverIdCard,
			PlateNumber:   order.DeliveryInfo.PlateNumber,
			PickupCode:    order.DeliveryInfo.PickupCode,
			WarehouseID:   order.DeliveryInfo.WarehouseID,
			WarehouseName: order.DeliveryInfo.WarehouseName,
			Address:       order.DeliveryInfo.Address,
			EstimatedTime: order.DeliveryInfo.EstimatedTime,
		}
	}
	return &dto.OrderInfo{
		ID:             order.ID,
		OrderNo:        order.OrderNo,
		BuyerID:        order.BuyerID,
		BuyerName:      order.BuyerName,
		SellerID:       order.SellerID,
		SellerName:     order.SellerName,
		Items:          items,
		TotalQuantity:  order.TotalQuantity,
		TotalAmount:    order.TotalAmount,
		TaxTotalAmount: order.TaxTotalAmount,
		Currency:       order.Currency,
		Status:         int(effectiveOrderStatus(order)),
		Expired:        isOrderExpired(order),
		ExpiresAt:      order.ExpiresAt,
		CreatedAt:      order.CreatedAt,
		PaymentInfo:    paymentInfo,
		DeliveryInfo:   deliveryInfo,
		ReceiptVoucher: order.ReceiptVoucher,
		ReceiptAt:      order.ReceiptAt,
		Timeline:       buildOrderTimeline(order),
	}
}

func buildOrderTimeline(order *orderEntity.Order) []dto.TimelineItem {
	if order == nil {
		return nil
	}

	status := effectiveOrderStatus(order)
	items := []dto.TimelineItem{
		{
			Key:         "created",
			Label:       "订单创建",
			Status:      "completed",
			Time:        &order.CreatedAt,
			Description: "买家已提交订单，系统锁定对应库存。",
		},
	}

	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		items = append(items, dto.TimelineItem{
			Key:         "payment_uploaded",
			Label:       "支付凭证上传",
			Status:      "completed",
			Time:        &order.PaymentInfo.PaymentTime,
			Description: "买家已上传支付凭证，等待卖家确认收款。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "payment_uploaded",
			Label:       "支付凭证上传",
			Status:      "skipped",
			Description: "订单已关闭，未上传支付凭证。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "payment_uploaded",
			Label:       "支付凭证上传",
			Status:      "pending",
			Description: "等待买家在支付截止前上传凭证。",
		})
	}

	if status >= orderEntity.OrderStatusPendingDelivery && status != orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "payment_confirmed",
			Label:       "卖家确认收款",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "卖家已确认收款，订单进入提货准备阶段。",
		})
	} else if status == orderEntity.OrderStatusPaid {
		items = append(items, dto.TimelineItem{
			Key:         "payment_confirmed",
			Label:       "卖家确认收款",
			Status:      "current",
			Description: "等待卖家核对支付凭证并确认收款。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "payment_confirmed",
			Label:       "卖家确认收款",
			Status:      "skipped",
			Description: "订单已关闭，无需确认收款。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "payment_confirmed",
			Label:       "卖家确认收款",
			Status:      "pending",
			Description: "买家上传支付凭证后由卖家确认。",
		})
	}

	if order.DeliveryInfo != nil && order.DeliveryInfo.PickupCode != "" {
		items = append(items, dto.TimelineItem{
			Key:         "pickup_info",
			Label:       "提货信息录入",
			Status:      "completed",
			Description: "买家已录入承运车辆、司机和送达地址，系统已生成提货码。",
		})
	} else if status == orderEntity.OrderStatusPendingDelivery {
		items = append(items, dto.TimelineItem{
			Key:         "pickup_info",
			Label:       "提货信息录入",
			Status:      "current",
			Description: "等待买家录入提货车辆与送达地址。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "pickup_info",
			Label:       "提货信息录入",
			Status:      "skipped",
			Description: "订单已关闭，无需录入提货信息。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "pickup_info",
			Label:       "提货信息录入",
			Status:      "pending",
			Description: "卖家确认收款后，买家可录入提货信息。",
		})
	}

	if status >= orderEntity.OrderStatusDelivering && status != orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "outbound",
			Label:       "卖家确认出库",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "卖家已核验提货码并确认货物出库。",
		})
	} else if status == orderEntity.OrderStatusPendingDelivery && order.DeliveryInfo != nil && order.DeliveryInfo.PickupCode != "" {
		items = append(items, dto.TimelineItem{
			Key:         "outbound",
			Label:       "卖家确认出库",
			Status:      "current",
			Description: "等待卖家核验提货码并确认出库。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "outbound",
			Label:       "卖家确认出库",
			Status:      "skipped",
			Description: "订单已关闭，无需出库。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "outbound",
			Label:       "卖家确认出库",
			Status:      "pending",
			Description: "提货信息录入后，由卖家确认出库。",
		})
	}

	if status == orderEntity.OrderStatusCompleted {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "completed",
			Time:        order.ReceiptAt,
			Description: "买家已确认收货并上传签收凭证，订单完成归档。",
		})
	} else if status == orderEntity.OrderStatusDelivering || status == orderEntity.OrderStatusDelivered {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "current",
			Description: "等待买家确认收货并上传签收凭证。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "skipped",
			Description: "订单已关闭，无需签收。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "pending",
			Description: "货物出库后，由买家确认收货。",
		})
	}

	if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "closed",
			Label:       "订单关闭",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "订单已关闭，锁定库存已释放。",
		})
	}

	return items
}

func (s *OrderService) reconcileOrderState(ctx context.Context, order *orderEntity.Order) error {
	if order == nil {
		return nil
	}
	if order.Status == orderEntity.OrderStatusCancelled || order.Status == orderEntity.OrderStatusCompleted {
		return nil
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
		return nil
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		return nil
	}
	if order.ExpiresAt == nil || order.ExpiresAt.After(time.Now()) {
		return nil
	}
	return s.cancelOrderWithUnlock(ctx, order, order.UpdatedBy)
}

func (s *OrderService) cancelOrderWithUnlock(ctx context.Context, order *orderEntity.Order, operatorID uint64) error {
	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			product, findErr := s.productRepo.FindByID(ctx, item.ProductID)
			if findErr != nil {
				return findErr
			}
			if product == nil {
				return ecode.ErrBadRequest.WithMessage("订单商品不存在，无法取消")
			}
			if !product.UnlockStock(item.Quantity) {
				return ecode.ErrBadRequest.WithMessage("订单锁定库存异常，无法取消")
			}
			product.UpdatedBy = operatorID
			if saveErr := tx.Save(product).Error; saveErr != nil {
				return saveErr
			}
		}

		order.Status = orderEntity.OrderStatusCancelled
		order.UpdatedBy = operatorID
		if saveErr := tx.Save(order).Error; saveErr != nil {
			return saveErr
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func effectiveOrderStatus(order *orderEntity.Order) orderEntity.OrderStatus {
	if order == nil {
		return orderEntity.OrderStatusDraft
	}
	if order.Status != orderEntity.OrderStatusDraft {
		if order.Status == orderEntity.OrderStatusPendingPayment && order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
			return orderEntity.OrderStatusPaid
		}
		return order.Status
	}
	// Draft 分支：当前 Create 方法不会产出 Draft 订单（Status 直接设为 PendingPayment），
	// 此分支仅用于数据修复/迁移场景中可能存在的 Draft 订单，根据已有字段倒推实际状态。
	if order.ReceiptAt != nil || order.ReceiptVoucher != "" {
		return orderEntity.OrderStatusCompleted
	}
	if order.DeliveryInfo != nil && order.DeliveryInfo.PickupCode != "" {
		return orderEntity.OrderStatusPendingDelivery
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		return orderEntity.OrderStatusPaid
	}
	return orderEntity.OrderStatusPendingPayment
}

func isOrderExpired(order *orderEntity.Order) bool {
	if order == nil {
		return false
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
		return false
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		return false
	}
	if order.ExpiresAt == nil || order.ExpiresAt.After(time.Now()) {
		return false
	}
	return true
}

func generatePickupCode() string {
	return fmt.Sprintf("PK%s", time.Now().Format("0102150405"))
}
