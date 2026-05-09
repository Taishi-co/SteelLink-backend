package service

import (
	"context"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	offerentity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// ProductService 商品服务
type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo *persistence.ProductCategoryRepository
}

// NewProductService 创建商品服务
func NewProductService() *ProductService {
	return &ProductService{
		productRepo:  persistence.NewProductRepository(),
		categoryRepo: persistence.NewProductCategoryRepository(),
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
	if err := s.categoryRepo.EnsureByName(ctx, product.Category, operatorID); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建商品品类失败")
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建商品失败")
	}
	invalidateProductCache(ctx)

	return toProductInfo(product), nil
}

// Update 更新卖家自有商品。商品关键字段变化后重新进入待审核状态。
func (s *ProductService) Update(ctx context.Context, id uint64, req *dto.CreateProductRequest, operatorID, enterpriseID uint64) (*dto.ProductInfo, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业用户才能修改商品")
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

	goodsType, err := normalizeGoodsType(req.GoodsType)
	if err != nil {
		return nil, err
	}
	warehouseInfo, warehouseServiceID, warehouseEnterpriseID, err := resolveProductWarehouse(ctx, req)
	if err != nil {
		return nil, err
	}

	product.Name = req.Name
	product.GoodsType = goodsType
	product.Category = req.Category
	product.SubCategory = req.SubCategory
	product.SteelMill = req.SteelMill
	product.Spec = &entity.ProductSpec{
		Material:   req.Spec.Material,
		Spec:       req.Spec.Spec,
		Thickness:  req.Spec.Thickness,
		Width:      req.Spec.Width,
		Length:     req.Spec.Length,
		Surface:    req.Spec.Surface,
		Edge:       req.Spec.Edge,
		Processing: req.Spec.Processing,
	}
	product.WarehouseServiceID = warehouseServiceID
	product.WarehouseEnterpriseID = warehouseEnterpriseID
	product.WarehouseInfo = warehouseInfo
	product.BundleNo = req.BundleNo
	product.Quantity = req.Quantity
	product.AvailableQty = req.Quantity - product.LockedQty
	if product.AvailableQty < 0 {
		product.AvailableQty = 0
	}
	product.UnitPrice = req.UnitPrice
	product.MinOrderQty = req.MinOrderQty
	product.Images = entity.StringList(req.Images)
	product.QualityCert = req.QualityCert
	product.Description = req.Description
	product.DeliveryDays = req.DeliveryDays
	product.QuoteValidity = req.QuoteValidity
	if product.DeliveryDays == 0 {
		product.DeliveryDays = 3
	}
	if product.QuoteValidity == 0 {
		product.QuoteValidity = 72
	}
	product.Status = entity.ProductStatusPending
	product.AuditRemark = ""
	product.ApprovedAt = nil
	product.UpdatedBy = operatorID

	if err := s.categoryRepo.EnsureByName(ctx, product.Category, operatorID); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建商品品类失败")
	}
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新商品失败")
	}
	invalidateProductCache(ctx)

	return toProductInfo(product), nil
}

// Delete 删除卖家自有商品。存在锁定库存时拒绝删除，避免影响已生成订单。
func (s *ProductService) Delete(ctx context.Context, id uint64, operatorID, enterpriseID uint64) error {
	if enterpriseID == 0 {
		return ecode.ErrBadRequest.WithMessage("企业用户才能删除商品")
	}

	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if product == nil {
		return ecode.ErrNotFound.WithMessage("商品不存在")
	}
	if product.EnterpriseID != enterpriseID {
		return ecode.ErrForbidden.WithMessage("无权删除该商品")
	}
	if product.LockedQty > 0 {
		return ecode.ErrBadRequest.WithMessage("商品存在锁定库存，不能删除")
	}
	product.UpdatedBy = operatorID
	if err := s.productRepo.Delete(ctx, product); err != nil {
		return ecode.ErrDatabase.WithMessage("删除商品失败")
	}
	invalidateProductCache(ctx)
	return nil
}

func (s *ProductService) ListCategories(ctx context.Context, req *dto.ProductCategoryListRequest) ([]*dto.ProductCategoryInfo, error) {
	key := cache.Key("product:categories", req)
	var cached []*dto.ProductCategoryInfo
	if cache.GetJSON(ctx, key, &cached) {
		return cached, nil
	}
	categories, err := s.categoryRepo.List(ctx, !req.IncludeDisabled)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	list := make([]*dto.ProductCategoryInfo, 0, len(categories))
	for _, category := range categories {
		list = append(list, toProductCategoryInfo(category))
	}
	cache.SetJSON(ctx, key, list, cache.ProductListTTL)
	return list, nil
}

func (s *ProductService) CreateCategory(ctx context.Context, req *dto.UpsertProductCategoryRequest, operatorID uint64) (*dto.ProductCategoryInfo, error) {
	category := toProductCategory(req, operatorID)
	if category.DisplayName == "" {
		category.DisplayName = category.Name
	}
	if category.Status != entity.ProductCategoryDisabled {
		category.Status = entity.ProductCategoryEnabled
	}
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建商品品类失败")
	}
	invalidateProductCache(ctx)
	return toProductCategoryInfo(category), nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, id uint64, req *dto.UpsertProductCategoryRequest, operatorID uint64) (*dto.ProductCategoryInfo, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if category == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品品类不存在")
	}
	category.ParentID = req.ParentID
	category.Name = strings.TrimSpace(req.Name)
	category.DisplayName = categoryDisplayName(req)
	if category.DisplayName == "" {
		category.DisplayName = category.Name
	}
	category.Code = strings.TrimSpace(req.Code)
	category.Icon = strings.TrimSpace(req.Icon)
	category.ImageURL = categoryImageURL(req)
	category.Sort = req.Sort
	if req.Status == entity.ProductCategoryDisabled {
		category.Status = entity.ProductCategoryDisabled
	} else {
		category.Status = entity.ProductCategoryEnabled
	}
	category.UpdatedBy = operatorID
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新商品品类失败")
	}
	invalidateProductCache(ctx)
	return toProductCategoryInfo(category), nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, id uint64) error {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if category == nil {
		return ecode.ErrNotFound.WithMessage("商品品类不存在")
	}
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return ecode.ErrDatabase.WithMessage("删除商品品类失败")
	}
	invalidateProductCache(ctx)
	return nil
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
	key := "product:public-detail:" + cache.Hash(id)
	var cached dto.ProductInfo
	if cache.GetJSON(ctx, key, &cached) {
		return &cached, nil
	}
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil || !product.IsOnSale() {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}

	result := toProductInfo(product)
	cache.SetJSON(ctx, key, result, cache.ProductDetailTTL)
	return result, nil
}

// List 获取商品列表
func (s *ProductService) List(ctx context.Context, req *dto.ProductListRequest) (*dto.ProductListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	key := cache.Key("product:list", req)
	var cached dto.ProductListResponse
	if cache.GetJSON(ctx, key, &cached) {
		return &cached, nil
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

	result := &dto.ProductListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	cache.SetJSON(ctx, key, result, cache.ProductListTTL)
	return result, nil
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
	invalidateProductCache(ctx)

	return toProductInfo(product), nil
}

// UpdateStatus 更新商品审核/上下架状态
func (s *ProductService) UpdateStatus(ctx context.Context, id uint64, status int, auditRemark string, operatorID uint64) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if product == nil {
		return ecode.ErrNotFound.WithMessage("商品不存在")
	}
	productStatus := entity.ProductStatus(status)
	switch productStatus {
	case entity.ProductStatusPending, entity.ProductStatusOnSale, entity.ProductStatusOffSale, entity.ProductStatusRejected:
	default:
		return ecode.ErrBadRequest.WithMessage("商品状态无效")
	}

	if err := s.productRepo.UpdateStatus(ctx, id, productStatus, auditRemark, operatorID); err != nil {
		return ecode.ErrDatabase.WithMessage("更新商品状态失败")
	}
	invalidateProductCache(ctx)

	return nil
}

func invalidateProductCache(ctx context.Context) {
	cache.DeleteByPattern(ctx, "product:list:*")
	cache.DeleteByPattern(ctx, "product:public-detail:*")
	cache.DeleteByPattern(ctx, "product:categories:*")
	cache.DeleteByPattern(ctx, "ai:match-products:*")
	cache.DeleteByPattern(ctx, "ai:match-logistics:*")
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

func toProductCategory(req *dto.UpsertProductCategoryRequest, operatorID uint64) *entity.ProductCategory {
	name := strings.TrimSpace(req.Name)
	displayName := categoryDisplayName(req)
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = name
	}
	return &entity.ProductCategory{
		ParentID:    req.ParentID,
		Name:        name,
		DisplayName: displayName,
		Code:        code,
		Icon:        strings.TrimSpace(req.Icon),
		ImageURL:    categoryImageURL(req),
		Sort:        req.Sort,
		Status:      req.Status,
		CreatedBy:   operatorID,
		UpdatedBy:   operatorID,
	}
}

func categoryDisplayName(req *dto.UpsertProductCategoryRequest) string {
	if value := strings.TrimSpace(req.DisplayName); value != "" {
		return value
	}
	return strings.TrimSpace(req.DisplayNameCamel)
}

func categoryImageURL(req *dto.UpsertProductCategoryRequest) string {
	if value := strings.TrimSpace(req.ImageURL); value != "" {
		return value
	}
	return strings.TrimSpace(req.ImageURLCamel)
}

func toProductCategoryInfo(category *entity.ProductCategory) *dto.ProductCategoryInfo {
	if category == nil {
		return nil
	}
	displayName := category.DisplayName
	if displayName == "" {
		displayName = category.Name
	}
	return &dto.ProductCategoryInfo{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		DisplayName: displayName,
		Code:        category.Code,
		Icon:        category.Icon,
		ImageURL:    category.ImageURL,
		Sort:        category.Sort,
		Status:      category.Status,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
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
