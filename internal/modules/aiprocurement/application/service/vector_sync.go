package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	serviceOfferEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
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
	startedAt := time.Now()
	lockKey := "ai:vector-sync:lock:" + target
	lockToken := fmt.Sprintf("%d", startedAt.UnixNano())
	if !cache.TryLock(ctx, lockKey, lockToken, cache.VectorSyncLockTTL) {
		status := readVectorSyncStatus(ctx, target)
		if status == nil {
			status = &dto.VectorSyncResponse{Target: target, Status: "running", Message: "vector sync is already running"}
		}
		status.Status = "running"
		status.Message = "vector sync is already running"
		return status, nil
	}
	defer cache.Unlock(ctx, lockKey, lockToken)

	cache.SetJSON(ctx, vectorSyncStatusKey(target), &dto.VectorSyncResponse{
		Target:    target,
		Status:    "running",
		StartedAt: startedAt,
		Message:   "vector sync started",
	}, cache.VectorSyncStatusTTL)

	var result *dto.VectorSyncResponse
	var err error
	switch target {
	case "products":
		result, err = s.syncProductVectors(ctx, req)
	case "suppliers", "supplier":
		result = reservedVectorSync(target)
	case "finance", "finance-products", "finance_products":
		result, err = s.syncFinanceVectors(ctx, req)
	case "all":
		result, err = s.syncProductVectors(ctx, req)
		if err == nil {
			financeResult, financeErr := s.syncFinanceVectors(ctx, req)
			if financeErr != nil {
				err = financeErr
			} else {
				result.Synced += financeResult.Synced
				result.ReservedTargets = append(result.ReservedTargets, "suppliers")
				result.Message = "product and finance product vectors synced; supplier sync is reserved"
			}
		}
	default:
		return nil, ecode.ErrBadRequest.WithMessagef("unsupported vector sync target %q", req.Target)
	}
	completedAt := time.Now()
	if result == nil {
		result = &dto.VectorSyncResponse{Target: target}
	}
	result.StartedAt = startedAt
	result.CompletedAt = completedAt
	result.DurationMs = completedAt.Sub(startedAt).Milliseconds()
	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = err.Error()
		result.Message = "vector sync failed"
		cache.SetJSON(ctx, vectorSyncStatusKey(target), result, cache.VectorSyncStatusTTL)
		return nil, err
	}
	result.Status = "succeeded"
	cache.SetJSON(ctx, vectorSyncStatusKey(target), result, cache.VectorSyncStatusTTL)
	return result, nil
}

func (s *ProcurementService) syncFinanceVectors(ctx context.Context, req *dto.VectorSyncRequest) (*dto.VectorSyncResponse, error) {
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
		if err := store.EnsureFinanceSchema(ctx, cfg.Dimension); err != nil {
			return nil, ecode.ErrDatabase.WithMessagef("ensure pgvector finance schema failed: %v", err)
		}
	}
	embedder := embedding.NewClient(cfg)

	var synced int
	db := database.GetDB()
	watermark, err := store.GetFinanceSyncWatermark(ctx)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessagef("read pgvector finance sync watermark failed: %v", err)
	}
	watermark, err = normalizeFinanceSyncWatermark(ctx, db, store, watermark)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessagef("initialize finance sync watermark failed: %v", err)
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
		services, err := listFinanceServicesForVectorSync(ctx, db, watermark, currentBatchSize)
		if err != nil {
			return nil, ecode.ErrDatabase.WithMessagef("list finance services failed: %v", err)
		}
		if len(services) == 0 {
			break
		}
		for _, item := range services {
			text := buildFinanceEmbeddingText(item)
			vector, err := embedder.Embed(ctx, text)
			if err != nil {
				return nil, ecode.ErrInternal.WithMessagef("embed finance service %d failed: %v", item.ID, err)
			}
			if err := store.UpsertFinance(ctx, vectorstore.FinanceEmbedding{
				FinanceServiceID:    item.ID,
				ProductName:         item.Name,
				FinanceType:         item.FinanceType,
				FundingEnterpriseID: item.EnterpriseID,
				MinAmount:           item.MinAmount,
				MaxAmount:           item.MaxAmount,
				TermDays:            item.TermDays,
				AnnualRate:          item.AnnualRate,
				SourceUpdatedAt:     item.UpdatedAt,
				Embedding:           vector,
			}); err != nil {
				return nil, ecode.ErrDatabase.WithMessagef("upsert finance service %d vector failed: %v", item.ID, err)
			}
			synced++
			watermark.SourceUpdatedAt = item.UpdatedAt
			watermark.FinanceServiceID = item.ID
			watermark.Initialized = true
			if err := store.SaveFinanceSyncWatermark(ctx, watermark); err != nil {
				return nil, ecode.ErrDatabase.WithMessagef("save finance sync watermark failed: %v", err)
			}
		}
	}

	return &dto.VectorSyncResponse{
		Target:              "finance",
		Synced:              synced,
		LastID:              watermark.FinanceServiceID,
		LastSourceUpdatedAt: watermark.SourceUpdatedAt,
		Status:              "succeeded",
		Message:             fmt.Sprintf("finance product vector incremental sync complete count=%d", synced),
	}, nil
}

func (s *ProcurementService) GetVectorSyncStatus(ctx context.Context, target string) *dto.VectorSyncResponse {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		target = "products"
	}
	if status := readVectorSyncStatus(ctx, target); status != nil {
		return status
	}
	return &dto.VectorSyncResponse{
		Target:  target,
		Status:  "unknown",
		Message: "no vector sync status has been recorded",
	}
}

func reservedVectorSync(target string) *dto.VectorSyncResponse {
	return &dto.VectorSyncResponse{
		Target:          target,
		Synced:          0,
		LastID:          0,
		Status:          "succeeded",
		ReservedTargets: []string{target},
		Message:         "reserved target; no-op in this version",
	}
}

func readVectorSyncStatus(ctx context.Context, target string) *dto.VectorSyncResponse {
	var status dto.VectorSyncResponse
	if cache.GetJSON(ctx, vectorSyncStatusKey(target), &status) {
		return &status
	}
	return nil
}

func vectorSyncStatusKey(target string) string {
	return "ai:vector-sync:status:" + cache.HashString(target)
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
		Status:              "succeeded",
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

func normalizeFinanceSyncWatermark(ctx context.Context, db *gorm.DB, store *vectorstore.PGVectorClient, watermark vectorstore.FinanceSyncWatermark) (vectorstore.FinanceSyncWatermark, error) {
	if watermark.Initialized {
		return watermark, nil
	}
	if isLegacyZeroWatermark(watermark.SourceUpdatedAt) && watermark.FinanceServiceID == 0 {
		return watermark, nil
	}
	if !isLegacyZeroWatermark(watermark.SourceUpdatedAt) && watermark.FinanceServiceID > 0 {
		watermark.Initialized = true
		if err := store.SaveFinanceSyncWatermark(ctx, watermark); err != nil {
			return watermark, err
		}
		return watermark, nil
	}
	current, err := latestFinanceWatermark(ctx, db, watermark.FinanceServiceID)
	if err != nil {
		return watermark, err
	}
	if current.SourceUpdatedAt.IsZero() {
		return watermark, nil
	}
	current.Initialized = true
	if err := store.SaveFinanceSyncWatermark(ctx, current); err != nil {
		return watermark, err
	}
	return current, nil
}

func latestFinanceWatermark(ctx context.Context, db *gorm.DB, maxFinanceServiceID uint64) (vectorstore.FinanceSyncWatermark, error) {
	var item serviceOfferEntity.FinanceService
	query := db.WithContext(ctx).Order("updated_at DESC, id DESC")
	if maxFinanceServiceID > 0 {
		query = query.Where("id <= ?", maxFinanceServiceID)
	}
	err := query.First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return vectorstore.FinanceSyncWatermark{}, nil
		}
		return vectorstore.FinanceSyncWatermark{}, err
	}
	return vectorstore.FinanceSyncWatermark{
		SourceUpdatedAt:  item.UpdatedAt,
		FinanceServiceID: item.ID,
		Initialized:      true,
	}, nil
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

func listFinanceServicesForVectorSync(ctx context.Context, db *gorm.DB, watermark vectorstore.FinanceSyncWatermark, limit int) ([]*serviceOfferEntity.FinanceService, error) {
	var services []*serviceOfferEntity.FinanceService
	err := db.WithContext(ctx).
		Where("(updated_at > ?) OR (updated_at = ? AND id > ?)", watermark.SourceUpdatedAt, watermark.SourceUpdatedAt, watermark.FinanceServiceID).
		Order("updated_at ASC, id ASC").
		Limit(limit).
		Find(&services).Error
	return services, err
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

func buildFinanceEmbeddingText(item *serviceOfferEntity.FinanceService) string {
	parts := []string{
		"融资产品:" + item.Name,
		"类型:" + item.FinanceType,
		fmt.Sprintf("额度:%.2f-%.2f", item.MinAmount, item.MaxAmount),
		fmt.Sprintf("期限:%d天", item.TermDays),
		fmt.Sprintf("年化费率:%.4f", item.AnnualRate),
		"手续费:" + item.FeeRule,
		"风控:" + item.RiskControlRule,
		"材料:" + strings.Join([]string(item.RequiredDocuments), " "),
		"适用场景:" + strings.Join([]string(item.SupportedScenarios), " "),
		"描述:" + item.Description,
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
