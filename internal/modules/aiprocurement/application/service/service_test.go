package service

import (
	"testing"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	orderDTO "github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
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
