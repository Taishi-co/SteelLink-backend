package service

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

func (s *ProductService) ParseProductImport(ctx context.Context, header *multipart.FileHeader, hint string) (*dto.ProductImportParseResponse, error) {
	source, err := readProductImportSource(header)
	if err != nil {
		return nil, err
	}

	items, modelWarnings, usedModel, modelErr := parseProductImportWithModel(ctx, source, hint)
	warnings := append([]string{}, source.Warnings...)
	warnings = append(warnings, modelWarnings...)
	if modelErr != nil {
		if source.IsImage {
			return nil, ecode.ErrBadRequest.WithMessage(modelErr.Error())
		}
		warnings = append(warnings, "AI 解析失败，已使用规则解析: "+modelErr.Error())
	}
	if len(items) == 0 && !source.IsImage {
		items = parseProductImportRows(source.Text)
	}
	if len(items) == 0 {
		if source.IsImage {
			return nil, ecode.ErrBadRequest.WithMessage("未识别到商品，请确认已启用支持图片识别的 AI 模型")
		}
		warnings = append(warnings, "未识别到商品，请检查表头是否包含品种、牌号、规格、产地、存放地等字段")
	}

	return &dto.ProductImportParseResponse{
		SourceFilename: source.Filename,
		SourceType:     source.SourceType,
		UsedModel:      usedModel,
		Items:          items,
		Warnings:       compactImportStrings(warnings),
	}, nil
}

func (s *ProductService) ConfirmProductImport(ctx context.Context, req *dto.ConfirmProductImportRequest, operatorID, enterpriseID uint64) (*dto.ProductImportConfirmResponse, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业用户才能批量导入商品")
	}
	if req == nil || len(req.Items) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请先确认要导入的商品")
	}

	result := &dto.ProductImportConfirmResponse{
		Items: make([]dto.ProductImportConfirmItemResult, 0, len(req.Items)),
	}
	for index, item := range req.Items {
		rowNo := item.RowNo
		if rowNo == 0 {
			rowNo = index + 1
		}
		product, err := s.createImportedProduct(ctx, item.Product, req.SubmitForReview, operatorID, enterpriseID)
		if err != nil {
			result.FailedCount++
			result.Items = append(result.Items, dto.ProductImportConfirmItemResult{
				RowNo:  rowNo,
				Status: "failed",
				Error:  err.Error(),
			})
			continue
		}
		result.SuccessCount++
		result.Items = append(result.Items, dto.ProductImportConfirmItemResult{
			RowNo:   rowNo,
			Status:  "created",
			Product: product,
		})
	}
	if result.SuccessCount > 0 {
		invalidateProductCache(ctx)
	}
	return result, nil
}

func (s *ProductService) createImportedProduct(ctx context.Context, item dto.ProductImportProduct, submitForReview bool, operatorID, enterpriseID uint64) (*dto.ProductInfo, error) {
	item = normalizeImportProduct(item)
	if err := validateImportProduct(item, submitForReview); err != nil {
		return nil, err
	}

	createReq := importProductToCreateRequest(item)
	goodsType, err := normalizeGoodsType(createReq.GoodsType)
	if err != nil {
		return nil, err
	}
	warehouseInfo, warehouseServiceID, warehouseEnterpriseID, err := resolveProductWarehouse(ctx, createReq)
	if err != nil {
		return nil, err
	}
	taxonomy, err := s.resolveProductTaxonomy(ctx, createReq, operatorID)
	if err != nil {
		return nil, err
	}
	spec := normalizeProductSpec(createReq.Spec)
	spec.Material = taxonomy.Material

	status := entity.ProductStatusDraft
	if submitForReview {
		status = entity.ProductStatusPending
	}
	product := &entity.Product{
		Name:                  strings.TrimSpace(createReq.Name),
		GoodsType:             goodsType,
		CategoryID:            taxonomy.CategoryID,
		GradeID:               taxonomy.GradeID,
		SteelMillID:           taxonomy.SteelMillID,
		Category:              taxonomy.Category,
		SubCategory:           strings.TrimSpace(createReq.SubCategory),
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
		BundleNo:              createReq.BundleNo,
		Quantity:              createReq.Quantity,
		AvailableQty:          createReq.Quantity,
		UnitPrice:             createReq.UnitPrice,
		MinOrderQty:           createReq.MinOrderQty,
		Status:                status,
		EnterpriseID:          enterpriseID,
		Images:                entity.StringList(createReq.Images),
		QualityCert:           strings.TrimSpace(createReq.QualityCert),
		Description:           createReq.Description,
		DeliveryDays:          createReq.DeliveryDays,
		QuoteValidity:         createReq.QuoteValidity,
		CreatedBy:             operatorID,
		UpdatedBy:             operatorID,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("批量创建商品失败")
	}
	return toProductInfo(product), nil
}

func importProductToCreateRequest(item dto.ProductImportProduct) *dto.CreateProductRequest {
	return &dto.CreateProductRequest{
		Name:               item.Name,
		GoodsType:          item.GoodsType,
		CategoryID:         item.CategoryID,
		GradeID:            item.GradeID,
		SteelMillID:        item.SteelMillID,
		Category:           item.Category,
		SubCategory:        item.SubCategory,
		SteelMill:          item.SteelMill,
		Spec:               item.Spec,
		WarehouseServiceID: item.WarehouseServiceID,
		WarehouseInfo:      item.WarehouseInfo,
		BundleNo:           item.BundleNo,
		Quantity:           item.Quantity,
		UnitPrice:          item.UnitPrice,
		MinOrderQty:        item.MinOrderQty,
		Images:             item.Images,
		QualityCert:        item.QualityCert,
		Description:        item.Description,
		DeliveryDays:       item.DeliveryDays,
		QuoteValidity:      item.QuoteValidity,
	}
}

func validateImportProduct(item dto.ProductImportProduct, submitForReview bool) error {
	if strings.TrimSpace(item.Category) == "" {
		return ecode.ErrBadRequest.WithMessage("缺少品种")
	}
	if strings.TrimSpace(item.Spec.Material) == "" {
		return ecode.ErrBadRequest.WithMessage("缺少牌号")
	}
	if strings.TrimSpace(item.SteelMill) == "" {
		return ecode.ErrBadRequest.WithMessage("缺少产地/钢厂")
	}
	if strings.TrimSpace(item.Name) == "" {
		return ecode.ErrBadRequest.WithMessage("缺少商品名称")
	}
	if !submitForReview {
		return nil
	}
	if strings.TrimSpace(item.Spec.Spec) == "" {
		return ecode.ErrBadRequest.WithMessage("提交审核前需要填写规格")
	}
	if item.Quantity <= 0 {
		return ecode.ErrBadRequest.WithMessage("提交审核前需要填写库存数量")
	}
	if item.UnitPrice <= 0 {
		return ecode.ErrBadRequest.WithMessage("提交审核前需要填写单价")
	}
	if item.MinOrderQty <= 0 {
		return ecode.ErrBadRequest.WithMessage("提交审核前需要填写最小起订量")
	}
	if len(item.Images) == 0 {
		return ecode.ErrBadRequest.WithMessage("提交审核前至少需要 1 张商品主图")
	}
	if strings.TrimSpace(item.WarehouseInfo.WarehouseName) == "" {
		return ecode.ErrBadRequest.WithMessage("提交审核前需要填写存放地")
	}
	return nil
}
