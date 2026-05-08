package persistence

import (
	"context"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductCategoryRepository struct {
	db *gorm.DB
}

func NewProductCategoryRepository() *ProductCategoryRepository {
	return &ProductCategoryRepository{db: database.GetDB()}
}

func (r *ProductCategoryRepository) List(ctx context.Context, activeOnly bool) ([]*entity.ProductCategory, error) {
	var categories []*entity.ProductCategory
	db := r.db.WithContext(ctx).Model(&entity.ProductCategory{})
	if activeOnly {
		db = db.Where("status = ?", entity.ProductCategoryEnabled)
	}
	if err := db.Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *ProductCategoryRepository) FindByName(ctx context.Context, name string) (*entity.ProductCategory, error) {
	var category entity.ProductCategory
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *ProductCategoryRepository) Update(ctx context.Context, category *entity.ProductCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *ProductCategoryRepository) FindByID(ctx context.Context, id uint64) (*entity.ProductCategory, error) {
	var category entity.ProductCategory
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.ProductCategory{}, id).Error
}

func (r *ProductCategoryRepository) EnsureByName(ctx context.Context, name string, operatorID uint64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	category := &entity.ProductCategory{
		Name:        name,
		DisplayName: name,
		Code:        normalizeCategoryCode(name),
		Status:      entity.ProductCategoryEnabled,
		CreatedBy:   operatorID,
		UpdatedBy:   operatorID,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(category).Error
}

func SeedProductCategoriesFromProducts(ctx context.Context) error {
	db := database.GetDB()
	rows, err := db.WithContext(ctx).Raw(`
		SELECT DISTINCT category
		FROM products
		WHERE category IS NOT NULL AND TRIM(category) <> ''
		ORDER BY category ASC`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	repo := NewProductCategoryRepository()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if err := repo.EnsureByName(ctx, name, 0); err != nil {
			return err
		}
	}
	return rows.Err()
}

func normalizeCategoryCode(name string) string {
	code := strings.TrimSpace(strings.ToLower(name))
	code = strings.ReplaceAll(code, " ", "_")
	code = strings.ReplaceAll(code, "/", "_")
	return code
}
