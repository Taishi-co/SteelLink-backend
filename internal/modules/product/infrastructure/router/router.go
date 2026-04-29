package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	productHandler := handler.NewProductHandler()

	r.GET("/products", middleware.OptionalAuth(), productHandler.List)
	r.GET("/products/:id", middleware.OptionalAuth(), productHandler.GetByID)

	products := r.Group("/products")
	products.Use(middleware.AuthRequired(), middleware.RequireUserType("enterprise"))
	{
		products.POST("", productHandler.Create)
		products.PUT("/:id/display", productHandler.UpdateDisplay)
		products.PUT("/:id/status", productHandler.UpdateStatus)
	}
}
