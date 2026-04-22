package router

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/trade/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	tradeHandler := handler.NewTradeHandler()

	r.GET("/products", tradeHandler.ListProducts)
	r.GET("/products/:id", tradeHandler.GetProduct)

	seller := r.Group("/seller")
	seller.Use(middleware.AuthRequired())
	{
		seller.GET("/products", tradeHandler.ListSellerProducts)
		seller.POST("/products", tradeHandler.CreateSellerProduct)
		seller.PUT("/products/:id", tradeHandler.UpdateSellerProduct)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	{
		admin.GET("/products", tradeHandler.ListAdminProducts)
		admin.PUT("/products/:id/audit", tradeHandler.AuditProduct)
	}

	orders := r.Group("/orders")
	orders.Use(middleware.AuthRequired())
	{
		orders.POST("", tradeHandler.CreateOrder)
		orders.GET("", tradeHandler.ListOrders)
		orders.GET("/:id", tradeHandler.GetOrder)
	}
}
