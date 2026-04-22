package router

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
)

// RegisterRoutes 注册商品模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	productHandler := handler.NewProductHandler()

	r.GET("/products", middleware.OptionalAuth(), productHandler.List)
	r.GET("/products/:id", middleware.OptionalAuth(), productHandler.GetByID)

	products := r.Group("/products")
	products.Use(middleware.AuthRequired())
	{
		products.POST("", productHandler.Create)
		products.PUT("/:id/status", productHandler.UpdateStatus)
	}
}
