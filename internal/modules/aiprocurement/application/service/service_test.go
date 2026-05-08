package service

import (
	"testing"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
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
