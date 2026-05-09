package persistence

import (
	"context"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func AutoMigrate() error {
	db := database.GetDB()
	if err := db.AutoMigrate(
		&entity.Product{},
		&entity.ProductCategory{},
	); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return SeedProductCategoriesFromProducts(ctx)
}
