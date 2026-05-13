package persistence

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func AutoMigrate() error {
	return database.GetDB().AutoMigrate(
		&entity.Order{},
		&entity.OrderEvent{},
		&entity.LogisticsBidding{},
		&entity.LogisticsBidQuote{},
		&entity.ShipmentTask{},
		&entity.LogisticsContract{},
		&entity.LogisticsSettlement{},
		&entity.OrderFinancingRequest{},
		&entity.PledgeReleaseEvent{},
		&entity.BreachRecord{},
		&entity.LitigationRecord{},
	)
}
