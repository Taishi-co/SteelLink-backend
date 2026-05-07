package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	orderHandler := handler.NewOrderHandler()

	orders := r.Group("/orders")
	orders.Use(middleware.AuthRequired())
	{
		fulfillmentParticipant := middleware.RequireFulfillmentParticipant()
		buyerService := middleware.RequireTradeCapability("procurement")
		sellerService := middleware.RequireTradeCapability("sales")
		warehouseService := middleware.RequireProviderService("warehouse")
		logisticsService := middleware.RequireProviderService("logistics")
		financeService := middleware.RequireProviderService("finance")
		buyerOrSellerService := middleware.RequireTradeCapability("procurement", "sales")

		orders.GET("", fulfillmentParticipant, orderHandler.List)
		orders.POST("", buyerService, orderHandler.Create)
		orders.GET("/:id", fulfillmentParticipant, orderHandler.GetByID)
		orders.GET("/:id/events", fulfillmentParticipant, orderHandler.ListEvents)
		orders.POST("/:id/cancel", buyerService, orderHandler.Cancel)
		orders.POST("/:id/confirm", sellerService, orderHandler.ConfirmOrder)
		orders.POST("/:id/warehouse-confirm", warehouseService, orderHandler.WarehouseConfirm)
		orders.POST("/:id/warehouse-check", warehouseService, orderHandler.WarehouseConfirm)
		orders.POST("/:id/seller-confirm", sellerService, orderHandler.SellerConfirm)
		orders.POST("/:id/buyer-confirm", buyerService, orderHandler.BuyerConfirm)
		orders.POST("/:id/sign-contract", buyerOrSellerService, orderHandler.SignContract)
		orders.POST("/:id/contracts/sign", buyerOrSellerService, orderHandler.SignContract)
		orders.POST("/:id/contract", buyerOrSellerService, orderHandler.AttachContract)
		orders.POST("/:id/payment-voucher", buyerService, orderHandler.UploadPaymentVoucher)
		orders.POST("/:id/payments/direct", buyerService, orderHandler.DirectPayment)
		orders.POST("/:id/financing-requests", buyerService, orderHandler.CreateFinancingRequest)
		orders.POST("/:id/financing-requests/risk-result", financeService, orderHandler.SubmitRiskResult)
		orders.POST("/:id/financing-requests/buyer-confirm", buyerService, orderHandler.BuyerConfirmFinancing)
		orders.POST("/:id/financing-requests/debt-contract", financeService, orderHandler.GenerateDebtContract)
		orders.POST("/:id/financing-requests/debt-sign", buyerService, orderHandler.BuyerSignDebtAndPayFee)
		orders.POST("/:id/financing-requests/disburse", financeService, orderHandler.DisburseFinancing)
		orders.POST("/:id/pledges", financeService, orderHandler.RecordPledge)
		orders.POST("/:id/pledges/release", financeService, orderHandler.ReleasePledge)
		orders.POST("/:id/confirm-payment", sellerService, orderHandler.ConfirmPayment)
		orders.POST("/:id/logistics-biddings", sellerService, orderHandler.CreateLogisticsBidding)
		orders.POST("/:id/logistics-quotes", logisticsService, orderHandler.CreateLogisticsQuote)
		orders.POST("/:id/logistics-confirm", buyerOrSellerService, orderHandler.ConfirmLogisticsQuote)
		orders.POST("/:id/shipments/:action", logisticsService, orderHandler.ShipmentProgress)
		orders.POST("/:id/delivery-info", buyerService, orderHandler.SetDeliveryInfo)
		orders.POST("/:id/confirm-outbound", sellerService, orderHandler.ConfirmOutbound)
		orders.POST("/:id/confirm-receipt", buyerService, orderHandler.ConfirmReceipt)
		orders.POST("/:id/receipts/confirm", buyerService, orderHandler.ConfirmReceipt)
		orders.POST("/:id/reconcile", sellerService, orderHandler.Reconcile)
		orders.POST("/:id/archive", buyerOrSellerService, orderHandler.Archive)
	}
}
