package handler

import (
	"fmt"
	"net/url"
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

func (h *Handler) CreateWarehouseReceiptPledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能申请仓单质押"))
		return
	}
	var req dto.CreateWarehouseReceiptPledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if !isHTTPURL(req.ReceiptVoucherURL) {
		response.BadRequest(c, "仓单文件URL必须是有效的HTTP/HTTPS地址")
		return
	}

	var finance serviceOfferEntity.FinanceService
	if err := h.db.WithContext(c.Request.Context()).
		Where("id = ? AND status = ?", req.FinanceServiceID, serviceOfferEntity.OfferStatusApproved).
		First(&finance).Error; err != nil {
		response.NotFound(c, "资金产品不存在或未审核通过")
		return
	}
	if req.RequestedAmount < finance.MinAmount || (finance.MaxAmount > 0 && req.RequestedAmount > finance.MaxAmount) {
		response.BadRequest(c, "申请金额不在资金产品额度范围内")
		return
	}

	var warehouse serviceOfferEntity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).
		Where("id = ? AND status = ?", req.WarehouseServiceID, serviceOfferEntity.OfferStatusApproved).
		First(&warehouse).Error; err != nil {
		response.NotFound(c, "监管仓库不存在或未审核通过")
		return
	}
	if !warehouse.PledgeSupervision {
		response.BadRequest(c, "所选仓库未声明支持质押监管")
		return
	}
	if err := h.ensureFinanceWarehouseAuthorized(c, finance.ID, finance.EnterpriseID, warehouse.ID); err != nil {
		response.Error(c, err)
		return
	}

	now := time.Now()
	item := entity.WarehouseReceiptPledgeApplication{
		ApplicationNo:         "WRP" + now.Format("20060102150405") + randomDigits(4),
		ApplicantEnterpriseID: claims.EnterpriseID,
		ApplicantUserID:       claims.UserID,
		ApplicantRole:         resolveApplicantRole(c.GetHeader("X-Trade-Role"), claims.ServiceCapabilities),
		FinanceServiceID:      finance.ID,
		FundingEnterpriseID:   finance.EnterpriseID,
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		WarehouseName:         warehouse.WarehouseName,
		ReceiptNo:             strings.TrimSpace(req.ReceiptNo),
		ReceiptVoucherURL:     strings.TrimSpace(req.ReceiptVoucherURL),
		GoodsName:             strings.TrimSpace(req.GoodsName),
		Quantity:              req.Quantity,
		QualityGrade:          strings.TrimSpace(req.QualityGrade),
		RequestedAmount:       req.RequestedAmount,
		Status:                entity.WarehouseReceiptPledgeSubmitted,
		SubmittedAt:           &now,
		Remark:                strings.TrimSpace(req.Remark),
		CreatedBy:             claims.UserID,
		UpdatedBy:             claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建仓单质押申请失败"))
		return
	}
	response.Success(c, h.warehouseReceiptPledgeToInfo(c, &item))
}

func (h *Handler) ListMyWarehouseReceiptPledges(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能查看仓单质押申请"))
		return
	}
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
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseReceiptPledgeApplication{}).
		Where("applicant_enterprise_id = ?", claims.EnterpriseID)
	query = applyWarehouseReceiptPledgeFilters(query, req)
	var total int64
	query.Count(&total)
	var rows []entity.WarehouseReceiptPledgeApplication
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询仓单质押申请失败"))
		return
	}
	list := make([]dto.WarehouseReceiptPledgeInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, h.warehouseReceiptPledgeToInfo(c, &row))
	}
	response.Success(c, dto.PageResponse[dto.WarehouseReceiptPledgeInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
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
	case "warehouse":
		query = query.Where("warehouse_service_id IN (?)", h.db.Model(&serviceOfferEntity.WarehouseService{}).Select("id").Where("enterprise_id = ?", claims.EnterpriseID))
	default:
		query = query.Where("buyer_enterprise_id = ? OR carrier_enterprise_id = ? OR logistics_service_id IN (?) OR warehouse_service_id IN (?)",
			claims.EnterpriseID,
			claims.EnterpriseID,
			h.db.Model(&serviceOfferEntity.LogisticsService{}).Select("id").Where("enterprise_id = ?", claims.EnterpriseID),
			h.db.Model(&serviceOfferEntity.WarehouseService{}).Select("id").Where("enterprise_id = ?", claims.EnterpriseID),
		)
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
	if err := h.ensureLogisticsStatusDocuments(c, &item, entity.LogisticsOrderPickupScheduled); err != nil {
		response.Error(c, err)
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
	if err := h.db.WithContext(c.Request.Context()).Where("id = ?", id).First(&item).Error; err != nil {
		response.NotFound(c, "物流委托不存在")
		return
	}
	if err := h.ensureLogisticsStatusActor(c, &item, targetStatus, claims.EnterpriseID); err != nil {
		response.Error(c, err)
		return
	}
	if err := h.ensureLogisticsStatusDocuments(c, &item, targetStatus); err != nil {
		response.Error(c, err)
		return
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
					actorService := "procurement"
					if item.OrderID > 0 && claims.EnterpriseID != order.BuyerID {
						var financing orderEntity.OrderFinancingRequest
						if err := tx.Where("order_id = ? AND funding_enterprise_id = ?", item.OrderID, claims.EnterpriseID).First(&financing).Error; err == nil {
							actorService = "finance"
						}
					}
					orderEvent := orderEntity.OrderEvent{
						OrderID:           order.ID,
						OrderNo:           order.OrderNo,
						EventType:         actorService + "_received",
						ActorUserID:       claims.UserID,
						ActorEnterpriseID: claims.EnterpriseID,
						ActorService:      actorService,
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

func (h *Handler) ensureLogisticsStatusActor(c *gin.Context, item *entity.LogisticsOrder, targetStatus entity.LogisticsOrderStatus, enterpriseID uint64) *ecode.Error {
	if item == nil {
		return ecode.ErrNotFound.WithMessage("物流委托不存在")
	}
	switch targetStatus {
	case entity.LogisticsOrderPickupScheduled, entity.LogisticsOrderLoadingConfirmed, entity.LogisticsOrderInTransit, entity.LogisticsOrderArrivalConfirmed:
		if item.CarrierEnterpriseID != enterpriseID {
			return ecode.ErrForbidden.WithMessage("当前节点需由物流方推进")
		}
	case entity.LogisticsOrderArrivalSignedIn, entity.LogisticsOrderOutboundConfirmed, entity.LogisticsOrderInboundCompleted:
		if item.WarehouseServiceID == 0 {
			return ecode.ErrBadRequest.WithMessage("当前物流单未绑定平台仓库，无法由仓储方执行该节点")
		}
		var warehouse serviceOfferEntity.WarehouseService
		if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", item.WarehouseServiceID, enterpriseID).First(&warehouse).Error; err != nil {
			return ecode.ErrForbidden.WithMessage("当前节点需由绑定仓储方推进")
		}
	case entity.LogisticsOrderReceiptSigned:
		if entity.NormalizeLogisticsOrderStatus(item.Status) != entity.LogisticsOrderInboundCompleted {
			return ecode.ErrBadRequest.WithMessage("需收货仓验货入库完成后，买方或资金方才能确认收货")
		}
		if item.OrderID == 0 {
			if item.BuyerEnterpriseID == enterpriseID {
				return nil
			}
			return ecode.ErrForbidden.WithMessage("当前节点需由买方或资金方确认")
		}
		var order orderEntity.Order
		if err := h.db.WithContext(c.Request.Context()).First(&order, item.OrderID).Error; err != nil {
			return ecode.ErrNotFound.WithMessage("关联订单不存在")
		}
		if order.FinanceID > 0 {
			var financing orderEntity.OrderFinancingRequest
			if err := h.db.WithContext(c.Request.Context()).
				Where("order_id = ? AND funding_enterprise_id = ? AND status IN ?", item.OrderID, enterpriseID, []string{
					string(orderEntity.FinancingRequestStatusPledged),
					string(orderEntity.FinancingRequestStatusReleased),
				}).
				First(&financing).Error; err != nil {
				return ecode.ErrForbidden.WithMessage("托盘垫资订单需由资金方确认收货")
			}
			return nil
		}
		if order.BuyerID != enterpriseID {
			return ecode.ErrForbidden.WithMessage("当前节点需由买方确认收货")
		}
	case entity.LogisticsOrderCompleted:
		if entity.NormalizeLogisticsOrderStatus(item.Status) != entity.LogisticsOrderReceiptSigned {
			return ecode.ErrBadRequest.WithMessage("需买方或资金方确认收货后，物流方才能归档")
		}
		if item.CarrierEnterpriseID != enterpriseID {
			return ecode.ErrForbidden.WithMessage("履约完成需由物流方归档")
		}
	default:
		if item.CarrierEnterpriseID != enterpriseID {
			return ecode.ErrForbidden.WithMessage("当前节点需由物流方推进")
		}
	}
	return nil
}

func (h *Handler) ensureLogisticsStatusDocuments(c *gin.Context, item *entity.LogisticsOrder, targetStatus entity.LogisticsOrderStatus) *ecode.Error {
	if targetStatus == entity.LogisticsOrderPickupScheduled {
		var count int64
		if err := h.db.WithContext(c.Request.Context()).
			Model(&entity.FulfillmentDocument{}).
			Where("logistics_order_id = ? AND doc_type = ? AND uploaded_role = ?", item.ID, entity.DocTypePickupInstruction, "seller").
			Count(&count).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("查询卖方提货指令失败")
		}
		if count == 0 {
			return ecode.ErrBadRequest.WithMessage("当前节点缺少卖方提货指令：pickup_instruction")
		}
		return nil
	}
	requiredGroups := requiredLogisticsDocTypeGroups(targetStatus)
	required := flattenDocTypeGroups(requiredGroups)
	if len(required) == 0 || item == nil || item.ID == 0 {
		return nil
	}
	var docs []entity.FulfillmentDocument
	if err := h.db.WithContext(c.Request.Context()).
		Model(&entity.FulfillmentDocument{}).
		Where("logistics_order_id = ? AND doc_type IN ?", item.ID, required).
		Find(&docs).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("查询物流履约单据失败")
	}
	existing := make(map[string]struct{}, len(docs))
	for _, doc := range docs {
		existing[string(doc.DocType)] = struct{}{}
	}
	missing := make([]string, 0, len(requiredGroups))
	for _, group := range requiredGroups {
		satisfied := false
		for _, docType := range group {
			if _, ok := existing[docType]; ok {
				satisfied = true
				break
			}
		}
		if !satisfied {
			missing = append(missing, strings.Join(group, "/"))
		}
	}
	if len(missing) > 0 {
		return ecode.ErrBadRequest.WithMessage("当前节点缺少必要履约单据：" + strings.Join(missing, "、"))
	}
	return nil
}

func requiredLogisticsDocTypeGroups(status entity.LogisticsOrderStatus) [][]string {
	required := requiredLogisticsDocTypes(status)
	groups := make([][]string, 0, len(required))
	for _, docType := range required {
		groups = append(groups, []string{docType})
	}
	if status == entity.LogisticsOrderLoadingConfirmed {
		return [][]string{{string(entity.DocTypeLoadingVoucher), string(entity.DocTypePickupOrder)}}
	}
	return groups
}

func flattenDocTypeGroups(groups [][]string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(groups))
	for _, group := range groups {
		for _, docType := range group {
			if _, ok := seen[docType]; ok {
				continue
			}
			seen[docType] = struct{}{}
			result = append(result, docType)
		}
	}
	return result
}

func requiredLogisticsDocTypes(status entity.LogisticsOrderStatus) []string {
	switch status {
	case entity.LogisticsOrderPickupScheduled:
		return []string{string(entity.DocTypePickupInstruction)}
	case entity.LogisticsOrderOutboundConfirmed:
		return []string{string(entity.DocTypeOutboundOrder), string(entity.DocTypeSealPhoto)}
	case entity.LogisticsOrderLoadingConfirmed:
		return []string{string(entity.DocTypeLoadingVoucher)}
	case entity.LogisticsOrderInTransit:
		return []string{string(entity.DocTypeGPSTrack)}
	case entity.LogisticsOrderArrivalConfirmed:
		return []string{string(entity.DocTypeDeliveryPhoto)}
	case entity.LogisticsOrderReceiptSigned:
		return nil
	case entity.LogisticsOrderInboundCompleted:
		return []string{string(entity.DocTypeSealPhoto), string(entity.DocTypeInboundOrder), string(entity.DocTypeReceiptOrder)}
	default:
		return nil
	}
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
	requirements := make([]dto.LogisticsStatusRequirementInfo, 0, len(logisticsStatusRequirementDefinitions()))
	satisfiedDocs := map[string]struct{}{}
	if item != nil && item.ID > 0 {
		var docs []entity.FulfillmentDocument
		database.GetDB().Where("logistics_order_id = ?", item.ID).Find(&docs)
		for _, doc := range docs {
			satisfiedDocs[string(doc.DocType)] = struct{}{}
		}
	}
	for _, requirement := range logisticsStatusRequirementDefinitions() {
		satisfied := true
		for _, docType := range requirement.RequiredDoc {
			if _, ok := satisfiedDocs[docType]; !ok {
				satisfied = false
				break
			}
		}
		requirement.Satisfied = satisfied
		requirements = append(requirements, requirement)
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
		CreatedAt:          item.CreatedAt, UpdatedAt: item.UpdatedAt, TrackingEvents: eventInfos, StatusRequirements: requirements,
	}
}

func logisticsStatusRequirementDefinitions() []dto.LogisticsStatusRequirementInfo {
	return []dto.LogisticsStatusRequirementInfo{
		{Status: string(entity.LogisticsOrderPickupScheduled), Label: "提货预约", Actor: "logistics", Description: "物流方查看卖方提货指令后预约提货", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderPickupScheduled)},
		{Status: string(entity.LogisticsOrderArrivalSignedIn), Label: "到场签到", Actor: "warehouse", Description: "仓储方确认车辆到场"},
		{Status: string(entity.LogisticsOrderOutboundConfirmed), Label: "出库确认", Actor: "warehouse", Description: "仓储方上传出库单和封条照片", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderOutboundConfirmed)},
		{Status: string(entity.LogisticsOrderLoadingConfirmed), Label: "装货确认", Actor: "logistics", Description: "物流方上传装车凭证并确认装货", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderLoadingConfirmed)},
		{Status: string(entity.LogisticsOrderInTransit), Label: "运输中", Actor: "logistics", Description: "物流方上传 GPS 或在途轨迹", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderInTransit)},
		{Status: string(entity.LogisticsOrderArrivalConfirmed), Label: "到货确认", Actor: "logistics", Description: "物流方上传到货照片", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderArrivalConfirmed)},
		{Status: string(entity.LogisticsOrderInboundCompleted), Label: "收货仓验货入库", Actor: "warehouse", Description: "收货仓上传封条照片、入库单和收货单", RequiredDoc: requiredLogisticsDocTypes(entity.LogisticsOrderInboundCompleted)},
		{Status: string(entity.LogisticsOrderReceiptSigned), Label: "确认收货", Actor: "buyer_or_finance", Description: "普通订单买方确认，融资订单资金方确认"},
		{Status: string(entity.LogisticsOrderCompleted), Label: "履约完成", Actor: "logistics", Description: "买方/资金方确认收货后，物流方归档物流履约"},
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
		entity.LogisticsOrderInboundCompleted:  "收货仓已完成验货入库",
		entity.LogisticsOrderReceiptSigned:     "买方/资金方已确认收货",
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

	role, actorErr := h.resolveFulfillmentDocumentActor(c, &logisticsOrder, claims.EnterpriseID)
	if actorErr != nil {
		response.Error(c, actorErr)
		return
	}
	if entity.FulfillmentDocType(req.DocType) == entity.DocTypePickupInstruction {
		isSeller, sellerErr := h.isLogisticsOrderSeller(c, &logisticsOrder, claims.EnterpriseID)
		if sellerErr != nil {
			response.Error(c, sellerErr)
			return
		}
		if !isSeller {
			response.Error(c, ecode.ErrForbidden.WithMessage("提货指令需由卖方上传"))
			return
		}
		role = "seller"
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

func (h *Handler) resolveFulfillmentDocumentActor(c *gin.Context, logisticsOrder *entity.LogisticsOrder, enterpriseID uint64) (string, *ecode.Error) {
	if logisticsOrder == nil || logisticsOrder.ID == 0 {
		return "", ecode.ErrNotFound.WithMessage("物流委托单不存在")
	}
	if enterpriseID == logisticsOrder.CarrierEnterpriseID {
		return "logistics", nil
	}
	if enterpriseID == logisticsOrder.BuyerEnterpriseID {
		return "buyer", nil
	}
	if logisticsOrder.OrderID > 0 {
		var order orderEntity.Order
		if err := h.db.WithContext(c.Request.Context()).First(&order, logisticsOrder.OrderID).Error; err == nil {
			if order.SellerID == enterpriseID || order.SellerEnterpriseID == enterpriseID {
				return "seller", nil
			}
		}
	}
	if logisticsOrder.WarehouseServiceID > 0 {
		var warehouse serviceOfferEntity.WarehouseService
		if err := h.db.WithContext(c.Request.Context()).
			Where("id = ? AND enterprise_id = ?", logisticsOrder.WarehouseServiceID, enterpriseID).
			First(&warehouse).Error; err == nil {
			return "warehouse", nil
		}
	}
	if logisticsOrder.OrderID > 0 {
		var order orderEntity.Order
		if err := h.db.WithContext(c.Request.Context()).First(&order, logisticsOrder.OrderID).Error; err == nil {
			var financing orderEntity.OrderFinancingRequest
			if err := h.db.WithContext(c.Request.Context()).
				Where("order_id = ? AND funding_enterprise_id = ?", logisticsOrder.OrderID, enterpriseID).
				First(&financing).Error; err == nil {
				return "finance", nil
			}
		}
	}
	return "", ecode.ErrForbidden.WithMessage("无权上传该物流委托单据")
}

func (h *Handler) isLogisticsOrderSeller(c *gin.Context, logisticsOrder *entity.LogisticsOrder, enterpriseID uint64) (bool, *ecode.Error) {
	if logisticsOrder == nil || logisticsOrder.ID == 0 {
		return false, ecode.ErrNotFound.WithMessage("物流委托单不存在")
	}
	if logisticsOrder.OrderID == 0 {
		return false, nil
	}
	var order orderEntity.Order
	if err := h.db.WithContext(c.Request.Context()).First(&order, logisticsOrder.OrderID).Error; err != nil {
		return false, ecode.ErrNotFound.WithMessage("关联交易订单不存在")
	}
	return order.SellerID == enterpriseID || order.SellerEnterpriseID == enterpriseID, nil
}

func (h *Handler) ListFulfillmentDocuments(c *gin.Context) {
	logisticsOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的物流委托ID")
		return
	}
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var logisticsOrder entity.LogisticsOrder
	if err := h.db.WithContext(c.Request.Context()).First(&logisticsOrder, logisticsOrderID).Error; err != nil {
		response.Error(c, ecode.ErrNotFound.WithMessage("物流委托单不存在"))
		return
	}
	if _, actorErr := h.resolveFulfillmentDocumentActor(c, &logisticsOrder, claims.EnterpriseID); actorErr != nil {
		response.Error(c, actorErr)
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
			item.LogisticsOrderID = logisticsOrder.ID
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

func (h *Handler) ListFinanceWarehouseReceiptPledges(c *gin.Context) {
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
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseReceiptPledgeApplication{}).
		Where("funding_enterprise_id = ?", claims.EnterpriseID)
	query = applyWarehouseReceiptPledgeFilters(query, req)
	var total int64
	query.Count(&total)
	var rows []entity.WarehouseReceiptPledgeApplication
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("查询仓单质押申请失败"))
		return
	}
	list := make([]dto.WarehouseReceiptPledgeInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, h.warehouseReceiptPledgeToInfo(c, &row))
	}
	response.Success(c, dto.PageResponse[dto.WarehouseReceiptPledgeInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
}

func (h *Handler) ReviewWarehouseReceiptPledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	item, ok := h.getFinanceWarehouseReceiptPledge(c)
	if !ok {
		return
	}
	if item.Status != entity.WarehouseReceiptPledgeSubmitted {
		response.BadRequest(c, "当前状态不可进行风控审核")
		return
	}
	var req dto.ReviewWarehouseReceiptPledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	now := time.Now()
	if req.Approved {
		if req.ApprovedAmount <= 0 {
			response.BadRequest(c, "审批通过时审批金额必须大于0")
			return
		}
		item.Status = entity.WarehouseReceiptPledgeRiskApproved
		item.ApprovedAmount = req.ApprovedAmount
		item.TermDays = req.TermDays
		item.AnnualRate = req.AnnualRate
	} else {
		item.Status = entity.WarehouseReceiptPledgeRiskRejected
	}
	item.RiskRemark = strings.TrimSpace(req.RiskRemark)
	item.RiskReviewedAt = &now
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("提交仓单质押风控结果失败"))
		return
	}
	response.Success(c, h.warehouseReceiptPledgeToInfo(c, item))
}

func (h *Handler) SignWarehouseReceiptPledgeAgreement(c *gin.Context) {
	claims := middleware.GetClaims(c)
	item, ok := h.getFinanceWarehouseReceiptPledge(c)
	if !ok {
		return
	}
	if item.Status != entity.WarehouseReceiptPledgeRiskApproved {
		response.BadRequest(c, "当前状态不可签署质押监管协议")
		return
	}
	var req dto.SignWarehouseReceiptPledgeAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if !isHTTPURL(req.PledgeAgreementURL) || !isHTTPURL(req.SupervisionAgreementURL) {
		response.BadRequest(c, "协议URL必须是有效的HTTP/HTTPS地址")
		return
	}
	now := time.Now()
	item.PledgeAgreementURL = strings.TrimSpace(req.PledgeAgreementURL)
	item.SupervisionAgreementURL = strings.TrimSpace(req.SupervisionAgreementURL)
	item.Status = entity.WarehouseReceiptPledgeAgreementSigned
	item.AgreementSignedAt = &now
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("登记质押监管协议失败"))
		return
	}
	response.Success(c, h.warehouseReceiptPledgeToInfo(c, item))
}

func (h *Handler) DisburseWarehouseReceiptPledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	item, ok := h.getFinanceWarehouseReceiptPledge(c)
	if !ok {
		return
	}
	if item.Status != entity.WarehouseReceiptPledgeAgreementSigned {
		response.BadRequest(c, "当前状态不可放款")
		return
	}
	var req dto.DisburseWarehouseReceiptPledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if !isHTTPURL(req.DisbursementVoucherURL) {
		response.BadRequest(c, "放款凭证URL必须是有效的HTTP/HTTPS地址")
		return
	}
	now := time.Now()
	item.DisbursementVoucherURL = strings.TrimSpace(req.DisbursementVoucherURL)
	if strings.TrimSpace(req.Remark) != "" {
		item.Remark = strings.TrimSpace(req.Remark)
	}
	item.Status = entity.WarehouseReceiptPledgeDisbursed
	item.DisbursedAt = &now
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(item).Error; err != nil {
			return err
		}
		return h.freezeWarehouseReceiptInventory(c, tx, item, claims.UserID)
	}); err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("确认仓单质押放款失败"))
		return
	}
	response.Success(c, h.warehouseReceiptPledgeToInfo(c, item))
}

func (h *Handler) ReleaseWarehouseReceiptPledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	item, ok := h.getFinanceWarehouseReceiptPledge(c)
	if !ok {
		return
	}
	if item.Status != entity.WarehouseReceiptPledgeDisbursed && item.Status != entity.WarehouseReceiptPledgeReleasing {
		response.BadRequest(c, "当前状态不可释放货权")
		return
	}
	var req dto.ReleaseWarehouseReceiptPledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.ReleaseVoucherURL) != "" && !isHTTPURL(req.ReleaseVoucherURL) {
		response.BadRequest(c, "释放凭证URL必须是有效的HTTP/HTTPS地址")
		return
	}
	now := time.Now()
	item.ReleaseVoucherURL = strings.TrimSpace(req.ReleaseVoucherURL)
	item.ReleaseRemark = strings.TrimSpace(req.ReleaseRemark)
	item.Status = entity.WarehouseReceiptPledgeReleased
	item.ReleasedAt = &now
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(item).Error; err != nil {
			return err
		}
		return h.releaseWarehouseReceiptInventory(c, tx, item, claims.UserID)
	}); err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("释放仓单质押货权失败"))
		return
	}
	response.Success(c, h.warehouseReceiptPledgeToInfo(c, item))
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

func (h *Handler) ensureFinanceWarehouseAuthorized(c *gin.Context, financeServiceID, financeEnterpriseID, warehouseServiceID uint64) *ecode.Error {
	var authorizedCount int64
	if err := h.db.WithContext(c.Request.Context()).
		Model(&serviceOfferEntity.FinanceAuthorizedWarehouse{}).
		Where("finance_service_id = ? AND finance_enterprise_id = ? AND status = ?", financeServiceID, financeEnterpriseID, "active").
		Count(&authorizedCount).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("查询资方监管仓授权失败")
	}
	if authorizedCount == 0 {
		return nil
	}
	var matchedCount int64
	if err := h.db.WithContext(c.Request.Context()).
		Model(&serviceOfferEntity.FinanceAuthorizedWarehouse{}).
		Where("finance_service_id = ? AND finance_enterprise_id = ? AND warehouse_service_id = ? AND status = ?", financeServiceID, financeEnterpriseID, warehouseServiceID, "active").
		Count(&matchedCount).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("查询资方监管仓授权失败")
	}
	if matchedCount == 0 {
		return ecode.ErrBadRequest.WithMessage("所选仓库不在该资金产品认可的监管仓范围内")
	}
	return nil
}

func applyWarehouseReceiptPledgeFilters(query *gorm.DB, req dto.ListRequest) *gorm.DB {
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.WarehouseServiceID > 0 {
		query = query.Where("warehouse_service_id = ?", req.WarehouseServiceID)
	}
	if req.FinanceServiceID > 0 {
		query = query.Where("finance_service_id = ?", req.FinanceServiceID)
	}
	return query
}

func (h *Handler) getFinanceWarehouseReceiptPledge(c *gin.Context) (*entity.WarehouseReceiptPledgeApplication, bool) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能处理仓单质押"))
		return nil, false
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的仓单质押申请ID")
		return nil, false
	}
	var item entity.WarehouseReceiptPledgeApplication
	if err := h.db.WithContext(c.Request.Context()).
		Where("id = ? AND funding_enterprise_id = ?", id, claims.EnterpriseID).
		First(&item).Error; err != nil {
		response.Error(c, ecode.ErrForbidden.WithMessage("仓单质押申请不存在或无权处理"))
		return nil, false
	}
	return &item, true
}

func (h *Handler) freezeWarehouseReceiptInventory(c *gin.Context, tx *gorm.DB, item *entity.WarehouseReceiptPledgeApplication, operatorID uint64) error {
	if item == nil {
		return nil
	}
	var snapshots []entity.WarehouseInventorySnapshot
	query := tx.WithContext(c.Request.Context()).
		Where("warehouse_service_id = ? AND product_name = ?", item.WarehouseServiceID, item.GoodsName)
	if item.ReceiptNo != "" {
		query = query.Where("(external_ref_no = ? OR bundle_no = ? OR batch_no = ?)", item.ReceiptNo, item.ReceiptNo, item.ReceiptNo)
	}
	if err := query.Find(&snapshots).Error; err != nil {
		return err
	}
	for i := range snapshots {
		snapshots[i].PledgeStatus = "frozen"
		snapshots[i].FrozenQuantity = item.Quantity
		snapshots[i].UpdatedBy = operatorID
		if strings.TrimSpace(snapshots[i].Remark) == "" {
			snapshots[i].Remark = "仓单质押冻结：" + item.ApplicationNo
		} else if !strings.Contains(snapshots[i].Remark, item.ApplicationNo) {
			snapshots[i].Remark = strings.TrimSpace(snapshots[i].Remark) + "；仓单质押冻结：" + item.ApplicationNo
		}
		if err := tx.Save(&snapshots[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) releaseWarehouseReceiptInventory(c *gin.Context, tx *gorm.DB, item *entity.WarehouseReceiptPledgeApplication, operatorID uint64) error {
	if item == nil {
		return nil
	}
	var snapshots []entity.WarehouseInventorySnapshot
	query := tx.WithContext(c.Request.Context()).
		Where("warehouse_service_id = ? AND product_name = ?", item.WarehouseServiceID, item.GoodsName)
	if item.ReceiptNo != "" {
		query = query.Where("(external_ref_no = ? OR bundle_no = ? OR batch_no = ?)", item.ReceiptNo, item.ReceiptNo, item.ReceiptNo)
	}
	if err := query.Find(&snapshots).Error; err != nil {
		return err
	}
	for i := range snapshots {
		snapshots[i].PledgeStatus = "released"
		snapshots[i].FrozenQuantity = 0
		snapshots[i].UpdatedBy = operatorID
		if strings.TrimSpace(snapshots[i].Remark) == "" {
			snapshots[i].Remark = "仓单质押释放：" + item.ApplicationNo
		} else if !strings.Contains(snapshots[i].Remark, item.ApplicationNo+"释放") {
			snapshots[i].Remark = strings.TrimSpace(snapshots[i].Remark) + "；仓单质押释放：" + item.ApplicationNo
		}
		if err := tx.Save(&snapshots[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) warehouseReceiptPledgeToInfo(c *gin.Context, item *entity.WarehouseReceiptPledgeApplication) dto.WarehouseReceiptPledgeInfo {
	if item == nil {
		return dto.WarehouseReceiptPledgeInfo{}
	}
	info := dto.WarehouseReceiptPledgeInfo{
		ID:                      item.ID,
		ApplicationNo:           item.ApplicationNo,
		ApplicantEnterpriseID:   item.ApplicantEnterpriseID,
		ApplicantUserID:         item.ApplicantUserID,
		ApplicantRole:           item.ApplicantRole,
		FinanceServiceID:        item.FinanceServiceID,
		FundingEnterpriseID:     item.FundingEnterpriseID,
		WarehouseServiceID:      item.WarehouseServiceID,
		WarehouseEnterpriseID:   item.WarehouseEnterpriseID,
		WarehouseName:           item.WarehouseName,
		ReceiptNo:               item.ReceiptNo,
		ReceiptVoucherURL:       item.ReceiptVoucherURL,
		GoodsName:               item.GoodsName,
		Quantity:                item.Quantity,
		QualityGrade:            item.QualityGrade,
		RequestedAmount:         item.RequestedAmount,
		ApprovedAmount:          item.ApprovedAmount,
		TermDays:                item.TermDays,
		AnnualRate:              item.AnnualRate,
		RiskRemark:              item.RiskRemark,
		PledgeAgreementURL:      item.PledgeAgreementURL,
		SupervisionAgreementURL: item.SupervisionAgreementURL,
		DisbursementVoucherURL:  item.DisbursementVoucherURL,
		ReleaseVoucherURL:       item.ReleaseVoucherURL,
		ReleaseRemark:           item.ReleaseRemark,
		Status:                  string(item.Status),
		SubmittedAt:             item.SubmittedAt,
		RiskReviewedAt:          item.RiskReviewedAt,
		AgreementSignedAt:       item.AgreementSignedAt,
		DisbursedAt:             item.DisbursedAt,
		ReleasedAt:              item.ReleasedAt,
		Remark:                  item.Remark,
		CreatedAt:               item.CreatedAt,
		UpdatedAt:               item.UpdatedAt,
	}
	var finance serviceOfferEntity.FinanceService
	if err := h.db.WithContext(c.Request.Context()).Select("name").First(&finance, item.FinanceServiceID).Error; err == nil {
		info.FinanceServiceName = finance.Name
	}
	return info
}

func resolveApplicantRole(tradeRole string, capabilities []string) string {
	switch strings.TrimSpace(tradeRole) {
	case "buyer":
		return "procurement"
	case "seller":
		return "sales"
	}
	for _, capability := range capabilities {
		if capability == "sales" {
			return "sales"
		}
	}
	for _, capability := range capabilities {
		if capability == "procurement" {
			return "procurement"
		}
	}
	return ""
}

func isHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
