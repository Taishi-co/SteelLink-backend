package persistence

import (
	"context"
	"errors"
	"time"

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

func (r *productRepository) Delete(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Delete(product).Error
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

	if query.CategoryID > 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}
	if query.GradeID > 0 {
		db = db.Where("grade_id = ?", query.GradeID)
	}
	if query.SteelMillID > 0 {
		db = db.Where("steel_mill_id = ?", query.SteelMillID)
	}
	if query.Category != "" {
		db = db.Where("category = ? OR sub_category = ?", query.Category, query.Category)
	}
	if query.SubCategory != "" {
		db = db.Where("sub_category = ?", query.SubCategory)
	}
	if query.Material != "" {
		db = db.Where("JSON_UNQUOTE(JSON_EXTRACT(spec, '$.material')) = ?", query.Material)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR steel_mill LIKE ? OR bundle_no LIKE ? OR product_no LIKE ? OR spec_raw LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(spec, '$.material')) LIKE ?", like, like, like, like, like, like)
	}
	if query.SteelMill != "" {
		db = db.Where("steel_mill = ?", query.SteelMill)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.EnterpriseID > 0 {
		db = db.Where("enterprise_id = ?", query.EnterpriseID)
	}
	if query.MinUnitPrice > 0 {
		db = db.Where("unit_price >= ?", query.MinUnitPrice)
	}
	if query.MaxUnitPrice > 0 {
		db = db.Where("unit_price <= ?", query.MaxUnitPrice)
	}
	if query.MinAvailable > 0 {
		db = db.Where("available_qty >= ?", query.MinAvailable)
	}
	if query.Province != "" {
		db = db.Where("JSON_UNQUOTE(JSON_EXTRACT(warehouse_info, '$.province')) = ?", query.Province)
	}
	if query.City != "" {
		db = db.Where("JSON_UNQUOTE(JSON_EXTRACT(warehouse_info, '$.city')) = ?", query.City)
	}
	if query.GoodsType != "" {
		db = db.Where("goods_type = ?", query.GoodsType)
	}
	if query.WarehouseServiceID > 0 {
		db = db.Where("warehouse_service_id = ?", query.WarehouseServiceID)
	}
	if query.HeightSpecType != "" {
		db = db.Where("height_spec_type = ?", query.HeightSpecType)
	}
	if query.MinHeightSpec > 0 {
		db = db.Where("height_spec_value >= ?", query.MinHeightSpec)
	}
	if query.MaxHeightSpec > 0 {
		db = db.Where("height_spec_value <= ?", query.MaxHeightSpec)
	}
	if query.WidthSpecType != "" {
		db = db.Where("width_spec_type = ?", query.WidthSpecType)
	}
	if query.MinWidthSpec > 0 {
		db = db.Where("width_spec_value >= ?", query.MinWidthSpec)
	}
	if query.MaxWidthSpec > 0 {
		db = db.Where("width_spec_value <= ?", query.MaxWidthSpec)
	}
	if query.LengthType != "" {
		db = db.Where("length_type = ?", query.LengthType)
	}
	if query.MinLength > 0 {
		db = db.Where("length_value >= ?", query.MinLength)
	}
	if query.MaxLength > 0 {
		db = db.Where("length_value <= ?", query.MaxLength)
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

	orderBy := "id DESC"
	switch query.SortBy {
	case "price_asc":
		orderBy = "unit_price ASC, id DESC"
	case "price_desc":
		orderBy = "unit_price DESC, id DESC"
	case "stock_desc":
		orderBy = "available_qty DESC, id DESC"
	case "newest":
		orderBy = "created_at DESC, id DESC"
	case "updated_desc":
		orderBy = "updated_at DESC, id DESC"
	}

	if err := db.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) UpdateStatus(ctx context.Context, id uint64, status entity.ProductStatus, auditRemark string, updatedBy uint64) error {
	updates := map[string]interface{}{
		"status":       status,
		"audit_remark": auditRemark,
		"updated_by":   updatedBy,
	}
	if status == entity.ProductStatusOnSale {
		now := time.Now()
		updates["approved_at"] = &now
	} else if status == entity.ProductStatusPending || status == entity.ProductStatusRejected {
		updates["approved_at"] = nil
	}
	return r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).Updates(updates).Error
}
