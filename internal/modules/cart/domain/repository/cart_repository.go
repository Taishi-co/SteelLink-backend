package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/entity"
)

type CartRepository interface {
	FindByOwnerAndProduct(ctx context.Context, userID, productID uint64) (*entity.CartItem, error)
	FindByID(ctx context.Context, id uint64) (*entity.CartItem, error)
	ListByOwner(ctx context.Context, userID uint64) ([]*entity.CartItem, error)
	Create(ctx context.Context, item *entity.CartItem) error
	Update(ctx context.Context, item *entity.CartItem) error
	Delete(ctx context.Context, id uint64) error
	ClearByOwner(ctx context.Context, userID uint64) error
}
