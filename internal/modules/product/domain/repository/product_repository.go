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
	UpdateStatus(ctx context.Context, id uint64, status entity.ProductStatus, auditRemark string, updatedBy uint64) error
}

// ProductQuery 商品查询条件
type ProductQuery struct {
	Page               int
	PageSize           int
	CategoryID         uint64
	GradeID            uint64
	SteelMillID        uint64
	Category           string
	SubCategory        string
	Material           string
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
	HeightSpecType     string
	MinHeightSpec      float64
	MaxHeightSpec      float64
	WidthSpecType      string
	MinWidthSpec       float64
	MaxWidthSpec       float64
	LengthType         string
	MinLength          float64
	MaxLength          float64
	SortBy             string
}
