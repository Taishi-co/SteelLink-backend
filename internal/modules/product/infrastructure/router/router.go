package router

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	productHandler := handler.NewProductHandler()

	r.GET("/products", middleware.OptionalAuth(), productHandler.List)
	r.GET("/products/:id", middleware.OptionalAuth(), productHandler.GetByID)

	products := r.Group("/products")
	products.Use(middleware.AuthRequired(), middleware.RequireUserType("enterprise"))
	{
		products.POST("", productHandler.Create)
		products.PUT("/:id/status", productHandler.UpdateStatus)
	}
}
