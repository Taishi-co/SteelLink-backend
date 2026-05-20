package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.FavoriteRepository = (*favoriteRepository)(nil)

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository() repository.FavoriteRepository {
	return &favoriteRepository{db: database.GetDB()}
}

func (r *favoriteRepository) FindByOwnerAndProduct(ctx context.Context, userID, productID uint64) (*entity.Favorite, error) {
	var favorite entity.Favorite
	if err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepository) FindByID(ctx context.Context, id uint64) (*entity.Favorite, error) {
	var favorite entity.Favorite
	if err := r.db.WithContext(ctx).First(&favorite, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepository) ListByOwner(ctx context.Context, userID uint64) ([]*entity.Favorite, error) {
	var favorites []*entity.Favorite
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

func (r *favoriteRepository) CountByOwner(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Favorite{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *favoriteRepository) Create(ctx context.Context, favorite *entity.Favorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *favoriteRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Favorite{}, id).Error
}

func (r *favoriteRepository) DeleteByOwnerAndProduct(ctx context.Context, userID, productID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).Delete(&entity.Favorite{}).Error
}
