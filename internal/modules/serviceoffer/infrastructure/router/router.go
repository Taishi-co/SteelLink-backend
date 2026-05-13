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

	buyerOrSeller := middleware.RequireTradeCapability("procurement", "sales")

	warehouseQuery := r.Group("/warehouse-services")
	warehouseQuery.Use(middleware.AuthRequired(), buyerOrSeller)
	{
		warehouseQuery.GET("/my-goods", h.ListWarehousesForMyGoods)
	}

	warehouse := r.Group("/warehouse-services")
	warehouse.Use(middleware.AuthRequired(), middleware.RequireProviderService("warehouse"))
	{
		warehouse.POST("", h.CreateWarehouse)
		warehouse.GET("/mine", h.ListMyWarehouses)
		warehouse.PUT("/:id", h.UpdateWarehouse)
		warehouse.DELETE("/:id", h.DeleteWarehouse)
	}
	r.PUT("/warehouse-services/:id/audit", middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"), h.AuditWarehouse)

	logisticsQuery := r.Group("/logistics-services")
	logisticsQuery.Use(middleware.AuthRequired(), buyerOrSeller)
	{
		logisticsQuery.GET("/my-shipments", h.ListLogisticsForMyShipments)
	}

	logistics := r.Group("/logistics-services")
	logistics.Use(middleware.AuthRequired(), middleware.RequireProviderService("logistics"))
	{
		logistics.POST("", h.CreateLogistics)
		logistics.GET("/mine", h.ListMyLogistics)
		logistics.PUT("/:id", h.UpdateLogistics)
		logistics.DELETE("/:id", h.DeleteLogistics)
	}
	r.PUT("/logistics-services/:id/audit", middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"), h.AuditLogistics)

	financeQuery := r.Group("/finance-services")
	financeQuery.Use(middleware.AuthRequired(), buyerOrSeller)
	{
		financeQuery.GET("/available-for-order/:order_id", h.ListFinanceAvailableForOrder)
	}

	finance := r.Group("/finance-services")
	finance.Use(middleware.AuthRequired(), middleware.RequireProviderService("finance"))
	{
		finance.POST("", h.CreateFinance)
		finance.GET("/mine", h.ListMyFinance)
		finance.PUT("/:id", h.UpdateFinance)
		finance.DELETE("/:id", h.DeleteFinance)
		finance.POST("/:id/authorize-warehouse", h.AuthorizeWarehouse)
		finance.DELETE("/authorized-warehouses/:auth_id", h.RevokeWarehouse)
		finance.GET("/:id/authorized-warehouses", h.ListAuthorizedWarehouses)
		finance.POST("/:id/authorize-logistics", h.AuthorizeLogistics)
		finance.DELETE("/authorized-logistics/:auth_id", h.RevokeLogistics)
	}
	r.PUT("/finance-services/:id/audit", middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"), h.AuditFinance)
}
