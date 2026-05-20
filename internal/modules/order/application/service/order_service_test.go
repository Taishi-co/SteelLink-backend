package service

import (
	"testing"

	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
)

func TestInitialOrderStatusRequiresDestinationWarehouseSelectionWhenMissing(t *testing.T) {
	status := initialOrderStatus(&dto.CreateOrderRequest{})

	if status != orderEntity.OrderStatusPendingDestWarehouse {
		t.Fatalf("expected pending destination warehouse status, got %d", status)
	}
}

func TestInitialOrderStatusGoesToWarehouseConfirmWhenDestinationProvided(t *testing.T) {
	status := initialOrderStatus(&dto.CreateOrderRequest{DestWarehouseServiceID: 8801})

	if status != orderEntity.OrderStatusPendingWarehouseConfirm {
		t.Fatalf("expected pending warehouse confirm status, got %d", status)
	}
}
