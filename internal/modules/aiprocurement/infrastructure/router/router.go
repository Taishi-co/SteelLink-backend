package router

import (
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	procurementHandler := handler.NewProcurementHandler()

	internalGroup := r.Group("/internal/ai-procurements")
	{
		internalGroup.POST("/vector-sync", procurementHandler.SyncVectors)
		internalGroup.GET("/vector-sync/status", procurementHandler.VectorSyncStatus)
	}

	group := r.Group("/ai-procurements")
	group.Use(middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"))
	{
		aiRateLimit := middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "ai-procurement",
			Limit:  60,
			Window: time.Hour,
			KeyFn:  middleware.RateLimitByUserOrIP,
		})
		group.POST("/parse", aiRateLimit, procurementHandler.Parse)
		group.POST("/confirm-instruction", procurementHandler.ConfirmInstruction)
		group.POST("/match-products", aiRateLimit, procurementHandler.MatchProducts)
		group.POST("/match-logistics-warehouse", aiRateLimit, procurementHandler.MatchLogisticsWarehouse)
		group.POST("/generate-order", procurementHandler.GenerateOrder)
		group.POST("", procurementHandler.Create)
		group.GET("/mine", procurementHandler.ListMine)
		group.GET("/:id", procurementHandler.GetByID)
		group.POST("/:id/match-finance", procurementHandler.MatchFinance)
		group.POST("/:id/submit-review", procurementHandler.SubmitReview)
	}
}
