package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	procurementHandler := handler.NewProcurementHandler()

	internalGroup := r.Group("/internal/ai-procurements")
	{
		internalGroup.POST("/vector-sync", procurementHandler.SyncVectors)
	}

	group := r.Group("/ai-procurements")
	group.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement"))
	{
		group.POST("/parse", procurementHandler.Parse)
		group.POST("/confirm-instruction", procurementHandler.ConfirmInstruction)
		group.POST("/match-products", procurementHandler.MatchProducts)
		group.POST("/generate-order", procurementHandler.GenerateOrder)
		group.POST("", procurementHandler.Create)
		group.GET("/mine", procurementHandler.ListMine)
		group.GET("/:id", procurementHandler.GetByID)
		group.POST("/:id/match-finance", procurementHandler.MatchFinance)
		group.POST("/:id/submit-review", procurementHandler.SubmitReview)
	}
}
