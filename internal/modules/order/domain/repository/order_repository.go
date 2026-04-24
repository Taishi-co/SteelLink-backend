package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByID(ctx context.Context, id uint64) (*entity.Order, error)
	ListByEnterprise(ctx context.Context, enterpriseID uint64) ([]*entity.Order, error)
	ListByBuyer(ctx context.Context, buyerID uint64) ([]*entity.Order, error)
	ListBySeller(ctx context.Context, sellerID uint64) ([]*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
}
