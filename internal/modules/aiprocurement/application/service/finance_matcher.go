package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	serviceOfferEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/embedding"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/vectorstore"
	"go.uber.org/zap"
)

func (s *ProcurementService) matchFinancePlans(ctx context.Context, parsed entity.ParsedInstruction, amount float64, selectedFinanceServiceID uint64) entity.FinancePlans {
	if !parsed.FinancingRequired {
		return entity.FinancePlans{}
	}
	cfg := config.GetAIProcurement().VectorSearch
	logAIProcurementInfo("AI procurement finance matching started",
		zap.Float64("amount", amount),
		zap.String("usage", parsed.Usage),
		zap.Bool("vector_enabled", cfg.Enabled),
		zap.String("vector_provider", cfg.Provider),
		zap.String("finance_product_collection", cfg.FinanceProductCollection),
		zap.Uint64("selected_finance_service_id", selectedFinanceServiceID),
	)
	services := s.matchFinanceByVector(ctx, parsed, amount, cfg)
	if len(services) == 0 {
		services = s.matchFinanceByRule(ctx, parsed, amount)
	}
	plans := financePlansFromServices(services, parsed, amount, selectedFinanceServiceID)
	if selectedFinanceServiceID > 0 {
		plans = ensureSelectedFinancePlan(ctx, plans, selectedFinanceServiceID, parsed, amount)
	}
	sort.SliceStable(plans, func(i, j int) bool {
		if plans[i].Selected != plans[j].Selected {
			return plans[i].Selected
		}
		if plans[i].AIRecommended != plans[j].AIRecommended {
			return plans[i].AIRecommended
		}
		return plans[i].Rate < plans[j].Rate
	})
	if len(plans) > 3 {
		plans = plans[:3]
	}
	return plans
}

func (s *ProcurementService) matchFinanceByVector(ctx context.Context, parsed entity.ParsedInstruction, amount float64, cfg config.VectorSearchConfig) []*serviceOfferEntity.FinanceService {
	if !cfg.Enabled || (cfg.Provider != "" && cfg.Provider != "pgvector") {
		logAIProcurementInfo("AI procurement finance vector matching skipped",
			zap.String("reason", vectorMatcherSkipReason(cfg)),
		)
		return nil
	}
	vector, err := embedding.NewClient(cfg).Embed(ctx, buildFinanceQueryText(parsed, amount))
	if err != nil {
		logAIProcurementWarn("AI procurement finance vector embedding failed; falling back to rule matching", zap.Error(err))
		return nil
	}
	hits, err := vectorstore.NewPGVectorClient(cfg).SearchFinance(ctx, vector, cfg.TopK)
	if err != nil {
		logAIProcurementWarn("AI procurement finance vector search failed; falling back to rule matching", zap.Error(err))
		return nil
	}
	if len(hits) == 0 {
		logAIProcurementInfo("AI procurement finance vector matching returned no hits")
		return nil
	}
	ids := make([]uint64, 0, len(hits))
	for _, hit := range hits {
		if amountCovered(hit.MinAmount, hit.MaxAmount, amount) {
			ids = append(ids, hit.FinanceServiceID)
		}
	}
	services := loadApprovedFinanceServicesByIDs(ctx, ids)
	if len(services) > 0 {
		logAIProcurementInfo("AI procurement finance vector matching succeeded", zap.Int("matched_count", len(services)))
	}
	return services
}

func (s *ProcurementService) matchFinanceByRule(ctx context.Context, parsed entity.ParsedInstruction, amount float64) []*serviceOfferEntity.FinanceService {
	db := database.GetDB()
	if db == nil {
		return nil
	}
	var services []*serviceOfferEntity.FinanceService
	query := db.WithContext(ctx).
		Where("status = ?", serviceOfferEntity.OfferStatusApproved).
		Where("(min_amount = 0 OR min_amount <= ?) AND (max_amount = 0 OR max_amount >= ?)", amount, amount)
	if err := query.Order("annual_rate ASC, term_days DESC, id ASC").Limit(10).Find(&services).Error; err != nil {
		logAIProcurementWarn("AI procurement finance rule query failed", zap.Error(err))
		return nil
	}
	logAIProcurementInfo("AI procurement finance rule query finished", zap.Int("matched_count", len(services)))
	return services
}

func financePlansFromServices(services []*serviceOfferEntity.FinanceService, parsed entity.ParsedInstruction, amount float64, selectedFinanceServiceID uint64) entity.FinancePlans {
	plans := make(entity.FinancePlans, 0, len(services))
	seen := map[uint64]struct{}{}
	for idx, item := range services {
		if item == nil || item.ID == 0 {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		recommendedAmount := amount
		if item.MaxAmount > 0 {
			recommendedAmount = math.Min(recommendedAmount, item.MaxAmount)
		}
		if item.MinAmount > 0 && recommendedAmount < item.MinAmount {
			recommendedAmount = item.MinAmount
		}
		selected := selectedFinanceServiceID == item.ID
		plans = append(plans, entity.FinancePlan{
			FinanceServiceID:    item.ID,
			FundingEnterpriseID: item.EnterpriseID,
			Provider:            fmt.Sprintf("资金方企业%d", item.EnterpriseID),
			ProductName:         item.Name,
			FinanceType:         item.FinanceType,
			Amount:              recommendedAmount,
			MinAmount:           item.MinAmount,
			MaxAmount:           item.MaxAmount,
			Rate:                item.AnnualRate,
			TermDays:            item.TermDays,
			RequiredDocuments:   []string(item.RequiredDocuments),
			SupportedScenarios:  []string(item.SupportedScenarios),
			Matched:             true,
			AIRecommended:       idx == 0 || selected,
			Selected:            selected,
			Reason:              financeMatchReason(parsed, amount, item, idx == 0 || selected),
		})
	}
	return plans
}

func ensureSelectedFinancePlan(ctx context.Context, plans entity.FinancePlans, selectedFinanceServiceID uint64, parsed entity.ParsedInstruction, amount float64) entity.FinancePlans {
	for idx := range plans {
		if plans[idx].FinanceServiceID == selectedFinanceServiceID {
			plans[idx].Selected = true
			plans[idx].AIRecommended = true
			return plans
		}
	}
	services := loadApprovedFinanceServicesByIDs(ctx, []uint64{selectedFinanceServiceID})
	if len(services) == 0 {
		return plans
	}
	selected := financePlansFromServices(services, parsed, amount, selectedFinanceServiceID)
	return append(selected, plans...)
}

func loadApprovedFinanceServicesByIDs(ctx context.Context, ids []uint64) []*serviceOfferEntity.FinanceService {
	if len(ids) == 0 {
		return nil
	}
	db := database.GetDB()
	if db == nil {
		return nil
	}
	var services []*serviceOfferEntity.FinanceService
	if err := db.WithContext(ctx).
		Where("id IN ? AND status = ?", ids, serviceOfferEntity.OfferStatusApproved).
		Find(&services).Error; err != nil {
		logAIProcurementWarn("AI procurement finance service lookup failed", zap.Error(err))
		return nil
	}
	byID := make(map[uint64]*serviceOfferEntity.FinanceService, len(services))
	for _, service := range services {
		byID[service.ID] = service
	}
	ordered := make([]*serviceOfferEntity.FinanceService, 0, len(services))
	for _, id := range ids {
		if service := byID[id]; service != nil {
			ordered = append(ordered, service)
		}
	}
	return ordered
}

func financeMatchReason(parsed entity.ParsedInstruction, amount float64, item *serviceOfferEntity.FinanceService, recommended bool) string {
	reasons := []string{"额度覆盖当前订单"}
	if parsed.Usage != "" && containsFinanceScenario(item, parsed.Usage) {
		reasons = append(reasons, "适用场景匹配")
	}
	if item.TermDays > 0 {
		reasons = append(reasons, fmt.Sprintf("账期%d天", item.TermDays))
	}
	if item.AnnualRate > 0 {
		reasons = append(reasons, fmt.Sprintf("年化%.1f%%", item.AnnualRate))
	}
	if recommended {
		reasons = append(reasons, "AI推荐优先展示")
	}
	return strings.Join(reasons, "、")
}

func containsFinanceScenario(item *serviceOfferEntity.FinanceService, usage string) bool {
	for _, scenario := range item.SupportedScenarios {
		if strings.Contains(scenario, usage) || strings.Contains(usage, scenario) {
			return true
		}
	}
	return false
}

func amountCovered(minAmount, maxAmount, amount float64) bool {
	return (minAmount <= 0 || amount >= minAmount) && (maxAmount <= 0 || amount <= maxAmount)
}

func buildFinanceQueryText(parsed entity.ParsedInstruction, amount float64) string {
	parts := []string{
		"融资需求:" + parsed.ProductName,
		"品类:" + parsed.Category,
		fmt.Sprintf("订单金额:%.2f", amount),
		"用途:" + parsed.Usage,
		"关键词:" + strings.Join(parsed.Keywords, " "),
	}
	return strings.Join(compactEmbeddingParts(parts), "；")
}
