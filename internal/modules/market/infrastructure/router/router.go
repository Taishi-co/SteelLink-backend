package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	marketHandler := handler.NewMarketHandler()

	r.GET("/biddings/stats", marketHandler.GetBiddingStats)
	r.GET("/biddings", middleware.OptionalAuth(), marketHandler.ListBiddings)
	r.GET("/biddings/:id", middleware.OptionalAuth(), marketHandler.GetBidding)
	r.POST("/biddings", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.CreateBidding)
	r.POST("/biddings/:id/quotes", middleware.AuthRequired(), middleware.RequireTradeCapability("sales"), marketHandler.SubmitBiddingQuote)
	r.POST("/biddings/:id/award", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.AwardBidding)
	r.POST("/biddings/:id/orders", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.CreateBiddingOrder)

	r.GET("/procurements/stats", marketHandler.GetProcurementStats)
	r.GET("/procurements", marketHandler.ListProcurements)
	r.GET("/procurements/mine", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.ListMyJoinedProcurements)
	r.GET("/procurements/seller/mine", middleware.AuthRequired(), middleware.RequireTradeCapability("sales"), marketHandler.ListSellerProcurements)
	r.POST("/procurements/requests", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.CreateProcurementRequest)
	r.GET("/procurements/requests/mine", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.ListMyProcurementRequests)
	r.GET("/procurements/invitations/mine", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.ListMyProcurementInvitations)
	r.POST("/procurements/invitations/:id/accept", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.AcceptProcurementInvitation)
	r.GET("/procurements/:id", marketHandler.GetProcurement)
	r.POST("/procurements/:id/seller-response", middleware.AuthRequired(), middleware.RequireTradeCapability("sales"), marketHandler.RespondProcurement)
	r.POST("/procurements/:id/join", middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"), marketHandler.JoinProcurement)
}
