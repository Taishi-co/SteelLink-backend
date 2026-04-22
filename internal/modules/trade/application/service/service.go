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
	if err := query.Order("id DESC").Offset((req.Page-1)*req.PageSize).Limit(req.PageSize).Find(&products).Error; err != nil {
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
			Status:        tradeentity.OrderStatusPendingPayment,
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

func (s *TradeService) ListOrders(ctx context.Context, buyerID uint64, req *tradedto.ProductListRequest) (*tradedto.OrderListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var orders []tradeentity.Order
	var total int64
	query := s.db.WithContext(ctx).Model(&tradeentity.Order{}).Where("buyer_id = ?", buyerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Preload("Items").Order("id DESC").Offset((req.Page-1)*req.PageSize).Limit(req.PageSize).Find(&orders).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Preload("Items").Where("id = ? AND buyer_id = ?", orderRefID, buyerID).First(&order).Error; err != nil {
		return nil, err
	}
	info := toOrderInfo(order)
	return &info, nil
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
	if err := query.Order("id DESC").Offset((req.Page-1)*req.PageSize).Limit(req.PageSize).Find(&products).Error; err != nil {
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
		ID:            order.ID,
		OrderID:       order.OrderID,
		BuyerID:       order.BuyerID,
		BuyerUID:      order.BuyerUID,
		SellerID:      order.SellerID,
		SellerUID:     order.SellerUID,
		ProductID:     order.ProductID,
		ProductSKUID:  order.ProductSKUID,
		Status:        string(order.Status),
		TotalAmount:   order.TotalAmount,
		TotalQuantity: order.TotalQuantity,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
		Items:         items,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
