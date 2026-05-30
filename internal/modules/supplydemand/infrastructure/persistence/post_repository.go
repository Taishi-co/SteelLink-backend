package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.PostRepository = (*postRepository)(nil)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository() repository.PostRepository {
	return &postRepository{db: database.GetDB()}
}

func (r *postRepository) Create(ctx context.Context, post *entity.SupplyDemandPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) FindByID(ctx context.Context, id uint64) (*entity.SupplyDemandPost, error) {
	var post entity.SupplyDemandPost
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, query *repository.PostQuery) ([]*entity.SupplyDemandPost, int64, error) {
	var posts []*entity.SupplyDemandPost
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.SupplyDemandPost{})

	if query.PublicOnly {
		db = db.Where("status IN ?", []entity.PostStatus{entity.PostStatusPublished, entity.PostStatusMatched})
	} else if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("title LIKE ? OR category LIKE ? OR material LIKE ? OR spec LIKE ? OR description LIKE ?", like, like, like, like, like)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Province != "" {
		db = db.Where("province = ?", query.Province)
	}
	if query.City != "" {
		db = db.Where("city = ?", query.City)
	}
	if query.CreatedBy > 0 {
		db = db.Where("created_by = ?", query.CreatedBy)
	}
	if query.EnterpriseID > 0 {
		db = db.Where("enterprise_id = ?", query.EnterpriseID)
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

	orderBy := "created_at DESC, id DESC"
	switch query.SortBy {
	case "matched":
		orderBy = "match_count DESC, created_at DESC, id DESC"
	case "amount_desc":
		orderBy = "match_amount DESC, created_at DESC, id DESC"
	}

	if err := db.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepository) Stats(ctx context.Context) (*repository.PostStats, error) {
	db := r.db.WithContext(ctx).Model(&entity.SupplyDemandPost{})
	stats := &repository.PostStats{}

	if err := db.Where("status IN ?", []entity.PostStatus{entity.PostStatusPublished, entity.PostStatusMatched}).Count(&stats.PublishedCount).Error; err != nil {
		return nil, err
	}
	if err := db.Where("status = ?", entity.PostStatusMatched).Count(&stats.MatchedCount).Error; err != nil {
		return nil, err
	}
	if err := db.Select("COALESCE(SUM(match_amount), 0)").Where("status = ?", entity.PostStatusMatched).Scan(&stats.MatchedAmount).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	if err := db.Where("status IN ? AND created_at >= ?", []entity.PostStatus{entity.PostStatusPublished, entity.PostStatusMatched}, today).Count(&stats.TodayNewCount).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *postRepository) UpdateStatus(ctx context.Context, id uint64, status entity.PostStatus, updatedBy uint64) error {
	return r.db.WithContext(ctx).Model(&entity.SupplyDemandPost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_by": updatedBy,
	}).Error
}
