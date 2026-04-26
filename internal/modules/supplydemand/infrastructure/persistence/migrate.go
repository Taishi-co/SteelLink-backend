package persistence

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

func AutoMigrate() error {
	return database.GetDB().AutoMigrate(&entity.SupplyDemandPost{})
}
