package service

import (
	"context"
	"strconv"
	"strings"

	notificationService "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/application/service"
	notificationEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
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
	productRepo   repository.ProductRepository
	categoryRepo  *persistence.ProductCategoryRepository
	gradeRepo     *persistence.ProductGradeRepository
	steelMillRepo *persistence.ProductSteelMillRepository
}

// NewProductService 创建商品服务
func NewProductService() *ProductService {
	return &ProductService{
		productRepo:   persistence.NewProductRepository(),
		categoryRepo:  persistence.NewProductCategoryRepository(),
		gradeRepo:     persistence.NewProductGradeRepository(),
		steelMillRepo: persistence.NewProductSteelMillRepository(),
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
	taxonomy, err := s.resolveProductTaxonomy(ctx, req, operatorID)
	if err != nil {
		return nil, err
	}
	spec := normalizeProductSpec(req.Spec)
	spec.Material = taxonomy.Material

	product := &entity.Product{
		Name:                  strings.TrimSpace(req.Name),
		GoodsType:             goodsType,
		CategoryID:            taxonomy.CategoryID,
		GradeID:               taxonomy.GradeID,
		SteelMillID:           taxonomy.SteelMillID,
		Category:              taxonomy.Category,
		SubCategory:           strings.TrimSpace(req.SubCategory),
		SteelMill:             taxonomy.SteelMill,
		Spec:                  spec,
		SpecRaw:               spec.Spec,
		HeightSpecType:        spec.HeightSpecType,
		HeightSpecValue:       spec.HeightSpecValue,
		HeightSpecUnit:        spec.HeightSpecUnit,
		WidthSpecType:         spec.WidthSpecType,
		WidthSpecValue:        spec.WidthSpecValue,
		WidthSpecUnit:         spec.WidthSpecUnit,
		LengthType:            spec.LengthType,
		LengthValue:           spec.LengthValue,
		LengthUnit:            spec.LengthUnit,
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
		QualityCert:           strings.TrimSpace(req.QualityCert),
		TitleVoucherURL:       strings.TrimSpace(req.TitleVoucherURL),
		TitleVoucherRemark:    strings.TrimSpace(req.TitleVoucherRemark),
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
	taxonomy, err := s.resolveProductTaxonomy(ctx, req, operatorID)
	if err != nil {
		return nil, err
	}
	spec := normalizeProductSpec(req.Spec)
	spec.Material = taxonomy.Material

	product.Name = strings.TrimSpace(req.Name)
	product.GoodsType = goodsType
	product.CategoryID = taxonomy.CategoryID
	product.GradeID = taxonomy.GradeID
	product.SteelMillID = taxonomy.SteelMillID
	product.Category = taxonomy.Category
	product.SubCategory = strings.TrimSpace(req.SubCategory)
	product.SteelMill = taxonomy.SteelMill
	product.Spec = spec
	product.SpecRaw = spec.Spec
	product.HeightSpecType = spec.HeightSpecType
	product.HeightSpecValue = spec.HeightSpecValue
	product.HeightSpecUnit = spec.HeightSpecUnit
	product.WidthSpecType = spec.WidthSpecType
	product.WidthSpecValue = spec.WidthSpecValue
	product.WidthSpecUnit = spec.WidthSpecUnit
	product.LengthType = spec.LengthType
	product.LengthValue = spec.LengthValue
	product.LengthUnit = spec.LengthUnit
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
	product.QualityCert = strings.TrimSpace(req.QualityCert)
	product.TitleVoucherURL = strings.TrimSpace(req.TitleVoucherURL)
	product.TitleVoucherRemark = strings.TrimSpace(req.TitleVoucherRemark)
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

func (s *ProductService) ListGrades(ctx context.Context, req *dto.ProductGradeListRequest) ([]*dto.ProductGradeInfo, error) {
	categoryID := req.CategoryID
	if categoryID == 0 && strings.TrimSpace(req.Category) != "" {
		category, err := s.categoryRepo.FindByName(ctx, strings.TrimSpace(req.Category))
		if err != nil {
			return nil, ecode.ErrDatabase
		}
		if category != nil {
			categoryID = category.ID
		}
	}
	key := cache.Key("product:grades", struct {
		CategoryID      uint64
		IncludeDisabled bool
	}{CategoryID: categoryID, IncludeDisabled: req.IncludeDisabled})
	var cached []*dto.ProductGradeInfo
	if cache.GetJSON(ctx, key, &cached) {
		return cached, nil
	}
	grades, err := s.gradeRepo.List(ctx, persistence.ProductGradeQuery{
		CategoryID: categoryID,
		ActiveOnly: !req.IncludeDisabled,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	list := make([]*dto.ProductGradeInfo, 0, len(grades))
	for _, grade := range grades {
		list = append(list, toProductGradeInfo(grade))
	}
	cache.SetJSON(ctx, key, list, cache.ProductListTTL)
	return list, nil
}

func (s *ProductService) ListSteelMills(ctx context.Context, req *dto.ProductSteelMillListRequest) ([]*dto.ProductSteelMillInfo, error) {
	key := cache.Key("product:steel-mills", req)
	var cached []*dto.ProductSteelMillInfo
	if cache.GetJSON(ctx, key, &cached) {
		return cached, nil
	}
	mills, err := s.steelMillRepo.List(ctx, !req.IncludeDisabled)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	list := make([]*dto.ProductSteelMillInfo, 0, len(mills))
	for _, mill := range mills {
		list = append(list, toProductSteelMillInfo(mill))
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
		CategoryID:         req.CategoryID,
		GradeID:            req.GradeID,
		SteelMillID:        req.SteelMillID,
		Category:           req.Category,
		SubCategory:        req.SubCategory,
		Material:           req.Material,
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
		HeightSpecType:     req.HeightSpecType,
		MinHeightSpec:      req.MinHeightSpec,
		MaxHeightSpec:      req.MaxHeightSpec,
		WidthSpecType:      req.WidthSpecType,
		MinWidthSpec:       req.MinWidthSpec,
		MaxWidthSpec:       req.MaxWidthSpec,
		LengthType:         req.LengthType,
		MinLength:          req.MinLength,
		MaxLength:          req.MaxLength,
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
	invalidateProductCache(ctx)

	if err := s.productRepo.UpdateStatus(ctx, id, productStatus, auditRemark, operatorID); err != nil {
		return ecode.ErrDatabase.WithMessage("更新商品状态失败")
	}
	invalidateProductCache(ctx)
	if err := notifyProductAuditStatus(ctx, product, productStatus, auditRemark, operatorID); err != nil {
		return err
	}

	return nil
}

func notifyProductAuditStatus(ctx context.Context, product *entity.Product, status entity.ProductStatus, auditRemark string, operatorID uint64) error {
	if product == nil {
		return nil
	}
	var title string
	switch status {
	case entity.ProductStatusOnSale:
		title = "商品审核通过"
	case entity.ProductStatusRejected:
		title = "商品审核未通过"
	default:
		return nil
	}
	content := "您的商品「" + product.Name + "」已审核通过。"
	if status == entity.ProductStatusRejected {
		content = "您的商品「" + product.Name + "」审核未通过。"
	}
	if strings.TrimSpace(auditRemark) != "" {
		content += " 审核备注：" + strings.TrimSpace(auditRemark)
	}
	return notificationService.NewNotificationService().CreateForEnterprise(ctx, product.EnterpriseID, notificationService.CreateNotificationInput{
		SenderUserID: operatorID,
		Type:         notificationEntity.NotificationTypeAudit,
		Title:        title,
		Content:      content,
		LinkURL:      "/seller/products",
		Metadata: notificationEntity.NotificationMetadata{
			"product_id": product.ID,
			"status":     int(status),
		},
	})
}

func invalidateProductCache(ctx context.Context) {
	cache.DeleteByPattern(ctx, "product:list:*")
	cache.DeleteByPattern(ctx, "product:public-detail:*")
	cache.DeleteByPattern(ctx, "product:categories:*")
	cache.DeleteByPattern(ctx, "product:grades:*")
	cache.DeleteByPattern(ctx, "product:steel-mills:*")
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
			Material:        product.Spec.Material,
			Spec:            product.Spec.Spec,
			Thickness:       product.Spec.Thickness,
			Width:           product.Spec.Width,
			Length:          product.Spec.Length,
			HeightSpecType:  product.Spec.HeightSpecType,
			HeightSpecValue: product.Spec.HeightSpecValue,
			HeightSpecUnit:  product.Spec.HeightSpecUnit,
			WidthSpecType:   product.Spec.WidthSpecType,
			WidthSpecValue:  product.Spec.WidthSpecValue,
			WidthSpecUnit:   product.Spec.WidthSpecUnit,
			LengthType:      product.Spec.LengthType,
			LengthValue:     product.Spec.LengthValue,
			LengthUnit:      product.Spec.LengthUnit,
			Surface:         product.Spec.Surface,
			Edge:            product.Spec.Edge,
			Processing:      product.Spec.Processing,
		}
	}
	if spec.Spec == "" {
		spec.Spec = product.SpecRaw
	}
	if spec.HeightSpecType == "" {
		spec.HeightSpecType = product.HeightSpecType
		spec.HeightSpecValue = product.HeightSpecValue
		spec.HeightSpecUnit = product.HeightSpecUnit
	}
	if spec.WidthSpecType == "" {
		spec.WidthSpecType = product.WidthSpecType
		spec.WidthSpecValue = product.WidthSpecValue
		spec.WidthSpecUnit = product.WidthSpecUnit
	}
	if spec.LengthType == "" {
		spec.LengthType = product.LengthType
		spec.LengthValue = product.LengthValue
		spec.LengthUnit = product.LengthUnit
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
		CategoryID:            product.CategoryID,
		GradeID:               product.GradeID,
		SteelMillID:           product.SteelMillID,
		Category:              product.Category,
		SubCategory:           product.SubCategory,
		SteelMill:             product.SteelMill,
		Spec:                  spec,
		SpecRaw:               product.SpecRaw,
		HeightSpecType:        product.HeightSpecType,
		HeightSpecValue:       product.HeightSpecValue,
		HeightSpecUnit:        product.HeightSpecUnit,
		WidthSpecType:         product.WidthSpecType,
		WidthSpecValue:        product.WidthSpecValue,
		WidthSpecUnit:         product.WidthSpecUnit,
		LengthType:            product.LengthType,
		LengthValue:           product.LengthValue,
		LengthUnit:            product.LengthUnit,
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
		TitleVoucherURL:       product.TitleVoucherURL,
		TitleVoucherRemark:    product.TitleVoucherRemark,
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

func toProductGradeInfo(grade *entity.ProductGrade) *dto.ProductGradeInfo {
	if grade == nil {
		return nil
	}
	displayName := grade.DisplayName
	if displayName == "" {
		displayName = grade.Name
	}
	return &dto.ProductGradeInfo{
		ID:          grade.ID,
		CategoryID:  grade.CategoryID,
		Name:        grade.Name,
		DisplayName: displayName,
		Code:        grade.Code,
		Aliases:     []string(grade.Aliases),
		Standard:    grade.Standard,
		Sort:        grade.Sort,
		Status:      grade.Status,
		CreatedAt:   grade.CreatedAt,
		UpdatedAt:   grade.UpdatedAt,
	}
}

func toProductSteelMillInfo(mill *entity.ProductSteelMill) *dto.ProductSteelMillInfo {
	if mill == nil {
		return nil
	}
	displayName := mill.DisplayName
	if displayName == "" {
		displayName = mill.Name
	}
	return &dto.ProductSteelMillInfo{
		ID:          mill.ID,
		Name:        mill.Name,
		DisplayName: displayName,
		Code:        mill.Code,
		Aliases:     []string(mill.Aliases),
		Letter:      mill.Letter,
		LogoURL:     mill.LogoURL,
		Sort:        mill.Sort,
		Status:      mill.Status,
		CreatedAt:   mill.CreatedAt,
		UpdatedAt:   mill.UpdatedAt,
	}
}

type productTaxonomy struct {
	CategoryID  uint64
	GradeID     uint64
	SteelMillID uint64
	Category    string
	Material    string
	SteelMill   string
}

func (s *ProductService) resolveProductTaxonomy(ctx context.Context, req *dto.CreateProductRequest, operatorID uint64) (productTaxonomy, error) {
	result := productTaxonomy{
		CategoryID:  req.CategoryID,
		GradeID:     req.GradeID,
		SteelMillID: req.SteelMillID,
		Category:    strings.TrimSpace(req.Category),
		Material:    strings.TrimSpace(req.Spec.Material),
		SteelMill:   strings.TrimSpace(req.SteelMill),
	}

	if result.CategoryID > 0 {
		category, err := s.categoryRepo.FindByID(ctx, result.CategoryID)
		if err != nil {
			return result, ecode.ErrDatabase
		}
		if category == nil {
			return result, ecode.ErrBadRequest.WithMessage("绑定的品种不存在")
		}
		if result.Category == "" {
			result.Category = category.Name
		}
	}
	category, err := s.categoryRepo.EnsureByNameReturning(ctx, result.Category, operatorID)
	if err != nil {
		return result, ecode.ErrDatabase.WithMessage("创建商品品种失败")
	}
	if category != nil {
		result.CategoryID = category.ID
		result.Category = category.Name
	}

	if result.GradeID > 0 {
		grade, err := s.gradeRepo.FindByID(ctx, result.GradeID)
		if err != nil {
			return result, ecode.ErrDatabase
		}
		if grade == nil {
			return result, ecode.ErrBadRequest.WithMessage("绑定的牌号不存在")
		}
		if result.Material == "" {
			result.Material = grade.Name
		}
		if result.CategoryID == 0 && grade.CategoryID > 0 {
			result.CategoryID = grade.CategoryID
		}
	}
	grade, err := s.gradeRepo.EnsureByName(ctx, result.CategoryID, result.Material, operatorID)
	if err != nil {
		return result, ecode.ErrDatabase.WithMessage("创建商品牌号失败")
	}
	if grade != nil {
		result.GradeID = grade.ID
		result.Material = grade.Name
	}

	if result.SteelMillID > 0 {
		mill, err := s.steelMillRepo.FindByID(ctx, result.SteelMillID)
		if err != nil {
			return result, ecode.ErrDatabase
		}
		if mill == nil {
			return result, ecode.ErrBadRequest.WithMessage("绑定的钢厂不存在")
		}
		if result.SteelMill == "" {
			result.SteelMill = mill.Name
		}
	}
	mill, err := s.steelMillRepo.EnsureByName(ctx, result.SteelMill, operatorID)
	if err != nil {
		return result, ecode.ErrDatabase.WithMessage("创建商品钢厂失败")
	}
	if mill != nil {
		result.SteelMillID = mill.ID
		result.SteelMill = mill.Name
	}
	return result, nil
}

func normalizeProductSpec(req dto.ProductSpec) *entity.ProductSpec {
	spec := &entity.ProductSpec{
		Material:        strings.TrimSpace(req.Material),
		Spec:            strings.TrimSpace(req.Spec),
		Thickness:       strings.TrimSpace(req.Thickness),
		Width:           strings.TrimSpace(req.Width),
		Length:          strings.TrimSpace(req.Length),
		HeightSpecType:  normalizeSpecType(req.HeightSpecType, "thickness"),
		HeightSpecValue: req.HeightSpecValue,
		HeightSpecUnit:  normalizeSpecUnit(req.HeightSpecUnit),
		WidthSpecType:   normalizeSpecType(req.WidthSpecType, "width"),
		WidthSpecValue:  req.WidthSpecValue,
		WidthSpecUnit:   normalizeSpecUnit(req.WidthSpecUnit),
		LengthType:      normalizeSpecType(req.LengthType, "custom"),
		LengthValue:     req.LengthValue,
		LengthUnit:      normalizeSpecUnit(req.LengthUnit),
		Surface:         strings.TrimSpace(req.Surface),
		Edge:            strings.TrimSpace(req.Edge),
		Processing:      strings.TrimSpace(req.Processing),
	}
	if spec.HeightSpecValue <= 0 {
		spec.HeightSpecValue = parseSpecNumber(spec.Thickness)
	}
	if spec.WidthSpecValue <= 0 {
		spec.WidthSpecValue = parseSpecNumber(spec.Width)
	}
	if spec.LengthValue <= 0 {
		spec.LengthValue = parseSpecNumber(spec.Length)
	}
	if spec.Thickness == "" && spec.HeightSpecType == "thickness" && spec.HeightSpecValue > 0 {
		spec.Thickness = formatSpecNumber(spec.HeightSpecValue)
	}
	if spec.Width == "" && spec.WidthSpecType == "width" && spec.WidthSpecValue > 0 {
		spec.Width = formatSpecNumber(spec.WidthSpecValue)
	}
	if spec.Length == "" && spec.LengthValue > 0 {
		spec.Length = formatSpecNumber(spec.LengthValue)
	}
	return spec
}

func normalizeSpecType(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func normalizeSpecUnit(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "mm"
	}
	return value
}

func parseSpecNumber(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func formatSpecNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
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
