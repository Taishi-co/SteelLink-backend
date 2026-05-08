package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	productHandler := handler.NewProductHandler()

	r.GET("/product-categories", productHandler.ListCategories)
	r.GET("/products", middleware.OptionalAuth(), productHandler.List)
	r.GET("/products/:id", middleware.OptionalAuth(), productHandler.GetByID)

	products := r.Group("/products")
	products.Use(middleware.AuthRequired(), middleware.RequireTradeCapability("sales"))
	{
		products.POST("", productHandler.Create)
		products.PUT("/:id/display", productHandler.UpdateDisplay)
	}

	r.PUT("/products/:id/status", middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"), productHandler.UpdateStatus)

	categories := r.Group("/product-categories")
	categories.Use(middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"))
	{
		categories.POST("", productHandler.CreateCategory)
		categories.PUT("/:id", productHandler.UpdateCategory)
		categories.DELETE("/:id", productHandler.DeleteCategory)
	}
}
