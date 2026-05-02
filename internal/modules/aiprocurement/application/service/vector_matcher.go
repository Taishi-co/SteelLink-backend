package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/embedding"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/vectorstore"
)

type productVectorMatcher struct {
	cfg         config.VectorSearchConfig
	embedder    *embedding.Client
	vectorStore *vectorstore.MilvusClient
	productRepo productRepository.ProductRepository
}

func newProductVectorMatcher(productRepo productRepository.ProductRepository) *productVectorMatcher {
	globalCfg := config.Get()
	if globalCfg == nil {
		return nil
	}
	cfg := globalCfg.AIProcurement.VectorSearch
	if !cfg.Enabled {
		return nil
	}
	return &productVectorMatcher{
		cfg:         cfg,
		embedder:    embedding.NewClient(cfg),
		vectorStore: vectorstore.NewMilvusClient(cfg),
		productRepo: productRepo,
	}
}

func (m *productVectorMatcher) Match(ctx context.Context, parsed entity.ParsedInstruction) (entity.MatchedProducts, error) {
	if m == nil {
		return nil, nil
	}
	queryText := buildVectorQueryText(parsed)
	vector, err := m.embedder.Embed(ctx, queryText)
	if err != nil {
		return nil, err
	}
	hits, err := m.vectorStore.SearchProducts(ctx, vector, m.cfg.TopK)
	if err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return nil, nil
	}

	matches := make(entity.MatchedProducts, 0, len(hits))
	seen := map[uint64]struct{}{}
	for _, hit := range hits {
		if _, ok := seen[hit.ProductID]; ok {
			continue
		}
		seen[hit.ProductID] = struct{}{}

		product, err := m.productRepo.FindByID(ctx, hit.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil || !product.IsOnSale() {
			continue
		}
		if parsed.Category != "" && product.Category != parsed.Category && product.SubCategory != parsed.Category {
			continue
		}

		qty := parsed.Quantity
		if product.AvailableQty > 0 && product.AvailableQty < qty {
			qty = product.AvailableQty
		}
		score := hybridScore(hit.Score, keywordScore(parsed, product), m.cfg)
		if score < minScore(m.cfg.MinScore) {
			continue
		}
		matches = append(matches, entity.MatchedProduct{
			ProductID:    product.ID,
			ProductNo:    product.ProductNo,
			ProductName:  product.Name,
			Category:     product.SubCategory,
			SteelMill:    product.SteelMill,
			Quantity:     qty,
			UnitPrice:    product.UnitPrice,
			TotalAmount:  qty * product.UnitPrice,
			Score:        int(math.Round(score)),
			MatchReason:  fmt.Sprintf("向量召回匹配，结合商品语义、品类和库存重排，Milvus相似度 %.2f", hit.Score),
			EnterpriseID: product.EnterpriseID,
		})
	}
	return matches, nil
}

func buildVectorQueryText(parsed entity.ParsedInstruction) string {
	parts := []string{
		"采购商品:" + parsed.ProductName,
		"品类:" + parsed.Category,
		"数量:" + fmt.Sprintf("%.2f%s", parsed.Quantity, parsed.Unit),
		"用途:" + parsed.Usage,
	}
	if parsed.FinancingRequired {
		parts = append(parts, "需要融资")
	}
	if len(parsed.Keywords) > 0 {
		parts = append(parts, "关键词:"+strings.Join(parsed.Keywords, " "))
	}
	return strings.Join(parts, "；")
}

func keywordScore(parsed entity.ParsedInstruction, product *productEntity.Product) float64 {
	score := 60.0
	if parsed.ProductName != "" && (strings.Contains(product.Name, parsed.ProductName) || strings.Contains(parsed.ProductName, product.Name)) {
		score += 25
	}
	if parsed.Category != "" && (product.Category == parsed.Category || product.SubCategory == parsed.Category) {
		score += 10
	}
	if parsed.Quantity > 0 && product.AvailableQty >= parsed.Quantity {
		score += 5
	}
	if score > 100 {
		return 100
	}
	return score
}

func hybridScore(vectorScore float64, keyword float64, cfg config.VectorSearchConfig) float64 {
	normalizedVector := vectorScore
	if normalizedVector <= 1 {
		normalizedVector = normalizedVector * 100
	}
	if normalizedVector > 100 {
		normalizedVector = 100
	}
	if normalizedVector < 0 {
		normalizedVector = 0
	}
	totalWeight := cfg.HybridKeywordWeight + cfg.HybridVectorWeight
	if totalWeight <= 0 {
		return normalizedVector
	}
	return (keyword*cfg.HybridKeywordWeight + normalizedVector*cfg.HybridVectorWeight) / totalWeight
}

func minScore(value float64) float64 {
	if value <= 1 {
		return value * 100
	}
	return value
}
