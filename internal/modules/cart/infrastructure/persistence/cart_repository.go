package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.CartRepository = (*cartRepository)(nil)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository() repository.CartRepository {
	return &cartRepository{db: database.GetDB()}
}

func (r *cartRepository) FindByOwnerAndProduct(ctx context.Context, userID, productID uint64) (*entity.CartItem, error) {
	var item entity.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) FindByID(ctx context.Context, id uint64) (*entity.CartItem, error) {
	var item entity.CartItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) ListByOwner(ctx context.Context, userID uint64) ([]*entity.CartItem, error) {
	var items []*entity.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartRepository) Create(ctx context.Context, item *entity.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *cartRepository) Update(ctx context.Context, item *entity.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *cartRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.CartItem{}, id).Error
}

func (r *cartRepository) ClearByOwner(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.CartItem{}).Error
}
