package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tradeentity "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/domain/entity"
	tradedto "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/ui/dto"
	userentity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

type TradeService struct {
	db *gorm.DB
}

func NewTradeService() *TradeService {
	return &TradeService{db: database.GetDB()}
}

func (s *TradeService) ListPublicProducts(ctx context.Context, req *tradedto.ProductListRequest) (*tradedto.ProductListResponse, error) {
	return s.listProducts(ctx, req, false, 0)
}

func (s *TradeService) ListSellerProducts(ctx context.Context, sellerID uint64, req *tradedto.ProductListRequest) (*tradedto.ProductListResponse, error) {
	return s.listProducts(ctx, req, true, sellerID)
}

func (s *TradeService) ListAdminProducts(ctx context.Context, req *tradedto.ProductListRequest) (*tradedto.ProductListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var products []tradeentity.Product
	var total int64

	query := s.db.WithContext(ctx).Model(&tradeentity.Product{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR sku_id LIKE ? OR seller_uid LIKE ?", like, like, like)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&products).Error; err != nil {
		return nil, err
	}

	items, err := s.toProductItems(ctx, products, true)
	if err != nil {
		return nil, err
	}

	return &tradedto.ProductListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *TradeService) GetPublicProduct(ctx context.Context, id uint64) (*tradedto.ProductItem, error) {
	var product tradeentity.Product
	if err := s.db.WithContext(ctx).Where("id = ? AND status = ?", id, tradeentity.ProductStatusApproved).First(&product).Error; err != nil {
		return nil, err
	}

	items, err := s.toProductItems(ctx, []tradeentity.Product{product}, false)
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (s *TradeService) SaveSellerProduct(ctx context.Context, sellerID uint64, req *tradedto.UpsertProductRequest, productID *uint64) (*tradedto.ProductItem, error) {
	var seller userentity.User
	if err := s.db.WithContext(ctx).First(&seller, sellerID).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	var product tradeentity.Product

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if productID != nil {
			if err := tx.Where("id = ? AND seller_id = ?", *productID, sellerID).First(&product).Error; err != nil {
				return err
			}
		} else {
			product = tradeentity.Product{
				SellerID:   seller.ID,
				SellerUID:  seller.UID,
				SellerName: seller.Username,
			}
		}

		product.Name = req.Name
		product.Category = req.Category
		product.Specs = req.Specs
		product.Price = req.Price
		product.Stock = req.Stock
		product.VisibleStock = req.Stock
		product.Description = req.Description
		product.ImageURLs = req.ImageURLs
		product.QualificationURLs = req.QualificationURLs
		product.Status = tradeentity.ProductStatusPendingReview
		product.AuditRemark = ""
		product.SubmittedAt = &now
		product.SellerUID = seller.UID
		if seller.Username != "" {
			product.SellerName = seller.Username
		} else {
			product.SellerName = seller.Phone
		}

		if productID != nil {
			if err := tx.Save(&product).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&product).Error; err != nil {
				return err
			}
		}

		action := "seller_submit"
		operatorName := seller.Username
		if productID != nil {
			action = "seller_resubmit"
		}
		if operatorName == "" {
			operatorName = seller.Phone
		}

		logItem := tradeentity.ProductAuditLog{
			ProductID:    product.ID,
			ProductSKUID: product.SKUID,
			Action:       action,
			Status:       string(product.Status),
			OperatorID:   seller.ID,
			OperatorName: operatorName,
			Remark:       "卖家提交商品审核",
		}
		return tx.Create(&logItem).Error
	})
	if err != nil {
		return nil, err
	}

	items, err := s.toProductItems(ctx, []tradeentity.Product{product}, true)
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (s *TradeService) AuditProduct(ctx context.Context, operatorID, productID uint64, req *tradedto.AuditProductRequest) (*tradedto.ProductItem, error) {
	var operator userentity.User
	if err := s.db.WithContext(ctx).First(&operator, operatorID).Error; err != nil {
		return nil, err
	}

	var product tradeentity.Product
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&product, productID).Error; err != nil {
			return err
		}

		switch req.Status {
		case string(tradeentity.ProductStatusApproved):
			product.Status = tradeentity.ProductStatusApproved
			product.AuditRemark = strings.TrimSpace(req.Remark)
			now := time.Now()
			product.ApprovedAt = &now
		case string(tradeentity.ProductStatusRejected):
			product.Status = tradeentity.ProductStatusRejected
			product.AuditRemark = strings.TrimSpace(req.Remark)
		default:
			return fmt.Errorf("invalid audit status")
		}

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		operatorName := operator.Username
		if operatorName == "" {
			operatorName = operator.Phone
		}
		logItem := tradeentity.ProductAuditLog{
			ProductID:    product.ID,
			ProductSKUID: product.SKUID,
			Action:       "platform_audit",
			Status:       string(product.Status),
			OperatorID:   operator.ID,
			OperatorName: operatorName,
			Remark:       req.Remark,
		}
		return tx.Create(&logItem).Error
	})
	if err != nil {
		return nil, err
	}

	items, err := s.toProductItems(ctx, []tradeentity.Product{product}, true)
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (s *TradeService) CreateOrder(ctx context.Context, buyerID uint64, req *tradedto.CreateOrderRequest) (*tradedto.OrderInfo, error) {
	var buyer userentity.User
	if err := s.db.WithContext(ctx).First(&buyer, buyerID).Error; err != nil {
		return nil, err
	}

	var order tradeentity.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product tradeentity.Product
		if err := tx.Where("id = ? AND status = ?", req.ProductID, tradeentity.ProductStatusApproved).First(&product).Error; err != nil {
			return err
		}
		if product.VisibleStock < req.Quantity {
			return fmt.Errorf("insufficient stock")
		}

		order = tradeentity.Order{
			BuyerID:       buyer.ID,
			BuyerUID:      buyer.UID,
			SellerID:      product.SellerID,
			SellerUID:     product.SellerUID,
			ProductID:     product.ID,
			ProductSKUID:  product.SKUID,
			TotalAmount:   float64(req.Quantity) * product.Price,
			TotalQuantity: req.Quantity,
			Status:        tradeentity.OrderStatusPendingContract,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		orderItem := tradeentity.OrderItem{
			OrderRefID:      order.ID,
			ProductID:       product.ID,
			ProductSKUID:    product.SKUID,
			ProductName:     product.Name,
			ProductCategory: product.Category,
			Specs:           product.Specs,
			Price:           product.Price,
			Quantity:        req.Quantity,
			Amount:          float64(req.Quantity) * product.Price,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			return err
		}

		product.VisibleStock -= req.Quantity
		product.Stock -= req.Quantity
		product.Sales += req.Quantity
		return tx.Save(&product).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, buyerID, order.ID)
}

func (s *TradeService) SubmitOrderContract(ctx context.Context, buyerID, orderRefID uint64, req *tradedto.SubmitContractRequest) (*tradedto.OrderInfo, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order tradeentity.Order
		if err := tx.Where("id = ? AND buyer_id = ?", orderRefID, buyerID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != tradeentity.OrderStatusPendingContract {
			return fmt.Errorf("order status does not allow contract submission")
		}

		contract := tradeentity.OrderContract{
			OrderRefID:   order.ID,
			OrderID:      order.OrderID,
			BuyerID:      buyerID,
			ContractURL:  strings.TrimSpace(req.ContractURL),
			ContractHash: strings.TrimSpace(req.ContractHash),
			Status:       tradeentity.OrderContractStatusSubmitted,
			SubmittedBy:  buyerID,
			SubmittedAt:  time.Now(),
		}
		if err := tx.Where("order_ref_id = ?", order.ID).Assign(contract).FirstOrCreate(&contract).Error; err != nil {
			return err
		}

		return tx.Model(&order).Update("status", tradeentity.OrderStatusPendingPayment).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, buyerID, orderRefID)
}

func (s *TradeService) SubmitPaymentReceipt(ctx context.Context, buyerID, orderRefID uint64, req *tradedto.SubmitPaymentReceiptRequest) (*tradedto.OrderInfo, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order tradeentity.Order
		if err := tx.Where("id = ? AND buyer_id = ?", orderRefID, buyerID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != tradeentity.OrderStatusPendingPayment {
			return fmt.Errorf("order status does not allow payment receipt submission")
		}
		if req.Amount != order.TotalAmount {
			return fmt.Errorf("payment receipt amount does not match order amount")
		}

		receipt := tradeentity.PaymentReceipt{
			OrderRefID:  order.ID,
			OrderID:     order.OrderID,
			BuyerID:     buyerID,
			ReceiptURL:  strings.TrimSpace(req.ReceiptURL),
			VoucherNo:   strings.TrimSpace(req.VoucherNo),
			Amount:      req.Amount,
			Status:      tradeentity.PaymentReceiptStatusPending,
			AuditRemark: "",
			SubmittedBy: buyerID,
			SubmittedAt: time.Now(),
			AuditorID:   nil,
			AuditedAt:   nil,
		}
		if err := tx.Where("order_ref_id = ?", order.ID).Assign(receipt).FirstOrCreate(&receipt).Error; err != nil {
			return err
		}

		return tx.Model(&order).Update("status", tradeentity.OrderStatusPendingPaymentAudit).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, buyerID, orderRefID)
}

func (s *TradeService) ListOrders(ctx context.Context, buyerID uint64, req *tradedto.OrderListRequest) (*tradedto.OrderListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var orders []tradeentity.Order
	var total int64
	query := s.db.WithContext(ctx).Model(&tradeentity.Order{}).Where("buyer_id = ?", buyerID)
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.
		Preload("Items").
		Preload("Contract").
		Preload("PaymentReceipt").
		Preload("PickupCode").
		Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	list := make([]tradedto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		list = append(list, toOrderInfo(order))
	}

	return &tradedto.OrderListResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *TradeService) GetOrder(ctx context.Context, buyerID, orderRefID uint64) (*tradedto.OrderInfo, error) {
	var order tradeentity.Order
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Preload("Contract").
		Preload("PaymentReceipt").
		Preload("PickupCode").
		Where("id = ? AND buyer_id = ?", orderRefID, buyerID).
		First(&order).Error; err != nil {
		return nil, err
	}
	info := toOrderInfo(order)
	return &info, nil
}

func (s *TradeService) ListAdminOrders(ctx context.Context, req *tradedto.OrderListRequest) (*tradedto.OrderListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var orders []tradeentity.Order
	var total int64
	query := s.db.WithContext(ctx).Model(&tradeentity.Order{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.
		Preload("Items").
		Preload("Contract").
		Preload("PaymentReceipt").
		Preload("PickupCode").
		Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	list := make([]tradedto.OrderInfo, 0, len(orders))
	for _, order := range orders {
		list = append(list, toOrderInfo(order))
	}
	return &tradedto.OrderListResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *TradeService) GetAdminOrder(ctx context.Context, orderRefID uint64) (*tradedto.OrderInfo, error) {
	var order tradeentity.Order
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Preload("Contract").
		Preload("PaymentReceipt").
		Preload("PickupCode").
		Where("id = ?", orderRefID).
		First(&order).Error; err != nil {
		return nil, err
	}
	info := toOrderInfo(order)
	return &info, nil
}

func (s *TradeService) AuditOrderPayment(ctx context.Context, operatorID, orderRefID uint64, req *tradedto.AuditPaymentRequest) (*tradedto.OrderInfo, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order tradeentity.Order
		if err := tx.Where("id = ?", orderRefID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != tradeentity.OrderStatusPendingPaymentAudit {
			return fmt.Errorf("order status does not allow payment audit")
		}

		var receipt tradeentity.PaymentReceipt
		if err := tx.Where("order_ref_id = ?", order.ID).First(&receipt).Error; err != nil {
			return err
		}
		if receipt.Status != tradeentity.PaymentReceiptStatusPending {
			return fmt.Errorf("payment receipt is not pending audit")
		}

		now := time.Now()
		operatorIDCopy := operatorID
		receipt.AuditRemark = strings.TrimSpace(req.Remark)
		receipt.AuditorID = &operatorIDCopy
		receipt.AuditedAt = &now

		switch req.Status {
		case string(tradeentity.PaymentReceiptStatusApproved):
			receipt.Status = tradeentity.PaymentReceiptStatusApproved
			order.Status = tradeentity.OrderStatusPendingPickup
		case string(tradeentity.PaymentReceiptStatusRejected):
			receipt.Status = tradeentity.PaymentReceiptStatusRejected
			order.Status = tradeentity.OrderStatusPendingPayment
		default:
			return fmt.Errorf("invalid payment audit status")
		}

		if err := tx.Save(&receipt).Error; err != nil {
			return err
		}

		if order.Status == tradeentity.OrderStatusPendingPickup {
			var pickupCode tradeentity.PickupCode
			if err := tx.Where("order_ref_id = ?", order.ID).First(&pickupCode).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				pickupCode = tradeentity.PickupCode{
					OrderRefID:  order.ID,
					OrderID:     order.OrderID,
					Code:        generatePickupCode(),
					Status:      tradeentity.PickupCodeStatusActive,
					GeneratedBy: operatorID,
					GeneratedAt: now,
				}
				if err := tx.Create(&pickupCode).Error; err != nil {
					return err
				}
			}
		}

		return tx.Save(&order).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetAdminOrder(ctx, orderRefID)
}

func (s *TradeService) GetOrderPickupCode(ctx context.Context, buyerID, orderRefID uint64) (*tradedto.PickupCodeInfo, error) {
	var order tradeentity.Order
	if err := s.db.WithContext(ctx).Preload("PickupCode").Where("id = ? AND buyer_id = ?", orderRefID, buyerID).First(&order).Error; err != nil {
		return nil, err
	}
	if order.PickupCode == nil {
		return nil, gorm.ErrRecordNotFound
	}
	info := toPickupCodeInfo(*order.PickupCode)
	return &info, nil
}

func (s *TradeService) VerifyPickupCode(ctx context.Context, operatorID, orderRefID uint64, req *tradedto.VerifyPickupCodeRequest) (*tradedto.OrderInfo, error) {
	var buyerID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order tradeentity.Order
		if err := tx.Where("id = ?", orderRefID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != tradeentity.OrderStatusPendingPickup {
			return fmt.Errorf("order status does not allow pickup verification")
		}
		buyerID = order.BuyerID
		if operatorID != order.SellerID && operatorID != order.BuyerID {
			return fmt.Errorf("current user cannot verify this pickup code")
		}

		var pickupCode tradeentity.PickupCode
		if err := tx.Where("order_ref_id = ?", order.ID).First(&pickupCode).Error; err != nil {
			return err
		}
		if pickupCode.Status != tradeentity.PickupCodeStatusActive {
			return fmt.Errorf("pickup code is not active")
		}
		if strings.TrimSpace(req.Code) != pickupCode.Code {
			return fmt.Errorf("invalid pickup code")
		}

		now := time.Now()
		operatorIDCopy := operatorID
		pickupCode.Status = tradeentity.PickupCodeStatusVerified
		pickupCode.VerifiedBy = &operatorIDCopy
		pickupCode.VerifiedAt = &now
		if err := tx.Save(&pickupCode).Error; err != nil {
			return err
		}

		order.Status = tradeentity.OrderStatusCompleted
		return tx.Save(&order).Error
	})
	if err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, buyerID, orderRefID)
}

func (s *TradeService) listProducts(ctx context.Context, req *tradedto.ProductListRequest, sellerScope bool, sellerID uint64) (*tradedto.ProductListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var products []tradeentity.Product
	var total int64

	query := s.db.WithContext(ctx).Model(&tradeentity.Product{})
	if sellerScope {
		query = query.Where("seller_id = ?", sellerID)
	} else {
		query = query.Where("status = ?", tradeentity.ProductStatusApproved)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR sku_id LIKE ?", like, like)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if sellerScope && req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&products).Error; err != nil {
		return nil, err
	}

	items, err := s.toProductItems(ctx, products, sellerScope)
	if err != nil {
		return nil, err
	}

	return &tradedto.ProductListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *TradeService) toProductItems(ctx context.Context, products []tradeentity.Product, includeLogs bool) ([]tradedto.ProductItem, error) {
	items := make([]tradedto.ProductItem, 0, len(products))
	for _, product := range products {
		item := tradedto.ProductItem{
			ID:                product.ID,
			SKUID:             product.SKUID,
			SellerID:          product.SellerID,
			SellerUID:         product.SellerUID,
			SellerName:        product.SellerName,
			Name:              product.Name,
			Category:          product.Category,
			Specs:             product.Specs,
			Price:             product.Price,
			Unit:              product.Unit,
			Stock:             product.Stock,
			VisibleStock:      product.VisibleStock,
			Sales:             product.Sales,
			Description:       product.Description,
			ImageURLs:         []string(product.ImageURLs),
			QualificationURLs: []string(product.QualificationURLs),
			Status:            string(product.Status),
			AuditRemark:       product.AuditRemark,
			SubmittedAt:       product.SubmittedAt,
			ApprovedAt:        product.ApprovedAt,
			CreatedAt:         product.CreatedAt,
			UpdatedAt:         product.UpdatedAt,
		}

		if includeLogs {
			var logs []tradeentity.ProductAuditLog
			if err := s.db.WithContext(ctx).Where("product_id = ?", product.ID).Order("id ASC").Find(&logs).Error; err != nil {
				return nil, err
			}
			item.AuditLogs = make([]tradedto.ProductAuditLogItem, 0, len(logs))
			for _, logItem := range logs {
				item.AuditLogs = append(item.AuditLogs, tradedto.ProductAuditLogItem{
					ID:           logItem.ID,
					Action:       logItem.Action,
					Status:       logItem.Status,
					OperatorID:   logItem.OperatorID,
					OperatorName: logItem.OperatorName,
					Remark:       logItem.Remark,
					CreatedAt:    logItem.CreatedAt,
				})
			}
		}

		items = append(items, item)
	}
	return items, nil
}

func toOrderInfo(order tradeentity.Order) tradedto.OrderInfo {
	items := make([]tradedto.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, tradedto.OrderItem{
			ID:              item.ID,
			ProductID:       item.ProductID,
			ProductSKUID:    item.ProductSKUID,
			ProductName:     item.ProductName,
			ProductCategory: item.ProductCategory,
			Specs:           item.Specs,
			Price:           item.Price,
			Quantity:        item.Quantity,
			Amount:          item.Amount,
		})
	}

	return tradedto.OrderInfo{
		ID:             order.ID,
		OrderID:        order.OrderID,
		BuyerID:        order.BuyerID,
		BuyerUID:       order.BuyerUID,
		SellerID:       order.SellerID,
		SellerUID:      order.SellerUID,
		ProductID:      order.ProductID,
		ProductSKUID:   order.ProductSKUID,
		Status:         string(order.Status),
		TotalAmount:    order.TotalAmount,
		TotalQuantity:  order.TotalQuantity,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		Items:          items,
		Contract:       toOrderContractInfo(order.Contract),
		PaymentReceipt: toPaymentReceiptInfo(order.PaymentReceipt),
		PickupCode:     toPickupCodeInfoPtr(order.PickupCode),
	}
}

func toOrderContractInfo(contract *tradeentity.OrderContract) *tradedto.OrderContractInfo {
	if contract == nil {
		return nil
	}
	return &tradedto.OrderContractInfo{
		ID:           contract.ID,
		ContractURL:  contract.ContractURL,
		ContractHash: contract.ContractHash,
		Status:       string(contract.Status),
		SubmittedBy:  contract.SubmittedBy,
		SubmittedAt:  contract.SubmittedAt,
	}
}

func toPaymentReceiptInfo(receipt *tradeentity.PaymentReceipt) *tradedto.PaymentReceiptInfo {
	if receipt == nil {
		return nil
	}
	return &tradedto.PaymentReceiptInfo{
		ID:          receipt.ID,
		ReceiptURL:  receipt.ReceiptURL,
		VoucherNo:   receipt.VoucherNo,
		Amount:      receipt.Amount,
		Status:      string(receipt.Status),
		AuditRemark: receipt.AuditRemark,
		SubmittedBy: receipt.SubmittedBy,
		SubmittedAt: receipt.SubmittedAt,
		AuditorID:   receipt.AuditorID,
		AuditedAt:   receipt.AuditedAt,
	}
}

func toPickupCodeInfoPtr(code *tradeentity.PickupCode) *tradedto.PickupCodeInfo {
	if code == nil {
		return nil
	}
	info := toPickupCodeInfo(*code)
	return &info
}

func toPickupCodeInfo(code tradeentity.PickupCode) tradedto.PickupCodeInfo {
	return tradedto.PickupCodeInfo{
		ID:          code.ID,
		Code:        code.Code,
		Status:      string(code.Status),
		GeneratedBy: code.GeneratedBy,
		GeneratedAt: code.GeneratedAt,
		VerifiedBy:  code.VerifiedBy,
		VerifiedAt:  code.VerifiedAt,
	}
}

func generatePickupCode() string {
	return "PK" + time.Now().Format("20060102150405") + fmt.Sprintf("%04d", time.Now().Nanosecond()%10000)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
