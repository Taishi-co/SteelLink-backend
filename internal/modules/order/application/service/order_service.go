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
		BuyerID:    buyerEnterprise.ID,
		BuyerName:  buyerEnterprise.Name,
		SellerID:   sellerEnterprise.ID,
		SellerName: sellerEnterprise.Name,
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

func (s *OrderService) UploadPaymentVoucher(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.UploadPaymentVoucherRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权上传支付凭证")
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
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
	if order.SellerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认收款")
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
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
	if order.Status != orderEntity.OrderStatusPendingDelivery {
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
	if order.Status != orderEntity.OrderStatusPendingDelivery {
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
	if order.Status != orderEntity.OrderStatusDelivering {
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
		Status:         int(order.Status),
		ExpiresAt:      order.ExpiresAt,
		CreatedAt:      order.CreatedAt,
		PaymentInfo:    paymentInfo,
		DeliveryInfo:   deliveryInfo,
		ReceiptVoucher: order.ReceiptVoucher,
		ReceiptAt:      order.ReceiptAt,
	}
}

func generatePickupCode() string {
	return fmt.Sprintf("PK%s", time.Now().Format("0102150405"))
}
