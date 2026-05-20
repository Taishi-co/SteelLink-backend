package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/entity"
)

type FavoriteRepository interface {
	FindByOwnerAndProduct(ctx context.Context, userID, productID uint64) (*entity.Favorite, error)
	FindByID(ctx context.Context, id uint64) (*entity.Favorite, error)
	ListByOwner(ctx context.Context, userID uint64) ([]*entity.Favorite, error)
	CountByOwner(ctx context.Context, userID uint64) (int64, error)
	Create(ctx context.Context, favorite *entity.Favorite) error
	Delete(ctx context.Context, id uint64) error
	DeleteByOwnerAndProduct(ctx context.Context, userID, productID uint64) error
}
