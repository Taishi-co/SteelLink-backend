package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	orderHandler := handler.NewOrderHandler()

	orders := r.Group("/orders")
	orders.Use(middleware.AuthRequired(), middleware.RequireUserType("enterprise"))
	{
		orders.GET("", orderHandler.List)
		orders.POST("", orderHandler.Create)
		orders.GET("/:id", orderHandler.GetByID)
		orders.POST("/:id/cancel", orderHandler.Cancel)
		orders.POST("/:id/payment-voucher", orderHandler.UploadPaymentVoucher)
		orders.POST("/:id/confirm-payment", orderHandler.ConfirmPayment)
		orders.POST("/:id/delivery-info", orderHandler.SetDeliveryInfo)
		orders.POST("/:id/confirm-outbound", orderHandler.ConfirmOutbound)
		orders.POST("/:id/confirm-receipt", orderHandler.ConfirmReceipt)
	}
}
