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
	"github.com/Taishi-co/SteelLink-backend/internal/modules/notification/application/realtime"
	notificationEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	orderRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/repository"
	orderPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/order/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	fulfillmentEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/domain/entity"
	offerEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	userEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
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

func (s *OrderService) PreviewWarehousePreData(ctx context.Context, req *dto.DraftOrderPreviewRequest) ([]*dto.WarehousePreDataResponse, error) {
	results := make([]*dto.WarehousePreDataResponse, 0, len(req.Items))
	for _, item := range req.Items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, ecode.ErrDatabase
		}
		if product == nil {
			return nil, ecode.ErrNotFound.WithMessage("商品不存在")
		}
		deliveryInfo := buildInitialDeliveryInfo(product)
		sellerEnterprise, _ := s.enterpriseRepo.FindByID(ctx, product.EnterpriseID)
		sellerName := ""
		if sellerEnterprise != nil {
			sellerName = sellerEnterprise.Name
		}
		canConfirm := deliveryInfo != nil && deliveryInfo.WarehouseEnterpriseID != 0
		result := &dto.WarehousePreDataResponse{
			ProductID:           product.ID,
			ProductName:         product.Name,
			SellerEnterpriseID:  product.EnterpriseID,
			SellerName:          sellerName,
			CanWarehouseConfirm: canConfirm,
		}
		if deliveryInfo != nil {
			result.WarehouseID = deliveryInfo.WarehouseID
			result.WarehouseServiceID = deliveryInfo.WarehouseServiceID
			result.WarehouseEnterpriseID = deliveryInfo.WarehouseEnterpriseID
			result.WarehouseName = deliveryInfo.WarehouseName
			result.WarehouseCode = deliveryInfo.WarehouseCode
			result.WarehouseProvince = deliveryInfo.WarehouseProvince
			result.WarehouseCity = deliveryInfo.WarehouseCity
			result.WarehouseDistrict = deliveryInfo.WarehouseDistrict
			result.WarehouseAddress = deliveryInfo.WarehouseAddress
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *OrderService) ListProductPool(ctx context.Context) ([]*dto.ProductPoolItem, error) {
	onSale := 2
	products, _, err := s.productRepo.List(ctx, &productRepository.ProductQuery{
		Status:   &onSale,
		PageSize: 200,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	items := make([]*dto.ProductPoolItem, 0, len(products))
	for _, product := range products {
		sellerEnterprise, _ := s.enterpriseRepo.FindByID(ctx, product.EnterpriseID)
		sellerName := ""
		if sellerEnterprise != nil {
			sellerName = sellerEnterprise.Name
		}
		hasWarehouse := product.WarehouseInfo != nil && product.WarehouseEnterpriseID != 0
		item := &dto.ProductPoolItem{
			ID:                  product.ID,
			ProductNo:           product.ProductNo,
			Name:                product.Name,
			Category:            product.Category,
			SubCategory:         product.SubCategory,
			SteelMill:           product.SteelMill,
			UnitPrice:           product.UnitPrice,
			AvailableQuantity:   product.AvailableQty,
			HasWarehouseBinding: hasWarehouse,
			SellerEnterpriseID:  product.EnterpriseID,
			SellerName:          sellerName,
		}
		if hasWarehouse && product.WarehouseInfo != nil {
			item.WarehouseName = product.WarehouseInfo.WarehouseName
			item.WarehouseAddress = product.WarehouseInfo.Address
		}
		items = append(items, item)
	}
	return items, nil
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

	db := database.GetDB()
	if req.DestWarehouseServiceID > 0 {
		var destWarehouse offerEntity.WarehouseService
		if err := db.WithContext(ctx).Where("id = ? AND status = ?", req.DestWarehouseServiceID, offerEntity.OfferStatusApproved).First(&destWarehouse).Error; err != nil {
			return nil, ecode.ErrBadRequest.WithMessage("目的地仓储资源不存在或未审核通过")
		}
		initialDeliveryInfo.DestWarehouseServiceID = destWarehouse.ID
		initialDeliveryInfo.DestWarehouseEnterpriseID = destWarehouse.EnterpriseID
		initialDeliveryInfo.DestWarehouseName = destWarehouse.WarehouseName
		initialDeliveryInfo.DestWarehouseProvince = destWarehouse.Province
		initialDeliveryInfo.DestWarehouseCity = destWarehouse.City
		initialDeliveryInfo.DestWarehouseDistrict = destWarehouse.District
		initialDeliveryInfo.DestWarehouseAddress = destWarehouse.Address
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
		Status:             orderEntity.OrderStatusPendingDestWarehouse,
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

	return toOrderInfo(order, nil), nil
}

type productOrderItem struct {
	Product  *productEntity.Product
	Quantity float64
}

func createEnterpriseNotification(tx *gorm.DB, ctx context.Context, enterpriseID, senderUserID uint64, notificationType notificationEntity.NotificationType, title, content, linkURL string, metadata notificationEntity.NotificationMetadata) error {
	if enterpriseID == 0 {
		return nil
	}
	var enterprise userEntity.Enterprise
	if err := tx.WithContext(ctx).Select("id, enterprise_code").First(&enterprise, enterpriseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if enterprise.EnterpriseCode == "" {
		return nil
	}
	var users []userEntity.User
	if err := tx.WithContext(ctx).
		Select("id").
		Where("enterprise_code = ? AND status = ? AND deleted_at IS NULL", enterprise.EnterpriseCode, userEntity.StatusActive).
		Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}
	notifications := make([]notificationEntity.Notification, 0, len(users))
	for _, user := range users {
		notifications = append(notifications, notificationEntity.Notification{
			RecipientUserID:       user.ID,
			RecipientEnterpriseID: enterpriseID,
			SenderUserID:          senderUserID,
			Type:                  notificationType,
			Title:                 title,
			Content:               content,
			LinkURL:               linkURL,
			Status:                notificationEntity.NotificationStatusUnread,
			Metadata:              metadata,
		})
	}
	if err := tx.WithContext(ctx).Create(&notifications).Error; err != nil {
		return err
	}
	for _, notification := range notifications {
		realtime.DefaultHub.Publish(notification)
	}
	return nil
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

func (s *OrderService) GetByID(ctx context.Context, id, enterpriseID uint64, allowAll bool) (*dto.OrderInfo, error) {
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
	if !allowAll && (enterpriseID == 0 || (order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID))) {
		funding, fundingErr := s.isFundingParticipant(ctx, order.ID, enterpriseID)
		if fundingErr != nil {
			return nil, fundingErr
		}
		if !funding {
			return nil, ecode.ErrForbidden.WithMessage("无权查看该订单")
		}
	}

	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) isFundingParticipant(ctx context.Context, orderID, enterpriseID uint64) (bool, error) {
	if orderID == 0 || enterpriseID == 0 {
		return false, nil
	}
	var count int64
	if err := database.GetDB().WithContext(ctx).
		Model(&orderEntity.OrderFinancingRequest{}).
		Where("order_id = ? AND funding_enterprise_id = ?", orderID, enterpriseID).
		Count(&count).Error; err != nil {
		return false, ecode.ErrDatabase.WithMessage("查询订单资金方权限失败")
	}
	return count > 0, nil
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
		result = append(result, toOrderInfo(order, nil))
	}
	return result, nil
}

func (s *OrderService) ListAll(ctx context.Context, keyword string, status *int, page, pageSize int) (*dto.OrderListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	db := database.GetDB().WithContext(ctx).Model(&orderEntity.Order{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("order_no LIKE ? OR buyer_name LIKE ? OR seller_name LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询订单总数失败")
	}
	var orders []*orderEntity.Order
	if err := db.Order("id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&orders).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询订单列表失败")
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		result = append(result, toOrderInfo(order, nil))
	}
	return &dto.OrderListResponse{List: result, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *OrderService) ListFundingFinancingRequests(ctx context.Context, fundingEnterpriseID uint64) ([]*dto.FinancingRequestInfo, error) {
	if fundingEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看融资申请")
	}
	var requests []orderEntity.OrderFinancingRequest
	if err := database.GetDB().WithContext(ctx).
		Where("funding_enterprise_id = ?", fundingEnterpriseID).
		Order("id desc").
		Find(&requests).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询资金方融资申请失败")
	}
	result := make([]*dto.FinancingRequestInfo, 0, len(requests))
	for i := range requests {
		result = append(result, toFinancingRequestInfo(&requests[i]))
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
		order.Status != orderEntity.OrderStatusPendingWarehouseSelect &&
		order.Status != orderEntity.OrderStatusPendingWarehouseReselect &&
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

	return toOrderInfo(order, nil), nil
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
	if order.DeliveryInfo == nil || order.DeliveryInfo.DestWarehouseEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("订单未绑定目的地仓储方，无法仓库确认")
	}
	if order.DeliveryInfo.DestWarehouseEnterpriseID != warehouseEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权确认该订单仓库条件")
	}
	if order.Status != orderEntity.OrderStatusPendingWarehouseConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可仓库确认")
	}
	if req != nil && !req.Confirmed {
		fromStatus := order.Status
		order.Status = orderEntity.OrderStatusPendingWarehouseReselect
		order.UpdatedBy = operatorID
		if order.Remark == nil {
			order.Remark = orderEntity.Remark{}
		}
		order.Remark["warehouse_reject_note"] = strings.TrimSpace(req.ConditionNote)
		order.Remark["warehouse_reject_voucher_url"] = strings.TrimSpace(req.VoucherURL)
		order.Remark["admin_attention"] = true
		if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(order).Error; err != nil {
				return err
			}
			return createOrderEvent(tx, ctx, order, "warehouse_rejected", operatorID, warehouseEnterpriseID, "warehouse", fromStatus, order.Status, orderEntity.OrderEventPayload{
				"condition_note":  strings.TrimSpace(req.ConditionNote),
				"voucher_url":     strings.TrimSpace(req.VoucherURL),
				"admin_attention": true,
				"next_action":     "buyer_reselect_warehouse",
			})
		}); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("仓库拒绝确认失败")
		}
		return toOrderInfo(order, nil), nil
	}
	fromStatus := order.Status
	db := database.GetDB()
	order.Status = orderEntity.OrderStatusPendingSellerConfirm
	order.UpdatedBy = operatorID
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		payload := orderEntity.OrderEventPayload{}
		if req != nil {
			payload["condition_note"] = strings.TrimSpace(req.ConditionNote)
			payload["voucher_url"] = strings.TrimSpace(req.VoucherURL)
		}
		if err := createOrderEvent(tx, ctx, order, "warehouse_confirmed", operatorID, warehouseEnterpriseID, "warehouse", fromStatus, order.Status, payload); err != nil {
			return err
		}
		return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
			"仓库已确认订单",
			"仓储方已确认订单「"+order.OrderNo+"」，请卖方继续确认订单。",
			"/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "event": "warehouse_confirmed"},
		)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("仓库确认失败")
	}
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) ReselectWarehouse(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.WarehouseReselectRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权为该订单选择仓库")
	}
	if order.Status != orderEntity.OrderStatusPendingWarehouseReselect && order.Status != orderEntity.OrderStatusPendingWarehouseSelect {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可选择仓库")
	}
	var warehouse offerEntity.WarehouseService
	db := database.GetDB()
	if err := db.WithContext(ctx).Where("id = ? AND status = ?", req.WarehouseServiceID, offerEntity.OfferStatusApproved).First(&warehouse).Error; err != nil {
		return nil, ecode.ErrBadRequest.WithMessage("仓储资源不存在或未审核通过")
	}
	if order.DeliveryInfo == nil {
		order.DeliveryInfo = &orderEntity.DeliveryInfo{}
	}
	order.DeliveryInfo.WarehouseServiceID = warehouse.ID
	order.DeliveryInfo.WarehouseEnterpriseID = warehouse.EnterpriseID
	order.DeliveryInfo.WarehouseName = warehouse.WarehouseName
	order.DeliveryInfo.WarehouseProvince = warehouse.Province
	order.DeliveryInfo.WarehouseCity = warehouse.City
	order.DeliveryInfo.WarehouseDistrict = warehouse.District
	order.DeliveryInfo.WarehouseAddress = warehouse.Address
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingWarehouseConfirm
	order.UpdatedBy = operatorID
	if order.Remark == nil {
		order.Remark = orderEntity.Remark{}
	}
	order.Remark["warehouse_reselect_note"] = strings.TrimSpace(req.ConfirmNote)
	order.Remark["admin_attention"] = false
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "warehouse_reselected", operatorID, buyerEnterpriseID, "procurement", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"warehouse_service_id":    warehouse.ID,
			"warehouse_enterprise_id": warehouse.EnterpriseID,
			"warehouse_name":          warehouse.WarehouseName,
			"confirm_note":            strings.TrimSpace(req.ConfirmNote),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("重新选择仓库失败")
	}
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) SelectDestWarehouse(ctx context.Context, id, buyerEnterpriseID, operatorID uint64, req *dto.DestWarehouseSelectRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != buyerEnterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权为该订单选择目的地仓库")
	}
	if order.Status != orderEntity.OrderStatusPendingDestWarehouse {
		return nil, ecode.ErrBadRequest.WithMessage("订单状态不允许选择目的地仓库")
	}
	if order.DeliveryInfo == nil || order.DeliveryInfo.WarehouseEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("订单未绑定提货仓库，无法选择目的地仓库")
	}
	if order.DeliveryInfo.DestWarehouseEnterpriseID != 0 {
		return nil, ecode.ErrBadRequest.WithMessage("订单已选择目的地仓库，如需更改请联系管理员")
	}
	var warehouse offerEntity.WarehouseService
	db := database.GetDB()
	if err := db.WithContext(ctx).Where("id = ? AND status = ?", req.DestWarehouseServiceID, offerEntity.OfferStatusApproved).First(&warehouse).Error; err != nil {
		return nil, ecode.ErrBadRequest.WithMessage("目的地仓储资源不存在或未审核通过")
	}
	order.DeliveryInfo.DestWarehouseServiceID = warehouse.ID
	order.DeliveryInfo.DestWarehouseEnterpriseID = warehouse.EnterpriseID
	order.DeliveryInfo.DestWarehouseName = warehouse.WarehouseName
	order.DeliveryInfo.DestWarehouseProvince = warehouse.Province
	order.DeliveryInfo.DestWarehouseCity = warehouse.City
	order.DeliveryInfo.DestWarehouseDistrict = warehouse.District
	order.DeliveryInfo.DestWarehouseAddress = warehouse.Address
	order.Status = orderEntity.OrderStatusPendingWarehouseConfirm
	order.UpdatedBy = operatorID
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if err := ensureWarehouseMovementTx(tx, order, fulfillmentEntity.WarehouseMovementOutbound, order.DeliveryInfo.WarehouseServiceID, order.DeliveryInfo.WarehouseEnterpriseID, operatorID, "", "主仓待确认出库安排"); err != nil {
			return err
		}
		if err := ensureWarehouseMovementTx(tx, order, fulfillmentEntity.WarehouseMovementInbound, order.DeliveryInfo.DestWarehouseServiceID, order.DeliveryInfo.DestWarehouseEnterpriseID, operatorID, "", "目的仓待确认入库安排"); err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "dest_warehouse_selected", operatorID, buyerEnterpriseID, "procurement", orderEntity.OrderStatusPendingDestWarehouse, orderEntity.OrderStatusPendingWarehouseConfirm, orderEntity.OrderEventPayload{
			"dest_warehouse_service_id":    warehouse.ID,
			"dest_warehouse_enterprise_id": warehouse.EnterpriseID,
			"dest_warehouse_name":          warehouse.WarehouseName,
			"dest_confirm_note":            strings.TrimSpace(req.DestConfirmNote),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("选择目的地仓库失败")
	}
	return toOrderInfo(order, nil), nil
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
		if err := createOrderEvent(tx, ctx, order, "seller_confirmed", operatorID, sellerEnterpriseID, "sales", fromStatus, order.Status, payload); err != nil {
			return err
		}
		return createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
			"卖方已确认订单",
			"卖方已确认订单「"+order.OrderNo+"」，请买方继续确认交易条件。",
			"/dashboard/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "event": "seller_confirmed"},
		)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("卖方确认订单失败")
	}
	return toOrderInfo(order, nil), nil
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
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) SignContract(ctx context.Context, id, enterpriseID, operatorID uint64, tradeRole string) (*dto.OrderInfo, error) {
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
	signer, actorService, signErr := resolveContractSigner(order, enterpriseID, tradeRole)
	if signErr != nil {
		return nil, signErr
	}
	if signer == "buyer" {
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
		// 设置过期时间
		expiresAt := time.Now().Add(24 * time.Hour)
		order.ExpiresAt = &expiresAt
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
		return createOrderEvent(tx, ctx, order, eventType, operatorID, enterpriseID, actorService, fromStatus, toStatus, orderEntity.OrderEventPayload{
			"signer": signer,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("签署合同失败")
	}
	return toOrderInfo(order, nil), nil
}

func resolveContractSigner(order *orderEntity.Order, enterpriseID uint64, tradeRole string) (string, string, *ecode.Error) {
	isBuyer := order != nil && order.BuyerID == enterpriseID
	isSeller := isOrderSeller(order, enterpriseID)

	switch strings.ToLower(strings.TrimSpace(tradeRole)) {
	case "buyer":
		if !isBuyer {
			return "", "", ecode.ErrForbidden.WithMessage("当前身份不是该订单买方")
		}
		return "buyer", "procurement", nil
	case "seller":
		if !isSeller {
			return "", "", ecode.ErrForbidden.WithMessage("当前身份不是该订单卖方")
		}
		return "seller", "sales", nil
	case "":
		if isBuyer && !isSeller {
			return "buyer", "procurement", nil
		}
		if isSeller && !isBuyer {
			return "seller", "sales", nil
		}
		if isBuyer && isSeller {
			return "", "", ecode.ErrBadRequest.WithMessage("无法判定当前签署身份，请从对应买家或卖家中心重新发起签署")
		}
		return "", "", ecode.ErrForbidden.WithMessage("无权签署该订单合同")
	default:
		return "", "", ecode.ErrBadRequest.WithMessage("无效的签署身份")
	}
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
	return toOrderInfo(order, nil), nil
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
		result = append(result, toOrderInfo(order, nil))
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
		result = append(result, toOrderInfo(order, nil))
	}
	return result, nil
}

func (s *OrderService) ListByWarehouse(ctx context.Context, warehouseEnterpriseID uint64, status *int) ([]*dto.OrderInfo, error) {
	if warehouseEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看仓储订单")
	}
	query := database.GetDB().WithContext(ctx).
		Where("JSON_EXTRACT(delivery_info, '$.warehouse_enterprise_id') = ? OR JSON_EXTRACT(delivery_info, '$.dest_warehouse_enterprise_id') = ?", warehouseEnterpriseID, warehouseEnterpriseID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var orders []*orderEntity.Order
	if err := query.Order("id desc").Find(&orders).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询仓储订单失败")
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		result = append(result, toOrderInfo(order, nil))
	}
	return result, nil
}

func (s *OrderService) ListByLogistics(ctx context.Context, logisticsEnterpriseID uint64, status *int, bidding bool) ([]*dto.OrderInfo, error) {
	if logisticsEnterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看物流订单")
	}
	db := database.GetDB().WithContext(ctx)
	var orders []*orderEntity.Order
	if bidding {
		var biddingOrders []*orderEntity.Order
		if err := db.Where("status = ?", orderEntity.OrderStatusLogisticsBidding).
			Order("id desc").Find(&biddingOrders).Error; err != nil {
			return nil, ecode.ErrDatabase.WithMessage("查询物流竞价订单失败")
		}
		var pendingOrders []*orderEntity.Order
		var quotedOrderIDs []uint64
		db.Model(&orderEntity.LogisticsBidQuote{}).
			Where("logistics_enterprise_id = ?", logisticsEnterpriseID).
			Pluck("order_id", &quotedOrderIDs)
		if len(quotedOrderIDs) > 0 {
			var biddingIDs []uint64
			db.Model(&orderEntity.LogisticsBidding{}).
				Where("order_id IN ? AND selected_quote_id > 0 AND (responder_confirmed = true OR initiator_enterprise_id = ?)", quotedOrderIDs, logisticsEnterpriseID).
				Pluck("id", &biddingIDs)
			if len(biddingIDs) > 0 {
				if err := db.Where("status = ? AND id IN (SELECT order_id FROM logistics_biddings WHERE id IN ?)", orderEntity.OrderStatusPendingLogisticsConfirm, biddingIDs).
					Order("id desc").Find(&pendingOrders).Error; err != nil {
					return nil, ecode.ErrDatabase.WithMessage("查询待接单物流订单失败")
				}
			}
		}
		orders = append(biddingOrders, pendingOrders...)
	} else if status != nil && *status == int(orderEntity.OrderStatusLogisticsBidding) {
		if err := db.Where("status = ?", orderEntity.OrderStatusLogisticsBidding).
			Order("id desc").Find(&orders).Error; err != nil {
			return nil, ecode.ErrDatabase.WithMessage("查询物流竞价订单失败")
		}
	} else if status != nil && *status == int(orderEntity.OrderStatusPendingLogisticsConfirm) {
		var quotedOrderIDs []uint64
		db.Model(&orderEntity.LogisticsBidQuote{}).
			Where("logistics_enterprise_id = ?", logisticsEnterpriseID).
			Pluck("order_id", &quotedOrderIDs)
		if len(quotedOrderIDs) == 0 {
			return []*dto.OrderInfo{}, nil
		}
		var biddingIDs []uint64
		db.Model(&orderEntity.LogisticsBidding{}).
			Where("order_id IN ? AND selected_quote_id > 0 AND (responder_confirmed = true OR initiator_enterprise_id = ?)", quotedOrderIDs, logisticsEnterpriseID).
			Pluck("id", &biddingIDs)
		if len(biddingIDs) == 0 {
			return []*dto.OrderInfo{}, nil
		}
		if err := db.Where("status = ? AND id IN (SELECT order_id FROM logistics_biddings WHERE id IN ?)", orderEntity.OrderStatusPendingLogisticsConfirm, biddingIDs).
			Order("id desc").Find(&orders).Error; err != nil {
			return nil, ecode.ErrDatabase.WithMessage("查询待接单物流订单失败")
		}
	} else {
		quotedOrderIDs := make([]uint64, 0)
		db.Model(&orderEntity.LogisticsBidQuote{}).
			Where("logistics_enterprise_id = ?", logisticsEnterpriseID).
			Pluck("order_id", &quotedOrderIDs)

		conditions := []string{"id IN (SELECT order_id FROM shipment_tasks WHERE logistics_enterprise_id = ?)"}
		args := []interface{}{logisticsEnterpriseID}
		if len(quotedOrderIDs) > 0 {
			conditions = append(conditions, "id IN ?")
			args = append(args, quotedOrderIDs)
		}
		query := db.Where(strings.Join(conditions, " OR "), args...)
		if status != nil {
			query = query.Where("status = ?", *status)
		}
		if err := query.Order("id desc").Find(&orders).Error; err != nil {
			return nil, ecode.ErrDatabase.WithMessage("查询物流订单失败")
		}
	}
	result := make([]*dto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		info := s.toOrderDetailInfo(ctx, order)
		result = append(result, info)
	}
	return result, nil
}

func (s *OrderService) UploadPaymentVoucher(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.UploadPaymentVoucherRequest) (*dto.OrderInfo, error) {
	db := database.GetDB()
	var resultOrder *orderEntity.Order

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中重新获取订单，确保数据一致性
		var order orderEntity.Order
		if err := tx.First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrNotFound.WithMessage("订单不存在")
			}
			return fmt.Errorf("查询订单失败: %w", err)
		}

		// 权限检查
		if order.BuyerID != enterpriseID {
			return ecode.ErrForbidden.WithMessage("无权上传支付凭证")
		}

		// 检查当前状态，但不调用 reconcileOrderState，
		// 因为我们正在主动上传支付凭证，应该优先处理这个操作
		if order.Status != orderEntity.OrderStatusPendingPayment {
			// 如果订单已经是已支付状态，也允许操作继续，当作重复提交的情况
			if order.Status != orderEntity.OrderStatusPaid {
				return ecode.ErrBadRequest.WithMessage("当前订单状态不可上传支付凭证")
			}
		}

		// 验证支付凭证URL
		if !isHTTPURL(req.VoucherURL) {
			return ecode.ErrBadRequest.WithMessage("支付凭证地址必须是有效的 HTTP/HTTPS URL")
		}

		now := time.Now()

		// 即使订单已经是已支付状态，也更新支付信息，保持数据最新
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

		// 确保状态更新为已支付
		fromStatus := order.Status
		order.Status = orderEntity.OrderStatusPaid
		order.ExpiresAt = nil // 清除过期时间
		order.UpdatedBy = operatorID

		// 保存订单
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("保存订单失败: %w", err)
		}

		// 创建订单事件
		if err := createOrderEvent(tx, ctx, &order, "direct_payment_confirmed", operatorID, enterpriseID, "procurement", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"transaction_no": strings.TrimSpace(req.TransactionNo),
			"voucher_url":    strings.TrimSpace(req.VoucherURL),
		}); err != nil {
			return fmt.Errorf("创建订单事件失败: %w", err)
		}

		resultOrder = &order
		return nil
	}); err != nil {
		if e, ok := err.(*ecode.Error); ok {
			return nil, e
		}
		return nil, ecode.ErrDatabase.WithMessage(err.Error())
	}

	return toOrderInfo(resultOrder, nil), nil
}

func (s *OrderService) DirectPayment(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.DirectPaymentRequest) (*dto.OrderInfo, error) {
	result, err := s.UploadPaymentVoucher(ctx, id, enterpriseID, operatorID, &dto.UploadPaymentVoucherRequest{
		VoucherURL:    req.VoucherURL,
		TransactionNo: req.TransactionNo,
		PayerAccount:  req.PayerAccount,
		PayeeAccount:  req.PayeeAccount,
		CapturedAt:    req.CapturedAt,
		GPSLatitude:   req.GPSLatitude,
		GPSLongitude:  req.GPSLongitude,
	})
	if err != nil {
		return nil, err
	}
	order, findErr := s.orderRepo.FindByID(ctx, id)
	if findErr != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}

	var titleVoucherURL string
	var titleVoucherRemark string

	if len(order.Items) > 0 {
		product, prodErr := s.productRepo.FindByID(ctx, order.Items[0].ProductID)
		if prodErr == nil && product != nil {
			titleVoucherURL = strings.TrimSpace(product.TitleVoucherURL)
			titleVoucherRemark = strings.TrimSpace(product.TitleVoucherRemark)
		}
	}

	if titleVoucherURL == "" {
		return nil, ecode.ErrBadRequest.WithMessage("商品未上传货权/仓单监管凭证，无法完成直采支付。请联系卖方在商品上架时认证货权凭证")
	}

	if order.Remark == nil {
		order.Remark = orderEntity.Remark{}
	}
	order.Remark["direct_title_voucher_url"] = titleVoucherURL
	order.Remark["direct_title_remark"] = titleVoucherRemark
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("保存订单货权监管信息失败: %w", err)
		}
		if err := createOrderEvent(tx, ctx, order, "direct_title_control_recorded", operatorID, enterpriseID, "procurement", order.Status, order.Status, orderEntity.OrderEventPayload{
			"title_voucher_url": titleVoucherURL,
			"title_remark":      titleVoucherRemark,
			"source":            "product_certification",
		}); err != nil {
			return fmt.Errorf("创建货权监管事件失败: %w", err)
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage(err.Error())
	}
	return result, nil
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
	if len(financeService.SupportedScenarios) > 0 {
		matched := false
		for _, item := range order.Items {
			for _, scenario := range financeService.SupportedScenarios {
				if strings.Contains(item.Category, scenario) || strings.Contains(scenario, item.Category) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return nil, ecode.ErrBadRequest.WithMessage("该融资产品不适用于当前订单商品品类")
		}
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
		order.Status = orderEntity.OrderStatusCancelled
		order.UpdatedBy = operatorID
		if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, item := range order.Items {
				product, findErr := s.productRepo.FindByID(ctx, item.ProductID)
				if findErr != nil {
					return findErr
				}
				if product == nil {
					return ecode.ErrBadRequest.WithMessage("订单商品不存在，无法终止融资订单")
				}
				if product.LockedQty >= item.Quantity {
					if !product.UnlockStock(item.Quantity) {
						return ecode.ErrBadRequest.WithMessage("订单锁定库存异常，无法终止融资订单")
					}
				} else {
					product.ReturnStock(item.Quantity)
				}
				product.UpdatedBy = operatorID
				if err := tx.Save(product).Error; err != nil {
					return err
				}
			}
			if order.Remark == nil {
				order.Remark = orderEntity.Remark{}
			}
			order.Remark["cancel_reason"] = "financing_risk_rejected"
			order.Remark["cancelled_by"] = "finance"
			if err := tx.Save(financing).Error; err != nil {
				return err
			}
			if err := tx.Save(order).Error; err != nil {
				return err
			}
			return createOrderEvent(tx, ctx, order, "financing_risk_rejected", operatorID, fundingEnterpriseID, "finance", fromStatus, order.Status, orderEntity.OrderEventPayload{
				"risk_remark":   financing.RiskRemark,
				"cancel_reason": "financing_risk_rejected",
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
	if financing.Status != orderEntity.FinancingRequestStatusRiskApproved {
		return nil, ecode.ErrBadRequest.WithMessage("当前融资申请状态不可买方确认")
	}
	fromStatus := order.Status
	if !req.Confirmed {
		financing.Status = orderEntity.FinancingRequestStatusRiskRejected
		financing.RiskResult = "buyer_declined"
		financing.UpdatedBy = operatorID
		order.Status = orderEntity.OrderStatusPendingPayment
		order.UpdatedBy = operatorID
		order.FinanceID = 0
		if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(financing).Error; err != nil {
				return err
			}
			if err := tx.Save(order).Error; err != nil {
				return err
			}
			return createOrderEvent(tx, ctx, order, "financing_buyer_declined", operatorID, buyerEnterpriseID, "procurement", fromStatus, order.Status, orderEntity.OrderEventPayload{
				"remark": strings.TrimSpace(req.Remark),
			})
		}); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("拒绝融资方案失败")
		}
		return toFinancingRequestInfo(financing), nil
	}
	now := time.Now()
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
	if req.FundingCommission < 0 {
		return nil, ecode.ErrBadRequest.WithMessage("资金方分佣金额不能小于0")
	}
	if !isHTTPURL(req.CommissionVoucher) {
		return nil, ecode.ErrBadRequest.WithMessage("资金方分佣凭证URL必须是有效的HTTP/HTTPS地址")
	}
	now := time.Now()
	fromStatus := order.Status
	financing.Status = orderEntity.FinancingRequestStatusDisbursed
	financing.FundingCommission = req.FundingCommission
	financing.CommissionVoucher = strings.TrimSpace(req.CommissionVoucher)
	financing.CommissionPaidAt = &now
	financing.DisbursementVoucher = strings.TrimSpace(req.DisbursementVoucher)
	financing.DisbursedAt = &now
	financing.UpdatedBy = operatorID
	order.Status = orderEntity.OrderStatusPaid
	order.ExpiresAt = nil // 清除过期时间
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			product, findErr := s.productRepo.FindByID(ctx, item.ProductID)
			if findErr != nil {
				return findErr
			}
			if product == nil {
				return ecode.ErrBadRequest.WithMessage("订单商品不存在")
			}
			if !product.DeductStock(item.Quantity) {
				return ecode.ErrBadRequest.WithMessage("订单库存扣减异常")
			}
			product.UpdatedBy = operatorID
			if saveErr := tx.Save(product).Error; saveErr != nil {
				return saveErr
			}
		}
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "financing_disbursed", operatorID, fundingEnterpriseID, "finance", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"disbursement_voucher": financing.DisbursementVoucher,
			"funding_commission":   financing.FundingCommission,
			"commission_voucher":   financing.CommissionVoucher,
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
	fromStatus := order.Status
	financing.Status = orderEntity.FinancingRequestStatusPledged
	financing.PledgeVoucherURL = strings.TrimSpace(req.PledgeVoucherURL)
	financing.PledgeRemark = strings.TrimSpace(req.PledgeRemark)
	financing.ReleaseCondition = strings.TrimSpace(req.ReleaseCondition)
	financing.PledgedAt = &now
	financing.UpdatedBy = operatorID
	order.Status = orderEntity.OrderStatusPendingDelivery
	order.UpdatedBy = operatorID
	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(financing).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "ownership_pledge_recorded", operatorID, fundingEnterpriseID, "finance", fromStatus, order.Status, orderEntity.OrderEventPayload{
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
	return s.toOrderDetailInfo(ctx, order), nil
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
	return toOrderInfo(order, nil), nil
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
	now := time.Now()

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if err := createOrderEvent(tx, ctx, order, "outbound_confirmed", operatorID, enterpriseID, "sales", orderEntity.OrderStatusPendingDelivery, order.Status, orderEntity.OrderEventPayload{
			"pickup_code": order.DeliveryInfo.PickupCode,
		}); err != nil {
			return err
		}
		var task orderEntity.ShipmentTask
		if err := tx.Where("order_id = ?", order.ID).First(&task).Error; err == nil {
			task.Status = orderEntity.ShipmentStatusPickedUp
			task.PickedUpAt = &now
			if err := tx.Save(&task).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认出库失败")
	}
	return toOrderInfo(order, nil), nil
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
		isFunding, _ := s.isFundingParticipant(ctx, order.ID, enterpriseID)
		if order.FinanceID == 0 || !isFunding {
			return nil, ecode.ErrForbidden.WithMessage("无权确认收货")
		}
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
		if err := createOrderEvent(tx, ctx, order, "buyer_received", operatorID, enterpriseID, "procurement", orderEntity.OrderStatusDelivering, order.Status, orderEntity.OrderEventPayload{
			"receipt_voucher": strings.TrimSpace(req.ReceiptVoucher),
		}); err != nil {
			return err
		}
		var task orderEntity.ShipmentTask
		if err := tx.Where("order_id = ?", order.ID).First(&task).Error; err == nil {
			task.Status = orderEntity.ShipmentStatusSigned
			task.SignatureVoucher = strings.TrimSpace(req.ReceiptVoucher)
			task.SignedAt = &now
			if err := tx.Save(&task).Error; err != nil {
				return err
			}
			if err := ensureLogisticsOrderSyncTx(tx, order, nil, nil, &task, operatorID); err != nil {
				return err
			}
		}
		var logisticsContract orderEntity.LogisticsContract
		if err := tx.Where("order_id = ? AND status = ?", order.ID, orderEntity.LogisticsContractCompleted).Order("id desc").First(&logisticsContract).Error; err == nil {
			commission := logisticsContract.FreightAmount * logisticsContract.CommissionRate
			settlement := orderEntity.LogisticsSettlement{
				SettlementNo:        generateSettlementNo(),
				LogisticsContractID: logisticsContract.ID,
				OrderID:             order.ID,
				OrderNo:             order.OrderNo,
				PayerEnterpriseID:   order.BuyerID,
				PayeeEnterpriseID:   logisticsContract.LogisticsEnterpriseID,
				FreightAmount:       logisticsContract.FreightAmount,
				PlatformCommission:  commission,
				NetPayable:          logisticsContract.FreightAmount - commission,
				Status:              orderEntity.LogisticsSettlementPending,
				CreatedBy:           operatorID,
			}
			if err := tx.Create(&settlement).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认收货失败")
	}
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) CreateLogisticsBidding(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.CreateLogisticsBiddingRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	isBuyer := order.BuyerID == enterpriseID
	isSeller := isOrderSeller(order, enterpriseID)
	if !isBuyer && !isSeller {
		return nil, ecode.ErrForbidden.WithMessage("无权发起该订单物流竞价")
	}
	if order.Status != orderEntity.OrderStatusPaid && order.Status != orderEntity.OrderStatusPendingDelivery && order.Status != orderEntity.OrderStatusLogisticsBidding && order.Status != orderEntity.OrderStatusPendingLogisticsRebid {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可发起物流竞价")
	}
	if err := s.ensureTitleControlReadyForLogistics(ctx, order); err != nil {
		return nil, err
	}

	initiatorRole := "seller"
	responderID := order.BuyerID
	if isBuyer {
		initiatorRole = "buyer"
		responderID = order.SellerID
	}
	freightPayer := strings.TrimSpace(req.FreightPayer)
	if freightPayer == "" {
		freightPayer = initiatorRole
	}
	bidding := &orderEntity.LogisticsBidding{
		OrderID:               order.ID,
		OrderNo:               order.OrderNo,
		SellerEnterpriseID:    order.SellerID,
		WarehouseEnterpriseID: warehouseEnterpriseIDOf(order),
		InitiatorEnterpriseID: enterpriseID,
		InitiatorRole:         initiatorRole,
		ResponderEnterpriseID: responderID,
		PickupAddress:         strings.TrimSpace(req.PickupAddress),
		DestinationAddress:    strings.TrimSpace(req.DestinationAddress),
		Requirement:           strings.TrimSpace(req.Requirement),
		FreightPayer:          freightPayer,
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
			"bidding_id":      bidding.ID,
			"freight_payer":   bidding.FreightPayer,
			"previous_status": int(fromStatus),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("发起物流竞价失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) ReportLogisticsNoCapacity(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.CreateLogisticsBiddingRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权反馈该订单物流无运力")
	}
	if order.Status != orderEntity.OrderStatusLogisticsBidding && order.Status != orderEntity.OrderStatusPendingLogisticsConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可反馈物流无运力")
	}
	db := database.GetDB()
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusPendingLogisticsRebid
	order.UpdatedBy = operatorID
	if order.Remark == nil {
		order.Remark = orderEntity.Remark{}
	}
	reason := ""
	if req != nil {
		reason = strings.TrimSpace(req.NoCapacityReason)
	}
	order.Remark["logistics_no_capacity_reason"] = reason
	order.Remark["admin_attention"] = true
	bidding.Status = orderEntity.LogisticsBiddingStatusClosed
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bidding).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "logistics_no_capacity", operatorID, enterpriseID, "sales", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"bidding_id":      bidding.ID,
			"reason":          reason,
			"admin_attention": true,
			"next_action":     "seller_recreate_logistics_bidding",
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("反馈物流无运力失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
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
		if err := createOrderEvent(tx, ctx, &order, "logistics_quote_submitted", operatorID, enterpriseID, "logistics", order.Status, order.Status, orderEntity.OrderEventPayload{
			"bidding_id": quote.BiddingID,
			"quote_id":   quote.ID,
			"amount":     quote.Amount,
		}); err != nil {
			return err
		}
		return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
			"物流方已提交报价",
			"物流方已为订单「"+order.OrderNo+"」提交报价，请及时查看并确认。",
			"/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "quote_id": quote.ID, "event": "logistics_quote_submitted"},
		)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("物流报价失败")
	}
	return toLogisticsQuoteInfo(quote), nil
}

func (s *OrderService) RespondConfirmLogistics(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusPendingLogisticsConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可确认物流选择")
	}
	db := database.GetDB()
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	if bidding.ResponderEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有响应方可以确认物流选择")
	}
	if bidding.ResponderConfirmed {
		return nil, ecode.ErrBadRequest.WithMessage("响应方已确认")
	}
	if bidding.SelectedQuoteID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("发起方尚未选定物流方")
	}

	bidding.ResponderConfirmed = true
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bidding).Error; err != nil {
			return err
		}
		actorService := "procurement"
		if bidding.InitiatorRole == "buyer" {
			actorService = "sales"
		}
		if err := createOrderEvent(tx, ctx, order, "logistics_responder_confirmed", operatorID, enterpriseID, actorService, order.Status, order.Status, orderEntity.OrderEventPayload{
			"selected_quote_id": bidding.SelectedQuoteID,
		}); err != nil {
			return err
		}
		return createEnterpriseNotification(tx, ctx, bidding.InitiatorEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
			"物流选择已确认",
			"订单「"+order.OrderNo+"」响应方已确认物流方选择，等待物流方确认接单。",
			"/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "event": "logistics_responder_confirmed"},
		)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认物流选择失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) ConfirmLogisticsQuote(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.ConfirmLogisticsRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
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
	if quote.LogisticsEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有被选中的物流公司才能确认接单")
	}
	if bidding.SelectedQuoteID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("发起方尚未选定物流方")
	}
	if !bidding.ResponderConfirmed {
		return nil, ecode.ErrBadRequest.WithMessage("响应方尚未确认物流选择，请等待确认后再接单")
	}
	fromStatus := order.Status
	bidding.SelectedQuoteID = req.QuoteID
	bidding.Status = orderEntity.LogisticsBiddingStatusSelected
	order.Status = orderEntity.OrderStatusPendingLogisticsContract
	order.UpdatedBy = operatorID

	existingDelivery := order.DeliveryInfo
	pickupCode := ""
	if existingDelivery != nil && existingDelivery.PickupCode != "" {
		pickupCode = existingDelivery.PickupCode
	} else {
		pickupCode = generatePickupCode()
	}

	deliveryInfo := &orderEntity.DeliveryInfo{
		DriverName:    quote.DriverInfo,
		PlateNumber:   quote.VehicleInfo,
		PickupCode:    pickupCode,
		Address:       bidding.DestinationAddress,
		EstimatedTime: quote.EstimatedDeliveryAt,
	}
	if existingDelivery != nil {
		deliveryInfo.WarehouseID = existingDelivery.WarehouseID
		deliveryInfo.WarehouseServiceID = existingDelivery.WarehouseServiceID
		deliveryInfo.WarehouseEnterpriseID = existingDelivery.WarehouseEnterpriseID
		deliveryInfo.WarehouseName = existingDelivery.WarehouseName
		deliveryInfo.WarehouseCode = existingDelivery.WarehouseCode
		deliveryInfo.WarehouseProvince = existingDelivery.WarehouseProvince
		deliveryInfo.WarehouseCity = existingDelivery.WarehouseCity
		deliveryInfo.WarehouseDistrict = existingDelivery.WarehouseDistrict
		deliveryInfo.WarehouseAddress = existingDelivery.WarehouseAddress
		if existingDelivery.Address != "" && deliveryInfo.Address == "" {
			deliveryInfo.Address = existingDelivery.Address
		}
		if existingDelivery.DriverPhone != "" {
			deliveryInfo.DriverPhone = existingDelivery.DriverPhone
		}
		if existingDelivery.DriverIdCard != "" {
			deliveryInfo.DriverIdCard = existingDelivery.DriverIdCard
		}
	}
	if deliveryInfo.WarehouseAddress == "" && bidding.PickupAddress != "" {
		deliveryInfo.WarehouseAddress = bidding.PickupAddress
	}
	order.DeliveryInfo = deliveryInfo

	isFinancingOrder := order.FinanceID > 0

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bidding).Error; err != nil {
			return err
		}
		contract := orderEntity.LogisticsContract{
			ContractNo:            generateLogisticsContractNo(),
			OrderID:               order.ID,
			OrderNo:               order.OrderNo,
			BiddingID:             bidding.ID,
			QuoteID:               quote.ID,
			BuyerEnterpriseID:     bidding.InitiatorEnterpriseID,
			LogisticsEnterpriseID: quote.LogisticsEnterpriseID,
			SellerEnterpriseID:    bidding.ResponderEnterpriseID,
			FreightAmount:         quote.Amount,
			CommissionRate:        0.03,
			PickupWarehouseID:     deliveryInfo.WarehouseServiceID,
			PickupWarehouseName:   deliveryInfo.WarehouseName,
			DestinationAddress:    deliveryInfo.Address,
			IsFinancingOrder:      isFinancingOrder,
			Status:                orderEntity.LogisticsContractPending,
			CreatedBy:             operatorID,
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// 物流商确认接单后立即创建物流订单，确保在物流流程中显示
		if err := ensureLogisticsOrderSyncTx(tx, order, &bidding, &quote, nil, operatorID); err != nil {
			return err
		}
		if err := createOrderEvent(tx, ctx, order, "logistics_contract_created", operatorID, enterpriseID, "logistics", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"quote_id":    req.QuoteID,
			"contract_id": contract.ID,
			"contract_no": contract.ContractNo,
		}); err != nil {
			return err
		}
		notifyBuyerContent := "订单「" + order.OrderNo + "」物流已确认接单，请签署物流合同后开始运输。"
		if err := createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
			"物流已确认，请签署物流合同",
			notifyBuyerContent,
			"/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "contract_id": contract.ID, "event": "logistics_contract_created"},
		); err != nil {
			return err
		}
		return createEnterpriseNotification(tx, ctx, order.SellerID, operatorID, notificationEntity.NotificationTypeOrder,
			"物流车队已确认",
			"订单「"+order.OrderNo+"」物流方已确认接单，买方和物流方正在签署物流合同。",
			"/orders",
			notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "contract_id": contract.ID, "event": "logistics_contract_created"},
		)
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认物流资源失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) GetLogisticsContract(ctx context.Context, orderID uint64) (*dto.LogisticsContractInfo, error) {
	db := database.GetDB()
	var contract orderEntity.LogisticsContract
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Order("id desc").First(&contract).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流合同不存在")
	}
	return toLogisticsContractInfo(&contract), nil
}

func (s *OrderService) InitiatorSignLogisticsContract(ctx context.Context, orderID, enterpriseID, operatorID uint64, req *dto.SignLogisticsContractRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusPendingLogisticsContract {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可签署物流合同")
	}

	db := database.GetDB()
	var contract orderEntity.LogisticsContract
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&contract).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流合同不存在")
	}
	var bidding orderEntity.LogisticsBidding
	db.WithContext(ctx).Where("id = ?", contract.BiddingID).First(&bidding)
	if bidding.InitiatorEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有竞价发起方可以签署物流合同")
	}
	if contract.BuyerSignedAt != nil {
		return nil, ecode.ErrBadRequest.WithMessage("发起方已签署")
	}

	now := time.Now()
	contract.BuyerSignedAt = &now
	if req.ContractURL != "" {
		contract.ContractURL = req.ContractURL
	}
	if contract.LogisticsSignedAt != nil {
		contract.Status = orderEntity.LogisticsContractCompleted
		contract.CompletedAt = &now
	} else {
		contract.Status = orderEntity.LogisticsContractBuyerSigned
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&contract).Error; err != nil {
			return err
		}
		if err := createOrderEvent(tx, ctx, order, "logistics_contract_buyer_signed", operatorID, enterpriseID, "procurement", order.Status, order.Status, orderEntity.OrderEventPayload{
			"contract_id": contract.ID,
			"contract_no": contract.ContractNo,
		}); err != nil {
			return err
		}
		if contract.Status == orderEntity.LogisticsContractCompleted {
			return s.completeLogisticsContractTx(tx, ctx, order, &contract, operatorID)
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("签署物流合同失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) LogisticsSignLogisticsContract(ctx context.Context, orderID, enterpriseID, operatorID uint64, req *dto.SignLogisticsContractRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusPendingLogisticsContract {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可签署物流合同")
	}

	db := database.GetDB()
	var contract orderEntity.LogisticsContract
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&contract).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流合同不存在")
	}
	if contract.LogisticsEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有物流方可以签署")
	}
	if contract.LogisticsSignedAt != nil {
		return nil, ecode.ErrBadRequest.WithMessage("物流方已签署")
	}

	now := time.Now()
	contract.LogisticsSignedAt = &now
	if req.ContractURL != "" {
		contract.ContractURL = req.ContractURL
	}
	if contract.BuyerSignedAt != nil {
		contract.Status = orderEntity.LogisticsContractCompleted
		contract.CompletedAt = &now
	} else {
		contract.Status = orderEntity.LogisticsContractLogisticsSigned
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&contract).Error; err != nil {
			return err
		}
		if err := createOrderEvent(tx, ctx, order, "logistics_contract_logistics_signed", operatorID, enterpriseID, "logistics", order.Status, order.Status, orderEntity.OrderEventPayload{
			"contract_id": contract.ID,
			"contract_no": contract.ContractNo,
		}); err != nil {
			return err
		}
		if contract.Status == orderEntity.LogisticsContractCompleted {
			return s.completeLogisticsContractTx(tx, ctx, order, &contract, operatorID)
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("签署物流合同失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) completeLogisticsContractTx(tx *gorm.DB, ctx context.Context, order *orderEntity.Order, contract *orderEntity.LogisticsContract, operatorID uint64) error {
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusShipping
	order.UpdatedBy = operatorID

	task := orderEntity.ShipmentTask{
		OrderID:               order.ID,
		OrderNo:               order.OrderNo,
		LogisticsEnterpriseID: contract.LogisticsEnterpriseID,
		LogisticsQuoteID:      contract.QuoteID,
		Status:                orderEntity.ShipmentStatusPendingPickup,
	}
	if err := tx.Where("order_id = ?", order.ID).Assign(task).FirstOrCreate(&task).Error; err != nil {
		return err
	}
	if err := tx.Save(order).Error; err != nil {
		return err
	}

	var bidding orderEntity.LogisticsBidding
	if err := tx.Where("id = ?", contract.BiddingID).First(&bidding).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	var quote orderEntity.LogisticsBidQuote
	if err := tx.Where("id = ?", contract.QuoteID).First(&quote).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err := ensureLogisticsOrderSyncTx(tx, order, &bidding, &quote, &task, operatorID); err != nil {
		return err
	}

	if err := createOrderEvent(tx, ctx, order, "logistics_contract_completed", operatorID, contract.LogisticsEnterpriseID, "logistics", fromStatus, order.Status, orderEntity.OrderEventPayload{
		"contract_id": contract.ID,
		"contract_no": contract.ContractNo,
	}); err != nil {
		return err
	}

	if err := createEnterpriseNotification(tx, ctx, contract.LogisticsEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
		"物流合同已签署完成",
		"订单「"+order.OrderNo+"」物流合同双方已签署完成，请安排提货运输。提货仓库："+contract.PickupWarehouseName,
		"/orders",
		notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "contract_id": contract.ID},
	); err != nil {
		return err
	}
	return createEnterpriseNotification(tx, ctx, order.SellerID, operatorID, notificationEntity.NotificationTypeOrder,
		"物流合同签署完成，运输即将开始",
		"订单「"+order.OrderNo+"」物流合同已签署完成，物流方即将安排提货。",
		"/orders",
		notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "contract_id": contract.ID},
	)
}

func toLogisticsContractInfo(c *orderEntity.LogisticsContract) *dto.LogisticsContractInfo {
	return &dto.LogisticsContractInfo{
		ID:                    c.ID,
		ContractNo:            c.ContractNo,
		OrderID:               c.OrderID,
		BiddingID:             c.BiddingID,
		QuoteID:               c.QuoteID,
		BuyerEnterpriseID:     c.BuyerEnterpriseID,
		LogisticsEnterpriseID: c.LogisticsEnterpriseID,
		SellerEnterpriseID:    c.SellerEnterpriseID,
		ContractURL:           c.ContractURL,
		FreightAmount:         c.FreightAmount,
		CommissionRate:        c.CommissionRate,
		PickupWarehouseID:     c.PickupWarehouseID,
		PickupWarehouseName:   c.PickupWarehouseName,
		DestinationAddress:    c.DestinationAddress,
		IsFinancingOrder:      c.IsFinancingOrder,
		Status:                string(c.Status),
		BuyerSignedAt:         c.BuyerSignedAt,
		LogisticsSignedAt:     c.LogisticsSignedAt,
		CompletedAt:           c.CompletedAt,
		CreatedAt:             c.CreatedAt,
	}
}

func (s *OrderService) CloseLogisticsBidding(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("只有卖家可以结束物流竞价")
	}
	if order.Status != orderEntity.OrderStatusLogisticsBidding {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可结束物流竞价")
	}
	db := database.GetDB()
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ? AND status = ?", order.ID, orderEntity.LogisticsBiddingStatusOpen).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	bidding.Status = orderEntity.LogisticsBiddingStatusSelected
	if err := db.WithContext(ctx).Save(&bidding).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("结束物流竞价失败")
	}
	order.Status = orderEntity.OrderStatusPendingLogisticsConfirm
	order.UpdatedBy = operatorID
	if err := db.WithContext(ctx).Save(order).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新订单状态失败")
	}
	err = createOrderEvent(db.WithContext(ctx), ctx, order, "logistics_bidding_closed", operatorID, enterpriseID, "sales", orderEntity.OrderStatusLogisticsBidding, order.Status, orderEntity.OrderEventPayload{
		"bidding_id": bidding.ID,
	})
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建事件失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) ListLogisticsQuotes(ctx context.Context, id, enterpriseID uint64) ([]*dto.LogisticsQuoteInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权查看该订单物流报价")
	}
	db := database.GetDB()
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	var quotes []orderEntity.LogisticsBidQuote
	if err := db.WithContext(ctx).Where("bidding_id = ?", bidding.ID).Order("amount asc").Find(&quotes).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询物流报价失败")
	}
	result := make([]*dto.LogisticsQuoteInfo, 0, len(quotes))
	for _, quote := range quotes {
		result = append(result, toLogisticsQuoteInfo(&quote))
	}
	return result, nil
}

func (s *OrderService) SelectLogisticsQuote(ctx context.Context, id, enterpriseID, operatorID uint64, req *dto.ConfirmLogisticsRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	isSeller := isOrderSeller(order, enterpriseID)
	isBuyer := order.BuyerID == enterpriseID
	if !isSeller && !isBuyer {
		return nil, ecode.ErrForbidden.WithMessage("只有买方或卖方可以选择物流报价")
	}
	if order.Status != orderEntity.OrderStatusLogisticsBidding && order.Status != orderEntity.OrderStatusPendingLogisticsConfirm {
		return nil, ecode.ErrBadRequest.WithMessage("当前订单状态不可选择物流报价")
	}
	db := database.GetDB()
	var bidding orderEntity.LogisticsBidding
	if err := db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流竞价不存在")
	}
	if bidding.InitiatorEnterpriseID != 0 && bidding.InitiatorEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有竞价发起方可以选定物流方")
	}
	var quote orderEntity.LogisticsBidQuote
	if err := db.WithContext(ctx).Where("id = ? AND bidding_id = ?", req.QuoteID, bidding.ID).First(&quote).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流报价不存在")
	}

	fromStatus := order.Status
	bidding.SelectedQuoteID = req.QuoteID
	if isSeller {
		bidding.SellerConfirmedQuoteID = req.QuoteID
	} else {
		bidding.BuyerConfirmedQuoteID = req.QuoteID
	}
	if order.Status == orderEntity.OrderStatusLogisticsBidding {
		order.Status = orderEntity.OrderStatusPendingLogisticsConfirm
	}
	order.UpdatedBy = operatorID

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&bidding).Error; err != nil {
			return err
		}
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("选择物流报价失败")
	}

	err = createOrderEvent(db.WithContext(ctx), ctx, order, "logistics_quote_selected", operatorID, enterpriseID, "sales", fromStatus, order.Status, orderEntity.OrderEventPayload{
		"quote_id":   req.QuoteID,
		"bidding_id": bidding.ID,
	})
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建事件失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) UpdateShipmentProgress(ctx context.Context, id, enterpriseID, operatorID uint64, action string, req *dto.ShipmentProgressRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusDelivering && order.Status != orderEntity.OrderStatusShipping && order.Status != orderEntity.OrderStatusDelivered {
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
		if err := ensureLogisticsOrderSyncTx(tx, order, nil, nil, &task, operatorID); err != nil {
			return err
		}
		if err := createOrderEvent(tx, ctx, order, eventType, operatorID, enterpriseID, "logistics", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"voucher_url": strings.TrimSpace(req.VoucherURL),
			"remark":      strings.TrimSpace(req.Remark),
		}); err != nil {
			return err
		}
		metadata := notificationEntity.NotificationMetadata{"order_id": order.ID, "order_no": order.OrderNo, "event": eventType}
		switch eventType {
		case "shipment_picked_up":
			if err := createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流已提货",
				"订单「"+order.OrderNo+"」物流方已提货，货物正在运输中。",
				"/dashboard/orders", metadata); err != nil {
				return err
			}
			return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流已提货",
				"订单「"+order.OrderNo+"」物流方已提货，货物正在运输中。",
				"/orders", metadata)
		case "shipment_delivered":
			if err := createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流已送达",
				"订单「"+order.OrderNo+"」物流方已送达，请及时确认收货。",
				"/dashboard/orders", metadata); err != nil {
				return err
			}
			return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流已送达",
				"订单「"+order.OrderNo+"」物流方已送达，等待买方确认收货。",
				"/orders", metadata)
		case "shipment_signature_uploaded":
			if err := createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
				"签收凭证已上传",
				"订单「"+order.OrderNo+"」物流方已上传签收凭证。",
				"/dashboard/orders", metadata); err != nil {
				return err
			}
			return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
				"签收凭证已上传",
				"订单「"+order.OrderNo+"」物流方已上传签收凭证。",
				"/orders", metadata)
		default:
			if err := createEnterpriseNotification(tx, ctx, order.BuyerID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流进度已更新",
				"订单「"+order.OrderNo+"」物流进度已更新。",
				"/dashboard/orders", metadata); err != nil {
				return err
			}
			return createEnterpriseNotification(tx, ctx, order.SellerEnterpriseID, operatorID, notificationEntity.NotificationTypeOrder,
				"物流进度已更新",
				"订单「"+order.OrderNo+"」物流进度已更新。",
				"/orders", metadata)
		}
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("回传运输进度失败")
	}
	return s.toOrderDetailInfo(ctx, order), nil
}

func (s *OrderService) GetLogisticsSettlement(ctx context.Context, orderID uint64) (*dto.LogisticsSettlementInfo, error) {
	db := database.GetDB()
	var settlement orderEntity.LogisticsSettlement
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Order("id desc").First(&settlement).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流结算单不存在")
	}
	return toLogisticsSettlementInfo(&settlement), nil
}

func (s *OrderService) PayLogisticsSettlement(ctx context.Context, orderID, enterpriseID, operatorID uint64, req *dto.PayLogisticsSettlementRequest) (*dto.LogisticsSettlementInfo, error) {
	db := database.GetDB()
	var settlement orderEntity.LogisticsSettlement
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Order("id desc").First(&settlement).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流结算单不存在")
	}
	if settlement.PayerEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有付款方可以支付")
	}
	if settlement.Status != orderEntity.LogisticsSettlementPending {
		return nil, ecode.ErrBadRequest.WithMessage("结算单状态不可支付")
	}

	now := time.Now()
	settlement.PaymentVoucherURL = req.VoucherURL
	settlement.PaidAt = &now
	settlement.Status = orderEntity.LogisticsSettlementPaid
	if err := db.WithContext(ctx).Save(&settlement).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("支付失败")
	}
	return toLogisticsSettlementInfo(&settlement), nil
}

func (s *OrderService) PayLogisticsCommission(ctx context.Context, orderID, enterpriseID, operatorID uint64, req *dto.PayLogisticsSettlementRequest) (*dto.LogisticsSettlementInfo, error) {
	db := database.GetDB()
	var settlement orderEntity.LogisticsSettlement
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Order("id desc").First(&settlement).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("物流结算单不存在")
	}
	if settlement.PayerEnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有付款方可以支付抽佣")
	}
	if settlement.Status != orderEntity.LogisticsSettlementPaid {
		return nil, ecode.ErrBadRequest.WithMessage("请先支付物流费用")
	}

	now := time.Now()
	settlement.CommissionVoucherURL = req.VoucherURL
	settlement.CommissionPaidAt = &now
	settlement.Status = orderEntity.LogisticsSettlementCommitted
	if err := db.WithContext(ctx).Save(&settlement).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("支付抽佣失败")
	}
	return toLogisticsSettlementInfo(&settlement), nil
}

func toLogisticsSettlementInfo(s *orderEntity.LogisticsSettlement) *dto.LogisticsSettlementInfo {
	return &dto.LogisticsSettlementInfo{
		ID:                   s.ID,
		SettlementNo:         s.SettlementNo,
		LogisticsContractID:  s.LogisticsContractID,
		OrderID:              s.OrderID,
		PayerEnterpriseID:    s.PayerEnterpriseID,
		PayeeEnterpriseID:    s.PayeeEnterpriseID,
		FreightAmount:        s.FreightAmount,
		PlatformCommission:   s.PlatformCommission,
		NetPayable:           s.NetPayable,
		PaymentVoucherURL:    s.PaymentVoucherURL,
		CommissionVoucherURL: s.CommissionVoucherURL,
		Status:               string(s.Status),
		PaidAt:               s.PaidAt,
		CommissionPaidAt:     s.CommissionPaidAt,
		CreatedAt:            s.CreatedAt,
	}
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
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) Archive(ctx context.Context, id, enterpriseID, operatorID uint64) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !isOrderSeller(order, enterpriseID) {
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
	return toOrderInfo(order, nil), nil
}

func isOrderSeller(order *orderEntity.Order, enterpriseID uint64) bool {
	if order == nil || enterpriseID == 0 {
		return false
	}
	return order.SellerID == enterpriseID || order.SellerEnterpriseID == enterpriseID
}

func toOrderInfo(order *orderEntity.Order, events []orderEntity.OrderEvent) *dto.OrderInfo {
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
		Timeline:       buildOrderTimeline(order, events),
		Remark:         map[string]interface{}(order.Remark),
	}
}

func (s *OrderService) toOrderDetailInfo(ctx context.Context, order *orderEntity.Order) *dto.OrderInfo {
	info := toOrderInfo(order, nil)
	if info == nil {
		return nil
	}
	db := database.GetDB().WithContext(ctx)

	var financing orderEntity.OrderFinancingRequest
	financingQuery := db.Where("order_id = ?", order.ID)
	if order.FinanceID > 0 {
		financingQuery = financingQuery.Where("id = ?", order.FinanceID)
	}
	if err := financingQuery.Order("id desc").First(&financing).Error; err == nil {
		info.FinancingInfo = toFinancingRequestInfo(&financing)

		var pledgeEvents []orderEntity.PledgeReleaseEvent
		if err := db.Where("order_id = ?", order.ID).Order("id asc").Find(&pledgeEvents).Error; err == nil {
			info.PledgeEvents = make([]dto.PledgeEventInfo, 0, len(pledgeEvents))
			for _, event := range pledgeEvents {
				info.PledgeEvents = append(info.PledgeEvents, toPledgeEventInfo(event))
			}
		}
	}

	var bidding orderEntity.LogisticsBidding
	if err := db.Where("order_id = ?", order.ID).Order("id desc").First(&bidding).Error; err == nil {
		info.LogisticsBidding = toLogisticsBiddingInfo(&bidding)

		var quotes []orderEntity.LogisticsBidQuote
		if err := db.Where("bidding_id = ?", bidding.ID).Order("amount asc, id asc").Find(&quotes).Error; err == nil {
			info.LogisticsQuotes = make([]dto.LogisticsQuoteInfo, 0, len(quotes))
			for _, quote := range quotes {
				info.LogisticsQuotes = append(info.LogisticsQuotes, *toLogisticsQuoteInfo(&quote))
			}
		}
	}

	var shipment orderEntity.ShipmentTask
	if err := db.Where("order_id = ?", order.ID).Order("id desc").First(&shipment).Error; err == nil {
		info.ShipmentTask = toShipmentTaskInfo(&shipment)
	}

	var logisticsContract orderEntity.LogisticsContract
	if err := db.Where("order_id = ?", order.ID).Order("id desc").First(&logisticsContract).Error; err == nil {
		info.LogisticsContractDTO = toLogisticsContractInfo(&logisticsContract)
	}

	var settlement orderEntity.LogisticsSettlement
	if err := db.Where("order_id = ?", order.ID).Order("id desc").First(&settlement).Error; err == nil {
		info.LogisticsSettlement = toLogisticsSettlementInfo(&settlement)
	}

	var orderEvents []orderEntity.OrderEvent
	if err := db.Where("order_id = ?", order.ID).Order("id asc").Find(&orderEvents).Error; err == nil {
		info.Timeline = buildOrderTimeline(order, orderEvents)
	}

	events, err := s.ListEvents(ctx, order.ID, 0, true)
	if err == nil {
		info.Events = events
	}
	return info
}

func buildOrderTimeline(order *orderEntity.Order, events []orderEntity.OrderEvent) []dto.TimelineItem {
	if order == nil {
		return nil
	}

	status := order.Status
	isCancelled := status == orderEntity.OrderStatusCancelled
	isException := status == orderEntity.OrderStatusException

	items := []dto.TimelineItem{
		{
			Key:         "created",
			Label:       "订单创建",
			Status:      "completed",
			Time:        &order.CreatedAt,
			Description: "买家已提交订单，系统锁定对应库存。",
		},
	}

	eventTimeMap := make(map[string]time.Time)
	for _, event := range events {
		eventTimeMap[event.EventType] = event.CreatedAt
	}

	skippedOrPending := func(reached bool) string {
		if isCancelled || isException {
			return "skipped"
		}
		if reached {
			return "completed"
		}
		return "pending"
	}

	getEventTime := func(eventType string) *time.Time {
		if t, ok := eventTimeMap[eventType]; ok {
			return &t
		}
		return &order.UpdatedAt
	}

	addItem := func(key, label, desc string, reached bool, eventType string) {
		s := skippedOrPending(reached)
		var t *time.Time
		if s == "completed" {
			t = getEventTime(eventType)
		}
		items = append(items, dto.TimelineItem{
			Key:         key,
			Label:       label,
			Status:      s,
			Time:        t,
			Description: desc,
		})
	}

	destWarehouseReached := orderStatusReached(status, orderEntity.OrderStatusPendingWarehouseConfirm)
	if status == orderEntity.OrderStatusPendingDestWarehouse {
		items = append(items, dto.TimelineItem{
			Key:         "dest_warehouse_select",
			Label:       "选择目的地仓储",
			Status:      "current",
			Description: "请选择物流送达的目的地仓库。",
		})
	} else {
		addItem("dest_warehouse_select", "选择目的地仓储", "买家已选择目的地仓库。", destWarehouseReached, "dest_warehouse_selected")
	}

	warehouseReached := orderStatusReached(status, orderEntity.OrderStatusPendingSellerConfirm)
	if status == orderEntity.OrderStatusPendingWarehouseReselect {
		items = append(items, dto.TimelineItem{
			Key:         "warehouse_reselect",
			Label:       "重新选择仓储方",
			Status:      "current",
			Description: "仓储方无法承接，已反馈平台管理端，请买方重新选择仓储方。",
		})
	} else if status == orderEntity.OrderStatusPendingWarehouseConfirm {
		items = append(items, dto.TimelineItem{
			Key:         "warehouse_confirm",
			Label:       "仓储方确认仓库条件",
			Status:      "current",
			Description: "等待仓储方核验仓库条件并反馈。",
		})
	} else {
		addItem("warehouse_confirm", "仓储方确认仓库条件", "仓储方已确认仓库条件。", warehouseReached, "warehouse_confirmed")
	}

	sellerReached := orderStatusReached(status, orderEntity.OrderStatusPendingBuyerConfirm)
	if status == orderEntity.OrderStatusPendingSellerConfirm {
		items = append(items, dto.TimelineItem{
			Key:         "seller_confirm",
			Label:       "卖方确认条款",
			Status:      "current",
			Description: "等待卖方确认价格、数量和交付条款。",
		})
	} else {
		addItem("seller_confirm", "卖方确认条款", "卖方已确认订单条款。", sellerReached, "seller_confirmed")
	}

	buyerReached := orderStatusReached(status, orderEntity.OrderStatusPendingContract)
	if status == orderEntity.OrderStatusPendingBuyerConfirm {
		items = append(items, dto.TimelineItem{
			Key:         "buyer_confirm",
			Label:       "买方最终确认",
			Status:      "current",
			Description: "等待买方确认订单最终版本。",
		})
	} else {
		addItem("buyer_confirm", "买方最终确认", "买方已确认订单。", buyerReached, "buyer_confirmed")
	}

	contractReached := orderStatusReached(status, orderEntity.OrderStatusPendingPayment)
	if status == orderEntity.OrderStatusPendingContract {
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
	} else {
		addItem("contract_signed", "双方签署合同", "买卖双方已签署合同。", contractReached, "contract_signed")
	}

	paymentReached := orderStatusReached(status, orderEntity.OrderStatusPaid)
	if status == orderEntity.OrderStatusPendingPayment {
		items = append(items, dto.TimelineItem{
			Key:         "payment",
			Label:       "支付/融资选择",
			Status:      "current",
			Description: "当前待支付，请上传支付凭证或申请托盘融资。",
		})
	} else if status == orderEntity.OrderStatusFinancingReview || status == orderEntity.OrderStatusPendingFinancingContract {
		items = append(items, dto.TimelineItem{
			Key:         "payment",
			Label:       "已申请融资",
			Status:      "completed",
			Description: "买方已选择托盘融资方式。",
		})
	} else {
		paymentDesc := "已完成支付。"
		if !paymentReached && !isCancelled && !isException {
			paymentDesc = "等待买方支付或申请托盘融资。"
		}
		addItem("payment", "支付/融资", paymentDesc, paymentReached, "direct_payment_confirmed")
	}

	if status == orderEntity.OrderStatusFinancingReview || status == orderEntity.OrderStatusPendingFinancingContract || orderStatusReached(status, orderEntity.OrderStatusPaid) {
		financingReached := orderStatusReached(status, orderEntity.OrderStatusPaid)
		if status == orderEntity.OrderStatusFinancingReview {
			items = append(items, dto.TimelineItem{
				Key:         "financing_review",
				Label:       "风控审批中",
				Status:      "current",
				Description: "资金方正在进行风控审批，请耐心等待。",
			})
		} else if status == orderEntity.OrderStatusPendingFinancingContract {
			items = append(items, dto.TimelineItem{
				Key:         "financing_review",
				Label:       "风控审批通过",
				Status:      "completed",
				Description: "资金方风控审批已通过。",
			})
			items = append(items, dto.TimelineItem{
				Key:         "financing_contract",
				Label:       "签署融资合同并支付手续费",
				Status:      "current",
				Description: "请签署债务合同并上传手续费支付凭证。",
			})
		} else if financingReached && order.FinanceID > 0 {
			hasEvent := func(eventType string) bool {
				for _, e := range events {
					if e.EventType == eventType {
						return true
					}
				}
				return false
			}

			riskApproved := hasEvent("financing_risk_approved")
			contractGenerated := hasEvent("financing_debt_contract_generated")
			feePaid := hasEvent("financing_buyer_signed_fee_paid")
			disbursed := hasEvent("financing_disbursed")
			pledged := hasEvent("pledge_recorded")

			if riskApproved {
				items = append(items, dto.TimelineItem{Key: "financing_review", Label: "风控审批通过", Status: "completed", Description: "资金方风控审批已通过。"})
			}
			if contractGenerated {
				items = append(items, dto.TimelineItem{Key: "financing_contract", Label: "融资合同已生成", Status: "completed", Description: "资金方已生成债务合同。"})
			} else if riskApproved {
				items = append(items, dto.TimelineItem{Key: "financing_contract", Label: "待资金方生成合同", Status: "current", Description: "等待资金方生成债务合同。"})
			}
			if feePaid {
				items = append(items, dto.TimelineItem{Key: "financing_fee_paid", Label: "合同已签署·手续费已付", Status: "completed", Description: "买方已签署债务合同并支付手续费。"})
			} else if contractGenerated {
				items = append(items, dto.TimelineItem{Key: "financing_fee_paid", Label: "待买方签署并支付手续费", Status: "current", Description: "请买方签署债务合同并上传手续费支付凭证。"})
			}
			if disbursed {
				items = append(items, dto.TimelineItem{Key: "financing_disburse", Label: "资金方已放款", Status: "completed", Description: "资金方已完成放款。"})
			} else if feePaid {
				items = append(items, dto.TimelineItem{Key: "financing_disburse", Label: "待资金方放款", Status: "current", Description: "等待资金方确认放款。"})
			}
			if pledged {
				items = append(items, dto.TimelineItem{Key: "financing_pledge", Label: "货权质押已记录", Status: "completed", Description: "资金方已记录货权质押，可进入物流履约。"})
			} else if disbursed {
				items = append(items, dto.TimelineItem{Key: "financing_pledge", Label: "待记录货权质押", Status: "current", Description: "资金方需记录货权质押后，卖方才能发起物流。"})
			}
		} else if financingReached {
			addItem("financing", "托盘融资", "融资已完成。", financingReached, "financing_disbursed")
		}
	}

	logisticsReached := orderStatusReached(status, orderEntity.OrderStatusShipping)
	if status == orderEntity.OrderStatusLogisticsBidding {
		items = append(items, dto.TimelineItem{
			Key:         "logistics_bidding",
			Label:       "物流竞价",
			Status:      "current",
			Description: "卖方已发起物流竞价，等待物流方报价。",
		})
	} else if status == orderEntity.OrderStatusPendingLogisticsRebid {
		items = append(items, dto.TimelineItem{
			Key:         "logistics_rebid",
			Label:       "重新发起物流",
			Status:      "current",
			Description: "当前无合规运力，已反馈平台管理端，请卖方重新发起物流竞价。",
		})
	} else if status == orderEntity.OrderStatusPendingLogisticsConfirm {
		items = append(items, dto.TimelineItem{
			Key:         "logistics_confirm",
			Label:       "确认物流资源",
			Status:      "current",
			Description: "等待买卖双方确认物流报价。",
		})
	} else if status == orderEntity.OrderStatusPendingLogisticsContract {
		addItem("logistics", "物流竞价与确认", "物流资源已确认，运输任务已下发。", true, "logistics_quote_confirmed")
		items = append(items, dto.TimelineItem{
			Key:         "logistics_contract",
			Label:       "签署物流合同",
			Status:      "current",
			Description: "物流方已接单，等待签署物流合同。",
		})
	} else {
		addItem("logistics", "物流竞价与确认", "物流资源已确认，运输任务已下发。", logisticsReached, "logistics_quote_confirmed")
	}

	shippingReached := orderStatusReached(status, orderEntity.OrderStatusDelivered)
	if status == orderEntity.OrderStatusShipping {
		items = append(items, dto.TimelineItem{
			Key:         "shipping",
			Label:       "运输中",
			Status:      "current",
			Description: "物流方已提货，正在运输途中。",
		})
	} else if status == orderEntity.OrderStatusDelivered {
		items = append(items, dto.TimelineItem{
			Key:         "shipping",
			Label:       "运输中",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "物流方已送达，等待买方签收确认。",
		})
	} else {
		addItem("shipping", "运输交付", "物流方已送达货物。", shippingReached, "logistics_delivered")
	}

	receiptReached := orderStatusReached(status, orderEntity.OrderStatusReceived)
	if status == orderEntity.OrderStatusDelivered {
		items = append(items, dto.TimelineItem{
			Key:         "receipt",
			Label:       "买家确认收货",
			Status:      "current",
			Description: "等待买家确认收货并上传签收凭证。",
		})
	} else {
		receiptDesc := "买家已确认收货。"
		if !receiptReached && !isCancelled && !isException {
			receiptDesc = "等待买家确认收货。"
		}
		addItem("receipt", "买家确认收货", receiptDesc, receiptReached, "buyer_received")
		if receiptReached {
			items[len(items)-1].Time = order.ReceiptAt
		}
	}

	reconciledReached := orderStatusReached(status, orderEntity.OrderStatusReconciled)
	if status == orderEntity.OrderStatusReceived {
		items = append(items, dto.TimelineItem{
			Key:         "reconciled",
			Label:       "卖方对账确认",
			Status:      "current",
			Description: "等待卖方完成收款对账确认。",
		})
	} else {
		addItem("reconciled", "卖方对账确认", "卖方已完成收款对账确认。", reconciledReached, "seller_reconciled")
	}

	archivedReached := status == orderEntity.OrderStatusArchived
	if status == orderEntity.OrderStatusReconciled {
		items = append(items, dto.TimelineItem{
			Key:         "archived",
			Label:       "订单归档",
			Status:      "current",
			Description: "对账完成，等待归档。",
		})
	} else {
		addItem("archived", "订单归档", "订单已归档。", archivedReached, "order_archived")
	}

	if isCancelled {
		items = append(items, dto.TimelineItem{
			Key:         "closed",
			Label:       "订单关闭",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "订单已关闭，锁定库存已释放。",
		})
	}

	if isException {
		items = append(items, dto.TimelineItem{
			Key:         "exception",
			Label:       "订单异常",
			Status:      "completed",
			Time:        &order.UpdatedAt,
			Description: "订单进入异常状态，等待平台运营介入。",
		})
	}

	return items
}

func orderStatusReached(current, target orderEntity.OrderStatus) bool {
	return orderEntity.OrderStatusReached(current, target)
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

func generateLogisticsContractNo() string {
	return fmt.Sprintf("LC%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
}

func generateSettlementNo() string {
	return fmt.Sprintf("LS%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%10000)
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

func (s *OrderService) ensureTitleControlReadyForLogistics(ctx context.Context, order *orderEntity.Order) error {
	if order == nil {
		return nil
	}
	var financing orderEntity.OrderFinancingRequest
	query := database.GetDB().WithContext(ctx).Where("order_id = ?", order.ID)
	if order.FinanceID > 0 {
		query = query.Where("id = ?", order.FinanceID)
	}
	err := query.Order("id desc").First(&financing).Error
	if err == nil {
		if financing.Status != orderEntity.FinancingRequestStatusPledged && financing.Status != orderEntity.FinancingRequestStatusReleased {
			return ecode.ErrBadRequest.WithMessage("托盘垫资订单必须先记录货权质押，才能进入物流履约")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ecode.ErrDatabase.WithMessage("查询融资质押状态失败")
	}
	if order.PaymentInfo != nil && order.PaymentInfo.VoucherURL != "" {
		if len(order.Items) > 0 {
			product, prodErr := s.productRepo.FindByID(ctx, order.Items[0].ProductID)
			if prodErr != nil || product == nil || strings.TrimSpace(product.TitleVoucherURL) == "" {
				if order.Remark == nil || strings.TrimSpace(fmt.Sprint(order.Remark["direct_title_voucher_url"])) == "" {
					return ecode.ErrBadRequest.WithMessage("直采订单必须先记录货权/仓单监管凭证，才能进入物流履约。请确认卖方已在商品上架时认证货权凭证")
				}
			}
		} else if order.Remark == nil || strings.TrimSpace(fmt.Sprint(order.Remark["direct_title_voucher_url"])) == "" {
			return ecode.ErrBadRequest.WithMessage("直采订单必须先记录货权/仓单监管凭证，才能进入物流履约")
		}
	}
	return nil
}

func (s *OrderService) ListEvents(ctx context.Context, id, enterpriseID uint64, allowAll bool) ([]dto.OrderEventInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if !allowAll && order.BuyerID != enterpriseID && !isOrderSeller(order, enterpriseID) && warehouseEnterpriseIDOf(order) != enterpriseID {
		funding, fundingErr := s.isFundingParticipant(ctx, order.ID, enterpriseID)
		if fundingErr != nil {
			return nil, fundingErr
		}
		if !funding {
			return nil, ecode.ErrForbidden.WithMessage("无权查看该订单事件")
		}
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
		FreightPayer:           bidding.FreightPayer,
		InitiatorEnterpriseID:  bidding.InitiatorEnterpriseID,
		InitiatorRole:          bidding.InitiatorRole,
		ResponderEnterpriseID:  bidding.ResponderEnterpriseID,
		ResponderConfirmed:     bidding.ResponderConfirmed,
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

func toShipmentTaskInfo(task *orderEntity.ShipmentTask) *dto.ShipmentTaskInfo {
	if task == nil {
		return nil
	}
	return &dto.ShipmentTaskInfo{
		ID:                    task.ID,
		OrderID:               task.OrderID,
		OrderNo:               task.OrderNo,
		LogisticsEnterpriseID: task.LogisticsEnterpriseID,
		LogisticsQuoteID:      task.LogisticsQuoteID,
		Status:                string(task.Status),
		PickupVoucher:         task.PickupVoucher,
		DeliveryVoucher:       task.DeliveryVoucher,
		SignatureVoucher:      task.SignatureVoucher,
		PickedUpAt:            task.PickedUpAt,
		DeliveredAt:           task.DeliveredAt,
		SignedAt:              task.SignedAt,
		CreatedAt:             task.CreatedAt,
		UpdatedAt:             task.UpdatedAt,
	}
}

func toPledgeEventInfo(event orderEntity.PledgeReleaseEvent) dto.PledgeEventInfo {
	return dto.PledgeEventInfo{
		ID:                 event.ID,
		FinancingRequestID: event.FinancingRequestID,
		OrderID:            event.OrderID,
		OrderNo:            event.OrderNo,
		EventType:          event.EventType,
		Condition:          event.Condition,
		Reason:             event.Reason,
		VoucherURL:         event.VoucherURL,
		ActorUserID:        event.ActorUserID,
		ActorEnterpriseID:  event.ActorEnterpriseID,
		CreatedAt:          event.CreatedAt,
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
		FundingCommission:   item.FundingCommission,
		CommissionVoucher:   item.CommissionVoucher,
		CommissionPaidAt:    item.CommissionPaidAt,
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

func (s *OrderService) ReportBreach(ctx context.Context, orderID, reporterEnterpriseID, operatorID uint64, req *dto.ReportBreachRequest) (*dto.BreachRecordInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.BuyerID != reporterEnterpriseID && !isOrderSeller(order, reporterEnterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权对该订单发起违约举报")
	}
	if order.Status < orderEntity.OrderStatusPendingContract {
		return nil, ecode.ErrBadRequest.WithMessage("订单尚未签署合同，不可发起违约举报")
	}
	if order.Status == orderEntity.OrderStatusCancelled || order.Status == orderEntity.OrderStatusArchived {
		return nil, ecode.ErrBadRequest.WithMessage("订单已终止或归档，不可发起违约举报")
	}

	var breachEnterpriseID uint64
	breachParty := orderEntity.BreachParty(req.BreachParty)
	switch breachParty {
	case orderEntity.BreachPartyBuyer:
		breachEnterpriseID = order.BuyerID
	case orderEntity.BreachPartySeller:
		breachEnterpriseID = order.SellerID
	default:
		return nil, ecode.ErrBadRequest.WithMessage("违约方参数无效")
	}

	evidenceURLs := orderEntity.StringList{}
	for _, u := range req.EvidenceURLs {
		if u != "" {
			evidenceURLs = append(evidenceURLs, strings.TrimSpace(u))
		}
	}

	record := &orderEntity.BreachRecord{
		OrderID:              order.ID,
		OrderNo:              order.OrderNo,
		ReporterEnterpriseID: reporterEnterpriseID,
		BreachParty:          breachParty,
		BreachEnterpriseID:   breachEnterpriseID,
		BreachType:           strings.TrimSpace(req.BreachType),
		Description:          strings.TrimSpace(req.Description),
		EvidenceURLs:         evidenceURLs,
		Status:               orderEntity.BreachRecordStatusSubmitted,
		CreatedBy:            operatorID,
		UpdatedBy:            operatorID,
	}

	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusException
	order.UpdatedBy = operatorID

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "breach_reported", operatorID, reporterEnterpriseID, actorServiceForOrder(order, reporterEnterpriseID), fromStatus, order.Status, orderEntity.OrderEventPayload{
			"breach_record_id":     record.ID,
			"breach_party":         string(breachParty),
			"breach_enterprise_id": breachEnterpriseID,
			"breach_type":          record.BreachType,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("提交违约举报失败")
	}

	return toBreachRecordInfo(record), nil
}

func (s *OrderService) ListBreachRecords(ctx context.Context, orderID uint64) ([]*dto.BreachRecordInfo, error) {
	var records []orderEntity.BreachRecord
	if err := database.GetDB().WithContext(ctx).Where("order_id = ?", orderID).Order("id DESC").Find(&records).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询违约记录失败")
	}
	result := make([]*dto.BreachRecordInfo, 0, len(records))
	for i := range records {
		result = append(result, toBreachRecordInfo(&records[i]))
	}
	return result, nil
}

func (s *OrderService) MediateBreach(ctx context.Context, breachRecordID, operatorID uint64, req *dto.MediateBreachRequest) (*dto.BreachRecordInfo, error) {
	var record orderEntity.BreachRecord
	if err := database.GetDB().WithContext(ctx).First(&record, breachRecordID).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("违约记录不存在")
	}
	if record.Status != orderEntity.BreachRecordStatusSubmitted && record.Status != orderEntity.BreachRecordStatusUnderReview {
		return nil, ecode.ErrBadRequest.WithMessage("当前违约记录状态不可发起协调")
	}
	now := time.Now()
	record.Status = orderEntity.BreachRecordStatusMediating
	record.MediationResult = strings.TrimSpace(req.MediationResult)
	record.MediatedBy = operatorID
	record.MediatedAt = &now
	record.UpdatedBy = operatorID

	order, _ := s.orderRepo.FindByID(ctx, record.OrderID)

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		if order != nil {
			return createOrderEvent(tx, ctx, order, "breach_mediation_started", operatorID, 0, "admin", order.Status, order.Status, orderEntity.OrderEventPayload{
				"breach_record_id": record.ID,
				"mediation_result": record.MediationResult,
			})
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新协调结果失败")
	}

	return toBreachRecordInfo(&record), nil
}

func (s *OrderService) ResolveBreach(ctx context.Context, breachRecordID, operatorID uint64, req *dto.ResolveBreachRequest) (*dto.BreachRecordInfo, error) {
	var record orderEntity.BreachRecord
	if err := database.GetDB().WithContext(ctx).First(&record, breachRecordID).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("违约记录不存在")
	}
	if record.Status != orderEntity.BreachRecordStatusMediating && record.Status != orderEntity.BreachRecordStatusUnderReview && record.Status != orderEntity.BreachRecordStatusSubmitted {
		return nil, ecode.ErrBadRequest.WithMessage("当前违约记录状态不可处理")
	}
	now := time.Now()
	record.Status = orderEntity.BreachRecordStatusResolved
	record.ResolutionResult = strings.TrimSpace(req.ResolutionResult)
	record.ResolvedBy = operatorID
	record.ResolvedAt = &now
	record.UpdatedBy = operatorID

	order, _ := s.orderRepo.FindByID(ctx, record.OrderID)

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		if order != nil && order.Status == orderEntity.OrderStatusException {
			order.Status = orderEntity.OrderStatusFrozen
			order.UpdatedBy = operatorID
			if err := tx.Save(order).Error; err != nil {
				return err
			}
			return createOrderEvent(tx, ctx, order, "breach_resolved", operatorID, 0, "admin", orderEntity.OrderStatusException, order.Status, orderEntity.OrderEventPayload{
				"breach_record_id":  record.ID,
				"resolution_result": record.ResolutionResult,
			})
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("处理违约记录失败")
	}

	return toBreachRecordInfo(&record), nil
}

func (s *OrderService) FreezeOrder(ctx context.Context, orderID, operatorID uint64, req *dto.FreezeOrderRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status == orderEntity.OrderStatusFrozen {
		return nil, ecode.ErrBadRequest.WithMessage("订单已冻结")
	}
	if order.Status == orderEntity.OrderStatusCancelled || order.Status == orderEntity.OrderStatusArchived {
		return nil, ecode.ErrBadRequest.WithMessage("订单已终止或归档，不可冻结")
	}
	fromStatus := order.Status
	order.Status = orderEntity.OrderStatusFrozen
	order.UpdatedBy = operatorID

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "order_frozen", operatorID, 0, "admin", fromStatus, order.Status, orderEntity.OrderEventPayload{
			"reason": strings.TrimSpace(req.Reason),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("冻结订单失败")
	}
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) UnfreezeOrder(ctx context.Context, orderID, operatorID uint64, req *dto.UnfreezeOrderRequest) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("订单不存在")
	}
	if order.Status != orderEntity.OrderStatusFrozen {
		return nil, ecode.ErrBadRequest.WithMessage("订单当前未冻结，不可解冻")
	}

	var lastEvent orderEntity.OrderEvent
	if err := database.GetDB().WithContext(ctx).Where("order_id = ? AND event_type = ?", order.ID, "order_frozen").Order("id DESC").First(&lastEvent).Error; err == nil {
		if lastEvent.FromStatus >= orderEntity.OrderStatusPendingContract {
			order.Status = lastEvent.FromStatus
		} else {
			order.Status = orderEntity.OrderStatusException
		}
	} else {
		order.Status = orderEntity.OrderStatusException
	}
	order.UpdatedBy = operatorID

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "order_unfrozen", operatorID, 0, "admin", orderEntity.OrderStatusFrozen, order.Status, orderEntity.OrderEventPayload{
			"reason": strings.TrimSpace(req.Reason),
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("解冻订单失败")
	}
	return toOrderInfo(order, nil), nil
}

func (s *OrderService) CreateLitigation(ctx context.Context, breachRecordID, plaintiffEnterpriseID, operatorID uint64, req *dto.CreateLitigationRequest) (*dto.LitigationRecordInfo, error) {
	var breachRecord orderEntity.BreachRecord
	if err := database.GetDB().WithContext(ctx).First(&breachRecord, breachRecordID).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("违约记录不存在")
	}
	order, err := s.orderRepo.FindByID(ctx, breachRecord.OrderID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if order == nil {
		return nil, ecode.ErrNotFound.WithMessage("关联订单不存在")
	}
	if order.BuyerID != plaintiffEnterpriseID && !isOrderSeller(order, plaintiffEnterpriseID) {
		return nil, ecode.ErrForbidden.WithMessage("无权对该违约记录发起诉讼")
	}

	var defendantEnterpriseID uint64
	if breachRecord.BreachParty == orderEntity.BreachPartyBuyer {
		defendantEnterpriseID = order.BuyerID
	} else {
		defendantEnterpriseID = order.SellerID
	}

	evidenceURLs := orderEntity.StringList{}
	for _, u := range req.EvidenceURLs {
		if u != "" {
			evidenceURLs = append(evidenceURLs, strings.TrimSpace(u))
		}
	}

	litigation := &orderEntity.LitigationRecord{
		BreachRecordID:        breachRecordID,
		OrderID:               order.ID,
		OrderNo:               order.OrderNo,
		PlaintiffEnterpriseID: plaintiffEnterpriseID,
		DefendantEnterpriseID: defendantEnterpriseID,
		CaseNo:                strings.TrimSpace(req.CaseNo),
		Court:                 strings.TrimSpace(req.Court),
		ClaimAmount:           req.ClaimAmount,
		Description:           strings.TrimSpace(req.Description),
		EvidenceURLs:          evidenceURLs,
		Status:                orderEntity.LitigationStatusFiled,
		CreatedBy:             operatorID,
		UpdatedBy:             operatorID,
	}

	fromBreachStatus := breachRecord.Status
	breachRecord.Status = orderEntity.BreachRecordStatusLitigating
	breachRecord.UpdatedBy = operatorID

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(litigation).Error; err != nil {
			return err
		}
		if err := tx.Save(&breachRecord).Error; err != nil {
			return err
		}
		return createOrderEvent(tx, ctx, order, "litigation_filed", operatorID, plaintiffEnterpriseID, actorServiceForOrder(order, plaintiffEnterpriseID), order.Status, order.Status, orderEntity.OrderEventPayload{
			"litigation_id":      litigation.ID,
			"breach_record_id":   breachRecordID,
			"from_breach_status": string(fromBreachStatus),
			"to_breach_status":   string(breachRecord.Status),
			"claim_amount":       litigation.ClaimAmount,
		})
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("发起诉讼失败")
	}

	return toLitigationRecordInfo(litigation), nil
}

func (s *OrderService) UpdateLitigationStatus(ctx context.Context, litigationID, operatorID uint64, req *dto.UpdateLitigationStatusRequest) (*dto.LitigationRecordInfo, error) {
	var litigation orderEntity.LitigationRecord
	if err := database.GetDB().WithContext(ctx).First(&litigation, litigationID).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("诉讼记录不存在")
	}

	fromStatus := litigation.Status
	litigation.Status = orderEntity.LitigationStatus(req.Status)
	litigation.UpdatedBy = operatorID

	if req.CaseNo != "" {
		litigation.CaseNo = strings.TrimSpace(req.CaseNo)
	}
	if req.Court != "" {
		litigation.Court = strings.TrimSpace(req.Court)
	}
	if req.JudgmentResult != "" {
		litigation.JudgmentResult = strings.TrimSpace(req.JudgmentResult)
	}
	if req.JudgmentURL != "" {
		litigation.JudgmentURL = strings.TrimSpace(req.JudgmentURL)
	}

	now := time.Now()
	switch litigation.Status {
	case orderEntity.LitigationStatusJudged:
		litigation.JudgedAt = &now
	case orderEntity.LitigationStatusClosed:
		litigation.ClosedAt = &now
		if litigation.JudgedAt == nil {
			litigation.JudgedAt = &now
		}
	}

	order, _ := s.orderRepo.FindByID(ctx, litigation.OrderID)

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&litigation).Error; err != nil {
			return err
		}
		if order != nil {
			eventPayload := orderEntity.OrderEventPayload{
				"litigation_id": litigation.ID,
				"from_status":   string(fromStatus),
				"to_status":     string(litigation.Status),
			}
			if litigation.Status == orderEntity.LitigationStatusJudged {
				eventPayload["judgment_result"] = litigation.JudgmentResult
			}
			return createOrderEvent(tx, ctx, order, "litigation_status_updated", operatorID, 0, "admin", order.Status, order.Status, eventPayload)
		}
		return nil
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新诉讼状态失败")
	}

	if litigation.Status == orderEntity.LitigationStatusClosed && order != nil {
		var breachRecord orderEntity.BreachRecord
		if err := database.GetDB().WithContext(ctx).First(&breachRecord, litigation.BreachRecordID).Error; err == nil {
			now2 := time.Now()
			breachRecord.Status = orderEntity.BreachRecordStatusResolved
			breachRecord.ResolvedBy = operatorID
			breachRecord.ResolvedAt = &now2
			breachRecord.ResolutionResult = "诉讼结案: " + litigation.JudgmentResult
			breachRecord.UpdatedBy = operatorID
			database.GetDB().WithContext(ctx).Save(&breachRecord)
		}

		if order.Status == orderEntity.OrderStatusFrozen || order.Status == orderEntity.OrderStatusException {
			order.Status = orderEntity.OrderStatusCancelled
			order.UpdatedBy = operatorID
			database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := tx.Save(order).Error; err != nil {
					return err
				}
				return createOrderEvent(tx, ctx, order, "order_cancelled_after_litigation", operatorID, 0, "admin", orderEntity.OrderStatusException, order.Status, orderEntity.OrderEventPayload{
					"litigation_id": litigation.ID,
				})
			})
		}
	}

	return toLitigationRecordInfo(&litigation), nil
}

func (s *OrderService) ListLitigationRecords(ctx context.Context, orderID uint64) ([]*dto.LitigationRecordInfo, error) {
	var records []orderEntity.LitigationRecord
	if err := database.GetDB().WithContext(ctx).Where("order_id = ?", orderID).Order("id DESC").Find(&records).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询诉讼记录失败")
	}
	result := make([]*dto.LitigationRecordInfo, 0, len(records))
	for i := range records {
		result = append(result, toLitigationRecordInfo(&records[i]))
	}
	return result, nil
}

func toBreachRecordInfo(r *orderEntity.BreachRecord) *dto.BreachRecordInfo {
	if r == nil {
		return nil
	}
	evidenceURLs := []string{}
	if r.EvidenceURLs != nil {
		evidenceURLs = []string(r.EvidenceURLs)
	}
	return &dto.BreachRecordInfo{
		ID:                   r.ID,
		OrderID:              r.OrderID,
		OrderNo:              r.OrderNo,
		ReporterEnterpriseID: r.ReporterEnterpriseID,
		BreachParty:          string(r.BreachParty),
		BreachEnterpriseID:   r.BreachEnterpriseID,
		BreachType:           r.BreachType,
		Description:          r.Description,
		EvidenceURLs:         evidenceURLs,
		Status:               string(r.Status),
		MediationResult:      r.MediationResult,
		MediatedBy:           r.MediatedBy,
		MediatedAt:           r.MediatedAt,
		ResolutionResult:     r.ResolutionResult,
		ResolvedBy:           r.ResolvedBy,
		ResolvedAt:           r.ResolvedAt,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

func toLitigationRecordInfo(r *orderEntity.LitigationRecord) *dto.LitigationRecordInfo {
	if r == nil {
		return nil
	}
	evidenceURLs := []string{}
	if r.EvidenceURLs != nil {
		evidenceURLs = []string(r.EvidenceURLs)
	}
	return &dto.LitigationRecordInfo{
		ID:                    r.ID,
		BreachRecordID:        r.BreachRecordID,
		OrderID:               r.OrderID,
		OrderNo:               r.OrderNo,
		PlaintiffEnterpriseID: r.PlaintiffEnterpriseID,
		DefendantEnterpriseID: r.DefendantEnterpriseID,
		CaseNo:                r.CaseNo,
		Court:                 r.Court,
		ClaimAmount:           r.ClaimAmount,
		Description:           r.Description,
		EvidenceURLs:          evidenceURLs,
		Status:                string(r.Status),
		JudgmentResult:        r.JudgmentResult,
		JudgmentURL:           r.JudgmentURL,
		JudgedAt:              r.JudgedAt,
		ClosedAt:              r.ClosedAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}
