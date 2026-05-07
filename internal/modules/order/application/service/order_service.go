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
	offerEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
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
	if !isSingleWarehouseOrder(products) {
		return nil, ecode.ErrBadRequest.WithMessage("同一订单暂不支持跨仓库商品，请按仓库分别下单")
	}
	initialDeliveryInfo := buildInitialDeliveryInfo(products[0].Product)
	if initialDeliveryInfo == nil || initialDeliveryInfo.WarehouseEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("商品未绑定仓储方，无法进入仓库确认")
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
		Status:             orderEntity.OrderStatusPendingWarehouseConfirm,
		DeliveryInfo:       initialDeliveryInfo,
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
		if err := createOrderEvent(tx, ctx, order, "order_created", userID, buyerEnterprise.ID, "procurement", orderEntity.OrderStatusDraft, order.Status, orderEntity.OrderEventPayload{
			"source": strings.TrimSpace(req.Source),
		}); err != nil {
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

func isSingleWarehouseOrder(items []productOrderItem) bool {
	if len(items) <= 1 {
		return true
	}
	base := warehouseSignature(items[0].Product)
	for _, item := range items[1:] {
		if warehouseSignature(item.Product) != base {
			return false
		}
	}
	return true
}

func warehouseSignature(product *productEntity.Product) string {
	if product == nil {
		return ""
	}
	if product.WarehouseServiceID > 0 {
		return fmt.Sprintf("service:%d", product.WarehouseServiceID)
	}
	if product.WarehouseInfo == nil {
		return ""
	}
	if product.WarehouseInfo.WarehouseID > 0 {
		return fmt.Sprintf("warehouse:%d", product.WarehouseInfo.WarehouseID)
	}
	return strings.Join([]string{
		product.WarehouseInfo.WarehouseName,
		product.WarehouseInfo.WarehouseCode,
		product.WarehouseInfo.Province,
		product.WarehouseInfo.City,
		product.WarehouseInfo.District,
		product.WarehouseInfo.Address,
	}, "|")
}

func buildInitialDeliveryInfo(product *productEntity.Product) *orderEntity.DeliveryInfo {
	if product == nil || product.WarehouseInfo == nil {
		return nil
	}
	warehouseServiceID := product.WarehouseServiceID
	if warehouseServiceID == 0 {
		warehouseServiceID = product.WarehouseInfo.WarehouseServiceID
	}
	warehouseEnterpriseID := product.WarehouseEnterpriseID
	if warehouseEnterpriseID == 0 {
		warehouseEnterpriseID = product.WarehouseInfo.WarehouseEnterpriseID
	}
	return &orderEntity.DeliveryInfo{
		WarehouseID:           product.WarehouseInfo.WarehouseID,
		WarehouseServiceID:    warehouseServiceID,
		WarehouseEnterpriseID: warehouseEnterpriseID,
		WarehouseName:         product.WarehouseInfo.WarehouseName,
		WarehouseCode:         product.WarehouseInfo.WarehouseCode,
		WarehouseProvince:     product.WarehouseInfo.Province,
		WarehouseCity:         product.WarehouseInfo.City,
		WarehouseDistrict:     product.WarehouseInfo.District,
		WarehouseAddress:      product.WarehouseInfo.Address,
	}
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
	if enterpriseID == 0 || (order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID)) {
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
		order.Status != orderEntity.OrderStatusPendingWarehouseConfirm &&
		order.Status != orderEntity.OrderStatusPendingSellerConfirm &&
		order.Status != orderEntity.OrderStatusPendingBuyerConfirm &&
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

		fromStatus := order.Status
		order.Status = orderEntity.OrderStatusCancelled
		order.UpdatedBy = operatorID
		if saveErr := tx.Save(order).Error; saveErr != nil {
			return saveErr
		}
		if saveErr := createOrderEvent(tx, ctx, order, "order_cancelled", operatorID, enterpriseID, "procurement", fromStatus, order.Status, nil); saveErr != nil {
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
	return s.SellerConfirm(ctx, id, sellerEnterpriseID, operatorID, &dto.SellerConfirmRequest{})
}

func (s *OrderService) WarehouseConfirm(ctx context.Context, id, warehouseEnterpriseID, operatorID uint64, req *dto.WarehouseConfirmRequest) (*dto.OrderInfo, error) {
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
	if order.DeliveryInfo == nil || order.DeliveryInfo.WarehouseEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("订单未绑定仓储方，无法仓库确认")
	}
	if order.DeliveryInfo.WarehouseEnterpriseID != warehouseEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单仓库条件")
	}
	if order.Status != orderEntity.OrderStatusPendingWarehouseConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可仓库确认")
	}
	if req != nil && !req.Confirmed {
		return nil, ecode.ErrBadRequest.WithMessage("仓库条件未确认")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingSellerConfirm
	order.UpdatedBy = operatorID
	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		payload := orderEntity.OrderEventPayload{}
		if req != nil {
			payload["condition_note"] = strings.TrimSpace(req.ConditionNote)
			payload["voucher_url"] = strings.TrimSpace(req.VoucherURL)
		}
		return createOrderEvent(tx, ctx, order, "warehouse_confirmed", operatorID, warehouseEnterpriseID, "warehouse", fromStatus, order.Status, payload)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("仓库确认失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) SellerConfirm(ctx context.Context, id, sellerEnterpriseID, operatorID uint64, req *dto.SellerConfirmRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, sellerEnterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单")
	}
	if order.Status != orderEntity.OrderStatusPendingSellerConfirm && order.Status != orderEntity.OrderStatusPendingConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态无需卖方确认")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingBuyerConfirm
	order.UpdatedBy = operatorID
	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		payload := orderEntity.OrderEventPayload{}
		if req != nil {
			payload["terms_note"] = strings.TrimSpace(req.TermsNote)
		}
		return createOrderEvent(tx, ctx, order, "seller_confirmed", operatorID, sellerEnterpriseID, "sales", fromStatus, order.Status, payload)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("卖方确认订单失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) BuyerConfirm(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.BuyerConfirmRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权最终确认该订单")
	}
	if order.Status != orderEntity.OrderStatusPendingBuyerConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可买方最终确认")
	}
	if err := s.ensureContract(ctx, order); err != nil {
		return nil, err
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingContract
	order.UpdatedBy = operatorID
	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		payload := orderEntity.OrderEventPayload{}
		if req != nil {
			payload["confirm_note"] = strings.TrimSpace(req.ConfirmNote)
		}
		return createOrderEvent(tx, ctx, order, "buyer_confirmed", operatorID, buyerEnterpriseID, "procurement", fromStatus, order.Status, payload)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("买方最终确认失败")
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
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) {
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
	fromStatus := orderEntity.OrderStatusPendingContract
	toStatus := order.Status
	db := database.GetDB()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		eventType := "contract_signed"
		if toStatus == orderEntity.OrderStatusPendingPayment {
			eventType = "contract_completed"
		}
		return createOrderEvent(tx, ctx, order, eventType, operatorID, enterpriseID, actorServiceForOrder(order, enterpriseID), fromStatus, toStatus, orderEntity.OrderEventPayload{
			"signer": signer,
		})
	}); err != nil {
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
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) {
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

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "direct_payment_confirmed", operatorID, enterpriseID, "procurement", orderEntity.OrderStatusPendingPayment, order.Status, orderEntity.OrderEventPayload{
			"transaction_no": strings.TrimSpace(req.TransactionNo),
			"voucher_url":    strings.TrimSpace(req.VoucherURL),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("上传支付凭证失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) DirectPayment(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.DirectPaymentRequest) (*dto.OrderInfo, error) {
	return s.UploadPaymentVoucher(ctx, id, enterpriseID, operatorID, &dto.UploadPaymentVoucherRequest{
		VoucherURL:    req.VoucherURL,
		TransactionNo: req.TransactionNo,
		PayerAccount:  req.PayerAccount,
		PayeeAccount:  req.PayeeAccount,
		CapturedAt:    req.CapturedAt,
		GPSLatitude:   req.GPSLatitude,
		GPSLongitude:  req.GPSLongitude,
	})
}

func (s *OrderService) CreateFinancingRequest(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.CreateFinancingRequest) (*dto.FinancingRequestInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权申请该订单托盘融资")
	}
	if order.Status != orderEntity.OrderStatusPendingPayment {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可申请托盘融资")
	}
	requestedAmount := req.RequestedAmount
	if requestedAmount <= 0 {
		requestedAmount = order.TaxTotalAmount
		if requestedAmount <= 0 {
			requestedAmount = order.TotalAmount
		}
	}
	db := database.GetDB()
	var financeService offerEntity.FinanceService
	if err := db.WithContext(ctx).Where("id = ? AND status = ?", req.FinanceServiceID, offerEntity.OfferStatusApproved).First(&financeService).Error; err != nil {
		return nil, ecode.ErrBadRequest.WithMessage("金融服务产品不存在或未审核通过")
	}
	if financeService.MinAmount > 0 && requestedAmount < financeService.MinAmount {
		return nil, ecode.ErrBadRequest.WithMessage("申请金额低于金融服务产品最低额度")
	}
	if financeService.MaxAmount > 0 && requestedAmount > financeService.MaxAmount {
		return nil, ecode.ErrBadRequest.WithMessage("申请金额超过金融服务产品最高额度")
	}
	financing := &orderEntity.OrderFinancingRequest{
		OrderID:             order.ID,
		OrderNo:             order.OrderNo,
		BuyerEnterpriseID:   order.BuyerID,
		SellerEnterpriseID:  order.SellerID,
		FinanceServiceID:    financeService.ID,
		FundingEnterpriseID: financeService.EnterpriseID,
		RequestedAmount:     requestedAmount,
		TermDays:            financeService.TermDays,
		AnnualRate:          financeService.AnnualRate,
		Status:              orderEntity.FinancingRequestStatusSubmitted,
		CreatedBy:           operatorID,
		UpdatedBy:           operatorID,
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusFinancingReview
	order.UpdatedBy = operatorID
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(financing).Error; err != nil {
			return err
		}
		order.FinanceID = financing.ID
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_requested", operatorID, buyerEnterpriseID, "procurement", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"finance_service_id":    financing.FinanceServiceID,
			"requested_amount":      financing.RequestedAmount,
			"funding_enterprise_id": financing.FundingEnterpriseID,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建融资申请失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) SubmitRiskResult(ctx context.Context, id, fundingEnterpriseID, operatorID uint64, req *dto.RiskResultRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if financing.FundingEnterpriseID != fundingEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权回传该订单融资风控结果")
	}
	if financing.Status != orderEntity.FinancingRequestStatusSubmitted {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可回传风控结果")
	}
	fromStatus := order.Status
	financing.RiskRemark = strings.TrimSpace(req.RiskRemark)
	financing.UpdatedBy = operatorID
	if !req.Approved {
		financing.Status = orderEntity.FinancingRequestStatusRiskRejected
		financing.RiskResult = "rejected"
		order.Status = orderEntity.OrderStatusPendingPayment
		order.UpdatedBy = operatorID
		if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(financing).Error; err != nil {
				return err
			}
			if err := tx.Save(order).Error; err != nil {
				return err
			}
			return createOrderEvent(tx, ctx, order, "financing_risk_rejected", operatorID, fundingEnterpriseID, "finance", fromStatus, order.Status, orderEntity.OrderEventPayload{
				"risk_remark": financing.RiskRemark,
			})
		}); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("回传风控结果失败")
		}
		return toFinancingRequestInfo(financing), nil
	}
	if req.ApprovedAmount <= 0 {
		return nil, ecode.ErrBadRequest.WithMessage("授信金额必须大于0")
	}
	financing.Status = orderEntity.FinancingRequestStatusRiskApproved
	financing.RiskResult = "approved"
	financing.ApprovedAmount = req.ApprovedAmount
	financing.TermDays = req.TermDays
	financing.AnnualRate = req.AnnualRate
	financing.ServiceFee = req.ServiceFee
	if financing.TermDays <= 0 {
		financing.TermDays = 30
	}
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_risk_approved", operatorID, fundingEnterpriseID, "finance", order.Status, order.Status, orderEntity.OrderEventPayload{
			"approved_amount": financing.ApprovedAmount,
			"term_days":       financing.TermDays,
			"annual_rate":     financing.AnnualRate,
			"service_fee":     financing.ServiceFee,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("回传风控结果失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) BuyerConfirmFinancing(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.BuyerFinancingConfirmRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单融资方案")
	}
	if !req.Confirmed {
		return nil, ecode.ErrBadRequest.WithMessage("买方未确认融资金额和定价")
	}
	if financing.Status != orderEntity.FinancingRequestStatusRiskApproved {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可买方确认")
	}
	now := time.Now()
	fromStatus := order.Status
	financing.Status = orderEntity.FinancingRequestStatusBuyerConfirmed
	financing.BuyerConfirmedAt = &now
	financing.UpdatedBy = operatorID
	order.Status = orderEntity.OrderStatusPendingFinancingContract
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_buyer_confirmed", operatorID, buyerEnterpriseID, "procurement", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"remark": strings.TrimSpace(req.Remark),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认融资方案失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) GenerateDebtContract(ctx context.Context, id, fundingEnterpriseID, operatorID uint64, req *dto.DebtContractRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if financing.FundingEnterpriseID != fundingEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权生成该订单债务合同")
	}
	if financing.Status != orderEntity.FinancingRequestStatusBuyerConfirmed {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可生成债务合同")
	}
	if !isHTTPURL(req.DebtContractURL) {
		return nil, ecode.ErrBadRequest.WithMessage("债务合同URL必须是有效的HTTP/HTTPS地址")
	}
	contractNo := strings.TrimSpace(req.DebtContractNo)
	if contractNo == "" {
		contractNo = fmt.Sprintf("DEBT%s%s", time.Now().Format("20060102"), order.OrderNo)
	}
	financing.Status = orderEntity.FinancingRequestStatusContractGenerated
	financing.DebtContractNo = contractNo
	financing.DebtContractURL = strings.TrimSpace(req.DebtContractURL)
	financing.DebtContractHash = strings.TrimSpace(req.DebtContractHash)
	financing.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_debt_contract_generated", operatorID, fundingEnterpriseID, "finance", order.Status, order.Status, orderEntity.OrderEventPayload{
			"debt_contract_no":  financing.DebtContractNo,
			"debt_contract_url": financing.DebtContractURL,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("生成债务合同失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) BuyerSignDebtAndPayFee(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.BuyerDebtSignRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权签署该订单债务合同")
	}
	if financing.Status != orderEntity.FinancingRequestStatusContractGenerated {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可签署债务合同")
	}
	if !isHTTPURL(req.FeeVoucherURL) {
		return nil, ecode.ErrBadRequest.WithMessage("手续费凭证URL必须是有效的HTTP/HTTPS地址")
	}
	now := time.Now()
	financing.Status = orderEntity.FinancingRequestStatusBuyerSignedFeePaid
	financing.BuyerSignedAt = &now
	financing.FeeVoucherURL = strings.TrimSpace(req.FeeVoucherURL)
	financing.FeePaidAt = &now
	financing.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_buyer_signed_fee_paid", operatorID, buyerEnterpriseID, "procurement", order.Status, order.Status, orderEntity.OrderEventPayload{
			"fee_voucher_url": financing.FeeVoucherURL,
			"remark":          strings.TrimSpace(req.Remark),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("签署债务合同失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) DisburseFinancing(ctx context.Context, id, fundingEnterpriseID, operatorID uint64, req *dto.FinancingDisburseRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if financing.FundingEnterpriseID != fundingEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权放款该订单融资")
	}
	if financing.Status != orderEntity.FinancingRequestStatusBuyerSignedFeePaid {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可放款")
	}
	if !isHTTPURL(req.DisbursementVoucher) {
		return nil, ecode.ErrBadRequest.WithMessage("放款凭证URL必须是有效的HTTP/HTTPS地址")
	}
	now := time.Now()
	fromStatus := order.Status
	financing.Status = orderEntity.FinancingRequestStatusDisbursed
	financing.DisbursementVoucher = strings.TrimSpace(req.DisbursementVoucher)
	financing.DisbursedAt = &now
	financing.UpdatedBy = operatorID
	order.Status = orderEntity.OrderStatusPaid
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_disbursed", operatorID, fundingEnterpriseID, "finance", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"disbursement_voucher": financing.DisbursementVoucher,
			"remark":               strings.TrimSpace(req.Remark),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("融资放款失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) RecordPledge(ctx context.Context, id, fundingEnterpriseID, operatorID uint64, req *dto.PledgeRecordRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if financing.FundingEnterpriseID != fundingEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权记录该订单货权质押")
	}
	if financing.Status != orderEntity.FinancingRequestStatusDisbursed {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可记录货权质押")
	}
	if !isHTTPURL(req.PledgeVoucherURL) {
		return nil, ecode.ErrBadRequest.WithMessage("货权质押凭证URL必须是有效的HTTP/HTTPS地址")
	}
	now := time.Now()
	financing.Status = orderEntity.FinancingRequestStatusPledged
	financing.PledgeVoucherURL = strings.TrimSpace(req.PledgeVoucherURL)
	financing.PledgeRemark = strings.TrimSpace(req.PledgeRemark)
	financing.ReleaseCondition = strings.TrimSpace(req.ReleaseCondition)
	financing.PledgedAt = &now
	financing.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "ownership_pledge_recorded", operatorID, fundingEnterpriseID, "finance", order.Status, order.Status, orderEntity.OrderEventPayload{
			"pledge_voucher_url": financing.PledgeVoucherURL,
			"release_condition":  financing.ReleaseCondition,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("记录货权质押失败")
	}
	return toFinancingRequestInfo(financing), nil
}

func (s *OrderService) ReleasePledge(ctx context.Context, id, fundingEnterpriseID, operatorID uint64, req *dto.PledgeReleaseRequest) (*dto.FinancingRequestInfo, error) {
	order, financing, err := s.getOrderFinancing(ctx, id)
	if err != nil {
		return nil, err
	}
	if financing.FundingEnterpriseID != fundingEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权释放该订单货权质押")
	}
	if financing.Status != orderEntity.FinancingRequestStatusPledged {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可释放货权质押")
	}
	now := time.Now()
	financing.Status = orderEntity.FinancingRequestStatusReleased
	financing.ReleaseReason = strings.TrimSpace(req.Reason)
	financing.ReleasedAt = &now
	financing.UpdatedBy = operatorID
	releaseEvent := &orderEntity.PledgeReleaseEvent{
		FinancingRequestID: financing.ID,
		OrderID:            order.ID,
		OrderNo:            order.OrderNo,
		EventType:          "pledge_released",
		Condition:          financing.ReleaseCondition,
		Reason:             financing.ReleaseReason,
		VoucherURL:         strings.TrimSpace(req.VoucherURL),
		ActorUserID:        operatorID,
		ActorEnterpriseID:  fundingEnterpriseID,
	}
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		if err := tx.Create(releaseEvent).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "ownership_pledge_released", operatorID, fundingEnterpriseID, "finance", order.Status, order.Status, orderEntity.OrderEventPayload{
			"release_condition": financing.ReleaseCondition,
			"release_reason":    financing.ReleaseReason,
			"voucher_url":       releaseEvent.VoucherURL,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("释放货权质押失败")
	}
	return toFinancingRequestInfo(financing), nil
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
	if !isOrderSeller(order, enterpriseID) {
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
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "seller_payment_confirmed", operatorID, enterpriseID, "sales", currentStatus, order.Status, nil)
	}); err != nil {
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

	deliveryInfo := order.DeliveryInfo
	if deliveryInfo == nil {
		deliveryInfo = &orderEntity.DeliveryInfo{
			WarehouseID:        req.WarehouseID,
			WarehouseServiceID: req.WarehouseServiceID,
			WarehouseName:      req.WarehouseName,
		}
	}
	if deliveryInfo.WarehouseID == 0 {
		deliveryInfo.WarehouseID = req.WarehouseID
	}
	if deliveryInfo.WarehouseServiceID == 0 {
		deliveryInfo.WarehouseServiceID = req.WarehouseServiceID
	}
	if deliveryInfo.WarehouseName == "" {
		deliveryInfo.WarehouseName = req.WarehouseName
	}

	order.DeliveryInfo = &orderEntity.DeliveryInfo{
		DriverName:            req.DriverName,
		DriverIdCard:          req.DriverIDCard,
		DriverPhone:           req.DriverPhone,
		PlateNumber:           req.PlateNumber,
		PickupCode:            pickupCode,
		WarehouseID:           deliveryInfo.WarehouseID,
		WarehouseServiceID:    deliveryInfo.WarehouseServiceID,
		WarehouseEnterpriseID: deliveryInfo.WarehouseEnterpriseID,
		WarehouseName:         deliveryInfo.WarehouseName,
		WarehouseCode:         deliveryInfo.WarehouseCode,
		WarehouseProvince:     deliveryInfo.WarehouseProvince,
		WarehouseCity:         deliveryInfo.WarehouseCity,
		WarehouseDistrict:     deliveryInfo.WarehouseDistrict,
		WarehouseAddress:      deliveryInfo.WarehouseAddress,
		Address:               req.Address,
		EstimatedTime:         req.EstimatedTime,
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
	if !isOrderSeller(order, enterpriseID) {
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
	if order.Status != orderEntity.OrderStatusDelivering && order.Status != orderEntity.OrderStatusShipping && order.Status != orderEntity.OrderStatusDelivered {
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
	order.Status = orderEntity.OrderStatusReceived
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "buyer_received", operatorID, enterpriseID, "procurement", orderEntity.OrderStatusDelivering, order.Status, orderEntity.OrderEventPayload{
			"receipt_voucher": strings.TrimSpace(req.ReceiptVoucher),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认收货失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) CreateLogisticsBidding(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.CreateLogisticsBiddingRequest) (*dto.LogisticsBiddingInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权发起该订单物流竞价")
	}
	if order.Status != orderEntity.OrderStatusPaid && order.Status != orderEntity.OrderStatusPendingDelivery {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可发起物流竞价")
	}
	bidding := &orderEntity.LogisticsBidding{
		OrderID:               order.ID,
		OrderNo:               order.OrderNo,
		SellerEnterpriseID:    enterpriseID,
		WarehouseEnterpriseID: warehouseEnterpriseIDOf(order),
		PickupAddress:         strings.TrimSpace(req.PickupAddress),
		DestinationAddress:    strings.TrimSpace(req.DestinationAddress),
		Requirement:           strings.TrimSpace(req.Requirement),
		Status:                orderEntity.LogisticsBiddingStatusOpen,
		CreatedBy:             operatorID,
	}
	if bidding.PickupAddress == "" && order.DeliveryInfo != nil {
		bidding.PickupAddress = order.DeliveryInfo.WarehouseAddress
	}
	if bidding.DestinationAddress == "" && order.DeliveryInfo != nil {
		bidding.DestinationAddress = order.DeliveryInfo.Address
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusLogisticsBidding
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bidding).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "logistics_bidding_created", operatorID, enterpriseID, "sales", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"bidding_id": bidding.ID,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("发起物流竞价失败")
	}
	return toLogisticsBiddingInfo(bidding), nil
}

func (s *OrderService) CreateLogisticsQuote(ctx context.Context, orderID, enterpriseID, operatorID uint64, req *dto.CreateLogisticsQuoteRequest) (*dto.LogisticsQuoteInfo, error) {
	var order orderEntity.Order
	db := database.GetDB()
	if err := db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusLogisticsBidding {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可物流报价")
	}
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ? AND status = ?", order.ID, orderEntity.LogisticsBiddingStatusOpen).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	logisticsName := ""
	if enterprise, err := s.enterpriseRepo.FindByID(ctx, enterpriseID); err == nil && enterprise != nil {
		logisticsName = enterprise.Name
	}
	quote := &orderEntity.LogisticsBidQuote{
		BiddingID:             bidding.ID,
		OrderID:               order.ID,
		LogisticsEnterpriseID: enterpriseID,
		LogisticsName:         logisticsName,
		VehicleInfo:           strings.TrimSpace(req.VehicleInfo),
		DriverInfo:            strings.TrimSpace(req.DriverInfo),
		Amount:                req.Amount,
		EstimatedPickupAt:     strings.TrimSpace(req.EstimatedPickupAt),
		EstimatedDeliveryAt:   strings.TrimSpace(req.EstimatedDeliveryAt),
		Remark:                strings.TrimSpace(req.Remark),
		CreatedBy:             operatorID,
	}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(quote).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, &order, "logistics_quote_submitted", operatorID, enterpriseID, "logistics", order.Status, order.Status, orderEntity.OrderEventPayload{
			"bidding_id": quote.BiddingID,
			"quote_id":   quote.ID,
			"amount":     quote.Amount,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("物流报价失败")
	}
	return toLogisticsQuoteInfo(quote), nil
}

func (s *OrderService) ConfirmLogisticsQuote(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.ConfirmLogisticsRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单物流资源")
	}
	if order.Status != orderEntity.OrderStatusLogisticsBidding && order.Status != orderEntity.OrderStatusPendingLogisticsConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认物流资源")
	}
	var bidding orderEntity.LogisticsBidding
	db := database.GetDB()
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	var quote orderEntity.LogisticsBidQuote
	if err := db.WithContext(ctx).Where("id = ? AND bidding_id = ?", req.QuoteID, bidding.ID).First(&quote).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流报价不存在")
	}
	role := actorServiceForOrder(order, enterpriseID)
	if order.BuyerID == enterpriseID {
		bidding.BuyerConfirmedQuoteID = req.QuoteID
	} else {
		bidding.SellerConfirmedQuoteID = req.QuoteID
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingLogisticsConfirm
	if bidding.BuyerConfirmedQuoteID == req.QuoteID && bidding.SellerConfirmedQuoteID == req.QuoteID {
		bidding.SelectedQuoteID = req.QuoteID
		bidding.Status = orderEntity.LogisticsBiddingStatusSelected
		order.Status = orderEntity.OrderStatusShipping
	}
	order.UpdatedBy = operatorID
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bidding).Error; err != nil {
			return err
		}
		if order.Status == orderEntity.OrderStatusShipping {
			task := &orderEntity.ShipmentTask{
				OrderID:               order.ID,
				OrderNo:               order.OrderNo,
				LogisticsEnterpriseID: quote.LogisticsEnterpriseID,
				LogisticsQuoteID:      quote.ID,
				Status:                orderEntity.ShipmentStatusPendingPickup,
			}
			if err := tx.Where("order_id = ?", order.ID).Assign(task).FirstOrCreate(task).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "logistics_quote_confirmed", operatorID, enterpriseID, role, fromStatus, order.Status, orderEntity.OrderEventPayload{
			"quote_id": req.QuoteID,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认物流资源失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) UpdateShipmentProgress(ctx context.Context, id, enterpriseID, operatorID uint64, action string, req *dto.ShipmentProgressRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusShipping && order.Status != orderEntity.OrderStatusDelivered {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可回传运输进度")
	}
	fromStatus := order.Status
	db := database.GetDB()
	var task orderEntity.ShipmentTask
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).First(&task).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("运输任务不存在")
	}
	if task.LogisticsEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权回传该订单运输进度")
	}
	now := time.Now()
	eventType := ""
	switch action {
	case "pickup":
		task.Status = orderEntity.ShipmentStatusPickedUp
		task.PickupVoucher = strings.TrimSpace(req.VoucherURL)
		task.PickedUpAt = &now
		eventType = "shipment_picked_up"
	case "deliver":
		task.Status = orderEntity.ShipmentStatusDelivered
		task.DeliveryVoucher = strings.TrimSpace(req.VoucherURL)
		task.DeliveredAt = &now
		order.Status = orderEntity.OrderStatusDelivered
		eventType = "shipment_delivered"
	case "signature":
		task.Status = orderEntity.ShipmentStatusSigned
		task.SignatureVoucher = strings.TrimSpace(req.VoucherURL)
		task.SignedAt = &now
		eventType = "shipment_signature_uploaded"
	default:
		return nil, ecode.ErrBadRequest.WithMessage("运输动作无效")
	}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, eventType, operatorID, enterpriseID, "logistics", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"voucher_url": strings.TrimSpace(req.VoucherURL),
			"remark":      strings.TrimSpace(req.Remark),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("回传运输进度失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) Reconcile(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单对账")
	}
	if order.Status != orderEntity.OrderStatusReceived {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可对账")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusReconciled
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "seller_reconciled", operatorID, enterpriseID, "sales", fromStatus, order.Status, nil)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("对账确认失败")
	}
	return toOrderInfo(order), nil
}

func (s *OrderService) Archive(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权归档该订单")
	}
	if order.Status != orderEntity.OrderStatusReconciled {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可归档")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusArchived
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "order_archived", operatorID, enterpriseID, actorServiceForOrder(order, enterpriseID), fromStatus, order.Status, nil)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("订单归档失败")
	}
	return toOrderInfo(order), nil
}

func isOrderSeller(order *orderEntity.Order, enterpriseID uint64) bool {
	if order == nil || enterpriseID == 0 {
		return false
	}
	return order.SellerID == enterpriseID || order.SellerEnterpriseID == enterpriseID
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
			DriverName:            order.DeliveryInfo.DriverName,
			DriverPhone:           order.DeliveryInfo.DriverPhone,
			DriverIDCard:          order.DeliveryInfo.DriverIdCard,
			PlateNumber:           order.DeliveryInfo.PlateNumber,
			PickupCode:            order.DeliveryInfo.PickupCode,
			WarehouseID:           order.DeliveryInfo.WarehouseID,
			WarehouseServiceID:    order.DeliveryInfo.WarehouseServiceID,
			WarehouseEnterpriseID: order.DeliveryInfo.WarehouseEnterpriseID,
			WarehouseName:         order.DeliveryInfo.WarehouseName,
			WarehouseCode:         order.DeliveryInfo.WarehouseCode,
			WarehouseProvince:     order.DeliveryInfo.WarehouseProvince,
			WarehouseCity:         order.DeliveryInfo.WarehouseCity,
			WarehouseDistrict:     order.DeliveryInfo.WarehouseDistrict,
			WarehouseAddress:      order.DeliveryInfo.WarehouseAddress,
			Address:               order.DeliveryInfo.Address,
			EstimatedTime:         order.DeliveryInfo.EstimatedTime,
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
	sellerID := order.SellerID
	if sellerID == 0 {
		sellerID = order.SellerEnterpriseID
	}
	return &dto.OrderInfo{
		ID:             order.ID,
		OrderNo:        order.OrderNo,
		BuyerID:        order.BuyerID,
		BuyerName:      order.BuyerName,
		SellerID:       sellerID,
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

	if orderStatusReached(status, orderEntity.OrderStatusPendingContract) {
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

	if orderStatusReached(status, orderEntity.OrderStatusPendingPayment) {
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

	if orderStatusReached(status, orderEntity.OrderStatusPendingDelivery) && status != orderEntity.OrderStatusCancelled {
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

	if orderStatusReached(status, orderEntity.OrderStatusDelivering) && status != orderEntity.OrderStatusCancelled {
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

	if status == orderEntity.OrderStatusReceived || status == orderEntity.OrderStatusReconciled || status == orderEntity.OrderStatusArchived || status == orderEntity.OrderStatusCompleted {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "completed",
			Time:        order.ReceiptAt,
			Description: "买家已确认收货并上传签收凭证，订单完成归档。",
		})
	} else if status == orderEntity.OrderStatusDelivering || status == orderEntity.OrderStatusShipping || status == orderEntity.OrderStatusDelivered {
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
	if status == orderEntity.OrderStatusReconciled || status == orderEntity.OrderStatusArchived {
		items = append(items, dto.TimelineItem{
			Key:         "reconciled",
			Label:       "卖方对账确认",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "卖方已完成收款对账确认。",
		})
	}
	if status == orderEntity.OrderStatusArchived {
		items = append(items, dto.TimelineItem{
			Key:         "archived",
			Label:       "订单归档",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "订单已归档。",
		})
	}

	return items
}

func orderStatusReached(current, target orderEntity.OrderStatus) bool {
	rank := map[orderEntity.OrderStatus]int{
		orderEntity.OrderStatusDraft:                    0,
		orderEntity.OrderStatusPendingWarehouseConfirm:  1,
		orderEntity.OrderStatusPendingSellerConfirm:     2,
		orderEntity.OrderStatusPendingConfirm:           2,
		orderEntity.OrderStatusPendingBuyerConfirm:      3,
		orderEntity.OrderStatusPendingContract:          4,
		orderEntity.OrderStatusPendingPayment:           5,
		orderEntity.OrderStatusFinancingReview:          6,
		orderEntity.OrderStatusPendingFinancingContract: 7,
		orderEntity.OrderStatusPaid:                     8,
		orderEntity.OrderStatusLogisticsBidding:         9,
		orderEntity.OrderStatusPendingLogisticsConfirm:  10,
		orderEntity.OrderStatusPendingDelivery:          11,
		orderEntity.OrderStatusDelivering:               12,
		orderEntity.OrderStatusShipping:                 12,
		orderEntity.OrderStatusDelivered:                13,
		orderEntity.OrderStatusReceived:                 14,
		orderEntity.OrderStatusCompleted:                14,
		orderEntity.OrderStatusReconciled:               15,
		orderEntity.OrderStatusArchived:                 16,
	}
	currentRank, ok := rank[current]
	if !ok {
		return false
	}
	targetRank, ok := rank[target]
	if !ok {
		return false
	}
	return currentRank >= targetRank
}

func (s *OrderService) reconcileOrderState(ctx context.Context, order *orderEntity.Order) error {
	if order == nil {
		return nil
	}
	if order.Status == orderEntity.OrderStatusCancelled || order.Status == orderEntity.OrderStatusCompleted || order.Status == orderEntity.OrderStatusArchived {
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
			order.Status = orderEntity.OrderStatusReceived
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

func (s *OrderService) ListEvents(ctx context.Context, id, enterpriseID uint64) ([]dto.OrderEventInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) && warehouseEnterpriseIDOf(order) != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权查看该订单事件")
	}
	var events []orderEntity.OrderEvent
	if err := database.GetDB().WithContext(ctx).Where("order_id = ?", order.ID).Order("id asc").Find(&events).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询订单事件失败")
	}
	result := make([]dto.OrderEventInfo, 0, len(events))
	for _, event := range events {
		result = append(result, toOrderEventInfo(event))
	}
	return result, nil
}

func (s *OrderService) getOrderFinancing(ctx context.Context, orderID uint64) (*orderEntity.Order, *orderEntity.OrderFinancingRequest, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	var financing orderEntity.OrderFinancingRequest
	query := database.GetDB().WithContext(ctx).Where("order_id = ?", order.ID)
	if order.FinanceID > 0 {
		query = query.Where("id = ?", order.FinanceID)
	}
	if err := query.Order("id desc").First(&financing).Error; err != nil {
		return nil, nil, ecode.ErrNotFound.WithMessage("融资申请不存在")
	}
	return order, &financing, nil
}

func createOrderEvent(tx *gorm.DB, ctx context.Context, order *orderEntity.Order, eventType string, actorUserID, actorEnterpriseID uint64, actorService string, fromStatus, toStatus orderEntity.OrderStatus, payload orderEntity.OrderEventPayload) error {
	if tx == nil || order == nil {
		return nil
	}
	event := &orderEntity.OrderEvent{
		OrderID:           order.ID,
		OrderNo:           order.OrderNo,
		EventType:         eventType,
		ActorUserID:       actorUserID,
		ActorEnterpriseID: actorEnterpriseID,
		ActorService:      actorService,
		FromStatus:        fromStatus,
		ToStatus:          toStatus,
		Payload:           payload,
	}
	return tx.WithContext(ctx).Create(event).Error
}

func actorServiceForOrder(order *orderEntity.Order, enterpriseID uint64) string {
	if order == nil {
		return ""
	}
	if order.BuyerID == enterpriseID {
		return "procurement"
	}
	if isOrderSeller(order, enterpriseID) {
		return "sales"
	}
	if warehouseEnterpriseIDOf(order) == enterpriseID {
		return "warehouse"
	}
	return ""
}

func warehouseEnterpriseIDOf(order *orderEntity.Order) uint64 {
	if order == nil || order.DeliveryInfo == nil {
		return 0
	}
	return order.DeliveryInfo.WarehouseEnterpriseID
}

func toOrderEventInfo(event orderEntity.OrderEvent) dto.OrderEventInfo {
	payload := map[string]interface{}(nil)
	if event.Payload != nil {
		payload = map[string]interface{}(event.Payload)
	}
	return dto.OrderEventInfo{
		ID:                event.ID,
		EventType:         event.EventType,
		ActorUserID:       event.ActorUserID,
		ActorEnterpriseID: event.ActorEnterpriseID,
		ActorService:      event.ActorService,
		FromStatus:        int(event.FromStatus),
		ToStatus:          int(event.ToStatus),
		Payload:           payload,
		CreatedAt:         event.CreatedAt,
	}
}

func toLogisticsBiddingInfo(bidding *orderEntity.LogisticsBidding) *dto.LogisticsBiddingInfo {
	if bidding == nil {
		return nil
	}
	return &dto.LogisticsBiddingInfo{
		ID:                     bidding.ID,
		OrderID:                bidding.OrderID,
		OrderNo:                bidding.OrderNo,
		Status:                 string(bidding.Status),
		SelectedQuoteID:        bidding.SelectedQuoteID,
		BuyerConfirmedQuoteID:  bidding.BuyerConfirmedQuoteID,
		SellerConfirmedQuoteID: bidding.SellerConfirmedQuoteID,
		CreatedAt:              bidding.CreatedAt,
		UpdatedAt:              bidding.UpdatedAt,
	}
}

func toLogisticsQuoteInfo(quote *orderEntity.LogisticsBidQuote) *dto.LogisticsQuoteInfo {
	if quote == nil {
		return nil
	}
	return &dto.LogisticsQuoteInfo{
		ID:                    quote.ID,
		BiddingID:             quote.BiddingID,
		OrderID:               quote.OrderID,
		LogisticsEnterpriseID: quote.LogisticsEnterpriseID,
		LogisticsName:         quote.LogisticsName,
		VehicleInfo:           quote.VehicleInfo,
		DriverInfo:            quote.DriverInfo,
		Amount:                quote.Amount,
		EstimatedPickupAt:     quote.EstimatedPickupAt,
		EstimatedDeliveryAt:   quote.EstimatedDeliveryAt,
		Remark:                quote.Remark,
		CreatedAt:             quote.CreatedAt,
	}
}

func toFinancingRequestInfo(item *orderEntity.OrderFinancingRequest) *dto.FinancingRequestInfo {
	if item == nil {
		return nil
	}
	return &dto.FinancingRequestInfo{
		ID:                  item.ID,
		OrderID:             item.OrderID,
		OrderNo:             item.OrderNo,
		BuyerEnterpriseID:   item.BuyerEnterpriseID,
		SellerEnterpriseID:  item.SellerEnterpriseID,
		FinanceServiceID:    item.FinanceServiceID,
		FundingEnterpriseID: item.FundingEnterpriseID,
		RequestedAmount:     item.RequestedAmount,
		ApprovedAmount:      item.ApprovedAmount,
		TermDays:            item.TermDays,
		AnnualRate:          item.AnnualRate,
		ServiceFee:          item.ServiceFee,
		Status:              string(item.Status),
		RiskResult:          item.RiskResult,
		RiskRemark:          item.RiskRemark,
		BuyerConfirmedAt:    item.BuyerConfirmedAt,
		DebtContractNo:      item.DebtContractNo,
		DebtContractURL:     item.DebtContractURL,
		DebtContractHash:    item.DebtContractHash,
		BuyerSignedAt:       item.BuyerSignedAt,
		FeeVoucherURL:       item.FeeVoucherURL,
		FeePaidAt:           item.FeePaidAt,
		DisbursementVoucher: item.DisbursementVoucher,
		DisbursedAt:         item.DisbursedAt,
		PledgeVoucherURL:    item.PledgeVoucherURL,
		PledgeRemark:        item.PledgeRemark,
		PledgedAt:           item.PledgedAt,
		ReleaseCondition:    item.ReleaseCondition,
		ReleaseReason:       item.ReleaseReason,
		ReleasedAt:          item.ReleasedAt,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}
