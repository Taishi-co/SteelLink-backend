package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/embedding"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/vectorstore"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

func (s *ProcurementService) SyncVectors(ctx context.Context, req *dto.VectorSyncRequest) (*dto.VectorSyncResponse, error) {
	if req == nil {
		req = &dto.VectorSyncRequest{}
	}
	target := strings.ToLower(strings.TrimSpace(req.Target))
	if target == "" {
		target = "products"
	}
	switch target {
	case "products":
		return s.syncProductVectors(ctx, req)
	case "suppliers", "supplier":
		return reservedVectorSync(target), nil
	case "finance", "finance-products", "finance_products":
		return reservedVectorSync(target), nil
	case "all":
		result, err := s.syncProductVectors(ctx, req)
		if err != nil {
			return nil, err
		}
		result.ReservedTargets = append(result.ReservedTargets, "suppliers", "finance")
		result.Message = "product vectors synced; supplier and finance product sync are reserved"
		return result, nil
	default:
		return nil, ecode.ErrBadRequest.WithMessagef("unsupported vector sync target %q", req.Target)
	}
}

func reservedVectorSync(target string) *dto.VectorSyncResponse {
	return &dto.VectorSyncResponse{
		Target:          target,
		Synced:          0,
		LastID:          0,
		ReservedTargets: []string{target},
		Message:         "reserved target; no-op in this version",
	}
}

func (s *ProcurementService) syncProductVectors(ctx context.Context, req *dto.VectorSyncRequest) (*dto.VectorSyncResponse, error) {
	config.RefreshAIProcurementRuntimeConfig(ctx)
	cfg := config.GetAIProcurement().VectorSearch
	if !cfg.Enabled {
		return nil, ecode.ErrBadRequest.WithMessage("vector search is disabled in runtime config")
	}
	if cfg.Provider != "pgvector" {
		return nil, ecode.ErrBadRequest.WithMessagef("vector provider must be pgvector, got %q", cfg.Provider)
	}
	if cfg.ConnectionString() == "" {
		return nil, ecode.ErrBadRequest.WithMessagef("%s is not configured", cfg.ConnectionEnv)
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	store := vectorstore.NewPGVectorClient(cfg)
	if req.CreateSchema == nil || *req.CreateSchema {
		if err := store.EnsureProductSchema(ctx, cfg.Dimension); err != nil {
			return nil, ecode.ErrDatabase.WithMessagef("ensure pgvector product schema failed: %v", err)
		}
	}
	embedder := embedding.NewClient(cfg)

	var synced int
	db := database.GetDB()
	watermark, err := store.GetProductSyncWatermark(ctx)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessagef("read pgvector product sync watermark failed: %v", err)
	}
	watermark, err = normalizeProductSyncWatermark(ctx, db, store, watermark)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessagef("initialize product sync watermark failed: %v", err)
	}
	for {
		remaining := req.Limit - synced
		if req.Limit > 0 && remaining <= 0 {
			break
		}
		currentBatchSize := batchSize
		if req.Limit > 0 && remaining < currentBatchSize {
			currentBatchSize = remaining
		}
		products, err := listProductsForVectorSync(ctx, db, watermark, currentBatchSize)
		if err != nil {
			return nil, ecode.ErrDatabase.WithMessagef("list products failed: %v", err)
		}
		if len(products) == 0 {
			break
		}
		for _, product := range products {
			text := buildProductEmbeddingText(product)
			vector, err := embedder.Embed(ctx, text)
			if err != nil {
				return nil, ecode.ErrInternal.WithMessagef("embed product %d failed: %v", product.ID, err)
			}
			if err := store.UpsertProduct(ctx, vectorstore.ProductEmbedding{
				ProductID:       product.ID,
				ProductNo:       product.ProductNo,
				ProductName:     product.Name,
				Category:        productCategory(product),
				EnterpriseID:    product.EnterpriseID,
				SourceUpdatedAt: product.UpdatedAt,
				Embedding:       vector,
			}); err != nil {
				return nil, ecode.ErrDatabase.WithMessagef("upsert product %d vector failed: %v", product.ID, err)
			}
			synced++
			watermark.SourceUpdatedAt = product.UpdatedAt
			watermark.ProductID = product.ID
			watermark.Initialized = true
			if err := store.SaveProductSyncWatermark(ctx, watermark); err != nil {
				return nil, ecode.ErrDatabase.WithMessagef("save product sync watermark failed: %v", err)
			}
		}
	}

	return &dto.VectorSyncResponse{
		Target:              "products",
		Synced:              synced,
		LastID:              watermark.ProductID,
		LastSourceUpdatedAt: watermark.SourceUpdatedAt,
		Message:             fmt.Sprintf("product vector incremental sync complete count=%d", synced),
	}, nil
}

func normalizeProductSyncWatermark(ctx context.Context, db *gorm.DB, store *vectorstore.PGVectorClient, watermark vectorstore.ProductSyncWatermark) (vectorstore.ProductSyncWatermark, error) {
	if watermark.Initialized {
		return watermark, nil
	}
	if !isLegacyZeroWatermark(watermark.SourceUpdatedAt) && watermark.ProductID > 0 {
		return watermark, nil
	}
	current, err := latestProductWatermark(ctx, db, watermark.ProductID)
	if err != nil {
		return watermark, err
	}
	if current.SourceUpdatedAt.IsZero() {
		return watermark, nil
	}
	current.Initialized = true
	if err := store.SaveProductSyncWatermark(ctx, current); err != nil {
		return watermark, err
	}
	return current, nil
}

func latestProductWatermark(ctx context.Context, db *gorm.DB, maxProductID uint64) (vectorstore.ProductSyncWatermark, error) {
	var product productEntity.Product
	query := db.WithContext(ctx).Order("updated_at DESC, id DESC")
	if maxProductID > 0 {
		query = query.Where("id <= ?", maxProductID)
	}
	err := query.First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return vectorstore.ProductSyncWatermark{}, nil
		}
		return vectorstore.ProductSyncWatermark{}, err
	}
	return vectorstore.ProductSyncWatermark{
		SourceUpdatedAt: product.UpdatedAt,
		ProductID:       product.ID,
		Initialized:     true,
	}, nil
}

func isLegacyZeroWatermark(value time.Time) bool {
	return value.IsZero() || value.Year() <= 1970
}

func listProductsForVectorSync(ctx context.Context, db *gorm.DB, watermark vectorstore.ProductSyncWatermark, limit int) ([]*productEntity.Product, error) {
	var products []*productEntity.Product
	err := db.WithContext(ctx).
		Where("(updated_at > ?) OR (updated_at = ? AND id > ?)", watermark.SourceUpdatedAt, watermark.SourceUpdatedAt, watermark.ProductID).
		Order("updated_at ASC, id ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func buildProductEmbeddingText(product *productEntity.Product) string {
	parts := []string{
		"商品:" + product.Name,
		"品类:" + product.Category,
		"子品类:" + product.SubCategory,
		"钢厂:" + product.SteelMill,
		fmt.Sprintf("库存:%.2f", product.AvailableQty),
		fmt.Sprintf("价格:%.2f", product.UnitPrice),
		fmt.Sprintf("交期:%d天", product.DeliveryDays),
	}
	if product.Spec != nil {
		parts = append(parts,
			"材质:"+product.Spec.Material,
			"规格:"+product.Spec.Spec,
			"厚度:"+product.Spec.Thickness,
			"宽度:"+product.Spec.Width,
			"长度:"+product.Spec.Length,
			"表面:"+product.Spec.Surface,
			"加工:"+product.Spec.Processing,
		)
	}
	if product.WarehouseInfo != nil {
		parts = append(parts,
			"仓库:"+product.WarehouseInfo.WarehouseName,
			"省份:"+product.WarehouseInfo.Province,
			"城市:"+product.WarehouseInfo.City,
			"地址:"+product.WarehouseInfo.Address,
		)
	}
	if product.Description != "" {
		parts = append(parts, "描述:"+product.Description)
	}
	return strings.Join(compactEmbeddingParts(parts), "；")
}

func productCategory(product *productEntity.Product) string {
	if product.SubCategory != "" {
		return product.SubCategory
	}
	return product.Category
}

func compactEmbeddingParts(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && !strings.HasSuffix(part, ":") {
			out = append(out, part)
		}
	}
	return out
}
