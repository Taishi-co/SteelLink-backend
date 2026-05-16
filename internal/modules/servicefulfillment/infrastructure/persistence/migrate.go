package persistence

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func AutoMigrate() error {
	return database.GetDB().AutoMigrate(
		&entity.WarehouseInventoryMovement{},
		&entity.WarehouseInventorySnapshot{},
		&entity.LogisticsOrder{},
		&entity.LogisticsDriver{},
		&entity.LogisticsVehicle{},
		&entity.LogisticsTrackingEvent{},
		&entity.FulfillmentDocument{},
		&entity.FulfillmentGuarantee{},
		&entity.WarehouseReceiptPledgeApplication{},
	)
}
