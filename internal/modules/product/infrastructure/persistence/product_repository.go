package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.ProductRepository = (*productRepository)(nil)

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储
func NewProductRepository() repository.ProductRepository {
	return &productRepository{db: database.GetDB()}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uint64) (*entity.Product, error) {
	var product entity.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) List(ctx context.Context, query *repository.ProductQuery) ([]*entity.Product, int64, error) {
	var products []*entity.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Product{})

	if query.Category != "" {
		db = db.Where("category = ? OR sub_category = ?", query.Category, query.Category)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR steel_mill LIKE ? OR bundle_no LIKE ?", like, like, like)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.EnterpriseID > 0 {
		db = db.Where("enterprise_id = ?", query.EnterpriseID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	if err := db.Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) UpdateStatus(ctx context.Context, id uint64, status entity.ProductStatus, updatedBy uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_by": updatedBy,
	}).Error
}
