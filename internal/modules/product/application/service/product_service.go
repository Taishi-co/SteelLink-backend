package service

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	offerentity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// ProductService 商品服务
type ProductService struct {
	productRepo repository.ProductRepository
}

// NewProductService 创建商品服务
func NewProductService() *ProductService {
	return &ProductService{
		productRepo: persistence.NewProductRepository(),
	}
}

// Create 创建商品
func (s *ProductService) Create(ctx context.Context, req *dto.CreateProductRequest, operatorID, enterpriseID uint64) (*dto.ProductInfo, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业用户才能发布商品")
	}

	goodsType, err := normalizeGoodsType(req.GoodsType)
	if err != nil {
		return nil, err
	}
	warehouseInfo, warehouseServiceID, warehouseEnterpriseID, err := resolveProductWarehouse(ctx, req)
	if err != nil {
		return nil, err
	}

	product := &entity.Product{
		Name:        req.Name,
		GoodsType:   goodsType,
		Category:    req.Category,
		SubCategory: req.SubCategory,
		SteelMill:   req.SteelMill,
		Spec: &entity.ProductSpec{
			Material:   req.Spec.Material,
			Spec:       req.Spec.Spec,
			Thickness:  req.Spec.Thickness,
			Width:      req.Spec.Width,
			Length:     req.Spec.Length,
			Surface:    req.Spec.Surface,
			Edge:       req.Spec.Edge,
			Processing: req.Spec.Processing,
		},
		WarehouseServiceID:    warehouseServiceID,
		WarehouseEnterpriseID: warehouseEnterpriseID,
		WarehouseInfo:         warehouseInfo,
		BundleNo:              req.BundleNo,
		Quantity:              req.Quantity,
		AvailableQty:          req.Quantity,
		UnitPrice:             req.UnitPrice,
		MinOrderQty:           req.MinOrderQty,
		Status:                entity.ProductStatusPending,
		EnterpriseID:          enterpriseID,
		Images:                entity.StringList(req.Images),
		QualityCert:           req.QualityCert,
		Description:           req.Description,
		DeliveryDays:          req.DeliveryDays,
		QuoteValidity:         req.QuoteValidity,
		CreatedBy:             operatorID,
		UpdatedBy:             operatorID,
	}

	if product.DeliveryDays == 0 {
		product.DeliveryDays = 3
	}
	if product.QuoteValidity == 0 {
		product.QuoteValidity = 72
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建商品失败")
	}

	return toProductInfo(product), nil
}

// GetByID 获取商品详情
func (s *ProductService) GetByID(ctx context.Context, id uint64) (*dto.ProductInfo, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}

	return toProductInfo(product), nil
}

// GetPublicByID 获取对外可见商品详情
func (s *ProductService) GetPublicByID(ctx context.Context, id uint64) (*dto.ProductInfo, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil || !product.IsOnSale() {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}

	return toProductInfo(product), nil
}

// List 获取商品列表
func (s *ProductService) List(ctx context.Context, req *dto.ProductListRequest) (*dto.ProductListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	products, total, err := s.productRepo.List(ctx, &repository.ProductQuery{
		Page:               req.Page,
		PageSize:           req.PageSize,
		Category:           req.Category,
		SubCategory:        req.SubCategory,
		Keyword:            req.Keyword,
		SteelMill:          req.SteelMill,
		Status:             req.Status,
		EnterpriseID:       req.EnterpriseID,
		MinUnitPrice:       req.MinUnitPrice,
		MaxUnitPrice:       req.MaxUnitPrice,
		MinAvailable:       req.MinAvailable,
		Province:           req.Province,
		City:               req.City,
		GoodsType:          req.GoodsType,
		WarehouseServiceID: req.WarehouseServiceID,
		SortBy:             req.SortBy,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}

	list := make([]*dto.ProductInfo, 0, len(products))
	for _, product := range products {
		list = append(list, toProductInfo(product))
	}

	return &dto.ProductListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateDisplay 更新商品展示/审核补充资料，不允许修改捆包号、库存、价格、规格、仓库等关键交易字段。
func (s *ProductService) UpdateDisplay(ctx context.Context, id uint64, req *dto.UpdateProductDisplayRequest, operatorID, enterpriseID uint64) (*dto.ProductInfo, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业用户才能修改商品资料")
	}

	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}
	if product.EnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权修改该商品")
	}
	if len(req.Images) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请至少保留 1 张商品主图")
	}

	product.Images = entity.StringList(req.Images)
	product.QualityCert = req.QualityCert
	product.Description = req.Description
	if req.DeliveryDays > 0 {
		product.DeliveryDays = req.DeliveryDays
	}
	if req.QuoteValidity > 0 {
		product.QuoteValidity = req.QuoteValidity
	}
	product.Status = entity.ProductStatusPending
	product.AuditRemark = ""
	product.ApprovedAt = nil
	product.UpdatedBy = operatorID

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新商品资料失败")
	}

	return toProductInfo(product), nil
}

// UpdateStatus 更新商品审核/上下架状态
func (s *ProductService) UpdateStatus(ctx context.Context, id uint64, status int, operatorID uint64) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if product == nil {
		return ecode.ErrNotFound.WithMessage("商品不存在")
	}

	if err := s.productRepo.UpdateStatus(ctx, id, entity.ProductStatus(status), operatorID); err != nil {
		return ecode.ErrDatabase.WithMessage("更新商品状态失败")
	}

	return nil
}

func toProductInfo(product *entity.Product) *dto.ProductInfo {
	if product == nil {
		return nil
	}

	var spec dto.ProductSpec
	if product.Spec != nil {
		spec = dto.ProductSpec{
			Material:   product.Spec.Material,
			Spec:       product.Spec.Spec,
			Thickness:  product.Spec.Thickness,
			Width:      product.Spec.Width,
			Length:     product.Spec.Length,
			Surface:    product.Spec.Surface,
			Edge:       product.Spec.Edge,
			Processing: product.Spec.Processing,
		}
	}

	var warehouseInfo dto.WarehouseInfo
	if product.WarehouseInfo != nil {
		warehouseInfo = dto.WarehouseInfo{
			WarehouseID:           product.WarehouseInfo.WarehouseID,
			WarehouseServiceID:    product.WarehouseInfo.WarehouseServiceID,
			WarehouseEnterpriseID: product.WarehouseInfo.WarehouseEnterpriseID,
			WarehouseName:         product.WarehouseInfo.WarehouseName,
			WarehouseCode:         product.WarehouseInfo.WarehouseCode,
			Address:               product.WarehouseInfo.Address,
			Province:              product.WarehouseInfo.Province,
			City:                  product.WarehouseInfo.City,
			District:              product.WarehouseInfo.District,
		}
	}

	goodsType := product.GoodsType
	if goodsType == "" {
		goodsType = entity.GoodsTypeSpot
	}

	return &dto.ProductInfo{
		ID:                    product.ID,
		ProductNo:             product.ProductNo,
		Name:                  product.Name,
		GoodsType:             string(goodsType),
		Category:              product.Category,
		SubCategory:           product.SubCategory,
		SteelMill:             product.SteelMill,
		Spec:                  spec,
		WarehouseServiceID:    product.WarehouseServiceID,
		WarehouseEnterpriseID: product.WarehouseEnterpriseID,
		WarehouseInfo:         warehouseInfo,
		BundleNo:              product.BundleNo,
		Quantity:              product.Quantity,
		AvailableQty:          product.AvailableQty,
		LockedQty:             product.LockedQty,
		UnitPrice:             product.UnitPrice,
		MinOrderQty:           product.MinOrderQty,
		Currency:              product.Currency,
		TaxIncluded:           product.TaxIncluded,
		TaxRate:               product.TaxRate,
		Status:                int(product.Status),
		EnterpriseID:          product.EnterpriseID,
		AuditRemark:           product.AuditRemark,
		ApprovedAt:            product.ApprovedAt,
		QualityCert:           product.QualityCert,
		Images:                []string(product.Images),
		Description:           product.Description,
		DeliveryDays:          product.DeliveryDays,
		QuoteValidity:         product.QuoteValidity,
		CreatedAt:             product.CreatedAt,
		UpdatedAt:             product.UpdatedAt,
	}
}

func normalizeGoodsType(value string) (entity.GoodsType, error) {
	if value == "" {
		return entity.GoodsTypeSpot, nil
	}
	switch entity.GoodsType(value) {
	case entity.GoodsTypeSpot, entity.GoodsTypeFutures:
		return entity.GoodsType(value), nil
	default:
		return "", ecode.ErrBadRequest.WithMessage("货品类型只能为现货或期货")
	}
}

func resolveProductWarehouse(ctx context.Context, req *dto.CreateProductRequest) (*entity.WarehouseInfo, uint64, uint64, error) {
	warehouseServiceID := req.WarehouseServiceID
	if warehouseServiceID == 0 {
		warehouseServiceID = req.WarehouseInfo.WarehouseServiceID
	}

	if warehouseServiceID == 0 {
		warehouseInfo := &entity.WarehouseInfo{
			WarehouseID:           req.WarehouseInfo.WarehouseID,
			WarehouseEnterpriseID: req.WarehouseInfo.WarehouseEnterpriseID,
			WarehouseName:         req.WarehouseInfo.WarehouseName,
			WarehouseCode:         req.WarehouseInfo.WarehouseCode,
			Address:               req.WarehouseInfo.Address,
			Province:              req.WarehouseInfo.Province,
			City:                  req.WarehouseInfo.City,
			District:              req.WarehouseInfo.District,
		}
		return warehouseInfo, 0, req.WarehouseInfo.WarehouseEnterpriseID, nil
	}

	var warehouse offerentity.WarehouseService
	if err := database.GetDB().WithContext(ctx).First(&warehouse, warehouseServiceID).Error; err != nil {
		return nil, 0, 0, ecode.ErrBadRequest.WithMessage("绑定的仓储资源不存在")
	}
	if warehouse.Status != offerentity.OfferStatusApproved {
		return nil, 0, 0, ecode.ErrBadRequest.WithMessage("绑定的仓储资源尚未审核通过")
	}

	warehouseInfo := &entity.WarehouseInfo{
		WarehouseID:           warehouse.ID,
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		WarehouseName:         warehouse.WarehouseName,
		WarehouseCode:         req.WarehouseInfo.WarehouseCode,
		Address:               warehouse.Address,
		Province:              warehouse.Province,
		City:                  warehouse.City,
		District:              warehouse.District,
	}
	return warehouseInfo, warehouse.ID, warehouse.EnterpriseID, nil
}
