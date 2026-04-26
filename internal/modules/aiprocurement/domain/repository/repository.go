package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
)

type ProcurementRepository interface {
	Create(ctx context.Context, procurement *entity.AIProcurement) error
	Update(ctx context.Context, procurement *entity.AIProcurement) error
	FindByID(ctx context.Context, id uint64) (*entity.AIProcurement, error)
	List(ctx context.Context, query *ProcurementQuery) ([]*entity.AIProcurement, int64, error)
}

type ProcurementQuery struct {
	Page         int
	PageSize     int
	Keyword      string
	Status       entity.ProcurementStatus
	EnterpriseID uint64
	CreatedBy    uint64
}
