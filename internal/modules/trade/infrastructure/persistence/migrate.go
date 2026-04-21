package persistence

import (
	tradeentity "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func AutoMigrate() error {
	db := database.GetDB()
	return db.AutoMigrate(
		&tradeentity.Product{},
		&tradeentity.ProductAuditLog{},
		&tradeentity.Order{},
		&tradeentity.OrderItem{},
		&tradeentity.OrderContract{},
		&tradeentity.PaymentReceipt{},
		&tradeentity.PickupCode{},
	)
}
