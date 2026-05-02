package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.ProcurementRepository = (*procurementRepository)(nil)

type procurementRepository struct {
	db *gorm.DB
}

func NewProcurementRepository() repository.ProcurementRepository {
	return &procurementRepository{db: database.GetDB()}
}

func (r *procurementRepository) Create(ctx context.Context, procurement *entity.AIProcurement) error {
	return r.db.WithContext(ctx).Create(procurement).Error
}

func (r *procurementRepository) Update(ctx context.Context, procurement *entity.AIProcurement) error {
	return r.db.WithContext(ctx).Save(procurement).Error
}

func (r *procurementRepository) FindByID(ctx context.Context, id uint64) (*entity.AIProcurement, error) {
	var procurement entity.AIProcurement
	if err := r.db.WithContext(ctx).First(&procurement, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &procurement, nil
}

func (r *procurementRepository) List(ctx context.Context, query *repository.ProcurementQuery) ([]*entity.AIProcurement, int64, error) {
	var procurements []*entity.AIProcurement
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AIProcurement{})
	if query.EnterpriseID > 0 {
		db = db.Where("enterprise_id = ?", query.EnterpriseID)
	}
	if query.CreatedBy > 0 {
		db = db.Where("created_by = ?", query.CreatedBy)
	}
	if query.HistorySaved != nil {
		db = db.Where("history_saved = ?", *query.HistorySaved)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("procurement_no LIKE ? OR prompt LIKE ? OR product_name LIKE ? OR usage LIKE ?", like, like, like, like)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC, id DESC").Offset(offset).Limit(query.PageSize).Find(&procurements).Error; err != nil {
		return nil, 0, err
	}
	return procurements, total, nil
}
