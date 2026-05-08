package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
)

// ProductRepository 商品仓储接口
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id uint64) (*entity.Product, error)
	List(ctx context.Context, query *ProductQuery) ([]*entity.Product, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status entity.ProductStatus, updatedBy uint64) error
}

// ProductQuery 商品查询条件
type ProductQuery struct {
	Page               int
	PageSize           int
	Category           string
	SubCategory        string
	Keyword            string
	SteelMill          string
	Status             *int
	EnterpriseID       uint64
	MinUnitPrice       float64
	MaxUnitPrice       float64
	MinAvailable       float64
	Province           string
	City               string
	GoodsType          string
	WarehouseServiceID uint64
	SortBy             string
}
