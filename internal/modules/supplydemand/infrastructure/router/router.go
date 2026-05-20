package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	supplyDemandHandler := handler.NewSupplyDemandHandler()

	r.GET("/supply-demands", middleware.OptionalAuth(), supplyDemandHandler.List)
	r.GET("/supply-demands/stats", supplyDemandHandler.Stats)

	posts := r.Group("/supply-demands")
	posts.Use(middleware.AuthRequired())
	{
		posts.POST("", supplyDemandHandler.Create)
		posts.GET("/mine", supplyDemandHandler.ListMine)
		posts.PUT("/:id/status", supplyDemandHandler.UpdateStatus)
	}

	r.GET("/supply-demands/:id", middleware.OptionalAuth(), supplyDemandHandler.GetByID)
}
