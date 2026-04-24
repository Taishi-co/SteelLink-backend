package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.OrderRepository = (*orderRepository)(nil)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository() repository.OrderRepository {
	return &orderRepository{db: database.GetDB()}
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id uint64) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) ListByEnterprise(ctx context.Context, enterpriseID uint64) ([]*entity.Order, error) {
	var orders []*entity.Order
	if err := r.db.WithContext(ctx).
		Where("buyer_id = ? OR seller_id = ?", enterpriseID, enterpriseID).
		Order("id desc").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) ListByBuyer(ctx context.Context, buyerID uint64) ([]*entity.Order, error) {
	var orders []*entity.Order
	if err := r.db.WithContext(ctx).
		Where("buyer_id = ?", buyerID).
		Order("id desc").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) ListBySeller(ctx context.Context, sellerID uint64) ([]*entity.Order, error) {
	var orders []*entity.Order
	if err := r.db.WithContext(ctx).
		Where("seller_id = ?", sellerID).
		Order("id desc").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}
