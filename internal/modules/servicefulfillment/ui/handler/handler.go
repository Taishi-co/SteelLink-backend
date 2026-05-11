package handler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/servicefulfillment/ui/dto"
	serviceOfferEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler() *Handler {
	return &Handler{db: database.GetDB()}
}

func (h *Handler) BindProductWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}
	var req dto.BindProductWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var product productEntity.Product
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", productID, claims.EnterpriseID).First(&product).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("商品不存在或无权绑定仓库"))
		return
	}
	var warehouse serviceOfferEntity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).First(&warehouse, req.WarehouseServiceID).Error; err != nil {
		response.NotFound(c, "仓储资源不存在")
		return
	}

	product.WarehouseInfo = &productEntity.WarehouseInfo{
		WarehouseID:   warehouse.ID,
		WarehouseName: warehouse.WarehouseName,
		WarehouseCode: fmt.Sprintf("WH-%06d", warehouse.ID),
		Address:       warehouse.Address,
		Province:      warehouse.Province,
		City:          warehouse.City,
		District:      warehouse.District,
	}
	product.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&product).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("绑定商品仓库失败"))
		return
	}
	response.Success(c, product.WarehouseInfo)
}

func (h *Handler) ListWarehouseProducts(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的仓库ID")
		return
	}
	var products []productEntity.Product
	if err := h.db.WithContext(c.Request.Context()).
		Where("JSON_UNQUOTE(JSON_EXTRACT(warehouse_info, '$.warehouse_id')) = ?", strconv.FormatUint(warehouseID, 10)).
		Order("id DESC").
		Find(&products).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询仓库商品失败"))
		return
	}
	response.Success(c, gin.H{"list": products})
}

func (h *Handler) CreateWarehouseMovement(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.CreateWarehouseMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var warehouse serviceOfferEntity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", req.WarehouseServiceID, claims.EnterpriseID).First(&warehouse).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("仓储资源不存在或无权创建作业记录"))
		return
	}
	status := entity.WarehouseMovementStatus(strings.TrimSpace(req.Status))
	if status == "" {
		status = entity.WarehouseMovementPending
	}
	item := entity.WarehouseInventoryMovement{
		MovementNo:            "WMS" + time.Now().Format("20060102150405") + randomDigits(4),
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		RequestEnterpriseID:   claims.EnterpriseID,
		ProductID:             req.ProductID,
		ProductName:           req.ProductName,
		BundleNo:              req.BundleNo,
		OrderID:               req.OrderID,
		OrderNo:               req.OrderNo,
		MovementType:          entity.WarehouseMovementType(req.MovementType),
		Quantity:              req.Quantity,
		Status:                status,
		ScheduledAt:           req.ScheduledAt,
		Remark:                req.Remark,
		CreatedBy:             claims.UserID,
		UpdatedBy:             claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建出入库记录失败"))
		return
	}
	response.Success(c, movementToInfo(&item))
}

func (h *Handler) CreateWarehouseStorageRequest(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.CreateWarehouseStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var warehouse serviceOfferEntity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND status = ?", req.WarehouseServiceID, serviceOfferEntity.OfferStatusApproved).First(&warehouse).Error; err != nil {
		response.NotFound(c, "仓储资源不存在或未审核通过")
		return
	}
	remarkParts := []string{}
	if req.ContactName != "" || req.ContactPhone != "" {
		remarkParts = append(remarkParts, "联系人:"+strings.TrimSpace(req.ContactName+" "+req.ContactPhone))
	}
	if strings.TrimSpace(req.StorageRequirement) != "" {
		remarkParts = append(remarkParts, "仓储要求:"+strings.TrimSpace(req.StorageRequirement))
	}
	if strings.TrimSpace(req.Remark) != "" {
		remarkParts = append(remarkParts, strings.TrimSpace(req.Remark))
	}
	item := entity.WarehouseInventoryMovement{
		MovementNo:            "WMS" + time.Now().Format("20060102150405") + randomDigits(4),
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		RequestEnterpriseID:   claims.EnterpriseID,
		ProductID:             req.ProductID,
		ProductName:           req.ProductName,
		BundleNo:              req.BundleNo,
		OrderID:               req.OrderID,
		OrderNo:               req.OrderNo,
		MovementType:          entity.WarehouseMovementInbound,
		Quantity:              req.Quantity,
		Status:                entity.WarehouseMovementPending,
		ScheduledAt:           req.ScheduledAt,
		Remark:                strings.Join(remarkParts, "；"),
		CreatedBy:             claims.UserID,
		UpdatedBy:             claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建仓储托管需求失败"))
		return
	}
	response.Success(c, movementToInfo(&item))
}

func (h *Handler) ListWarehouseStorageRequests(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseInventoryMovement{}).Where("request_enterprise_id = ?", claims.EnterpriseID)
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	var total int64
	query.Count(&total)
	var rows []entity.WarehouseInventoryMovement
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.WarehouseMovementInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, movementToInfo(&row))
	}
	response.Success(c, dto.PageResponse[dto.WarehouseMovementInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

func (h *Handler) ListWarehouseMovements(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseInventoryMovement{})
	if req.OrderID > 0 {
		query = query.Where("order_id = ?", req.OrderID)
	}
	if req.WarehouseServiceID > 0 {
		query = query.Where("warehouse_service_id = ?", req.WarehouseServiceID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	switch req.Role {
	case "requester":
		query = query.Where("request_enterprise_id = ?", claims.EnterpriseID)
	case "warehouse":
		query = query.Where("warehouse_enterprise_id = ?", claims.EnterpriseID)
	}
	var total int64
	query.Count(&total)
	var rows []entity.WarehouseInventoryMovement
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.WarehouseMovementInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, movementToInfo(&row))
	}
	response.Success(c, dto.PageResponse[dto.WarehouseMovementInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

func (h *Handler) UpdateWarehouseMovementStatus(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.WarehouseInventoryMovement
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND warehouse_enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("出入库记录不存在或无权更新"))
		return
	}
	item.Status = entity.WarehouseMovementStatus(req.Status)
	if req.Remark != "" {
		item.Remark = req.Remark
	}
	if item.Status == entity.WarehouseMovementCompleted {
		now := time.Now()
		item.CompletedAt = &now
	}
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新出入库状态失败"))
		return
	}
	response.Success(c, movementToInfo(&item))
}

func (h *Handler) EstimateLogistics(c *gin.Context) {
	var req dto.EstimateLogisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var service serviceOfferEntity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).First(&service, req.LogisticsServiceID).Error; err != nil {
		response.NotFound(c, "物流服务不存在")
		return
	}
	unitPrice := parseUnitPrice(service.PriceRule)
	origin := req.Origin
	if origin == "" && req.WarehouseServiceID > 0 {
		var warehouse serviceOfferEntity.WarehouseService
		if err := h.db.WithContext(c.Request.Context()).First(&warehouse, req.WarehouseServiceID).Error; err == nil {
			origin = strings.TrimSpace(warehouse.Province + warehouse.City + warehouse.District + warehouse.Address)
		}
	}
	response.Success(c, dto.EstimateLogisticsResponse{
		LogisticsServiceID: req.LogisticsServiceID,
		WarehouseServiceID: req.WarehouseServiceID,
		Origin:             origin,
		Destination:        req.Destination,
		Weight:             req.Weight,
		UnitPrice:          unitPrice,
		EstimatedPrice:     unitPrice * req.Weight,
		PriceRule:          service.PriceRule,
	})
}

func (h *Handler) CreateLogisticsOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.CreateLogisticsOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var service serviceOfferEntity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).First(&service, req.LogisticsServiceID).Error; err != nil {
		response.NotFound(c, "物流服务不存在")
		return
	}
	unitPrice := parseUnitPrice(service.PriceRule)
	item := entity.LogisticsOrder{
		LogisticsNo:             "LG" + time.Now().Format("20060102150405") + randomDigits(4),
		OrderID:                 req.OrderID,
		OrderNo:                 req.OrderNo,
		BuyerEnterpriseID:       claims.EnterpriseID,
		LogisticsServiceID:      req.LogisticsServiceID,
		WarehouseServiceID:      req.WarehouseServiceID,
		Origin:                  req.Origin,
		Destination:             req.Destination,
		CargoName:               req.CargoName,
		Weight:                  req.Weight,
		VehicleType:             req.VehicleType,
		EstimatedPrice:          unitPrice * req.Weight,
		Status:                  entity.LogisticsOrderTaskCreated,
		CurrentResponsibleParty: entity.ResponsibilityPartyLogistics,
		GuaranteeType:           req.GuaranteeType,
		CreatedBy:               claims.UserID,
		UpdatedBy:               claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建物流委托失败"))
		return
	}
	response.Success(c, logisticsOrderToInfo(&item, nil))
}

func (h *Handler) ListLogisticsOrders(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.LogisticsOrder{})
	switch req.Role {
	case "carrier":
		query = query.Where("carrier_enterprise_id = ? OR logistics_service_id IN (?)", claims.EnterpriseID, h.db.Model(&serviceOfferEntity.LogisticsService{}).Select("id").Where("enterprise_id = ?", claims.EnterpriseID))
	case "buyer":
		query = query.Where("buyer_enterprise_id = ?", claims.EnterpriseID)
	default:
		query = query.Where("buyer_enterprise_id = ? OR carrier_enterprise_id = ? OR logistics_service_id IN (?)", claims.EnterpriseID, claims.EnterpriseID, h.db.Model(&serviceOfferEntity.LogisticsService{}).Select("id").Where("enterprise_id = ?", claims.EnterpriseID))
	}
	if req.OrderID > 0 {
		query = query.Where("order_id = ?", req.OrderID)
	}
	if req.LogisticsServiceID > 0 {
		query = query.Where("logistics_service_id = ?", req.LogisticsServiceID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	var total int64
	query.Count(&total)
	var rows []entity.LogisticsOrder
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.LogisticsOrderInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, logisticsOrderToInfo(&row, nil))
	}
	response.Success(c, dto.PageResponse[dto.LogisticsOrderInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

func (h *Handler) AcceptLogisticsOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.AcceptLogisticsOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "物流委托不存在")
		return
	}
	var service serviceOfferEntity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", item.LogisticsServiceID, claims.EnterpriseID).First(&service).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("无权承接该物流委托"))
		return
	}
	now := time.Now()
	item.CarrierEnterpriseID = claims.EnterpriseID
	item.DriverName = req.DriverName
	item.DriverPhone = req.DriverPhone
	item.PlateNumber = req.PlateNumber
	item.FinalPrice = req.FinalPrice
	if item.FinalPrice == 0 {
		item.FinalPrice = item.EstimatedPrice
	}
	item.Status = entity.LogisticsOrderPickupScheduled
	item.CurrentResponsibleParty = entity.ResponsibilityPartyWarehouse
	item.AcceptedBy = claims.UserID
	item.AcceptedAt = &now
	item.PickupAppointmentAt = &now
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		event := entity.LogisticsTrackingEvent{
			LogisticsOrderID: item.ID,
			EventType:        string(item.Status),
			Description:      buildLogisticsStatusDescription(item.Status, ""),
			OccurredAt:       now,
			CreatedBy:        claims.UserID,
		}
		return tx.Create(&event).Error
	}); err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("物流接单失败"))
		return
	}
	response.Success(c, logisticsOrderToInfo(&item, nil))
}

func (h *Handler) UpdateLogisticsOrderStatus(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND carrier_enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("物流委托不存在或无权更新"))
		return
	}
	currentStatus := entity.NormalizeLogisticsOrderStatus(item.Status)
	targetStatus := entity.NormalizeLogisticsOrderStatus(entity.LogisticsOrderStatus(req.Status))
	if !entity.CanTransitionLogisticsOrderStatus(currentStatus, targetStatus) {
		response.BadRequest(c, "当前物流状态不可推进到目标节点")
		return
	}
	now := time.Now()
	item.Status = targetStatus
	item.CurrentResponsibleParty = entity.ResponsibilityForLogisticsStatus(targetStatus)
	if strings.TrimSpace(req.Remark) != "" {
		item.HandoverNote = strings.TrimSpace(req.Remark)
	}
	applyLogisticsStatusTimestamp(&item, targetStatus, now)
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		event := entity.LogisticsTrackingEvent{
			LogisticsOrderID: item.ID,
			EventType:        string(targetStatus),
			Description:      buildLogisticsStatusDescription(targetStatus, req.Remark),
			OccurredAt:       now,
			CreatedBy:        claims.UserID,
		}
		return tx.Create(&event).Error
	}); err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新物流状态失败"))
		return
	}
	response.Success(c, logisticsOrderToInfo(&item, nil))
}

func (h *Handler) AddTrackingEvent(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.CreateTrackingEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var order entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND carrier_enterprise_id = ?", id, claims.EnterpriseID).First(&order).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("物流委托不存在或无权追加轨迹"))
		return
	}
	occurredAt := time.Now()
	if req.OccurredAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.OccurredAt); err == nil {
			occurredAt = parsed
		}
	}
	event := entity.LogisticsTrackingEvent{
		LogisticsOrderID: order.ID,
		EventType:        req.EventType,
		Location:         req.Location,
		Description:      req.Description,
		OccurredAt:       occurredAt,
		CreatedBy:        claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&event).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("追加物流轨迹失败"))
		return
	}
	response.Success(c, trackingEventToInfo(&event))
}

func (h *Handler) CreateGuarantee(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.CreateGuaranteeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.FulfillmentGuarantee{
		GuaranteeNo:      "FG" + time.Now().Format("20060102150405") + randomDigits(4),
		OrderID:          req.OrderID,
		LogisticsOrderID: req.LogisticsOrderID,
		EnterpriseID:     claims.EnterpriseID,
		GuaranteeType:    req.GuaranteeType,
		Status:           entity.GuaranteeStatusActive,
		CoverageAmount:   req.CoverageAmount,
		Description:      req.Description,
		CreatedBy:        claims.UserID,
		UpdatedBy:        claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建履约保障失败"))
		return
	}
	response.Success(c, guaranteeToInfo(&item))
}

func (h *Handler) UpdateGuaranteeStatus(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateGuaranteeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.FulfillmentGuarantee
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("保障记录不存在或无权更新"))
		return
	}
	item.Status = entity.GuaranteeStatus(req.Status)
	item.Resolution = req.Resolution
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新履约保障失败"))
		return
	}
	response.Success(c, guaranteeToInfo(&item))
}

func (h *Handler) GetOrderFulfillment(c *gin.Context) {
	claims := middleware.GetClaims(c)
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var order orderEntity.Order
	if err := h.db.WithContext(c.Request.Context()).First(&order, orderID).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.BuyerID != claims.EnterpriseID && order.SellerID != claims.EnterpriseID {
		var count int64
		h.db.WithContext(c.Request.Context()).Model(&entity.LogisticsOrder{}).Where("order_id = ? AND carrier_enterprise_id = ?", orderID, claims.EnterpriseID).Count(&count)
		if count == 0 {
			response.Error(c, ecode.ErrForbidden.WithMessage("无权查看订单履约信息"))
			return
		}
	}
	var movements []entity.WarehouseInventoryMovement
	h.db.WithContext(c.Request.Context()).Where("order_id = ?", orderID).Order("id DESC").Find(&movements)
	var logistics []entity.LogisticsOrder
	h.db.WithContext(c.Request.Context()).Where("order_id = ?", orderID).Order("id DESC").Find(&logistics)
	var guarantees []entity.FulfillmentGuarantee
	h.db.WithContext(c.Request.Context()).Where("order_id = ?", orderID).Order("id DESC").Find(&guarantees)

	moveInfos := make([]dto.WarehouseMovementInfo, 0, len(movements))
	for _, item := range movements {
		moveInfos = append(moveInfos, movementToInfo(&item))
	}
	logisticsInfos := make([]dto.LogisticsOrderInfo, 0, len(logistics))
	for _, item := range logistics {
		var events []entity.LogisticsTrackingEvent
		h.db.WithContext(c.Request.Context()).Where("logistics_order_id = ?", item.ID).Order("occurred_at ASC, id ASC").Find(&events)
		logisticsInfos = append(logisticsInfos, logisticsOrderToInfo(&item, events))
	}
	guaranteeInfos := make([]dto.GuaranteeInfo, 0, len(guarantees))
	for _, item := range guarantees {
		guaranteeInfos = append(guaranteeInfos, guaranteeToInfo(&item))
	}
	response.Success(c, dto.FulfillmentSummary{
		OrderID:        orderID,
		WarehouseMoves: moveInfos,
		Logistics:      logisticsInfos,
		Guarantees:     guaranteeInfos,
	})
}

func movementToInfo(item *entity.WarehouseInventoryMovement) dto.WarehouseMovementInfo {
	return dto.WarehouseMovementInfo{
		ID: item.ID, MovementNo: item.MovementNo, WarehouseServiceID: item.WarehouseServiceID,
		WarehouseEnterpriseID: item.WarehouseEnterpriseID, RequestEnterpriseID: item.RequestEnterpriseID, ProductID: item.ProductID, ProductName: item.ProductName,
		BundleNo: item.BundleNo, OrderID: item.OrderID, OrderNo: item.OrderNo, MovementType: string(item.MovementType),
		Quantity: item.Quantity, Status: string(item.Status), ScheduledAt: item.ScheduledAt, CompletedAt: item.CompletedAt,
		Remark: item.Remark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func logisticsOrderToInfo(item *entity.LogisticsOrder, events []entity.LogisticsTrackingEvent) dto.LogisticsOrderInfo {
	eventInfos := make([]dto.TrackingEventInfo, 0, len(events))
	for _, event := range events {
		eventInfos = append(eventInfos, trackingEventToInfo(&event))
	}
	return dto.LogisticsOrderInfo{
		ID: item.ID, LogisticsNo: item.LogisticsNo, OrderID: item.OrderID, OrderNo: item.OrderNo,
		BuyerEnterpriseID: item.BuyerEnterpriseID, LogisticsServiceID: item.LogisticsServiceID,
		CarrierEnterpriseID: item.CarrierEnterpriseID, WarehouseServiceID: item.WarehouseServiceID,
		Origin: item.Origin, Destination: item.Destination, CargoName: item.CargoName, Weight: item.Weight,
		VehicleType: item.VehicleType, EstimatedPrice: item.EstimatedPrice, FinalPrice: item.FinalPrice,
		Status: string(entity.NormalizeLogisticsOrderStatus(item.Status)), CurrentResponsibleParty: string(entity.ResponsibilityForLogisticsStatus(item.Status)),
		HandoverNote: item.HandoverNote, DriverName: item.DriverName, DriverPhone: item.DriverPhone,
		PlateNumber: item.PlateNumber, GuaranteeType: item.GuaranteeType, AcceptedAt: item.AcceptedAt,
		PickupAppointmentAt: item.PickupAppointmentAt, ArrivalSignedInAt: item.ArrivalSignedInAt,
		OutboundConfirmedAt: item.OutboundConfirmedAt, LoadingConfirmedAt: item.LoadingConfirmedAt,
		ArrivalConfirmedAt: item.ArrivalConfirmedAt, ReceiptSignedAt: item.ReceiptSignedAt,
		InboundCompletedAt: item.InboundCompletedAt,
		CreatedAt:          item.CreatedAt, UpdatedAt: item.UpdatedAt, TrackingEvents: eventInfos,
	}
}

func trackingEventToInfo(item *entity.LogisticsTrackingEvent) dto.TrackingEventInfo {
	return dto.TrackingEventInfo{
		ID: item.ID, LogisticsOrderID: item.LogisticsOrderID, EventType: item.EventType,
		Location: item.Location, Description: item.Description, OccurredAt: item.OccurredAt, CreatedAt: item.CreatedAt,
	}
}

func guaranteeToInfo(item *entity.FulfillmentGuarantee) dto.GuaranteeInfo {
	return dto.GuaranteeInfo{
		ID: item.ID, GuaranteeNo: item.GuaranteeNo, OrderID: item.OrderID, LogisticsOrderID: item.LogisticsOrderID,
		EnterpriseID: item.EnterpriseID, GuaranteeType: item.GuaranteeType, Status: string(item.Status),
		CoverageAmount: item.CoverageAmount, Description: item.Description, Resolution: item.Resolution,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func applyLogisticsStatusTimestamp(item *entity.LogisticsOrder, status entity.LogisticsOrderStatus, now time.Time) {
	switch entity.NormalizeLogisticsOrderStatus(status) {
	case entity.LogisticsOrderPickupScheduled:
		if item.PickupAppointmentAt == nil {
			item.PickupAppointmentAt = &now
		}
	case entity.LogisticsOrderArrivalSignedIn:
		if item.ArrivalSignedInAt == nil {
			item.ArrivalSignedInAt = &now
		}
	case entity.LogisticsOrderOutboundConfirmed:
		if item.OutboundConfirmedAt == nil {
			item.OutboundConfirmedAt = &now
		}
	case entity.LogisticsOrderLoadingConfirmed:
		if item.LoadingConfirmedAt == nil {
			item.LoadingConfirmedAt = &now
		}
	case entity.LogisticsOrderArrivalConfirmed:
		if item.ArrivalConfirmedAt == nil {
			item.ArrivalConfirmedAt = &now
		}
	case entity.LogisticsOrderReceiptSigned:
		if item.ReceiptSignedAt == nil {
			item.ReceiptSignedAt = &now
		}
	case entity.LogisticsOrderInboundCompleted:
		if item.InboundCompletedAt == nil {
			item.InboundCompletedAt = &now
		}
	}
}

func buildLogisticsStatusDescription(status entity.LogisticsOrderStatus, remark string) string {
	label := map[entity.LogisticsOrderStatus]string{
		entity.LogisticsOrderTaskCreated:       "已创建运输任务",
		entity.LogisticsOrderPickupScheduled:   "已完成提货预约",
		entity.LogisticsOrderArrivalSignedIn:   "承运车辆已到场签到",
		entity.LogisticsOrderOutboundConfirmed: "仓储方已确认出库",
		entity.LogisticsOrderLoadingConfirmed:  "物流方已确认装货",
		entity.LogisticsOrderInTransit:         "货物运输中",
		entity.LogisticsOrderArrivalConfirmed:  "货物已到货确认",
		entity.LogisticsOrderReceiptSigned:     "买方已完成验收签收",
		entity.LogisticsOrderInboundCompleted:  "目的地仓储已完成入库",
		entity.LogisticsOrderCompleted:         "物流履约已完成",
		entity.LogisticsOrderCancelled:         "物流履约已取消",
		entity.LogisticsOrderException:         "物流履约出现异常",
	}
	description := label[entity.NormalizeLogisticsOrderStatus(status)]
	if strings.TrimSpace(remark) == "" {
		return description
	}
	return description + "：" + strings.TrimSpace(remark)
}

func parseUnitPrice(priceRule string) float64 {
	re := regexp.MustCompile(`\d+(\.\d+)?`)
	value := re.FindString(priceRule)
	if value == "" {
		return 80
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return 80
	}
	return parsed
}

func randomDigits(length int) string {
	const digits = "0123456789"
	now := time.Now().UnixNano()
	result := make([]byte, length)
	for i := range result {
		result[i] = digits[(now+int64(i*7))%int64(len(digits))]
	}
	return string(result)
}
