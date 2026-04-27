package service

import (
	"context"
	"errors"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
	marketdto "github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

type MarketService struct {
	db *gorm.DB
}

func NewMarketService() *MarketService {
	return &MarketService{db: database.GetDB()}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (s *MarketService) GetBiddingStats(ctx context.Context) (*marketdto.BiddingStats, error) {
	var stats marketdto.BiddingStats
	err := s.db.WithContext(ctx).Model(&marketentity.BiddingActivity{}).
		Select(`
			COALESCE(SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END), 0) AS active_count,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status = ? THEN deal_amount ELSE 0 END), 0) AS total_transaction,
			COALESCE(SUM(participants), 0) AS participating_company
		`, []marketentity.BiddingStatus{marketentity.BiddingStatusOngoing, marketentity.BiddingStatusSprint}, marketentity.BiddingStatusSuccess, marketentity.BiddingStatusSuccess).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *MarketService) ListBiddings(ctx context.Context, req *marketdto.ListRequest) (*marketdto.BiddingListResponse, error) {
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
	for _, activity := range activities {
		items = append(items, toBiddingItem(activity))
	}
	return &marketdto.BiddingListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) GetBidding(ctx context.Context, id uint64) (*marketdto.BiddingActivityItem, error) {
	var activity marketentity.BiddingActivity
	if err := s.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	item := toBiddingItem(activity)
	return &item, nil
}

func (s *MarketService) GetProcurementStats(ctx context.Context) (*marketdto.ProcurementStats, error) {
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

func toBiddingItem(activity marketentity.BiddingActivity) marketdto.BiddingActivityItem {
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
		BuyerName:       activity.BuyerName,
		SupplierName:    activity.SupplierName,
		SavingRate:      activity.SavingRate,
		StartedAt:       activity.StartedAt,
		EndedAt:         activity.EndedAt,
		ClosedAt:        activity.ClosedAt,
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
