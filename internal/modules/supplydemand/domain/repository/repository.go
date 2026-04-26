package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/entity"
)

type PostRepository interface {
	Create(ctx context.Context, post *entity.SupplyDemandPost) error
	FindByID(ctx context.Context, id uint64) (*entity.SupplyDemandPost, error)
	List(ctx context.Context, query *PostQuery) ([]*entity.SupplyDemandPost, int64, error)
	Stats(ctx context.Context) (*PostStats, error)
	UpdateStatus(ctx context.Context, id uint64, status entity.PostStatus, updatedBy uint64) error
}

type PostQuery struct {
	Page         int
	PageSize     int
	Type         entity.PostType
	Status       *entity.PostStatus
	Keyword      string
	Category     string
	Province     string
	City         string
	CreatedBy    uint64
	EnterpriseID uint64
	PublicOnly   bool
	SortBy       string
}

type PostStats struct {
	PublishedCount int64
	MatchedCount   int64
	MatchedAmount  float64
	TodayNewCount  int64
}
