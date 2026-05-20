package persistence

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

func AutoMigrate() error {
	db := database.GetDB()
	if err := cleanupDuplicateFavorites(db); err != nil {
		return err
	}
	return db.AutoMigrate(&entity.Favorite{})
}

func cleanupDuplicateFavorites(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Favorite{}) {
		return nil
	}

	hasDeletedAt := db.Migrator().HasColumn(&entity.Favorite{}, "deleted_at")
	return db.Exec(duplicateFavoriteCleanupSQL(hasDeletedAt)).Error
}

func duplicateFavoriteCleanupSQL(hasDeletedAt bool) string {
	orderBy := "id ASC"
	if hasDeletedAt {
		orderBy = "CASE WHEN deleted_at IS NULL THEN 0 ELSE 1 END, id ASC"
	}

	return `
DELETE f FROM favorites AS f
JOIN (
	SELECT id
	FROM (
		SELECT id,
			ROW_NUMBER() OVER (
				PARTITION BY user_id, product_id
				ORDER BY ` + orderBy + `
			) AS rn
		FROM favorites
	) AS ranked
	WHERE rn > 1
) AS duplicate_favorites ON duplicate_favorites.id = f.id`
}
