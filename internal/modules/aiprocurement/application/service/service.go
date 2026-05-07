package service

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	orderService "github.com/Taishi-co/SteelLink-backend/internal/modules/order/application/service"
	orderDTO "github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	serviceOfferEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"go.uber.org/zap"
)

type ProcurementService struct {
	procurementRepo repository.ProcurementRepository
	productRepo     productRepository.ProductRepository
	orderSvc        *orderService.OrderService
}

func NewProcurementService() *ProcurementService {
	productRepo := productPersistence.NewProductRepository()
	return &ProcurementService{
		procurementRepo: persistence.NewProcurementRepository(),
		productRepo:     productRepo,
		orderSvc:        orderService.NewOrderService(),
	}
}

func (s *ProcurementService) Parse(ctx context.Context, prompt string) (*dto.ParseResponse, error) {
	config.RefreshAIProcurementRuntimeConfig(ctx)
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, ecode.ErrBadRequest.WithMessage("采购指令不能为空")
	}
	parsed, ok, err := parseInstructionWithModel(ctx, prompt)
	if err != nil || !ok {
		if err != nil {
			logAIProcurementWarn("AI procurement parse fallback to rule parser",
				zap.String("reason", "model_parser_error"),
				zap.Error(err),
			)
		} else {
			logAIProcurementInfo("AI procurement parse fallback to rule parser",
				zap.String("reason", "model_parser_unavailable"),
			)
		}
		parsed = parseInstructionByRule(prompt)
		logAIProcurementInfo("AI procurement rule parser succeeded",
			zap.String("product_name", parsed.ProductName),
			zap.String("category", parsed.Category),
			zap.Float64("confidence", parsed.Confidence),
		)
	}
	return &dto.ParseResponse{
		Parsed:          parsed,
		MatchedProducts: entity.MatchedProducts{},
		OrderDraft:      entity.OrderDraft{},
		FinancePlans:    entity.FinancePlans{},
		AuditItems:      buildAuditItems(parsed, nil, nil),
		Stages:          buildStages(entity.StageInstructionParsing, false, parsed.FinancingRequired),
	}, nil
}

func (s *ProcurementService) ConfirmInstruction(ctx context.Context, req *dto.ConfirmInstructionRequest) (*dto.ParseResponse, error) {
	parsed := normalizeParsedInput(req.Parsed, req.Prompt)
	return &dto.ParseResponse{
		Parsed:          parsed,
		MatchedProducts: entity.MatchedProducts{},
		OrderDraft:      entity.OrderDraft{},
		FinancePlans:    entity.FinancePlans{},
		AuditItems:      buildAuditItems(parsed, nil, nil),
		Stages:          buildStages(entity.StageProductMatching, false, parsed.FinancingRequired),
	}, nil
}

func (s *ProcurementService) MatchProducts(ctx context.Context, req *dto.MatchProductsRequest) (*dto.MatchProductsResponse, error) {
	config.RefreshAIProcurementRuntimeConfig(ctx)
	parsed := normalizeParsedInput(req.Parsed, "")
	matches, err := s.matchProducts(ctx, parsed)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("商品匹配失败")
	}
	return &dto.MatchProductsResponse{
		MatchedProducts: matches,
		Stages:          buildStages(entity.StageProductMatching, false, parsed.FinancingRequired),
	}, nil
}

func (s *ProcurementService) GenerateOrder(ctx context.Context, req *dto.GenerateOrderRequest, userID, enterpriseID uint64) (*dto.GenerateOrderResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能生成AI采购订单")
	}
	parsed := normalizeParsedInput(req.Parsed, req.Prompt)
	selectedProducts := normalizeSelectedProducts(req)
	if len(selectedProducts) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请选择召回商品后再生成订单")
	}
	for idx := range selectedProducts {
		selectedProducts[idx].Selected = true
		if selectedProducts[idx].Quantity <= 0 {
			return nil, ecode.ErrBadRequest.WithMessage("拼单商品数量必须大于0")
		}
		selectedProducts[idx].TotalAmount = selectedProducts[idx].Quantity * selectedProducts[idx].UnitPrice
	}
	selectedProducts = s.enrichMatchesWithLogistics(ctx, parsed, selectedProducts)
	orderInfos, err := s.createOrdersBySeller(ctx, selectedProducts, req.OrderForm, userID, enterpriseID)
	if err != nil {
		return nil, err
	}
	orderDraft := buildOrderDraftFromForm(parsed, selectedProducts, req.OrderForm, orderInfos)
	financePlans := entity.FinancePlans{}
	nextStage := entity.StageReviewSubmission
	status := entity.ProcurementStatusOrderGenerated
	if parsed.FinancingRequired {
		nextStage = entity.StageFinancingMatching
		status = entity.ProcurementStatusFinancingReady
	}
	procurement := &entity.AIProcurement{
		Prompt:            strings.TrimSpace(req.Prompt),
		ProductName:       parsed.ProductName,
		Category:          parsed.Category,
		Quantity:          totalSelectedQuantity(selectedProducts),
		Unit:              parsed.Unit,
		Usage:             parsed.Usage,
		FinancingRequired: parsed.FinancingRequired,
		CurrentStage:      nextStage,
		Status:            status,
		ParsedInstruction: parsed,
		MatchedProducts:   selectedProducts,
		OrderDraft:        orderDraft,
		FinancePlans:      financePlans,
		AuditItems:        buildAuditItems(parsed, selectedProducts, financePlans),
		HistorySaved:      false,
		EnterpriseID:      enterpriseID,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if err := s.procurementRepo.Create(ctx, procurement); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建AI采购流程记录失败")
	}
	return &dto.GenerateOrderResponse{Procurement: toProcurementInfo(procurement), OrderDraft: orderDraft}, nil
}

func (s *ProcurementService) MatchFinance(ctx context.Context, id, userID, enterpriseID uint64, req *dto.MatchFinanceRequest) (*dto.ProcurementInfo, error) {
	item, err := s.findOwned(ctx, id, userID, enterpriseID, "匹配融资")
	if err != nil {
		return nil, err
	}
	item.FinancingRequired = req.FinancingRequired
	item.ParsedInstruction.FinancingRequired = req.FinancingRequired
	if req.FinancingRequired {
		item.FinancePlans = buildFinancePlans(item.ParsedInstruction, item.OrderDraft.EstimatedAmt)
		item.CurrentStage = entity.StageFinancingMatching
		item.Status = entity.ProcurementStatusFinancingReady
	} else {
		item.FinancePlans = entity.FinancePlans{}
		item.CurrentStage = entity.StageReviewSubmission
		item.Status = entity.ProcurementStatusOrderGenerated
	}
	item.AuditItems = buildAuditItems(item.ParsedInstruction, item.MatchedProducts, item.FinancePlans)
	item.UpdatedBy = userID
	if err := s.procurementRepo.Update(ctx, item); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("融资匹配更新失败")
	}
	return toProcurementInfo(item), nil
}

func (s *ProcurementService) Create(ctx context.Context, req *dto.CreateProcurementRequest, userID, enterpriseID uint64) (*dto.ProcurementInfo, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能发起AI采购")
	}
	parsedResult, err := s.Parse(ctx, req.Prompt)
	if err != nil {
		return nil, err
	}
	matches, err := s.matchProducts(ctx, parsedResult.Parsed)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("商品匹配失败")
	}
	parsedResult.MatchedProducts = matches
	parsedResult.OrderDraft = buildOrderDraft(parsedResult.Parsed, matches)
	parsedResult.FinancePlans = buildFinancePlans(parsedResult.Parsed, parsedResult.OrderDraft.EstimatedAmt)
	parsedResult.AuditItems = buildAuditItems(parsedResult.Parsed, matches, parsedResult.FinancePlans)
	procurement := &entity.AIProcurement{
		Prompt:            strings.TrimSpace(req.Prompt),
		ProductName:       parsedResult.Parsed.ProductName,
		Category:          parsedResult.Parsed.Category,
		Quantity:          parsedResult.Parsed.Quantity,
		Unit:              parsedResult.Parsed.Unit,
		Usage:             parsedResult.Parsed.Usage,
		FinancingRequired: parsedResult.Parsed.FinancingRequired,
		CurrentStage:      entity.StageFinancingMatching,
		Status:            entity.ProcurementStatusFinancingReady,
		ParsedInstruction: parsedResult.Parsed,
		MatchedProducts:   parsedResult.MatchedProducts,
		OrderDraft:        parsedResult.OrderDraft,
		FinancePlans:      parsedResult.FinancePlans,
		AuditItems:        parsedResult.AuditItems,
		EnterpriseID:      enterpriseID,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if len(parsedResult.MatchedProducts) > 0 {
		procurement.CurrentStage = entity.StageOrderGeneration
		procurement.Status = entity.ProcurementStatusOrderGenerated
	}
	if procurement.FinancingRequired {
		procurement.CurrentStage = entity.StageFinancingMatching
		procurement.Status = entity.ProcurementStatusFinancingReady
	}
	if err := s.procurementRepo.Create(ctx, procurement); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建AI采购单失败")
	}
	return toProcurementInfo(procurement), nil
}

func (s *ProcurementService) ListMine(ctx context.Context, req *dto.ProcurementListRequest, userID, enterpriseID uint64) (*dto.ProcurementListResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	historySaved := true
	items, total, err := s.procurementRepo.List(ctx, &repository.ProcurementQuery{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Keyword:      req.Keyword,
		Status:       entity.ProcurementStatus(req.Status),
		EnterpriseID: enterpriseID,
		CreatedBy:    userID,
		HistorySaved: &historySaved,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	return &dto.ProcurementListResponse{List: toProcurementInfoList(items), Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *ProcurementService) GetByID(ctx context.Context, id, userID, enterpriseID uint64) (*dto.ProcurementInfo, error) {
	item, err := s.procurementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if item == nil {
		return nil, ecode.ErrNotFound.WithMessage("AI采购单不存在")
	}
	if item.CreatedBy != userID && item.EnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权查看该AI采购单")
	}
	return toProcurementInfo(item), nil
}

func (s *ProcurementService) SubmitReview(ctx context.Context, id, userID, enterpriseID uint64) (*dto.ProcurementInfo, error) {
	return s.SubmitReviewWithOptions(ctx, id, userID, enterpriseID, &dto.SubmitReviewRequest{SaveHistory: boolPtr(true)})
}

func (s *ProcurementService) SubmitReviewWithOptions(ctx context.Context, id, userID, enterpriseID uint64, req *dto.SubmitReviewRequest) (*dto.ProcurementInfo, error) {
	item, err := s.findOwned(ctx, id, userID, enterpriseID, "提交")
	if err != nil {
		return nil, err
	}
	if len(item.MatchedProducts) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("没有可提交审核的商品匹配结果")
	}
	if item.OrderDraft.OrderID == 0 && len(item.OrderDraft.Orders) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请先生成真实订单后再提交审核")
	}
	now := time.Now()
	item.CurrentStage = entity.StageReviewSubmission
	item.Status = entity.ProcurementStatusReviewSubmitted
	item.ReviewSubmittedAt = &now
	if req != nil && req.SaveHistory != nil {
		item.HistorySaved = *req.SaveHistory
	} else {
		item.HistorySaved = true
	}
	item.UpdatedBy = userID
	item.AuditItems = markAuditSubmitted(item.AuditItems)
	if err := s.procurementRepo.Update(ctx, item); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("提交审核失败")
	}
	return toProcurementInfo(item), nil
}

func (s *ProcurementService) findOwned(ctx context.Context, id, userID, enterpriseID uint64, action string) (*entity.AIProcurement, error) {
	item, err := s.procurementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if item == nil {
		return nil, ecode.ErrNotFound.WithMessage("AI采购单不存在")
	}
	if item.CreatedBy != userID && item.EnterpriseID != enterpriseID {
		return nil, ecode.ErrForbidden.WithMessage("无权" + action + "该AI采购单")
	}
	return item, nil
}

func parseInstructionByRule(prompt string) entity.ParsedInstruction {
	quantity, unit := extractQuantity(prompt)
	productName := extractProductName(prompt)
	keywords := buildKeywords(productName, prompt)
	return entity.ParsedInstruction{
		ProductName:       productName,
		Category:          inferCategory(productName, prompt),
		Quantity:          quantity,
		Unit:              unit,
		Usage:             extractUsage(prompt),
		FinancingRequired: strings.Contains(prompt, "融资") || strings.Contains(prompt, "账期"),
		Keywords:          keywords,
		Confidence:        0.86,
	}
}

func normalizeParsedInput(parsed entity.ParsedInstruction, prompt string) entity.ParsedInstruction {
	parsed.ProductName = strings.TrimSpace(parsed.ProductName)
	parsed.Category = strings.TrimSpace(parsed.Category)
	parsed.Unit = strings.TrimSpace(parsed.Unit)
	parsed.Usage = strings.TrimSpace(parsed.Usage)
	parsed.DeliveryAddress = strings.TrimSpace(parsed.DeliveryAddress)
	parsed.DeliveryAddressSource = strings.TrimSpace(parsed.DeliveryAddressSource)
	parsed.DeliveryWarehouseName = strings.TrimSpace(parsed.DeliveryWarehouseName)
	normalizeParsedInstruction(&parsed, prompt)
	return parsed
}

func (s *ProcurementService) matchProducts(ctx context.Context, parsed entity.ParsedInstruction) (entity.MatchedProducts, error) {
	cfg := config.GetAIProcurement().VectorSearch
	logAIProcurementInfo("AI procurement product matching started",
		zap.String("product_name", parsed.ProductName),
		zap.String("category", parsed.Category),
		zap.Float64("quantity", parsed.Quantity),
		zap.String("unit", parsed.Unit),
		zap.Bool("vector_enabled", cfg.Enabled),
		zap.String("vector_provider", cfg.Provider),
		zap.String("product_collection", cfg.ProductCollection),
		zap.Int("top_k", cfg.TopK),
	)
	if matcher := newProductVectorMatcher(s.productRepo); matcher != nil {
		matches, err := matcher.Match(ctx, parsed)
		if err == nil && len(matches) > 0 {
			matches = s.enrichMatchesWithLogistics(ctx, parsed, matches)
			logAIProcurementInfo("AI procurement vector matching succeeded",
				zap.Int("matched_count", len(matches)),
				zap.String("product_name", parsed.ProductName),
				zap.String("category", parsed.Category),
				zap.String("destination_address", parsed.DeliveryAddress),
			)
			return matches, nil
		}
		if err != nil {
			logAIProcurementWarn("AI procurement vector matching failed; falling back to rule matching",
				zap.String("product_name", parsed.ProductName),
				zap.String("category", parsed.Category),
				zap.Error(err),
			)
		} else {
			logAIProcurementInfo("AI procurement vector matching returned no matches; falling back to rule matching",
				zap.String("product_name", parsed.ProductName),
				zap.String("category", parsed.Category),
			)
		}
	} else {
		logAIProcurementInfo("AI procurement vector matching skipped; falling back to rule matching",
			zap.String("reason", vectorMatcherSkipReason(cfg)),
			zap.Bool("vector_enabled", cfg.Enabled),
			zap.String("vector_provider", cfg.Provider),
		)
	}
	matches, err := s.matchProductsByRule(ctx, parsed)
	if err != nil {
		return nil, err
	}
	return s.enrichMatchesWithLogistics(ctx, parsed, matches), nil
}

func (s *ProcurementService) matchProductsByRule(ctx context.Context, parsed entity.ParsedInstruction) (entity.MatchedProducts, error) {
	status := int(productEntity.ProductStatusOnSale)
	logAIProcurementInfo("AI procurement rule product query started",
		zap.String("stage", "keyword_category"),
		zap.String("keyword", parsed.ProductName),
		zap.String("category", parsed.Category),
		zap.Int("status", status),
	)
	products, _, err := s.productRepo.List(ctx, &productRepository.ProductQuery{
		Page:     1,
		PageSize: 10,
		Keyword:  parsed.ProductName,
		Category: parsed.Category,
		Status:   &status,
		SortBy:   "stock_desc",
	})
	if err != nil {
		logAIProcurementWarn("AI procurement rule product query failed",
			zap.String("stage", "keyword_category"),
			zap.String("keyword", parsed.ProductName),
			zap.String("category", parsed.Category),
			zap.Error(err),
		)
		return nil, err
	}
	logAIProcurementInfo("AI procurement rule product query finished",
		zap.String("stage", "keyword_category"),
		zap.Int("matched_count", len(products)),
		zap.String("keyword", parsed.ProductName),
		zap.String("category", parsed.Category),
	)
	if len(products) == 0 && parsed.Category != "" {
		logAIProcurementInfo("AI procurement rule product query started",
			zap.String("stage", "category_only"),
			zap.String("category", parsed.Category),
			zap.Int("status", status),
		)
		products, _, err = s.productRepo.List(ctx, &productRepository.ProductQuery{
			Page:     1,
			PageSize: 10,
			Category: parsed.Category,
			Status:   &status,
			SortBy:   "stock_desc",
		})
		if err != nil {
			logAIProcurementWarn("AI procurement rule product query failed",
				zap.String("stage", "category_only"),
				zap.String("category", parsed.Category),
				zap.Error(err),
			)
			return nil, err
		}
		logAIProcurementInfo("AI procurement rule product query finished",
			zap.String("stage", "category_only"),
			zap.Int("matched_count", len(products)),
			zap.String("category", parsed.Category),
		)
	}
	if len(products) == 0 {
		logAIProcurementWarn("AI procurement product matching returned reference plan",
			zap.String("reason", "no_spot_products_matched"),
			zap.String("product_name", parsed.ProductName),
			zap.String("category", parsed.Category),
		)
		return entity.MatchedProducts{
			{
				ProductName: parsed.ProductName,
				Category:    parsed.Category,
				Quantity:    parsed.Quantity,
				UnitPrice:   320,
				TotalAmount: parsed.Quantity * 320,
				Score:       72,
				MatchReason: "未命中现货商品，已生成参考采购方案",
			},
		}, nil
	}
	matches := make(entity.MatchedProducts, 0, len(products))
	for _, product := range products {
		qty := parsed.Quantity
		if product.AvailableQty > 0 && product.AvailableQty < qty {
			qty = product.AvailableQty
		}
		score := 82
		if strings.Contains(product.Name, parsed.ProductName) || strings.Contains(parsed.ProductName, product.Name) {
			score = 95
		}
		matches = append(matches, entity.MatchedProduct{
			ProductID:     product.ID,
			ProductNo:     product.ProductNo,
			ProductName:   product.Name,
			Category:      product.SubCategory,
			SteelMill:     product.SteelMill,
			Quantity:      qty,
			AvailableQty:  product.AvailableQty,
			UnitPrice:     product.UnitPrice,
			TotalAmount:   qty * product.UnitPrice,
			Score:         score,
			MatchReason:   "按商品名称、品类、可用库存综合匹配",
			EnterpriseID:  product.EnterpriseID,
			Spec:          structToMap(product.Spec),
			Warehouse:     structToMap(product.WarehouseInfo),
			BundleNo:      product.BundleNo,
			DeliveryDays:  product.DeliveryDays,
			QuoteValidity: product.QuoteValidity,
			QualityCert:   product.QualityCert,
			Description:   product.Description,
		})
	}
	logAIProcurementInfo("AI procurement rule matching succeeded",
		zap.Int("matched_count", len(matches)),
		zap.String("product_name", parsed.ProductName),
		zap.String("category", parsed.Category),
	)
	return matches, nil
}

func (s *ProcurementService) enrichMatchesWithLogistics(ctx context.Context, parsed entity.ParsedInstruction, matches entity.MatchedProducts) entity.MatchedProducts {
	if len(matches) == 0 {
		return matches
	}
	destination := strings.TrimSpace(parsed.DeliveryAddress)
	carriers := loadApprovedLogisticsServices(ctx)
	for idx := range matches {
		matches[idx].TotalAmount = matches[idx].Quantity * matches[idx].UnitPrice
		matches[idx].DestinationAddress = destination
		matches[idx].OriginWarehouseName = stringFromMap(matches[idx].Warehouse, "warehouse_name")
		matches[idx].OriginAddress = buildWarehouseAddress(matches[idx].Warehouse)
		if destination == "" {
			matches[idx].LandedTotal = matches[idx].TotalAmount
			matches[idx].UnitLandedPrice = safeDiv(matches[idx].LandedTotal, matches[idx].Quantity)
			continue
		}
		estimate := chooseBestLogisticsEstimate(matches[idx], carriers, destination)
		matches[idx].LogisticsServiceID = estimate.ServiceID
		matches[idx].LogisticsServiceName = estimate.ServiceName
		matches[idx].LogisticsRoute = estimate.Route
		matches[idx].FreightEstimate = estimate.Amount
		matches[idx].LandedTotal = matches[idx].TotalAmount + estimate.Amount
		matches[idx].UnitLandedPrice = safeDiv(matches[idx].LandedTotal, matches[idx].Quantity)
		if estimate.ServiceName != "" {
			matches[idx].MatchReason = strings.TrimSpace(matches[idx].MatchReason + "；已按收货地址估算平台物流到手价")
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		left := matches[i].LandedTotal
		right := matches[j].LandedTotal
		if left == 0 {
			left = matches[i].TotalAmount
		}
		if right == 0 {
			right = matches[j].TotalAmount
		}
		if math.Abs(left-right) > 0.0001 {
			return left < right
		}
		return matches[i].Score > matches[j].Score
	})
	if destination != "" {
		logAIProcurementInfo("AI procurement logistics landed-cost ranking completed",
			zap.Int("matched_count", len(matches)),
			zap.Int("carrier_count", len(carriers)),
			zap.String("destination_address", destination),
		)
	}
	return matches
}

type logisticsEstimate struct {
	ServiceID   uint64
	ServiceName string
	Route       string
	Amount      float64
}

func loadApprovedLogisticsServices(ctx context.Context) []serviceOfferEntity.LogisticsService {
	db := database.GetDB()
	if db == nil {
		return nil
	}
	var rows []serviceOfferEntity.LogisticsService
	if err := db.WithContext(ctx).Where("status = ?", serviceOfferEntity.OfferStatusApproved).Find(&rows).Error; err != nil {
		logAIProcurementWarn("AI procurement logistics services load failed", zap.Error(err))
		return nil
	}
	return rows
}

func chooseBestLogisticsEstimate(match entity.MatchedProduct, carriers []serviceOfferEntity.LogisticsService, destination string) logisticsEstimate {
	origin := strings.TrimSpace(match.OriginAddress)
	quantity := match.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	best := logisticsEstimate{Amount: quantity * 80}
	bestScore := -1
	for _, carrier := range carriers {
		rate := extractPriceRate(carrier.PriceRule)
		if rate <= 0 {
			rate = 80
		}
		amount := rate * quantity
		routeScore := logisticsRouteScore(origin, destination, carrier)
		if best.ServiceID == 0 || amount < best.Amount || (math.Abs(amount-best.Amount) < 0.0001 && routeScore > bestScore) {
			best = logisticsEstimate{
				ServiceID:   carrier.ID,
				ServiceName: carrier.Name,
				Route:       strings.Trim(strings.TrimSpace(carrier.RouteStart)+" → "+strings.TrimSpace(carrier.RouteEnd), " →"),
				Amount:      amount,
			}
			bestScore = routeScore
		}
	}
	return best
}

func logisticsRouteScore(origin, destination string, carrier serviceOfferEntity.LogisticsService) int {
	score := 0
	routeStart := strings.TrimSpace(carrier.RouteStart)
	routeEnd := strings.TrimSpace(carrier.RouteEnd)
	scope := strings.TrimSpace(carrier.ServiceScope)
	if routeStart != "" && strings.Contains(origin, routeStart) {
		score += 2
	}
	if routeEnd != "" && strings.Contains(destination, routeEnd) {
		score += 2
	}
	if scope != "" && (strings.Contains(scope, destination) || strings.Contains(scope, routeEnd)) {
		score++
	}
	return score
}

func extractPriceRate(rule string) float64 {
	re := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)`)
	matches := re.FindStringSubmatch(rule)
	if len(matches) < 2 {
		return 0
	}
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}
	return value
}

func buildWarehouseAddress(warehouse map[string]any) string {
	parts := []string{
		stringFromMap(warehouse, "province"),
		stringFromMap(warehouse, "city"),
		stringFromMap(warehouse, "district"),
		stringFromMap(warehouse, "address"),
	}
	return strings.TrimSpace(strings.Join(nonEmptyStrings(parts), ""))
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key]; ok && value != nil {
		return strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.TrimSpace(toString(value)), "<nil>", ""), "\""))
	}
	return ""
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.TrimSpace(fmtAny(value)), "<nil>", ""), "\""))
	}
}

func fmtAny(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func nonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func safeDiv(left, right float64) float64 {
	if right <= 0 {
		return 0
	}
	return left / right
}

func vectorMatcherSkipReason(cfg config.VectorSearchConfig) string {
	if !cfg.Enabled {
		return "vector_search_disabled"
	}
	if cfg.Provider != "" && cfg.Provider != "pgvector" {
		return "unsupported_vector_provider"
	}
	return "vector_matcher_unavailable"
}

func normalizeSelectedProducts(req *dto.GenerateOrderRequest) entity.MatchedProducts {
	if len(req.SelectedProducts) > 0 {
		selected := make(entity.MatchedProducts, 0, len(req.SelectedProducts))
		for _, item := range req.SelectedProducts {
			if item.ProductID == 0 {
				continue
			}
			selected = append(selected, item)
		}
		return selected
	}
	if req.SelectedProduct.ProductID == 0 {
		return entity.MatchedProducts{}
	}
	selected := req.SelectedProduct
	if selected.Quantity <= 0 {
		selected.Quantity = req.OrderForm.Quantity
	}
	return entity.MatchedProducts{selected}
}

func totalSelectedQuantity(selected entity.MatchedProducts) float64 {
	total := 0.0
	for _, item := range selected {
		total += item.Quantity
	}
	return total
}

func (s *ProcurementService) createOrdersBySeller(ctx context.Context, selected entity.MatchedProducts, form dto.OrderForm, userID, enterpriseID uint64) ([]*orderDTO.OrderInfo, error) {
	grouped := make(map[uint64]entity.MatchedProducts)
	order := make([]uint64, 0)
	for _, item := range selected {
		sellerID := item.EnterpriseID
		if sellerID == 0 {
			sellerID = item.ProductID
		}
		if _, exists := grouped[sellerID]; !exists {
			order = append(order, sellerID)
		}
		grouped[sellerID] = append(grouped[sellerID], item)
	}
	orders := make([]*orderDTO.OrderInfo, 0, len(grouped))
	for _, sellerID := range order {
		group := grouped[sellerID]
		orderItems := make([]orderDTO.CreateOrderItemRequest, 0, len(group))
		for _, item := range group {
			orderItems = append(orderItems, orderDTO.CreateOrderItemRequest{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
		orderInfo, err := s.orderSvc.Create(ctx, &orderDTO.CreateOrderRequest{
			Items:               orderItems,
			Source:              "ai_procurement",
			RestoreCartOnCancel: true,
			ContactName:         form.ContactName,
			ContactPhone:        form.ContactPhone,
			WarehouseName:       form.WarehouseName,
			DeliveryAddress:     form.DeliveryAddress,
			DeliveryMethod:      form.DeliveryMethod,
			PaymentMethod:       form.PaymentMethod,
			DeliveryNote:        form.DeliveryNote,
			Remark:              form.Remark,
		}, userID, enterpriseID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderInfo)
	}
	return orders, nil
}

func buildOrderDraft(parsed entity.ParsedInstruction, matches entity.MatchedProducts) entity.OrderDraft {
	amount := parsed.Quantity * 320
	if len(matches) > 0 && matches[0].TotalAmount > 0 {
		amount = matches[0].TotalAmount
	}
	return entity.OrderDraft{
		ProductName:   parsed.ProductName,
		Quantity:      parsed.Quantity,
		Unit:          parsed.Unit,
		EstimatedAmt:  amount,
		PaymentMethod: "线上合同 + 对公转账",
		DeliveryNote:  "按匹配供应商交付周期生成订单",
	}
}

func buildOrderDraftFromForm(parsed entity.ParsedInstruction, selected entity.MatchedProducts, form dto.OrderForm, orderInfos []*orderDTO.OrderInfo) entity.OrderDraft {
	quantity := totalSelectedQuantity(selected)
	amount := 0.0
	productID := uint64(0)
	productName := parsed.ProductName
	if len(selected) > 0 {
		productID = selected[0].ProductID
		if len(selected) == 1 {
			productName = selected[0].ProductName
		} else {
			productName = parsed.ProductName + "拼单"
		}
		for _, item := range selected {
			if item.LandedTotal > 0 {
				amount += item.LandedTotal
			} else {
				amount += item.Quantity * item.UnitPrice
			}
		}
	}
	orderID := uint64(0)
	orderNo := ""
	orderSummaries := make(entity.OrderSummaries, 0, len(orderInfos))
	orderDetail := map[string]any{}
	if len(orderInfos) > 0 {
		orderID = orderInfos[0].ID
		orderNo = orderInfos[0].OrderNo
		productAmount := 0.0
		orderDetails := make([]map[string]any, 0, len(orderInfos))
		for _, orderInfo := range orderInfos {
			if orderInfo == nil {
				continue
			}
			productAmount += orderInfo.TotalAmount
			orderSummaries = append(orderSummaries, entity.OrderSummary{
				OrderID:       orderInfo.ID,
				OrderNo:       orderInfo.OrderNo,
				SellerID:      orderInfo.SellerID,
				SellerName:    orderInfo.SellerName,
				Quantity:      orderInfo.TotalQuantity,
				TotalAmount:   orderInfo.TotalAmount,
				ProductCount:  len(orderInfo.Items),
				PaymentMethod: form.PaymentMethod,
			})
			orderDetails = append(orderDetails, structToMap(orderInfo))
		}
		amount = productAmount + sumFreightEstimate(selected)
		if len(orderSummaries) > 1 {
			orderNo = "MULTI"
		}
		orderDetail["orders"] = orderDetails
	}
	orderDetail["selected_products"] = selected
	orderDetail["product_amount"] = sumProductAmount(selected)
	orderDetail["freight_estimate"] = sumFreightEstimate(selected)
	deliveryNote := strings.TrimSpace(form.DeliveryNote)
	if deliveryNote == "" {
		deliveryNote = form.DeliveryMethod + "；" + form.DeliveryAddress
	}
	return entity.OrderDraft{
		OrderID:         orderID,
		OrderNo:         orderNo,
		Orders:          orderSummaries,
		ProductID:       productID,
		ProductName:     productName,
		Quantity:        quantity,
		Unit:            parsed.Unit,
		EstimatedAmt:    amount,
		PaymentMethod:   form.PaymentMethod,
		DeliveryNote:    deliveryNote,
		ContactName:     strings.TrimSpace(form.ContactName),
		ContactPhone:    strings.TrimSpace(form.ContactPhone),
		WarehouseName:   strings.TrimSpace(form.WarehouseName),
		DeliveryAddress: strings.TrimSpace(form.DeliveryAddress),
		DeliveryMethod:  strings.TrimSpace(form.DeliveryMethod),
		Remark:          strings.TrimSpace(form.Remark),
		OrderDetail:     orderDetail,
	}
}

func sumProductAmount(selected entity.MatchedProducts) float64 {
	total := 0.0
	for _, item := range selected {
		total += item.Quantity * item.UnitPrice
	}
	return total
}

func sumFreightEstimate(selected entity.MatchedProducts) float64 {
	total := 0.0
	for _, item := range selected {
		total += item.FreightEstimate
	}
	return total
}

func buildFinancePlans(parsed entity.ParsedInstruction, amount float64) entity.FinancePlans {
	if !parsed.FinancingRequired {
		return entity.FinancePlans{}
	}
	return entity.FinancePlans{
		{Provider: "钢链金融", ProductName: "订单融资", Amount: amount * 0.8, Rate: 6.2, TermDays: 90, Matched: true, Reason: "采购用途和订单金额满足订单融资初筛"},
		{Provider: "合作银行", ProductName: "短期经营贷", Amount: amount * 0.7, Rate: 6.8, TermDays: 120, Matched: true, Reason: "作为备选融资方案"},
	}
}

func buildAuditItems(parsed entity.ParsedInstruction, matches entity.MatchedProducts, financePlans entity.FinancePlans) entity.AuditItems {
	items := entity.AuditItems{
		{Code: "instruction", Name: "采购指令解析", Status: "passed", Required: true, Message: "已识别商品、数量、用途"},
		{Code: "product_match", Name: "商品匹配", Status: "passed", Required: true, Message: "已生成商品匹配方案"},
		{Code: "order_draft", Name: "订单草稿", Status: "passed", Required: true, Message: "已生成订单草稿"},
	}
	if len(matches) == 0 {
		items[1].Status = "pending"
		items[1].Message = "等待供应商报价或现货补充"
	}
	if parsed.FinancingRequired {
		status := "passed"
		message := "已匹配融资方案"
		if len(financePlans) == 0 {
			status = "pending"
			message = "等待融资产品匹配"
		}
		items = append(items, entity.AuditItem{Code: "finance", Name: "融资匹配", Status: status, Required: true, Message: message})
	}
	items = append(items, entity.AuditItem{Code: "review", Name: "审核提交", Status: "pending", Required: true, Message: "待用户确认后提交审核"})
	return items
}

func markAuditSubmitted(items entity.AuditItems) entity.AuditItems {
	for idx := range items {
		if items[idx].Code == "review" {
			items[idx].Status = "submitted"
			items[idx].Message = "已提交平台审核"
		}
	}
	return items
}

func extractQuantity(prompt string) (float64, string) {
	re := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)(吨|套|件|支|张|批|个|台)?`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) < 2 {
		return 1, "套"
	}
	qty, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || qty <= 0 {
		qty = 1
	}
	unit := "套"
	if len(matches) >= 3 && matches[2] != "" {
		unit = matches[2]
	}
	return qty, unit
}

func extractProductName(prompt string) string {
	re := regexp.MustCompile(`采购[0-9]*(?:\.[0-9]+)?(?:吨|套|件|支|张|批|个|台)?([^，,。；;]+)`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	for _, keyword := range []string{"工业轴承", "螺纹钢", "热轧卷", "冷轧卷", "钢板", "钢管"} {
		if strings.Contains(prompt, keyword) {
			return keyword
		}
	}
	return "采购物品"
}

func inferCategory(productName, prompt string) string {
	switch {
	case strings.Contains(productName, "轴承") || strings.Contains(prompt, "轴承"):
		return "工业备件"
	case strings.Contains(productName, "螺纹钢") || strings.Contains(productName, "钢筋"):
		return "螺纹钢"
	case strings.Contains(productName, "热轧"):
		return "热轧"
	case strings.Contains(productName, "冷轧"):
		return "冷轧"
	case strings.Contains(productName, "钢板"):
		return "中厚板"
	case strings.Contains(productName, "钢管"):
		return "钢管"
	default:
		return "钢材现货"
	}
}

func extractUsage(prompt string) string {
	re := regexp.MustCompile(`用于([^，,。；;]+)`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return "企业生产采购"
}

func buildKeywords(productName, prompt string) []string {
	keywords := []string{productName}
	for _, keyword := range []string{"融资", "维修", "现货", "急采", "设备"} {
		if strings.Contains(prompt, keyword) {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

func toProcurementInfoList(items []*entity.AIProcurement) []*dto.ProcurementInfo {
	result := make([]*dto.ProcurementInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toProcurementInfo(item))
	}
	return result
}

func toProcurementInfo(item *entity.AIProcurement) *dto.ProcurementInfo {
	if item == nil {
		return nil
	}
	return &dto.ProcurementInfo{
		ID:                item.ID,
		ProcurementNo:     item.ProcurementNo,
		Prompt:            item.Prompt,
		ProductName:       item.ProductName,
		Category:          item.Category,
		Quantity:          item.Quantity,
		Unit:              item.Unit,
		Usage:             item.Usage,
		FinancingRequired: item.FinancingRequired,
		CurrentStage:      item.CurrentStage,
		Status:            item.Status,
		ParsedInstruction: item.ParsedInstruction,
		MatchedProducts:   item.MatchedProducts,
		OrderDraft:        item.OrderDraft,
		FinancePlans:      item.FinancePlans,
		AuditItems:        item.AuditItems,
		Stages:            buildStages(item.CurrentStage, item.Status == entity.ProcurementStatusReviewSubmitted, item.FinancingRequired),
		HistorySaved:      item.HistorySaved,
		EnterpriseID:      item.EnterpriseID,
		CreatedBy:         item.CreatedBy,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		ReviewSubmittedAt: item.ReviewSubmittedAt,
	}
}

func buildStages(current entity.WorkflowStage, submitted bool, includeFinance bool) []dto.StageInfo {
	names := map[entity.WorkflowStage]string{
		entity.StageInstructionParsing: "指令解析",
		entity.StageProductMatching:    "商品匹配",
		entity.StageOrderGeneration:    "订单生成",
		entity.StageFinancingMatching:  "融资匹配",
		entity.StageReviewSubmission:   "审核提交",
	}
	descriptions := map[entity.WorkflowStage]string{
		entity.StageInstructionParsing: "识别采购品类、数量、用途和融资意图",
		entity.StageProductMatching:    "按商品、库存、价格和供应商能力匹配",
		entity.StageOrderGeneration:    "生成订单草稿和交易要素",
		entity.StageFinancingMatching:  "匹配订单融资和备选金融产品",
		entity.StageReviewSubmission:   "提交平台审核并进入后续履约",
	}
	workflowStages := entity.WorkflowStages()
	if !includeFinance {
		workflowStages = []entity.WorkflowStage{
			entity.StageInstructionParsing,
			entity.StageProductMatching,
			entity.StageOrderGeneration,
			entity.StageReviewSubmission,
		}
	}
	currentIndex := 0
	for idx, stage := range workflowStages {
		if stage == current {
			currentIndex = idx
			break
		}
	}
	stages := make([]dto.StageInfo, 0, len(workflowStages))
	for idx, stage := range workflowStages {
		status := "pending"
		if idx < currentIndex || submitted {
			status = "completed"
		}
		if idx == currentIndex && !submitted {
			status = "processing"
		}
		stages = append(stages, dto.StageInfo{Code: stage, Name: names[stage], Status: status, Description: descriptions[stage]})
	}
	return stages
}

func structToMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil
	}
	return result
}

func boolPtr(value bool) *bool {
	return &value
}
