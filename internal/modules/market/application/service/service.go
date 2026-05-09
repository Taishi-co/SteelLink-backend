package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
	marketdto "github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/dto"
	orderentity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

type BiddingViewScope string

const (
	BiddingViewPublic BiddingViewScope = "public"
	BiddingViewBuyer  BiddingViewScope = "buyer"
	BiddingViewSeller BiddingViewScope = "seller"
)

type MarketService struct {
	db             *gorm.DB
	enterpriseRepo userRepository.EnterpriseRepository
}

func NewMarketService() *MarketService {
	return &MarketService{
		db:             database.GetDB(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func invalidateMarketCache(ctx context.Context) {
	cache.Delete(ctx, "market:bidding:stats", "market:procurement:stats")
}

func (s *MarketService) GetBiddingStats(ctx context.Context) (*marketdto.BiddingStats, error) {
	var cached marketdto.BiddingStats
	if cache.GetJSON(ctx, "market:bidding:stats", &cached) {
		return &cached, nil
	}
	var stats marketdto.BiddingStats
	err := s.db.WithContext(ctx).Model(&marketentity.BiddingActivity{}).
		Select(`
			COALESCE(SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END), 0) AS active_count,
			COALESCE(SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status IN ? THEN deal_amount ELSE 0 END), 0) AS total_transaction,
			COALESCE(SUM(participants), 0) AS participating_company
		`,
			[]marketentity.BiddingStatus{marketentity.BiddingStatusBidding, marketentity.BiddingStatusOngoing, marketentity.BiddingStatusSprint},
			[]marketentity.BiddingStatus{marketentity.BiddingStatusAwarded, marketentity.BiddingStatusOrderCreated, marketentity.BiddingStatusSuccess},
			[]marketentity.BiddingStatus{marketentity.BiddingStatusAwarded, marketentity.BiddingStatusOrderCreated, marketentity.BiddingStatusSuccess},
		).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	cache.SetJSON(ctx, "market:bidding:stats", &stats, cache.MarketStatsTTL)
	return &stats, nil
}

func (s *MarketService) ListBiddings(ctx context.Context, req *marketdto.ListRequest, enterpriseID uint64) (*marketdto.BiddingListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := s.db.WithContext(ctx).Model(&marketentity.BiddingActivity{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR bidding_no LIKE ? OR buyer_name LIKE ? OR supplier_name LIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var activities []marketentity.BiddingActivity
	if err := query.Order("status ASC, ended_at ASC, id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&activities).Error; err != nil {
		return nil, err
	}

	items := make([]marketdto.BiddingActivityItem, 0, len(activities))
	for i := range activities {
		if err := s.reconcileBiddingState(ctx, &activities[i]); err != nil {
			return nil, err
		}
		scope := BiddingViewPublic
		var quotes []marketentity.BiddingQuote
		viewerEnterpriseID := uint64(0)
		if req.Scope == "workspace" && enterpriseID != 0 {
			viewerEnterpriseID = enterpriseID
			scope = resolveBiddingViewScope(activities[i], enterpriseID)
			var err error
			quotes, err = s.listBiddingQuotes(ctx, activities[i].ID)
			if err != nil {
				return nil, err
			}
		}
		items = append(items, toBiddingItem(activities[i], quotes, scope, viewerEnterpriseID))
	}
	return &marketdto.BiddingListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) GetBidding(ctx context.Context, id uint64, enterpriseID uint64) (*marketdto.BiddingActivityItem, error) {
	var activity marketentity.BiddingActivity
	if err := s.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	if err := s.reconcileBiddingState(ctx, &activity); err != nil {
		return nil, err
	}
	quotes, err := s.listBiddingQuotes(ctx, activity.ID)
	if err != nil {
		return nil, err
	}
	scope := resolveBiddingViewScope(activity, enterpriseID)
	item := toBiddingItem(activity, quotes, scope, enterpriseID)
	return &item, nil
}

func (s *MarketService) CreateBidding(ctx context.Context, userID, enterpriseID uint64, req *marketdto.CreateBiddingRequest) (*marketdto.BiddingActivityItem, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能发起竞价")
	}
	enterprise, err := s.enterpriseRepo.FindByID(ctx, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if enterprise == nil || !enterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("企业不存在或未通过审核")
	}
	endedAt, err := parseMarketTime(req.EndedAt)
	if err != nil {
		return nil, ecode.ErrBadRequest.WithMessage("截止时间格式错误")
	}
	now := time.Now()
	if !endedAt.After(now) {
		return nil, ecode.ErrBadRequest.WithMessage("截止时间必须晚于当前时间")
	}
	unit := req.Unit
	if unit == "" {
		unit = "吨"
	}
	activity := &marketentity.BiddingActivity{
		BiddingNo:      generateBiddingNo(),
		Title:          req.Title,
		Category:       req.Category,
		Quantity:       req.Quantity,
		Unit:           unit,
		StartPrice:     req.StartPrice,
		CurrentPrice:   req.StartPrice,
		Status:         marketentity.BiddingStatusBidding,
		BuyerID:        enterprise.ID,
		BuyerName:      enterprise.Name,
		DeliveryRegion: req.DeliveryRegion,
		Progress:       5,
		StartedAt:      now,
		EndedAt:        endedAt,
	}
	if err := s.db.WithContext(ctx).Create(activity).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建竞价失败")
	}
	invalidateMarketCache(ctx)
	item := toBiddingItem(*activity, nil, BiddingViewBuyer, enterpriseID)
	_ = userID
	return &item, nil
}

func (s *MarketService) SubmitBiddingQuote(ctx context.Context, userID, enterpriseID, biddingID uint64, req *marketdto.SubmitBiddingQuoteRequest) (*marketdto.BiddingActivityItem, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能报价")
	}
	enterprise, err := s.enterpriseRepo.FindByID(ctx, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if enterprise == nil || !enterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("企业不存在或未通过审核")
	}

	var activity marketentity.BiddingActivity
	if err := s.db.WithContext(ctx).First(&activity, biddingID).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("竞价活动不存在")
		}
		return nil, ecode.ErrDatabase
	}
	if err := s.reconcileBiddingState(ctx, &activity); err != nil {
		return nil, ecode.ErrDatabase
	}
	if activity.BuyerID == enterpriseID {
		return nil, ecode.ErrBadRequest.WithMessage("发起方不能参与自身竞价报价")
	}
	if activity.Status != marketentity.BiddingStatusBidding {
		return nil, ecode.ErrBadRequest.WithMessage("当前竞价状态不可报价")
	}
	if !activity.EndedAt.After(time.Now()) {
		return nil, ecode.ErrBadRequest.WithMessage("竞价已截止")
	}
	if req.Quantity <= 0 || req.Quantity > activity.Quantity {
		return nil, ecode.ErrBadRequest.WithMessage("报价数量必须大于0且不超过竞价需求量")
	}
	ceiling := activity.CurrentPrice
	if ceiling <= 0 {
		ceiling = activity.StartPrice
	}
	if req.UnitPrice > ceiling {
		return nil, ecode.ErrBadRequest.WithMessage("报价不能高于当前最低价")
	}

	var sameSupplierQuotes int64
	if err := s.db.WithContext(ctx).Model(&marketentity.BiddingQuote{}).
		Where("bidding_id = ? AND supplier_id = ?", biddingID, enterpriseID).
		Count(&sameSupplierQuotes).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	quote := &marketentity.BiddingQuote{
		BiddingID:    biddingID,
		SupplierID:   enterprise.ID,
		SupplierName: enterprise.Name,
		UnitPrice:    req.UnitPrice,
		Quantity:     req.Quantity,
		DeliveryDays: req.DeliveryDays,
		Remark:       req.Remark,
		Round:        int(sameSupplierQuotes) + 1,
		CreatedBy:    userID,
	}
	participantDelta := 1
	if sameSupplierQuotes > 0 {
		participantDelta = 0
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(quote).Error; err != nil {
			return err
		}
		return tx.Model(&marketentity.BiddingActivity{}).Where("id = ?", biddingID).Updates(map[string]interface{}{
			"current_price":     req.UnitPrice,
			"quote_count":       gorm.Expr("quote_count + 1"),
			"participants":      gorm.Expr("participants + ?", participantDelta),
			"price_change_rate": calcPriceChangeRate(activity.StartPrice, req.UnitPrice),
			"progress":          calcBiddingProgress(activity.StartedAt, activity.EndedAt),
			"updated_at":        time.Now(),
		}).Error
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("提交报价失败")
	}
	invalidateMarketCache(ctx)
	return s.GetBidding(ctx, biddingID, enterpriseID)
}

func (s *MarketService) AwardBidding(ctx context.Context, enterpriseID, biddingID uint64, req *marketdto.AwardBiddingRequest) (*marketdto.BiddingActivityItem, error) {
	var activity marketentity.BiddingActivity
	if err := s.db.WithContext(ctx).First(&activity, biddingID).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("竞价活动不存在")
		}
		return nil, ecode.ErrDatabase
	}
	if err := s.reconcileBiddingState(ctx, &activity); err != nil {
		return nil, ecode.ErrDatabase
	}
	if activity.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有竞价发起方可以定标")
	}
	if !canAwardBiddingStatus(activity.Status) {
		return nil, ecode.ErrBadRequest.WithMessage("竞价截止并进入待定标后才能定标")
	}
	var quote marketentity.BiddingQuote
	if err := s.db.WithContext(ctx).Where("id = ? AND bidding_id = ?", req.QuoteID, biddingID).First(&quote).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("报价不存在")
		}
		return nil, ecode.ErrDatabase
	}
	closedAt := time.Now()
	dealAmount := quote.UnitPrice * quote.Quantity
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&marketentity.BiddingQuote{}).Where("bidding_id = ?", biddingID).Update("is_awarded", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&quote).Update("is_awarded", true).Error; err != nil {
			return err
		}
		return tx.Model(&activity).Updates(map[string]interface{}{
			"status":            marketentity.BiddingStatusAwarded,
			"award_quote_id":    quote.ID,
			"award_supplier_id": quote.SupplierID,
			"supplier_name":     quote.SupplierName,
			"current_price":     quote.UnitPrice,
			"deal_amount":       dealAmount,
			"saving_rate":       calcSavingRate(activity.StartPrice, quote.UnitPrice),
			"progress":          100,
			"closed_at":         &closedAt,
			"updated_at":        closedAt,
		}).Error
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("定标失败")
	}
	invalidateMarketCache(ctx)
	return s.GetBidding(ctx, biddingID, enterpriseID)
}

func (s *MarketService) CreateBiddingOrder(ctx context.Context, enterpriseID, userID, biddingID uint64) (*marketdto.CreateBiddingOrderResponse, error) {
	var activity marketentity.BiddingActivity
	if err := s.db.WithContext(ctx).First(&activity, biddingID).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("竞价活动不存在")
		}
		return nil, ecode.ErrDatabase
	}
	if activity.BuyerID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("只有竞价发起方可以生成订单")
	}
	if activity.Status != marketentity.BiddingStatusAwarded {
		return nil, ecode.ErrBadRequest.WithMessage("只有已定标竞价才能生成订单")
	}
	if activity.OrderID != 0 {
		return nil, ecode.ErrBadRequest.WithMessage("竞价已生成订单")
	}
	var quote marketentity.BiddingQuote
	if err := s.db.WithContext(ctx).Where("id = ? AND bidding_id = ? AND is_awarded = ?", activity.AwardQuoteID, biddingID, true).First(&quote).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrBadRequest.WithMessage("中标报价不存在")
		}
		return nil, ecode.ErrDatabase
	}

	order := &orderentity.Order{
		BuyerID:            activity.BuyerID,
		BuyerName:          activity.BuyerName,
		SellerID:           quote.SupplierID,
		SellerEnterpriseID: quote.SupplierID,
		SellerName:         quote.SupplierName,
		Items: orderentity.OrderItems{{
			ProductName: activity.Title,
			Category:    activity.Category,
			Quantity:    quote.Quantity,
			UnitPrice:   quote.UnitPrice,
			TotalAmount: quote.Quantity * quote.UnitPrice,
			TaxRate:     0,
			TaxAmount:   0,
		}},
		Currency:  "CNY",
		Status:    orderentity.OrderStatusPendingConfirm,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return tx.Model(&activity).Updates(map[string]interface{}{
			"status":     marketentity.BiddingStatusOrderCreated,
			"order_id":   order.ID,
			"updated_at": time.Now(),
		}).Error
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("生成订单失败")
	}
	invalidateMarketCache(ctx)
	activity.Status = marketentity.BiddingStatusOrderCreated
	activity.OrderID = order.ID
	item, err := s.GetBidding(ctx, biddingID, enterpriseID)
	if err != nil {
		return nil, err
	}
	return &marketdto.CreateBiddingOrderResponse{Bidding: *item, OrderID: order.ID, OrderNo: order.OrderNo}, nil
}

func (s *MarketService) GetProcurementStats(ctx context.Context) (*marketdto.ProcurementStats, error) {
	var cached marketdto.ProcurementStats
	if cache.GetJSON(ctx, "market:procurement:stats", &cached) {
		return &cached, nil
	}
	var stats marketdto.ProcurementStats
	err := s.db.WithContext(ctx).Model(&marketentity.ProcurementActivity{}).
		Select(`
			COUNT(*) AS activity_count,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS preparing_count,
			COALESCE(SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status IN ? THEN budget_min ELSE 0 END), 0) AS total_transaction,
			COALESCE(SUM(participating_count), 0) AS participating_company,
			COALESCE(SUM(demand_quantity), 0) AS demand_quantity
		`, marketentity.ProcurementStatusPreparing, []marketentity.ProcurementStatus{marketentity.ProcurementStatusAwarded, marketentity.ProcurementStatusClosed}, []marketentity.ProcurementStatus{marketentity.ProcurementStatusAwarded, marketentity.ProcurementStatusClosed}).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	cache.SetJSON(ctx, "market:procurement:stats", &stats, cache.MarketStatsTTL)
	return &stats, nil
}

func (s *MarketService) ListProcurements(ctx context.Context, req *marketdto.ListRequest) (*marketdto.ProcurementListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := s.db.WithContext(ctx).Model(&marketentity.ProcurementActivity{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR procurement_no LIKE ? OR leader_name LIKE ? OR target LIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var activities []marketentity.ProcurementActivity
	if err := query.Order("progress ASC, id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&activities).Error; err != nil {
		return nil, err
	}

	items := make([]marketdto.ProcurementActivityItem, 0, len(activities))
	for _, activity := range activities {
		items = append(items, toProcurementItem(activity, nil))
	}
	return &marketdto.ProcurementListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) GetProcurement(ctx context.Context, id uint64) (*marketdto.ProcurementActivityItem, error) {
	var activity marketentity.ProcurementActivity
	if err := s.db.WithContext(ctx).Preload("Logs", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC, id ASC")
	}).First(&activity, id).Error; err != nil {
		return nil, err
	}
	item := toProcurementItem(activity, activity.Logs)
	return &item, nil
}

func (s *MarketService) listBiddingQuotes(ctx context.Context, biddingID uint64) ([]marketentity.BiddingQuote, error) {
	var quotes []marketentity.BiddingQuote
	if err := s.db.WithContext(ctx).
		Where("bidding_id = ?", biddingID).
		Order("unit_price ASC, id ASC").
		Find(&quotes).Error; err != nil {
		return nil, err
	}
	return quotes, nil
}

func (s *MarketService) reconcileBiddingState(ctx context.Context, activity *marketentity.BiddingActivity) error {
	normalized := normalizeBiddingStatus(activity.Status)
	updates := map[string]interface{}{}
	if normalized != activity.Status {
		updates["status"] = normalized
		activity.Status = normalized
	}
	if activity.Status == marketentity.BiddingStatusBidding && !activity.EndedAt.IsZero() && !activity.EndedAt.After(time.Now()) {
		if activity.QuoteCount > 0 {
			activity.Status = marketentity.BiddingStatusPendingAward
		} else {
			activity.Status = marketentity.BiddingStatusFailed
		}
		activity.Progress = 100
		updates["status"] = activity.Status
		updates["progress"] = activity.Progress
	}
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return s.db.WithContext(ctx).Model(activity).Updates(updates).Error
}

func normalizeBiddingStatus(status marketentity.BiddingStatus) marketentity.BiddingStatus {
	switch status {
	case marketentity.BiddingStatusOngoing, marketentity.BiddingStatusSprint:
		return marketentity.BiddingStatusBidding
	case marketentity.BiddingStatusSuccess:
		return marketentity.BiddingStatusAwarded
	default:
		return status
	}
}

func resolveBiddingViewScope(activity marketentity.BiddingActivity, enterpriseID uint64) BiddingViewScope {
	if enterpriseID == 0 {
		return BiddingViewPublic
	}
	if activity.BuyerID == enterpriseID {
		return BiddingViewBuyer
	}
	return BiddingViewSeller
}

func canAwardBiddingStatus(status marketentity.BiddingStatus) bool {
	return status == marketentity.BiddingStatusPendingAward
}

func toBiddingItem(activity marketentity.BiddingActivity, quotes []marketentity.BiddingQuote, scope BiddingViewScope, viewerEnterpriseID uint64) marketdto.BiddingActivityItem {
	quoteItems := make([]marketdto.BiddingQuoteItem, 0, len(quotes))
	for _, quote := range quotes {
		if scope == BiddingViewPublic {
			continue
		}
		if scope == BiddingViewSeller && quote.SupplierID != viewerEnterpriseID && !quote.IsAwarded {
			continue
		}
		item := marketdto.BiddingQuoteItem{
			ID:           quote.ID,
			BiddingID:    quote.BiddingID,
			SupplierID:   quote.SupplierID,
			SupplierName: quote.SupplierName,
			UnitPrice:    quote.UnitPrice,
			Quantity:     quote.Quantity,
			DeliveryDays: quote.DeliveryDays,
			Remark:       quote.Remark,
			Round:        quote.Round,
			IsAwarded:    quote.IsAwarded,
			CreatedAt:    quote.CreatedAt,
		}
		if scope == BiddingViewSeller && quote.SupplierID != viewerEnterpriseID {
			item.SupplierID = 0
			item.SupplierName = ""
			item.Remark = ""
		}
		quoteItems = append(quoteItems, item)
	}
	supplierName := activity.SupplierName
	awardSupplierID := activity.AwardSupplierID
	if scope == BiddingViewPublic && activity.Status != marketentity.BiddingStatusAwarded && activity.Status != marketentity.BiddingStatusOrderCreated {
		supplierName = ""
		awardSupplierID = 0
	}
	return marketdto.BiddingActivityItem{
		ID:              activity.ID,
		BiddingNo:       activity.BiddingNo,
		Title:           activity.Title,
		Category:        activity.Category,
		Quantity:        activity.Quantity,
		Unit:            activity.Unit,
		StartPrice:      activity.StartPrice,
		CurrentPrice:    activity.CurrentPrice,
		DealAmount:      activity.DealAmount,
		PriceChangeRate: activity.PriceChangeRate,
		Participants:    activity.Participants,
		Progress:        activity.Progress,
		Status:          string(activity.Status),
		BuyerID:         activity.BuyerID,
		BuyerName:       activity.BuyerName,
		DeliveryRegion:  activity.DeliveryRegion,
		QuoteCount:      activity.QuoteCount,
		AwardQuoteID:    activity.AwardQuoteID,
		AwardSupplierID: awardSupplierID,
		OrderID:         activity.OrderID,
		SupplierName:    supplierName,
		SavingRate:      activity.SavingRate,
		StartedAt:       activity.StartedAt,
		EndedAt:         activity.EndedAt,
		ClosedAt:        activity.ClosedAt,
		Quotes:          quoteItems,
	}
}

func toProcurementItem(activity marketentity.ProcurementActivity, logs []marketentity.ProcurementLog) marketdto.ProcurementActivityItem {
	logItems := make([]marketdto.ProcurementLogItem, 0, len(logs))
	for _, logItem := range logs {
		logItems = append(logItems, marketdto.ProcurementLogItem{
			ID:        logItem.ID,
			Content:   logItem.Content,
			Color:     logItem.Color,
			CreatedAt: logItem.CreatedAt,
		})
	}

	return marketdto.ProcurementActivityItem{
		ID:                 activity.ID,
		ProcurementNo:      activity.ProcurementNo,
		Title:              activity.Title,
		LeaderName:         activity.LeaderName,
		Status:             string(activity.Status),
		Category:           activity.Category,
		Target:             activity.Target,
		TargetQuantity:     activity.TargetQuantity,
		Unit:               activity.Unit,
		BudgetMin:          activity.BudgetMin,
		BudgetMax:          activity.BudgetMax,
		BudgetDisplay:      activity.BudgetDisplay,
		Stage:              activity.Stage,
		Progress:           activity.Progress,
		Region:             activity.Region,
		OpenDate:           activity.OpenDate,
		ParticipatingCount: activity.ParticipatingCount,
		DemandQuantity:     activity.DemandQuantity,
		QualifiedSuppliers: activity.QualifiedSuppliers,
		QuoteCount:         activity.QuoteCount,
		ContractCount:      activity.ContractCount,
		DeliveryDays:       activity.DeliveryDays,
		QuoteMode:          activity.QuoteMode,
		EvaluationMode:     activity.EvaluationMode,
		Documents:          []string(activity.Documents),
		Logs:               logItems,
		CreatedAt:          activity.CreatedAt,
		UpdatedAt:          activity.UpdatedAt,
	}
}

func parseMarketTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04", value)
}

func generateBiddingNo() string {
	return fmt.Sprintf("BD%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%1000000)
}

func calcPriceChangeRate(startPrice, currentPrice float64) float64 {
	if startPrice <= 0 {
		return 0
	}
	return (currentPrice - startPrice) / startPrice * 100
}

func calcSavingRate(startPrice, dealPrice float64) float64 {
	if startPrice <= 0 {
		return 0
	}
	return (startPrice - dealPrice) / startPrice * 100
}

func calcBiddingProgress(startedAt, endedAt time.Time) int {
	now := time.Now()
	if endedAt.Before(now) || endedAt.Equal(now) {
		return 100
	}
	total := endedAt.Sub(startedAt)
	if total <= 0 {
		return 60
	}
	elapsed := now.Sub(startedAt)
	if elapsed <= 0 {
		return 5
	}
	progress := int(elapsed.Seconds() / total.Seconds() * 100)
	if progress < 5 {
		return 5
	}
	if progress > 95 {
		return 95
	}
	return progress
}
