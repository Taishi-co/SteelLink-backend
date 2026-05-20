package service

import (
	"context"
	"testing"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	orderDTO "github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	serviceOfferEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func TestAIOrderGroupKeySplitsSameSellerByWarehouse(t *testing.T) {
	first := entity.MatchedProduct{
		ProductID: 1,
		Warehouse: map[string]any{
			"warehouse_service_id": 1001,
		},
	}
	second := entity.MatchedProduct{
		ProductID: 2,
		Warehouse: map[string]any{
			"warehouse_service_id": 1002,
		},
	}

	if aiOrderGroupKey(88, first) == aiOrderGroupKey(88, second) {
		t.Fatal("same seller with different warehouses must be split into different order groups")
	}
}

func TestAIOrderGroupKeyKeepsSameSellerWarehouseTogether(t *testing.T) {
	first := entity.MatchedProduct{
		ProductID: 1,
		Warehouse: map[string]any{
			"warehouse_id": 2001,
		},
	}
	second := entity.MatchedProduct{
		ProductID: 2,
		Warehouse: map[string]any{
			"warehouse_id": 2001,
		},
	}

	if aiOrderGroupKey(88, first) != aiOrderGroupKey(88, second) {
		t.Fatal("same seller with same warehouse should stay in one order group")
	}
}

func TestIsNegotiatedLogistics(t *testing.T) {
	if !isNegotiatedLogistics(entity.MatchedProduct{LogisticsMatchStatus: "negotiation"}) {
		t.Fatal("expected logistics_match_status=negotiation to be treated as negotiated logistics")
	}
	if !isNegotiatedLogistics(entity.MatchedProduct{LogisticsServiceName: "协商解决"}) {
		t.Fatal("expected 协商解决 logistics service name to be treated as negotiated logistics")
	}
}

func TestBuildAILogisticsOrderCreatesOrderForMatchedService(t *testing.T) {
	plan := aiOrderPlan{
		OrderInfo: &orderDTO.OrderInfo{ID: 99, OrderNo: "O202605092229592501"},
		Products: entity.MatchedProducts{
			{
				ProductID:          1,
				ProductName:        "热轧板卷",
				Quantity:           1,
				LogisticsServiceID: 9401,
				OriginAddress:      "上海宝山仓",
				DestinationAddress: "武汉青山",
				FreightEstimate:    80,
				Warehouse: map[string]any{
					"warehouse_service_id": 9301,
				},
			},
			{
				ProductID:          2,
				ProductName:        "中厚板",
				Quantity:           1,
				LogisticsServiceID: 9401,
				OriginAddress:      "上海宝山仓",
				DestinationAddress: "武汉青山",
				FreightEstimate:    80,
				Warehouse: map[string]any{
					"warehouse_service_id": 9301,
				},
			},
		},
	}

	order, ok := buildAILogisticsOrder(plan, dto.OrderForm{}, 7, 8, 0)
	if !ok {
		t.Fatal("expected matched logistics service to create a logistics order")
	}
	if order.OrderID != 99 || order.OrderNo != "O202605092229592501" {
		t.Fatalf("unexpected order linkage: %#v", order)
	}
	if order.LogisticsServiceID != 9401 || order.WarehouseServiceID != 9301 {
		t.Fatalf("unexpected service linkage: %#v", order)
	}
	if order.Weight != 2 {
		t.Fatalf("expected combined weight 2, got %v", order.Weight)
	}
}

func TestBuildAILogisticsOrderSkipsNegotiation(t *testing.T) {
	plan := aiOrderPlan{
		OrderInfo: &orderDTO.OrderInfo{ID: 99, OrderNo: "O202605092229592501"},
		Products: entity.MatchedProducts{
			{
				ProductID:            1,
				ProductName:          "热轧板卷",
				Quantity:             1,
				LogisticsMatchStatus: "negotiation",
				LogisticsServiceName: "协商解决",
				LogisticsServiceID:   9401,
				DestinationAddress:   "武汉青山",
				OriginWarehouseName:  "上海宝山仓",
			},
		},
	}

	if _, ok := buildAILogisticsOrder(plan, dto.OrderForm{}, 7, 8, 0); ok {
		t.Fatal("expected negotiated logistics to skip concrete carrier logistics order")
	}
}

func TestCreateLogisticsOrdersForAIOrdersSkipsWhenDBUnavailable(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	t.Cleanup(func() {
		database.DB = originalDB
	})

	err := (&ProcurementService{}).createLogisticsOrdersForAIOrders(
		context.Background(),
		[]aiOrderPlan{
			{
				OrderInfo: &orderDTO.OrderInfo{ID: 99, OrderNo: "O202605092229592501"},
				Products: entity.MatchedProducts{
					{
						ProductID:          1,
						ProductName:        "热轧板卷",
						Quantity:           1,
						LogisticsServiceID: 9401,
					},
				},
			},
		},
		dto.OrderForm{},
		7,
		8,
	)
	if err != nil {
		t.Fatalf("expected nil error when db is unavailable, got %v", err)
	}
}

func TestToProcurementInfoNormalizesEmptyCollections(t *testing.T) {
	info := toProcurementInfo(&entity.AIProcurement{
		ID:           1,
		CurrentStage: entity.StageReviewSubmission,
		Status:       entity.ProcurementStatusOrderGenerated,
	})

	if info == nil {
		t.Fatal("expected procurement info")
	}
	if info.ParsedInstruction.Keywords == nil {
		t.Fatal("expected parsed keywords to be an empty list")
	}
	if info.MatchedProducts == nil {
		t.Fatal("expected matched products to be an empty list")
	}
	if info.OrderDraft.Orders == nil {
		t.Fatal("expected order summaries to be an empty list")
	}
	if info.FinancePlans == nil {
		t.Fatal("expected finance plans to be an empty list")
	}
	if info.AuditItems == nil {
		t.Fatal("expected audit items to be an empty list")
	}
	if len(info.Stages) == 0 {
		t.Fatal("expected stages to be built")
	}
}

func TestFinancePlansFromServicesMarksSelectedAIRecommendation(t *testing.T) {
	plans := financePlansFromServices([]*serviceOfferEntity.FinanceService{
		{
			ID:                 11,
			EnterpriseID:       21,
			Name:               "订单融资A",
			FinanceType:        "订单融资",
			MinAmount:          1000,
			MaxAmount:          50000,
			TermDays:           90,
			AnnualRate:         6.2,
			RequiredDocuments:  serviceOfferEntity.StringList{"交易合同"},
			SupportedScenarios: serviceOfferEntity.StringList{"设备维修"},
		},
	}, entity.ParsedInstruction{Usage: "设备维修", FinancingRequired: true}, 12000, 11)

	if len(plans) != 1 {
		t.Fatalf("expected one finance plan, got %d", len(plans))
	}
	if plans[0].FinanceServiceID != 11 || !plans[0].Selected || !plans[0].AIRecommended {
		t.Fatalf("expected selected AI-recommended finance service, got %#v", plans[0])
	}
	if plans[0].Provider == "" || plans[0].ProductName != "订单融资A" {
		t.Fatalf("expected real finance product fields, got %#v", plans[0])
	}
}
