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
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseInventoryMovement{})
	switch req.Role {
	case "warehouse":
		query = query.Where("warehouse_enterprise_id = ?", claims.EnterpriseID)
	default:
		query = query.Where("request_enterprise_id = ?", claims.EnterpriseID)
	}
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
	default:
		query = query.Where("(request_enterprise_id = ? OR warehouse_enterprise_id = ?)", claims.EnterpriseID, claims.EnterpriseID)
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
	targetStatus := entity.NormalizeLogisticsOrderStatus(entity.LogisticsOrderStatus(req.Status))
	if targetStatus == entity.LogisticsOrderReceiptSigned {
		if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND buyer_enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
			response.Error(c, ecode.ErrForbidden.WithMessage("物流委托不存在或无权更新"))
			return
		}
	} else {
		if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND carrier_enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
			response.Error(c, ecode.ErrForbidden.WithMessage("物流委托不存在或无权更新"))
			return
		}
	}
	currentStatus := entity.NormalizeLogisticsOrderStatus(item.Status)
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
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		if targetStatus == entity.LogisticsOrderReceiptSigned && item.OrderID > 0 {
			var order orderEntity.Order
			if err := tx.Where("id = ?", item.OrderID).First(&order).Error; err == nil {
				if order.Status != orderEntity.OrderStatusReceived && order.Status != orderEntity.OrderStatusReconciled && order.Status != orderEntity.OrderStatusArchived && order.Status != orderEntity.OrderStatusCompleted {
					fromStatus := order.Status
					order.Status = orderEntity.OrderStatusReceived
					order.ReceiptAt = &now
					order.UpdatedBy = claims.UserID
					if err := tx.Save(&order).Error; err != nil {
						return err
					}
					orderEvent := orderEntity.OrderEvent{
						OrderID:           order.ID,
						OrderNo:           order.OrderNo,
						EventType:         "buyer_received",
						ActorUserID:       claims.UserID,
						ActorEnterpriseID: claims.EnterpriseID,
						ActorService:      "procurement",
						FromStatus:        fromStatus,
						ToStatus:          orderEntity.OrderStatusReceived,
					}
					if err := tx.Create(&orderEvent).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
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

func warehouseSnapshotToInfo(item *entity.WarehouseInventorySnapshot) dto.WarehouseInventorySnapshotInfo {
	return dto.WarehouseInventorySnapshotInfo{
		ID:                    item.ID,
		SnapshotNo:            item.SnapshotNo,
		WarehouseServiceID:    item.WarehouseServiceID,
		WarehouseEnterpriseID: item.WarehouseEnterpriseID,
		ProductName:           item.ProductName,
		BundleNo:              item.BundleNo,
		BatchNo:               item.BatchNo,
		LocationCode:          item.LocationCode,
		Quantity:              item.Quantity,
		FrozenQuantity:        item.FrozenQuantity,
		QualityStatus:         item.QualityStatus,
		PledgeStatus:          item.PledgeStatus,
		ExternalSystem:        item.ExternalSystem,
		ExternalRefNo:         item.ExternalRefNo,
		SyncStatus:            item.SyncStatus,
		Remark:                item.Remark,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
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

func logisticsDriverToInfo(item *entity.LogisticsDriver) dto.LogisticsDriverInfo {
	return dto.LogisticsDriverInfo{
		ID:                    item.ID,
		EnterpriseID:          item.EnterpriseID,
		Name:                  item.Name,
		Phone:                 item.Phone,
		IDCard:                item.IDCard,
		LicenseNo:             item.LicenseNo,
		LicenseExpireAt:       item.LicenseExpireAt,
		QualificationNo:       item.QualificationNo,
		QualificationExpireAt: item.QualificationExpireAt,
		Status:                item.Status,
		Remark:                item.Remark,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}

func logisticsVehicleToInfo(item *entity.LogisticsVehicle) dto.LogisticsVehicleInfo {
	return dto.LogisticsVehicleInfo{
		ID:                item.ID,
		EnterpriseID:      item.EnterpriseID,
		PlateNumber:       item.PlateNumber,
		VehicleType:       item.VehicleType,
		LoadCapacity:      item.LoadCapacity,
		VehicleLicense:    item.VehicleLicense,
		LicenseExpireAt:   item.LicenseExpireAt,
		InsuranceNo:       item.InsuranceNo,
		InsuranceExpireAt: item.InsuranceExpireAt,
		DriverID:          item.DriverID,
		Status:            item.Status,
		Remark:            item.Remark,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
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

func (h *Handler) UploadFulfillmentDocument(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	logisticsOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的物流委托ID")
		return
	}
	var req dto.UploadFulfillmentDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var logisticsOrder entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).First(&logisticsOrder, logisticsOrderID).Error; err != nil {
		response.Error(c, ecode.ErrNotFound.WithMessage("物流委托单不存在"))
		return
	}

	role := ""
	if claims.EnterpriseID == logisticsOrder.CarrierEnterpriseID {
		role = "logistics"
	} else if claims.EnterpriseID == logisticsOrder.BuyerEnterpriseID {
		role = "buyer"
	} else {
		role = "warehouse"
	}

	doc := entity.FulfillmentDocument{
		LogisticsOrderID: logisticsOrderID,
		OrderID:          logisticsOrder.OrderID,
		DocType:          entity.FulfillmentDocType(req.DocType),
		FileURL:          req.FileURL,
		FileHash:         req.FileHash,
		Remark:           req.Remark,
		GPSLatitude:      req.GPSLatitude,
		GPSLongitude:     req.GPSLongitude,
		CapturedAt:       req.CapturedAt,
		UploadedBy:       claims.UserID,
		EnterpriseID:     claims.EnterpriseID,
		UploadedRole:     role,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&doc).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("上传单据失败"))
		return
	}
	response.Success(c, gin.H{"id": doc.ID, "doc_type": doc.DocType, "file_url": doc.FileURL})
}

func (h *Handler) ListFulfillmentDocuments(c *gin.Context) {
	logisticsOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的物流委托ID")
		return
	}
	var docs []entity.FulfillmentDocument
	if err := h.db.WithContext(c.Request.Context()).Where("logistics_order_id = ?", logisticsOrderID).Order("created_at ASC").Find(&docs).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询单据失败"))
		return
	}
	result := make([]dto.FulfillmentDocumentInfo, 0, len(docs))
	for _, d := range docs {
		result = append(result, dto.FulfillmentDocumentInfo{
			ID:               d.ID,
			LogisticsOrderID: d.LogisticsOrderID,
			OrderID:          d.OrderID,
			DocType:          string(d.DocType),
			FileURL:          d.FileURL,
			Remark:           d.Remark,
			GPSLatitude:      d.GPSLatitude,
			GPSLongitude:     d.GPSLongitude,
			CapturedAt:       d.CapturedAt,
			UploadedBy:       d.UploadedBy,
			EnterpriseID:     d.EnterpriseID,
			UploadedRole:     d.UploadedRole,
			CreatedAt:        d.CreatedAt.Format(time.RFC3339),
		})
	}
	response.Success(c, gin.H{"list": result})
}

func (h *Handler) SealConfirm(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	logisticsOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的物流委托ID")
		return
	}
	var req dto.SealConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var logisticsOrder entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).First(&logisticsOrder, logisticsOrderID).Error; err != nil {
		response.Error(c, ecode.ErrNotFound.WithMessage("物流委托单不存在"))
		return
	}

	doc := entity.FulfillmentDocument{
		LogisticsOrderID: logisticsOrderID,
		OrderID:          logisticsOrder.OrderID,
		DocType:          entity.DocTypeSealPhoto,
		FileURL:          req.FileURL,
		Remark:           req.Remark,
		UploadedBy:       claims.UserID,
		EnterpriseID:     claims.EnterpriseID,
		UploadedRole:     "warehouse",
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&doc).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("铅封确认失败"))
		return
	}

	h.db.WithContext(c.Request.Context()).Create(&entity.LogisticsTrackingEvent{
		LogisticsOrderID: logisticsOrderID,
		EventType:        "seal_confirmed",
		Description:      "仓库已确认铅封",
		OccurredAt:       time.Now(),
		CreatedBy:        claims.UserID,
	})

	response.Success(c, gin.H{"message": "铅封确认成功", "document_id": doc.ID})
}

func (h *Handler) SubmitInspection(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	logisticsOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的物流委托ID")
		return
	}
	var req dto.InspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var logisticsOrder entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).First(&logisticsOrder, logisticsOrderID).Error; err != nil {
		response.Error(c, ecode.ErrNotFound.WithMessage("物流委托单不存在"))
		return
	}

	role := "buyer"
	if claims.EnterpriseID != logisticsOrder.BuyerEnterpriseID {
		role = "warehouse"
	}

	if req.FileURL != "" {
		doc := entity.FulfillmentDocument{
			LogisticsOrderID: logisticsOrderID,
			OrderID:          logisticsOrder.OrderID,
			DocType:          entity.DocTypeInspectionResult,
			FileURL:          req.FileURL,
			Remark:           req.Remark,
			UploadedBy:       claims.UserID,
			EnterpriseID:     claims.EnterpriseID,
			UploadedRole:     role,
		}
		h.db.WithContext(c.Request.Context()).Create(&doc)
	}

	sealStatus := "异常"
	if req.SealOK {
		sealStatus = "完好"
	}
	h.db.WithContext(c.Request.Context()).Create(&entity.LogisticsTrackingEvent{
		LogisticsOrderID: logisticsOrderID,
		EventType:        "inspection_completed",
		Description:      fmt.Sprintf("验货结果：%s，铅封：%s", req.Result, sealStatus),
		OccurredAt:       time.Now(),
		CreatedBy:        claims.UserID,
	})

	response.Success(c, gin.H{"message": "验货结果已提交", "result": req.Result, "seal_ok": req.SealOK})
}

func (h *Handler) ListWarehouseInventorySnapshots(c *gin.Context) {
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
		req.PageSize = 20
	}

	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseInventorySnapshot{}).
		Where("warehouse_enterprise_id = ?", claims.EnterpriseID)
	if req.WarehouseServiceID > 0 {
		query = query.Where("warehouse_service_id = ?", req.WarehouseServiceID)
	}
	if req.Status != "" {
		query = query.Where("pledge_status = ? OR quality_status = ? OR sync_status = ?", req.Status, req.Status, req.Status)
	}

	var total int64
	query.Count(&total)
	var rows []entity.WarehouseInventorySnapshot
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询库存快照失败"))
		return
	}
	list := make([]dto.WarehouseInventorySnapshotInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, warehouseSnapshotToInfo(&row))
	}
	response.Success(c, dto.PageResponse[dto.WarehouseInventorySnapshotInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

func (h *Handler) CreateWarehouseInventorySnapshot(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.UpsertWarehouseInventorySnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var warehouse serviceOfferEntity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", req.WarehouseServiceID, claims.EnterpriseID).First(&warehouse).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("仓储资源不存在或无权维护库存"))
		return
	}

	item := entity.WarehouseInventorySnapshot{
		SnapshotNo:            "INV" + time.Now().Format("20060102150405") + randomDigits(4),
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		ProductName:           req.ProductName,
		BundleNo:              req.BundleNo,
		BatchNo:               req.BatchNo,
		LocationCode:          req.LocationCode,
		Quantity:              req.Quantity,
		FrozenQuantity:        req.FrozenQuantity,
		QualityStatus:         defaultString(req.QualityStatus, "unchecked"),
		PledgeStatus:          defaultString(req.PledgeStatus, "none"),
		ExternalSystem:        req.ExternalSystem,
		ExternalRefNo:         req.ExternalRefNo,
		SyncStatus:            defaultString(req.SyncStatus, "manual"),
		Remark:                req.Remark,
		CreatedBy:             claims.UserID,
		UpdatedBy:             claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建库存快照失败"))
		return
	}
	response.Success(c, warehouseSnapshotToInfo(&item))
}

func (h *Handler) UpdateWarehouseInventorySnapshot(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的库存快照ID")
		return
	}
	var req dto.UpsertWarehouseInventorySnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var item entity.WarehouseInventorySnapshot
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND warehouse_enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("库存快照不存在或无权修改"))
		return
	}
	if req.WarehouseServiceID != item.WarehouseServiceID {
		var warehouse serviceOfferEntity.WarehouseService
		if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", req.WarehouseServiceID, claims.EnterpriseID).First(&warehouse).Error; err != nil {
			response.Error(c, ecode.ErrForbidden.WithMessage("仓储资源不存在或无权维护库存"))
			return
		}
		item.WarehouseServiceID = warehouse.ID
		item.WarehouseEnterpriseID = warehouse.EnterpriseID
	}
	item.ProductName = req.ProductName
	item.BundleNo = req.BundleNo
	item.BatchNo = req.BatchNo
	item.LocationCode = req.LocationCode
	item.Quantity = req.Quantity
	item.FrozenQuantity = req.FrozenQuantity
	item.QualityStatus = defaultString(req.QualityStatus, "unchecked")
	item.PledgeStatus = defaultString(req.PledgeStatus, "none")
	item.ExternalSystem = req.ExternalSystem
	item.ExternalRefNo = req.ExternalRefNo
	item.SyncStatus = defaultString(req.SyncStatus, "manual")
	item.Remark = req.Remark
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新库存快照失败"))
		return
	}
	response.Success(c, warehouseSnapshotToInfo(&item))
}

func (h *Handler) ListLogisticsDrivers(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var rows []entity.LogisticsDriver
	if err := h.db.WithContext(c.Request.Context()).Where("enterprise_id = ?", claims.EnterpriseID).Order("id DESC").Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询司机档案失败"))
		return
	}
	list := make([]dto.LogisticsDriverInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, logisticsDriverToInfo(&row))
	}
	response.Success(c, gin.H{"list": list})
}

func (h *Handler) CreateLogisticsDriver(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.UpsertLogisticsDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.LogisticsDriver{
		EnterpriseID:          claims.EnterpriseID,
		Name:                  req.Name,
		Phone:                 req.Phone,
		IDCard:                req.IDCard,
		LicenseNo:             req.LicenseNo,
		LicenseExpireAt:       req.LicenseExpireAt,
		QualificationNo:       req.QualificationNo,
		QualificationExpireAt: req.QualificationExpireAt,
		Status:                defaultString(req.Status, "active"),
		Remark:                req.Remark,
		CreatedBy:             claims.UserID,
		UpdatedBy:             claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建司机档案失败"))
		return
	}
	response.Success(c, logisticsDriverToInfo(&item))
}

func (h *Handler) UpdateLogisticsDriver(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的司机ID")
		return
	}
	var req dto.UpsertLogisticsDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.LogisticsDriver
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("司机档案不存在或无权修改"))
		return
	}
	item.Name = req.Name
	item.Phone = req.Phone
	item.IDCard = req.IDCard
	item.LicenseNo = req.LicenseNo
	item.LicenseExpireAt = req.LicenseExpireAt
	item.QualificationNo = req.QualificationNo
	item.QualificationExpireAt = req.QualificationExpireAt
	item.Status = defaultString(req.Status, "active")
	item.Remark = req.Remark
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新司机档案失败"))
		return
	}
	response.Success(c, logisticsDriverToInfo(&item))
}

func (h *Handler) ListLogisticsVehicles(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var rows []entity.LogisticsVehicle
	if err := h.db.WithContext(c.Request.Context()).Where("enterprise_id = ?", claims.EnterpriseID).Order("id DESC").Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询车辆档案失败"))
		return
	}
	list := make([]dto.LogisticsVehicleInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, logisticsVehicleToInfo(&row))
	}
	response.Success(c, gin.H{"list": list})
}

func (h *Handler) CreateLogisticsVehicle(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req dto.UpsertLogisticsVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.LogisticsVehicle{
		EnterpriseID:      claims.EnterpriseID,
		PlateNumber:       req.PlateNumber,
		VehicleType:       req.VehicleType,
		LoadCapacity:      req.LoadCapacity,
		VehicleLicense:    req.VehicleLicense,
		LicenseExpireAt:   req.LicenseExpireAt,
		InsuranceNo:       req.InsuranceNo,
		InsuranceExpireAt: req.InsuranceExpireAt,
		DriverID:          req.DriverID,
		Status:            defaultString(req.Status, "active"),
		Remark:            req.Remark,
		CreatedBy:         claims.UserID,
		UpdatedBy:         claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建车辆档案失败"))
		return
	}
	response.Success(c, logisticsVehicleToInfo(&item))
}

func (h *Handler) UpdateLogisticsVehicle(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的车辆ID")
		return
	}
	var req dto.UpsertLogisticsVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.LogisticsVehicle
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("车辆档案不存在或无权修改"))
		return
	}
	item.PlateNumber = req.PlateNumber
	item.VehicleType = req.VehicleType
	item.LoadCapacity = req.LoadCapacity
	item.VehicleLicense = req.VehicleLicense
	item.LicenseExpireAt = req.LicenseExpireAt
	item.InsuranceNo = req.InsuranceNo
	item.InsuranceExpireAt = req.InsuranceExpireAt
	item.DriverID = req.DriverID
	item.Status = defaultString(req.Status, "active")
	item.Remark = req.Remark
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新车辆档案失败"))
		return
	}
	response.Success(c, logisticsVehicleToInfo(&item))
}

func (h *Handler) ListCollateralMonitor(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var financings []orderEntity.OrderFinancingRequest
	if err := h.db.WithContext(c.Request.Context()).
		Where("funding_enterprise_id = ?", claims.EnterpriseID).
		Order("id DESC").
		Find(&financings).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询货权监控失败"))
		return
	}

	list := make([]dto.CollateralMonitorInfo, 0, len(financings))
	for _, financing := range financings {
		item := dto.CollateralMonitorInfo{
			FinancingRequestID: financing.ID,
			OrderID:            financing.OrderID,
			OrderNo:            financing.OrderNo,
			BuyerEnterpriseID:  financing.BuyerEnterpriseID,
			SellerEnterpriseID: financing.SellerEnterpriseID,
			FinanceServiceID:   financing.FinanceServiceID,
			RequestedAmount:    financing.RequestedAmount,
			ApprovedAmount:     financing.ApprovedAmount,
			FinancingStatus:    string(financing.Status),
			PledgeVoucherURL:   financing.PledgeVoucherURL,
			PledgeRemark:       financing.PledgeRemark,
			PledgedAt:          financing.PledgedAt,
			ReleaseCondition:   financing.ReleaseCondition,
			ReleasedAt:         financing.ReleasedAt,
		}

		var order orderEntity.Order
		if err := h.db.WithContext(c.Request.Context()).First(&order, financing.OrderID).Error; err == nil {
			item.OrderStatus = int(order.Status)
			if order.DeliveryInfo != nil {
				item.WarehouseServiceID = order.DeliveryInfo.DestWarehouseServiceID
				item.WarehouseName = order.DeliveryInfo.DestWarehouseName
				item.WarehouseAddress = order.DeliveryInfo.DestWarehouseAddress
				if item.WarehouseServiceID == 0 {
					item.WarehouseServiceID = order.DeliveryInfo.WarehouseServiceID
					item.WarehouseName = order.DeliveryInfo.WarehouseName
					item.WarehouseAddress = order.DeliveryInfo.WarehouseAddress
				}
			}
		}

		var logisticsOrder entity.LogisticsOrder
		if err := h.db.WithContext(c.Request.Context()).Where("order_id = ?", financing.OrderID).Order("id DESC").First(&logisticsOrder).Error; err == nil {
			item.LogisticsStatus = string(entity.NormalizeLogisticsOrderStatus(logisticsOrder.Status))
			var event entity.LogisticsTrackingEvent
			if err := h.db.WithContext(c.Request.Context()).Where("logistics_order_id = ?", logisticsOrder.ID).Order("occurred_at DESC, id DESC").First(&event).Error; err == nil {
				item.LastLogisticsEvent = event.Description
			}
			var docs []entity.FulfillmentDocument
			h.db.WithContext(c.Request.Context()).Where("logistics_order_id = ?", logisticsOrder.ID).Order("id DESC").Limit(5).Find(&docs)
			item.FulfillmentDocs = make([]dto.FulfillmentDocumentInfo, 0, len(docs))
			for _, doc := range docs {
				item.FulfillmentDocs = append(item.FulfillmentDocs, fulfillmentDocumentToInfo(&doc))
			}
		}

		if item.WarehouseServiceID > 0 {
			var snapshots []entity.WarehouseInventorySnapshot
			h.db.WithContext(c.Request.Context()).
				Where("warehouse_service_id = ?", item.WarehouseServiceID).
				Order("id DESC").
				Limit(5).
				Find(&snapshots)
			item.InventorySnapshots = make([]dto.WarehouseInventorySnapshotInfo, 0, len(snapshots))
			for _, snapshot := range snapshots {
				item.InventorySnapshots = append(item.InventorySnapshots, warehouseSnapshotToInfo(&snapshot))
			}
		}

		list = append(list, item)
	}
	response.Success(c, gin.H{"list": list})
}

func fulfillmentDocumentToInfo(d *entity.FulfillmentDocument) dto.FulfillmentDocumentInfo {
	return dto.FulfillmentDocumentInfo{
		ID:               d.ID,
		LogisticsOrderID: d.LogisticsOrderID,
		OrderID:          d.OrderID,
		DocType:          string(d.DocType),
		FileURL:          d.FileURL,
		Remark:           d.Remark,
		GPSLatitude:      d.GPSLatitude,
		GPSLongitude:     d.GPSLongitude,
		CapturedAt:       d.CapturedAt,
		UploadedBy:       d.UploadedBy,
		EnterpriseID:     d.EnterpriseID,
		UploadedRole:     d.UploadedRole,
		CreatedAt:        d.CreatedAt.Format(time.RFC3339),
	}
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
