package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := handler.NewHandler()

	product := r.Group("/products")
	product.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("sales"))
	{
		product.PUT("/:id/warehouse-binding", h.BindProductWarehouse)
	}

	r.GET("/warehouses/:id/products", middleware.AuthRequired(), h.ListWarehouseProducts)
	r.GET("/warehouse-storage-requests", middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement", "sales", "warehouse"), h.ListWarehouseStorageRequests)
	r.POST("/warehouse-storage-requests", middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement", "sales"), h.CreateWarehouseStorageRequest)

	movements := r.Group("/warehouse-movements")
	movements.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement", "sales", "warehouse"))
	{
		movements.GET("", h.ListWarehouseMovements)
		movements.POST("", middleware.RequireEnterpriseService("warehouse"), h.CreateWarehouseMovement)
		movements.PUT("/:id/status", middleware.RequireEnterpriseService("warehouse"), h.UpdateWarehouseMovementStatus)
	}

	inventory := r.Group("/warehouse-inventory-snapshots")
	inventory.Use(middleware.AuthRequired(), middleware.RequireProviderService("warehouse"))
	{
		inventory.GET("", h.ListWarehouseInventorySnapshots)
		inventory.POST("", h.CreateWarehouseInventorySnapshot)
		inventory.PUT("/:id", h.UpdateWarehouseInventorySnapshot)
	}

	logistics := r.Group("/logistics")
	logistics.Use(middleware.AuthRequired())
	{
		logistics.POST("/estimate", h.EstimateLogistics)
		logistics.GET("/orders", h.ListLogisticsOrders)
		logistics.POST("/orders", middleware.RequireEnterpriseService("procurement", "logistics"), h.CreateLogisticsOrder)
		logistics.POST("/orders/:id/accept", middleware.RequireEnterpriseService("logistics"), h.AcceptLogisticsOrder)
		logistics.PUT("/orders/:id/status", middleware.RequireEnterpriseService("logistics", "warehouse", "procurement", "finance"), h.UpdateLogisticsOrderStatus)
		logistics.POST("/orders/:id/tracking-events", middleware.RequireEnterpriseService("logistics"), h.AddTrackingEvent)
		logistics.POST("/orders/:id/documents", middleware.RequireEnterpriseService("logistics", "warehouse", "procurement", "sales", "finance"), h.UploadFulfillmentDocument)
		logistics.GET("/orders/:id/documents", middleware.RequireEnterpriseService("procurement", "sales", "warehouse", "logistics", "finance"), h.ListFulfillmentDocuments)
		logistics.POST("/orders/:id/seal-confirm", middleware.RequireEnterpriseService("warehouse"), h.SealConfirm)
		logistics.POST("/orders/:id/inspection", middleware.RequireEnterpriseService("procurement", "warehouse"), h.SubmitInspection)
		logistics.GET("/drivers", middleware.RequireEnterpriseService("logistics"), h.ListLogisticsDrivers)
		logistics.POST("/drivers", middleware.RequireEnterpriseService("logistics"), h.CreateLogisticsDriver)
		logistics.PUT("/drivers/:id", middleware.RequireEnterpriseService("logistics"), h.UpdateLogisticsDriver)
		logistics.GET("/vehicles", middleware.RequireEnterpriseService("logistics"), h.ListLogisticsVehicles)
		logistics.POST("/vehicles", middleware.RequireEnterpriseService("logistics"), h.CreateLogisticsVehicle)
		logistics.PUT("/vehicles/:id", middleware.RequireEnterpriseService("logistics"), h.UpdateLogisticsVehicle)
	}

	finance := r.Group("/finance")
	finance.Use(middleware.AuthRequired(), middleware.RequireProviderService("finance"))
	{
		finance.GET("/collateral-monitor", h.ListCollateralMonitor)
	}

	orders := r.Group("/orders")
	orders.Use(middleware.AuthRequired())
	{
		orders.GET("/:id/service-fulfillment", h.GetOrderFulfillment)
	}

	guarantees := r.Group("/fulfillment-guarantees")
	guarantees.Use(middleware.AuthRequired(), middleware.RequireEnterpriseService("procurement", "sales", "warehouse", "logistics"))
	{
		guarantees.POST("", h.CreateGuarantee)
		guarantees.PUT("/:id/status", h.UpdateGuaranteeStatus)
	}
}
