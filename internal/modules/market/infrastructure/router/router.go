package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/handler"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	marketHandler := handler.NewMarketHandler()

	r.GET("/biddings/stats", marketHandler.GetBiddingStats)
	r.GET("/biddings", marketHandler.ListBiddings)
	r.GET("/biddings/:id", marketHandler.GetBidding)

	r.GET("/procurements/stats", marketHandler.GetProcurementStats)
	r.GET("/procurements", marketHandler.ListProcurements)
	r.GET("/procurements/:id", marketHandler.GetProcurement)
}
