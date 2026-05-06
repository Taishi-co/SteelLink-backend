package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/trade/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	tradeHandler := handler.NewTradeHandler()

	r.GET("/products", tradeHandler.ListProducts)
	r.GET("/products/:id", tradeHandler.GetProduct)

	seller := r.Group("/seller")
	seller.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("sales"))
	{
		seller.GET("/products", tradeHandler.ListSellerProducts)
		seller.POST("/products", tradeHandler.CreateSellerProduct)
		seller.PUT("/products/:id", tradeHandler.UpdateSellerProduct)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"))
	{
		admin.GET("/products", tradeHandler.ListAdminProducts)
		admin.PUT("/products/:id/audit", tradeHandler.AuditProduct)
		admin.GET("/orders", tradeHandler.ListAdminOrders)
		admin.GET("/orders/:id", tradeHandler.GetAdminOrder)
		admin.PUT("/orders/:id/payment-audit", tradeHandler.AuditOrderPayment)
	}

	orders := r.Group("/orders")
	orders.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement", "sales"))
	{
		orders.POST("", tradeHandler.CreateOrder)
		orders.GET("", tradeHandler.ListOrders)
		orders.GET("/:id", tradeHandler.GetOrder)
		orders.POST("/:id/contract/submit", tradeHandler.SubmitOrderContract)
		orders.POST("/:id/contract/sign", tradeHandler.SignContract)
		orders.POST("/:id/payment-receipts", tradeHandler.SubmitPaymentReceipt)
		orders.GET("/:id/pickup-code", tradeHandler.GetOrderPickupCode)
		orders.POST("/:id/pickup-code/verify", tradeHandler.VerifyPickupCode)
	}
}
