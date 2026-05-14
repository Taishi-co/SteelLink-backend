package persistence

import (
	"context"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductGradeQuery struct {
	CategoryID uint64
	ActiveOnly bool
}

func SeedProductTaxonomyFromProducts(ctx context.Context) error {
	db := database.GetDB()
	var products []entity.Product
	if err := db.WithContext(ctx).Find(&products).Error; err != nil {
		return err
	}

	categoryRepo := NewProductCategoryRepository()
	gradeRepo := NewProductGradeRepository()
	millRepo := NewProductSteelMillRepository()
	for _, product := range products {
		updates := map[string]interface{}{}
		categoryID := product.CategoryID
		if strings.TrimSpace(product.Category) != "" {
			category, err := categoryRepo.EnsureByNameReturning(ctx, product.Category, 0)
			if err != nil {
				return err
			}
			if category != nil {
				categoryID = category.ID
				if product.CategoryID == 0 {
					updates["category_id"] = category.ID
				}
			}
		}
		if product.Spec != nil && strings.TrimSpace(product.Spec.Material) != "" {
			grade, err := gradeRepo.EnsureByName(ctx, categoryID, product.Spec.Material, 0)
			if err != nil {
				return err
			}
			if grade != nil && product.GradeID == 0 {
				updates["grade_id"] = grade.ID
			}
		}
		if strings.TrimSpace(product.SteelMill) != "" {
			mill, err := millRepo.EnsureByName(ctx, product.SteelMill, 0)
			if err != nil {
				return err
			}
			if mill != nil && product.SteelMillID == 0 {
				updates["steel_mill_id"] = mill.ID
			}
		}
		if product.SpecRaw == "" && product.Spec != nil && strings.TrimSpace(product.Spec.Spec) != "" {
			updates["spec_raw"] = strings.TrimSpace(product.Spec.Spec)
		}
		if len(updates) == 0 {
			continue
		}
		if err := db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", product.ID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

type ProductGradeRepository struct {
	db *gorm.DB
}

func NewProductGradeRepository() *ProductGradeRepository {
	return &ProductGradeRepository{db: database.GetDB()}
}

func (r *ProductGradeRepository) List(ctx context.Context, query ProductGradeQuery) ([]*entity.ProductGrade, error) {
	var grades []*entity.ProductGrade
	db := r.db.WithContext(ctx).Model(&entity.ProductGrade{})
	if query.ActiveOnly {
		db = db.Where("status = ?", entity.ProductTaxonomyEnabled)
	}
	if query.CategoryID > 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}
	if err := db.Order("sort ASC, id ASC").Find(&grades).Error; err != nil {
		return nil, err
	}
	return grades, nil
}

func (r *ProductGradeRepository) FindByID(ctx context.Context, id uint64) (*entity.ProductGrade, error) {
	var grade entity.ProductGrade
	if err := r.db.WithContext(ctx).First(&grade, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &grade, nil
}

func (r *ProductGradeRepository) FindByName(ctx context.Context, categoryID uint64, name string) (*entity.ProductGrade, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	var grade entity.ProductGrade
	db := r.db.WithContext(ctx).Where("name = ?", name)
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	if err := db.Order("category_id DESC, id ASC").First(&grade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &grade, nil
}

func (r *ProductGradeRepository) EnsureByName(ctx context.Context, categoryID uint64, name string, operatorID uint64) (*entity.ProductGrade, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	existing, err := r.FindByName(ctx, categoryID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	grade := &entity.ProductGrade{
		CategoryID:  categoryID,
		Name:        name,
		DisplayName: name,
		Code:        normalizeCategoryCode(name),
		Status:      entity.ProductTaxonomyEnabled,
		CreatedBy:   operatorID,
		UpdatedBy:   operatorID,
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "category_id"}, {Name: "name"}}, DoNothing: true}).
		Create(grade).Error; err != nil {
		return nil, err
	}
	return r.FindByName(ctx, categoryID, name)
}

type ProductSteelMillRepository struct {
	db *gorm.DB
}

func NewProductSteelMillRepository() *ProductSteelMillRepository {
	return &ProductSteelMillRepository{db: database.GetDB()}
}

func (r *ProductSteelMillRepository) List(ctx context.Context, activeOnly bool) ([]*entity.ProductSteelMill, error) {
	var mills []*entity.ProductSteelMill
	db := r.db.WithContext(ctx).Model(&entity.ProductSteelMill{})
	if activeOnly {
		db = db.Where("status = ?", entity.ProductTaxonomyEnabled)
	}
	if err := db.Order("sort ASC, id ASC").Find(&mills).Error; err != nil {
		return nil, err
	}
	return mills, nil
}

func (r *ProductSteelMillRepository) FindByID(ctx context.Context, id uint64) (*entity.ProductSteelMill, error) {
	var mill entity.ProductSteelMill
	if err := r.db.WithContext(ctx).First(&mill, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &mill, nil
}

func (r *ProductSteelMillRepository) FindByName(ctx context.Context, name string) (*entity.ProductSteelMill, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	var mill entity.ProductSteelMill
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&mill).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &mill, nil
}

func (r *ProductSteelMillRepository) EnsureByName(ctx context.Context, name string, operatorID uint64) (*entity.ProductSteelMill, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	existing, err := r.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	mill := &entity.ProductSteelMill{
		Name:        name,
		DisplayName: name,
		Code:        normalizeCategoryCode(name),
		Status:      entity.ProductTaxonomyEnabled,
		CreatedBy:   operatorID,
		UpdatedBy:   operatorID,
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(mill).Error; err != nil {
		return nil, err
	}
	return r.FindByName(ctx, name)
}
