package handler

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	userEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type PlatformStatsResponse struct {
	MerchantCount     int64   `json:"merchant_count"`
	SKUCount          int64   `json:"sku_count"`
	OrderCount        int64   `json:"order_count"`
	TransactionAmount float64 `json:"transaction_amount"`
}

// PlatformStats returns public aggregate metrics for the buyer home page.
func PlatformStats(c *gin.Context) {
	const key = "platform:stats"
	var cached PlatformStatsResponse
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}

	db := database.GetDB().WithContext(c.Request.Context())

	var stats PlatformStatsResponse
	if err := db.Model(&userEntity.Enterprise{}).
		Where("status = ?", userEntity.EnterpriseStatusApproved).
		Count(&stats.MerchantCount).Error; err != nil {
		response.InternalError(c, "查询入驻商家统计失败")
		return
	}

	if err := db.Model(&productEntity.Product{}).
		Where("status = ?", productEntity.ProductStatusOnSale).
		Count(&stats.SKUCount).Error; err != nil {
		response.InternalError(c, "查询SKU商品统计失败")
		return
	}

	validOrderStatuses := []entity.OrderStatus{
		entity.OrderStatusPendingConfirm,
		entity.OrderStatusPendingContract,
		entity.OrderStatusPendingPayment,
		entity.OrderStatusPaid,
		entity.OrderStatusPendingDelivery,
		entity.OrderStatusDelivering,
		entity.OrderStatusDelivered,
		entity.OrderStatusCompleted,
		entity.OrderStatusFrozen,
		entity.OrderStatusException,
		entity.OrderStatusPendingWarehouseConfirm,
		entity.OrderStatusPendingSellerConfirm,
		entity.OrderStatusPendingBuyerConfirm,
		entity.OrderStatusLogisticsBidding,
		entity.OrderStatusPendingLogisticsConfirm,
		entity.OrderStatusShipping,
		entity.OrderStatusReceived,
		entity.OrderStatusReconciled,
		entity.OrderStatusArchived,
		entity.OrderStatusFinancingReview,
		entity.OrderStatusPendingFinancingContract,
	}

	if err := db.Model(&entity.Order{}).
		Where("status IN ?", validOrderStatuses).
		Count(&stats.OrderCount).Error; err != nil {
		response.InternalError(c, "查询累计订单统计失败")
		return
	}

	if err := db.Model(&entity.Order{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", validOrderStatuses).
		Scan(&stats.TransactionAmount).Error; err != nil {
		response.InternalError(c, "查询交易金额统计失败")
		return
	}

	cache.SetJSON(c.Request.Context(), key, stats, cache.PlatformStatsTTL)
	response.Success(c, stats)
}
