package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	cartEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/entity"
	cartRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/repository"
	cartPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/cart/infrastructure/persistence"
	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	orderRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/repository"
	orderPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/order/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/contract"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

const (
	autoReceiptTimeout        = 7 * 24 * time.Hour
	autoReceiptVoucher        = "AUTO_CONFIRMED_BY_TIMEOUT"
	autoReceiptSource         = "system_timeout"
	autoReceiptReason         = "delivering_timeout"
	autoReceiptTimeoutDays    = 7
	autoReceiptTriggeredAtKey = "receipt_auto_triggered_at"
)

type OrderService struct {
	orderRepo      orderRepository.OrderRepository
	cartRepo       cartRepository.CartRepository
	productRepo    productRepository.ProductRepository
	enterpriseRepo userRepository.EnterpriseRepository
	contract       contract.Provider
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo:      orderPersistence.NewOrderRepository(),
		cartRepo:       cartPersistence.NewCartRepository(),
		productRepo:    productPersistence.NewProductRepository(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
		contract:       contract.NewProvider(),
	}
}

func (s *OrderService) Create(ctx context.Context, req *dto.CreateOrderRequest, userID, buyerEnterpriseID uint64) (*dto.OrderInfo, error) {
	if buyerEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能下单")
	}
	items, cartItemIDs, err := s.resolveCreateItems(ctx, req, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请选择下单商品")
	}

	buyerEnterprise, err := s.enterpriseRepo.FindByID(ctx, buyerEnterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if buyerEnterprise == nil || !buyerEnterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("买家企业不存在或未通过审核")
	}

	products := make([]productOrderItem, 0, len(items))
	var sellerEnterpriseID uint64
	var currency string
	for _, item := range items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, ecode.ErrDatabase
		}
		if product == nil {
			return nil, ecode.ErrNotFound.WithMessage("商品不存在")
		}
		if !product.CanOrder(item.Quantity) {
			return nil, ecode.ErrBadRequest.WithMessage("商品不可下单或库存不足")
		}
		if sellerEnterpriseID == 0 {
			sellerEnterpriseID = product.EnterpriseID
			currency = product.Currency
		}
		if product.EnterpriseID != sellerEnterpriseID {
			return nil, ecode.ErrBadRequest.WithMessage("同一订单暂不支持跨卖家拼单，请选择同一卖家的商品")
		}
		products = append(products, productOrderItem{Product: product, Quantity: item.Quantity})
	}

	sellerEnterprise, err := s.enterpriseRepo.FindByID(ctx, sellerEnterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if sellerEnterprise == nil {
		return nil, ecode.ErrBadRequest.WithMessage("卖家企业不存在")
	}

	orderItems := make(orderEntity.OrderItems, 0, len(products))
	totalQuantity := 0.0
	totalAmount := 0.0
	taxTotalAmount := 0.0
	for _, productItem := range products {
		product := productItem.Product
		itemTotal := productItem.Quantity * product.UnitPrice
		itemTaxAmount := 0.0
		if !product.TaxIncluded {
			itemTaxAmount = itemTotal * (product.TaxRate / 100)
		}
		orderItems = append(orderItems, orderEntity.OrderItem{
			ProductID:   product.ID,
			ProductNo:   product.ProductNo,
			ProductName: product.Name,
			SteelMill:   product.SteelMill,
			Category:    product.SubCategory,
			BundleNo:    product.BundleNo,
			Quantity:    productItem.Quantity,
			UnitPrice:   product.UnitPrice,
			TotalAmount: itemTotal,
			TaxRate:     product.TaxRate,
			TaxAmount:   itemTaxAmount,
		})
		totalQuantity += productItem.Quantity
		totalAmount += itemTotal
		taxTotalAmount += itemTotal + itemTaxAmount
	}

	order := &orderEntity.Order{
		BuyerID:            buyerEnterprise.ID,
		BuyerName:          buyerEnterprise.Name,
		SellerID:           sellerEnterprise.ID,
		SellerEnterpriseID: sellerEnterprise.ID,
		SellerName:         sellerEnterprise.Name,
		Items:              orderItems,
		TotalQuantity:      totalQuantity,
		TotalAmount:        totalAmount,
		TaxTotalAmount:     taxTotalAmount,
		Currency:           currency,
		Status:             orderEntity.OrderStatusPendingConfirm,
		Remark: orderEntity.Remark{
			"source":                 strings.TrimSpace(req.Source),
			"restore_cart_on_cancel": req.RestoreCartOnCancel || len(cartItemIDs) > 0,
			"contact_name":           strings.TrimSpace(req.ContactName),
			"contact_phone":          strings.TrimSpace(req.ContactPhone),
			"warehouse_name":         strings.TrimSpace(req.WarehouseName),
			"delivery_address":       strings.TrimSpace(req.DeliveryAddress),
			"delivery_method":        strings.TrimSpace(req.DeliveryMethod),
			"payment_method":         strings.TrimSpace(req.PaymentMethod),
			"delivery_note":          strings.TrimSpace(req.DeliveryNote),
			"remark":                 strings.TrimSpace(req.Remark),
		},
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, productItem := range products {
			if !productItem.Product.LockStock(productItem.Quantity) {
				return ecode.ErrBadRequest.WithMessage("库存不足")
			}
			productItem.Product.UpdatedBy = userID
			if err := tx.Save(productItem.Product).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, cartItemID := range cartItemIDs {
			if err := tx.Exec("DELETE FROM cart_items WHERE id = ? AND user_id = ?", cartItemID, userID).Error; err != nil {
				return err
			}
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

type productOrderItem struct {
	Product  *productEntity.Product
	Quantity float64
}

func (s *OrderService) resolveCreateItems(ctx context.Context, req *dto.CreateOrderRequest, userID uint64) ([]dto.CreateOrderItemRequest, []uint64, error) {
	if len(req.CartItemIDs) > 0 {
		items := make([]dto.CreateOrderItemRequest, 0, len(req.CartItemIDs))
		for _, cartItemID := range req.CartItemIDs {
			cartItem, err := s.cartRepo.FindByID(ctx, cartItemID)
			if err != nil {
				return nil, nil, ecode.ErrDatabase
			}
			if cartItem == nil {
				return nil, nil, ecode.ErrNotFound.WithMessage("购物车商品不存在")
			}
			if cartItem.UserID != userID {
				return nil, nil, ecode.ErrForbidden.WithMessage("无权结算该购物车商品")
			}
			items = append(items, dto.CreateOrderItemRequest{ProductID: cartItem.ProductID, Quantity: cartItem.Quantity})
		}
		return items, req.CartItemIDs, nil
	}
	if len(req.Items) > 0 {
		return req.Items, nil, nil
	}
	if req.ProductID > 0 && req.Quantity > 0 {
		return []dto.CreateOrderItemRequest{{ProductID: req.ProductID, Quantity: req.Quantity}}, nil, nil
	}
	return nil, nil, nil
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
	if order.Status != orderEntity.OrderStatusPendingPayment &&
		order.Status != orderEntity.OrderStatusPendingConfirm &&
		order.Status != orderEntity.OrderStatusPendingContract {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可取消")
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		return nil, ecode.ErrBadRequest.WithMessage("已上传支付凭证，无法直接取消订单")
	}

	db := database.GetDB()
	restoreCart := shouldRestoreCartOnCancel(order)
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
			if restoreCart {
				if saveErr := restoreOrderItemToCart(tx, ctx, order.CreatedBy, enterpriseID, item.ProductID, item.Quantity); saveErr != nil {
					return saveErr
				}
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

func shouldRestoreCartOnCancel(order *orderEntity.Order) bool {
	if order == nil || order.Remark == nil {
		return false
	}
	value, ok := order.Remark["restore_cart_on_cancel"]
	if !ok {
		return false
	}
	enabled, ok := value.(bool)
	return ok && enabled
}

func restoreOrderItemToCart(tx *gorm.DB, ctx context.Context, userID, enterpriseID, productID uint64, quantity float64) error {
	var item cartEntity.CartItem
	err := tx.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.WithContext(ctx).Create(&cartEntity.CartItem{
			UserID:       userID,
			EnterpriseID: enterpriseID,
			ProductID:    productID,
			Quantity:     quantity,
		}).Error
	}
	item.Quantity += quantity
	item.EnterpriseID = enterpriseID
	return tx.WithContext(ctx).Save(&item).Error
}

func (s *OrderService) ConfirmOrder(ctx context.Context, id, sellerEnterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.SellerID != sellerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单")
	}
	if order.Status != orderEntity.OrderStatusPendingConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态无需确认")
	}
	if err := s.ensureContract(ctx, order); err != nil {
		return nil, err
	}
	order.Status = orderEntity.OrderStatusPendingContract
	order.UpdatedBy = operatorID
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认订单失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) SignContract(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && order.SellerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权签署该订单合同")
	}
	if order.Status != orderEntity.OrderStatusPendingContract {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可签署合同")
	}
	if err := s.ensureContract(ctx, order); err != nil {
		return nil, err
	}
	signer := "seller"
	if order.BuyerID == enterpriseID {
		signer = "buyer"
		order.ContractInfo.BuyerSigned = true
	} else {
		order.ContractInfo.SellerSigned = true
	}
	signResult, err := s.contract.Sign(ctx, contract.SignRequest{
		ContractNo: order.ContractInfo.ContractNo,
		Signer:     signer,
	})
	if err != nil {
		return nil, ecode.ErrBadRequest.WithMessage("签署合同失败: " + err.Error())
	}
	order.UpdatedBy = operatorID
	if order.ContractInfo.BuyerSigned && order.ContractInfo.SellerSigned {
		order.ContractInfo.SignedAt = signResult.SignedAt
		order.Status = orderEntity.OrderStatusPendingPayment
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("签署合同失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) AttachContract(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.AttachContractRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && order.SellerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权绑定该订单合同")
	}
	if order.Status != orderEntity.OrderStatusPendingContract {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可绑定合同")
	}
	if !isHTTPURL(req.ContractURL) {
		return nil, ecode.ErrBadRequest.WithMessage("合同URL必须是有效的HTTP/HTTPS地址")
	}

	contractNo := req.ContractNo
	if contractNo == "" && order.ContractInfo != nil {
		contractNo = order.ContractInfo.ContractNo
	}
	if contractNo == "" {
		contractNo = fmt.Sprintf("HT%s%s", time.Now().Format("20060102"), order.OrderNo)
	}

	info := &orderEntity.ContractInfo{
		ContractNo:   contractNo,
		ContractURL:  req.ContractURL,
		ContractHash: req.ContractHash,
	}
	if order.ContractInfo != nil {
		info.BuyerSigned = order.ContractInfo.BuyerSigned
		info.SellerSigned = order.ContractInfo.SellerSigned
		info.SignedAt = order.ContractInfo.SignedAt
	}
	order.ContractInfo = info
	order.UpdatedBy = operatorID

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("绑定合同失败")
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
	if order.Status != orderEntity.OrderStatusPendingPayment {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可上传支付凭证")
	}
	if !isHTTPURL(req.VoucherURL) {
		return nil, ecode.ErrBadRequest.WithMessage("支付凭证地址必须是有效的 HTTP/HTTPS URL")
	}

	now := time.Now()

	order.PaymentInfo = &orderEntity.PaymentInfo{
		PaymentMethod: "bank_transfer",
		PaymentTime:   now,
		TransactionNo: req.TransactionNo,
		PayeeAccount:  req.PayeeAccount,
		PayerAccount:  req.PayerAccount,
		VoucherURL:    req.VoucherURL,
		UploadedBy:    operatorID,
		UploadedAt:    now,
		CapturedAt:    req.CapturedAt,
		GPSLatitude:   req.GPSLatitude,
		GPSLongitude:  req.GPSLongitude,
	}
	order.Status = orderEntity.OrderStatusPaid
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
	if err := checkOrderFrozenOrException(order); err != nil {
		return nil, err
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
	currentStatus := order.Status
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
	if err := checkOrderFrozenOrException(order); err != nil {
		return nil, err
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

func (s *OrderService) ConfirmOutbound(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.VerifyPickupCodeRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if err := checkOrderFrozenOrException(order); err != nil {
		return nil, err
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
	if req != nil && strings.TrimSpace(req.PickupCode) != "" && strings.TrimSpace(req.PickupCode) != order.DeliveryInfo.PickupCode {
		return nil, ecode.ErrBadRequest.WithMessage("提货码不正确")
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
	if err := checkOrderFrozenOrException(order); err != nil {
		return nil, err
	}
	if order.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认收货")
	}
	if order.Status != orderEntity.OrderStatusDelivering {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认收货")
	}
	if !isHTTPURL(req.ReceiptVoucher) {
		return nil, ecode.ErrBadRequest.WithMessage("签收凭证地址必须是有效的 HTTP/HTTPS URL")
	}

	now := time.Now()
	order.ReceiptVoucher = req.ReceiptVoucher
	order.ReceiptInfo = &orderEntity.VoucherEvidence{
		VoucherURL:   req.ReceiptVoucher,
		UploadedBy:   operatorID,
		UploadedAt:   now,
		CapturedAt:   req.CapturedAt,
		GPSLatitude:  req.GPSLatitude,
		GPSLongitude: req.GPSLongitude,
	}
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
			UploadedBy:    order.PaymentInfo.UploadedBy,
			UploadedAt:    order.PaymentInfo.UploadedAt,
			CapturedAt:    order.PaymentInfo.CapturedAt,
			GPSLatitude:   order.PaymentInfo.GPSLatitude,
			GPSLongitude:  order.PaymentInfo.GPSLongitude,
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
	var contractInfo *dto.ContractInfo
	if order.ContractInfo != nil {
		var signedAt *time.Time
		if !order.ContractInfo.SignedAt.IsZero() {
			signedAt = &order.ContractInfo.SignedAt
		}
		contractInfo = &dto.ContractInfo{
			ContractNo:   order.ContractInfo.ContractNo,
			ContractURL:  order.ContractInfo.ContractURL,
			ContractHash: order.ContractInfo.ContractHash,
			SignStatus:   contractSignStatus(order.ContractInfo),
			BuyerSigned:  order.ContractInfo.BuyerSigned,
			SellerSigned: order.ContractInfo.SellerSigned,
			SignedAt:     signedAt,
		}
	}
	var receiptInfo *dto.VoucherEvidence
	if order.ReceiptInfo != nil {
		receiptInfo = &dto.VoucherEvidence{
			VoucherURL:   order.ReceiptInfo.VoucherURL,
			UploadedBy:   order.ReceiptInfo.UploadedBy,
			UploadedAt:   order.ReceiptInfo.UploadedAt,
			CapturedAt:   order.ReceiptInfo.CapturedAt,
			GPSLatitude:  order.ReceiptInfo.GPSLatitude,
			GPSLongitude: order.ReceiptInfo.GPSLongitude,
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
		Expired:        isOrderExpired(order),
		ExpiresAt:      order.ExpiresAt,
		CreatedAt:      order.CreatedAt,
		ContractInfo:   contractInfo,
		PaymentInfo:    paymentInfo,
		DeliveryInfo:   deliveryInfo,
		ReceiptVoucher: order.ReceiptVoucher,
		ReceiptInfo:    receiptInfo,
		ReceiptAt:      order.ReceiptAt,
		Timeline:       buildOrderTimeline(order),
	}
}

func buildOrderTimeline(order *orderEntity.Order) []dto.TimelineItem {
	if order == nil {
		return nil
	}

	status := order.Status
	items := []dto.TimelineItem{
		{
			Key:         "created",
			Label:       "订单创建",
			Status:      "completed",
			Time:        &order.CreatedAt,
			Description: "买家已提交订单，系统锁定对应库存。",
		},
	}

	if status >= orderEntity.OrderStatusPendingContract {
		items = append(items, dto.TimelineItem{
			Key:         "order_confirmed",
			Label:       "卖家确认订单",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "卖家已确认订单，等待双方签署合同。",
		})
	} else if status == orderEntity.OrderStatusPendingConfirm {
		items = append(items, dto.TimelineItem{
			Key:         "order_confirmed",
			Label:       "卖家确认订单",
			Status:      "current",
			Description: "等待卖家确认订单。",
		})
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "order_confirmed",
			Label:       "卖家确认订单",
			Status:      "skipped",
			Description: "订单已关闭，卖家未确认。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "order_confirmed",
			Label:       "卖家确认订单",
			Status:      "pending",
			Description: "等待卖家确认订单。",
		})
	}

	if status >= orderEntity.OrderStatusPendingPayment {
		items = append(items, dto.TimelineItem{
			Key:         "contract_signed",
			Label:       "双方签署合同",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "买卖双方已签署合同，等待买家支付。",
		})
	} else if status == orderEntity.OrderStatusPendingContract {
		buyerSigned := order.ContractInfo != nil && order.ContractInfo.BuyerSigned
		sellerSigned := order.ContractInfo != nil && order.ContractInfo.SellerSigned
		if buyerSigned || sellerSigned {
			items = append(items, dto.TimelineItem{
				Key:    "contract_signed",
				Label:  "双方签署合同",
				Status: "current",
				Description: fmt.Sprintf("%s已签署，等待%s签署。", func() string {
					if buyerSigned {
						return "买家"
					}
					return "卖家"
				}(), func() string {
					if !buyerSigned {
						return "买家"
					}
					return "卖家"
				}()),
			})
		} else {
			items = append(items, dto.TimelineItem{
				Key:         "contract_signed",
				Label:       "双方签署合同",
				Status:      "current",
				Description: "等待买卖双方签署合同。",
			})
		}
	} else if status == orderEntity.OrderStatusCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "contract_signed",
			Label:       "双方签署合同",
			Status:      "skipped",
			Description: "订单已关闭，合同未签署。",
		})
	} else {
		items = append(items, dto.TimelineItem{
			Key:         "contract_signed",
			Label:       "双方签署合同",
			Status:      "pending",
			Description: "卖家确认订单后，双方签署合同。",
		})
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

	if order.Status == orderEntity.OrderStatusPendingPayment {
		if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
			return nil
		}
		if order.ExpiresAt == nil || order.ExpiresAt.After(time.Now()) {
			return nil
		}
		return s.cancelOrderWithUnlock(ctx, order, order.UpdatedBy)
	}

	if order.Status == orderEntity.OrderStatusDelivering {
		if order.UpdatedAt.Before(time.Now().Add(-autoReceiptTimeout)) {
			now := time.Now()
			order.ReceiptVoucher = autoReceiptVoucher
			order.ReceiptAt = &now
			order.Status = orderEntity.OrderStatusCompleted
			order.UpdatedBy = 0
			if order.Remark == nil {
				order.Remark = orderEntity.Remark{}
			}
			order.Remark["receipt_confirm_source"] = autoReceiptSource
			order.Remark["receipt_confirm_reason"] = autoReceiptReason
			order.Remark["receipt_timeout_days"] = autoReceiptTimeoutDays
			order.Remark[autoReceiptTriggeredAtKey] = now.Format(time.RFC3339)
			if err := s.orderRepo.Update(ctx, order); err != nil {
				return ecode.ErrDatabase.WithMessage("签收超时自动确认失败")
			}
		}
	}

	return nil
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

func isOrderExpired(order *orderEntity.Order) bool {
	if order == nil {
		return false
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
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

func (s *OrderService) ensureContract(ctx context.Context, order *orderEntity.Order) error {
	if order.ContractInfo != nil && order.ContractInfo.ContractNo != "" && order.ContractInfo.ContractURL != "" {
		return nil
	}

	doc, err := s.contract.Create(ctx, contract.OrderSnapshot{
		OrderNo:     order.OrderNo,
		BuyerName:   order.BuyerName,
		SellerName:  order.SellerName,
		Amount:      order.TotalAmount,
		Currency:    order.Currency,
		ConfirmedAt: time.Now(),
	})
	if err != nil {
		return ecode.ErrBadRequest.WithMessage("生成合同失败: " + err.Error())
	}

	info := &orderEntity.ContractInfo{
		ContractNo:   doc.ContractNo,
		ContractURL:  doc.ContractURL,
		ContractHash: doc.ContractHash,
	}
	if order.ContractInfo != nil {
		info.BuyerSigned = order.ContractInfo.BuyerSigned
		info.SellerSigned = order.ContractInfo.SellerSigned
		info.SignedAt = order.ContractInfo.SignedAt
	}
	order.ContractInfo = info
	return nil
}

func contractSignStatus(contract *orderEntity.ContractInfo) string {
	if contract == nil {
		return "pending"
	}
	if contract.BuyerSigned && contract.SellerSigned {
		return "signed"
	}
	if contract.BuyerSigned || contract.SellerSigned {
		return "partial"
	}
	return "pending"
}

func isHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func checkOrderFrozenOrException(order *orderEntity.Order) error {
	if order == nil {
		return nil
	}
	if order.Status == orderEntity.OrderStatusFrozen {
		return ecode.ErrBadRequest.WithMessage("订单已冻结，不可继续履约操作")
	}
	if order.Status == orderEntity.OrderStatusException {
		return ecode.ErrBadRequest.WithMessage("订单异常，不可继续履约操作")
	}
	if order.Status == orderEntity.OrderStatusCancelled {
		return ecode.ErrBadRequest.WithMessage("订单已取消，不可继续履约操作")
	}
	return nil
}
