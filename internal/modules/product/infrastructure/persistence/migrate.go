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
		&entity.ProductGrade{},
		&entity.ProductSteelMill{},
	); err != nil {
		return err
	}
	if err := ensureBundleNoNonUniqueIndex(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := SeedProductCategoriesFromProducts(ctx); err != nil {
		return err
	}
	return SeedProductTaxonomyFromProducts(ctx)
}

func ensureBundleNoNonUniqueIndex() error {
	db := database.GetDB()
	if db.Dialector.Name() != "mysql" {
		return nil
	}
	type indexRow struct {
		IndexName string `gorm:"column:index_name"`
	}
	var uniqueIndexes []indexRow
	if err := db.Raw(`
		SELECT DISTINCT index_name
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
			AND table_name = 'products'
			AND column_name = 'bundle_no'
			AND non_unique = 0
			AND index_name <> 'PRIMARY'`).Scan(&uniqueIndexes).Error; err != nil {
		return err
	}
	for _, index := range uniqueIndexes {
		if index.IndexName == "" {
			continue
		}
		if err := db.Migrator().DropIndex(&entity.Product{}, index.IndexName); err != nil {
			return err
		}
	}

	var indexCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
			AND table_name = 'products'
			AND column_name = 'bundle_no'
			AND non_unique = 1`).Scan(&indexCount).Error; err != nil {
		return err
	}
	if indexCount > 0 {
		return nil
	}
	return db.Migrator().CreateIndex(&entity.Product{}, "BundleNo")
}
