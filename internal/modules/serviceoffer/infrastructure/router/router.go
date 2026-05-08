package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := handler.NewHandler()

	r.GET("/warehouse-services", middleware.OptionalAuth(), h.ListWarehouses)
	r.GET("/warehouse-services/:id", middleware.OptionalAuth(), h.GetWarehouse)
	r.GET("/logistics-services", middleware.OptionalAuth(), h.ListLogistics)
	r.GET("/logistics-services/:id", middleware.OptionalAuth(), h.GetLogistics)
	r.GET("/finance-services", middleware.OptionalAuth(), h.ListFinance)
	r.GET("/finance-services/:id", middleware.OptionalAuth(), h.GetFinance)

	warehouse := r.Group("/warehouse-services")
	warehouse.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("sales", "warehouse"))
	{
		warehouse.POST("", h.CreateWarehouse)
		warehouse.GET("/mine", h.ListMyWarehouses)
		warehouse.PUT("/:id", h.UpdateWarehouse)
		warehouse.DELETE("/:id", h.DeleteWarehouse)
	}

	logistics := r.Group("/logistics-services")
	logistics.Use(middleware.AuthRequired(), middleware.RequireProviderService("logistics"))
	{
		logistics.POST("", h.CreateLogistics)
		logistics.GET("/mine", h.ListMyLogistics)
		logistics.PUT("/:id", h.UpdateLogistics)
		logistics.DELETE("/:id", h.DeleteLogistics)
	}

	finance := r.Group("/finance-services")
	finance.Use(middleware.AuthRequired(), middleware.RequireProviderService("finance"))
	{
		finance.POST("", h.CreateFinance)
		finance.GET("/mine", h.ListMyFinance)
	}
}
