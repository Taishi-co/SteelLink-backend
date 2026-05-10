package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
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

func (h *Handler) CreateWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能上架仓储资源"))
		return
	}
	var req dto.WarehouseServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.WarehouseService{
		EnterpriseID:       claims.EnterpriseID,
		WarehouseName:      req.WarehouseName,
		Province:           req.Province,
		City:               req.City,
		District:           req.District,
		Address:            req.Address,
		Capacity:           req.Capacity,
		AvailableCapacity:  req.AvailableCapacity,
		ServiceTypes:       entity.StringList(req.ServiceTypes),
		PriceRule:          req.PriceRule,
		LoadingCapacity:    req.LoadingCapacity,
		BusinessHours:      req.BusinessHours,
		PledgeSupervision:  req.PledgeSupervision,
		OutboundVoucherReq: req.OutboundVoucherReq,
		Images:             entity.StringList(req.Images),
		Certificates:       entity.StringList(req.Certificates),
		Description:        req.Description,
		Status:             entity.OfferStatusPending,
		CreatedBy:          claims.UserID,
		UpdatedBy:          claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建仓储资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, warehouseToInfo(&item))
}

func (h *Handler) ListWarehouses(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Status == nil {
		status := int(entity.OfferStatusApproved)
		req.Status = &status
	}
	key := cache.Key("service:warehouse:list", req)
	var cached dto.PageResponse[dto.WarehouseServiceInfo]
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.WarehouseService{})
	query = applyCommonQuery(query, req, "warehouse_name")
	var total int64
	query.Count(&total)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	var rows []entity.WarehouseService
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.WarehouseServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, warehouseToInfo(&row))
	}
	result := dto.PageResponse[dto.WarehouseServiceInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferListTTL)
	response.Success(c, result)
}

func (h *Handler) ListMyWarehouses(c *gin.Context) {
	claims := middleware.GetClaims(c)
	status := c.Query("status")
	query := h.db.WithContext(c.Request.Context()).Where("enterprise_id = ?", claims.EnterpriseID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var rows []entity.WarehouseService
	if err := query.Order("id DESC").Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.WarehouseServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, warehouseToInfo(&row))
	}
	response.Success(c, list)
}

func (h *Handler) UpdateWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能编辑仓储资源"))
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.WarehouseServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "仓储资源不存在或无权修改")
		return
	}
	item.WarehouseName = req.WarehouseName
	item.Province = req.Province
	item.City = req.City
	item.District = req.District
	item.Address = req.Address
	item.Capacity = req.Capacity
	item.AvailableCapacity = req.AvailableCapacity
	item.ServiceTypes = entity.StringList(req.ServiceTypes)
	item.PriceRule = req.PriceRule
	item.LoadingCapacity = req.LoadingCapacity
	item.BusinessHours = req.BusinessHours
	item.PledgeSupervision = req.PledgeSupervision
	item.OutboundVoucherReq = req.OutboundVoucherReq
	item.Images = entity.StringList(req.Images)
	item.Certificates = entity.StringList(req.Certificates)
	item.Description = req.Description
	item.Status = entity.OfferStatusPending
	item.AuditRemark = ""
	item.ApprovedAt = nil
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新仓储资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, warehouseToInfo(&item))
}

func (h *Handler) DeleteWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item entity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "仓储资源不存在或无权下线")
		return
	}
	item.Status = entity.OfferStatusOffline
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("下线仓储资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, warehouseToInfo(&item))
}

func (h *Handler) CreateLogistics(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能上架物流资源"))
		return
	}
	var req dto.LogisticsServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.LogisticsService{
		EnterpriseID: claims.EnterpriseID,
		Name:         req.Name,
		RouteStart:   req.RouteStart,
		RouteEnd:     req.RouteEnd,
		ServiceScope: req.ServiceScope,
		VehicleType:  req.VehicleType,
		LoadCapacity: req.LoadCapacity,
		PriceRule:    req.PriceRule,
		QuoteRules:   entity.JSONMap(req.QuoteRules),
		FleetName:    req.FleetName,
		VehicleCerts: entity.StringList(req.VehicleCerts),
		DriverCerts:  entity.StringList(req.DriverCerts),
		Images:       entity.StringList(req.Images),
		Certificates: entity.StringList(req.Certificates),
		Description:  req.Description,
		Status:       entity.OfferStatusPending,
		CreatedBy:    claims.UserID,
		UpdatedBy:    claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建物流资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, logisticsToInfo(&item))
}

func (h *Handler) ListLogistics(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Status == nil {
		status := int(entity.OfferStatusApproved)
		req.Status = &status
	}
	key := cache.Key("service:logistics:list", req)
	var cached dto.PageResponse[dto.LogisticsServiceInfo]
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.LogisticsService{})
	query = applyCommonQuery(query, req, "name", "route_start", "route_end", "service_scope", "vehicle_type")
	var total int64
	query.Count(&total)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	var rows []entity.LogisticsService
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.LogisticsServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, logisticsToInfo(&row))
	}
	result := dto.PageResponse[dto.LogisticsServiceInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferListTTL)
	response.Success(c, result)
}

func (h *Handler) ListMyLogistics(c *gin.Context) {
	claims := middleware.GetClaims(c)
	status := c.Query("status")
	query := h.db.WithContext(c.Request.Context()).Where("enterprise_id = ?", claims.EnterpriseID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var rows []entity.LogisticsService
	if err := query.Order("id DESC").Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.LogisticsServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, logisticsToInfo(&row))
	}
	response.Success(c, list)
}

func (h *Handler) UpdateLogistics(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能编辑物流资源"))
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.LogisticsServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "物流资源不存在或无权修改")
		return
	}
	item.Name = req.Name
	item.RouteStart = req.RouteStart
	item.RouteEnd = req.RouteEnd
	item.ServiceScope = req.ServiceScope
	item.VehicleType = req.VehicleType
	item.LoadCapacity = req.LoadCapacity
	item.PriceRule = req.PriceRule
	item.QuoteRules = entity.JSONMap(req.QuoteRules)
	item.FleetName = req.FleetName
	item.VehicleCerts = entity.StringList(req.VehicleCerts)
	item.DriverCerts = entity.StringList(req.DriverCerts)
	item.Images = entity.StringList(req.Images)
	item.Certificates = entity.StringList(req.Certificates)
	item.Description = req.Description
	item.Status = entity.OfferStatusPending
	item.AuditRemark = ""
	item.ApprovedAt = nil
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新物流资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, logisticsToInfo(&item))
}

func (h *Handler) DeleteLogistics(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item entity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "物流资源不存在或无权下线")
		return
	}
	item.Status = entity.OfferStatusOffline
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("下线物流资源失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, logisticsToInfo(&item))
}

func (h *Handler) CreateFinance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能上架金融服务"))
		return
	}
	var req dto.FinanceServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item := entity.FinanceService{
		EnterpriseID:       claims.EnterpriseID,
		Name:               req.Name,
		FinanceType:        req.FinanceType,
		MinAmount:          req.MinAmount,
		MaxAmount:          req.MaxAmount,
		TermDays:           req.TermDays,
		AnnualRate:         req.AnnualRate,
		FeeRule:            req.FeeRule,
		RiskControlRule:    req.RiskControlRule,
		RequiredDocuments:  entity.StringList(req.RequiredDocuments),
		SupportedScenarios: entity.StringList(req.SupportedScenarios),
		Description:        req.Description,
		Status:             entity.OfferStatusPending,
		CreatedBy:          claims.UserID,
		UpdatedBy:          claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建金融服务失败"))
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, financeToInfo(&item))
}

func (h *Handler) ListFinance(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Status == nil {
		status := int(entity.OfferStatusApproved)
		req.Status = &status
	}
	key := cache.Key("service:finance:list", req)
	var cached dto.PageResponse[dto.FinanceServiceInfo]
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	query := h.db.WithContext(c.Request.Context()).Model(&entity.FinanceService{})
	query = applyCommonQuery(query, req, "name")
	var total int64
	query.Count(&total)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	var rows []entity.FinanceService
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.FinanceServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, financeToInfo(&row))
	}
	result := dto.PageResponse[dto.FinanceServiceInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferListTTL)
	response.Success(c, result)
}

func (h *Handler) ListMyFinance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	status := c.Query("status")
	query := h.db.WithContext(c.Request.Context()).Where("enterprise_id = ?", claims.EnterpriseID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var rows []entity.FinanceService
	if err := query.Order("id DESC").Find(&rows).Error; err != nil {
		response.Error(c, ecode.ErrDatabase)
		return
	}
	list := make([]dto.FinanceServiceInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, financeToInfo(&row))
	}
	response.Success(c, list)
}

func (h *Handler) UpdateFinance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.EnterpriseID == 0 {
		response.Error(c, ecode.ErrForbidden.WithMessage("企业用户才能编辑金融服务"))
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.FinanceServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	var item entity.FinanceService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "金融服务不存在或无权修改")
		return
	}
	item.Name = req.Name
	item.FinanceType = req.FinanceType
	item.MinAmount = req.MinAmount
	item.MaxAmount = req.MaxAmount
	item.TermDays = req.TermDays
	item.AnnualRate = req.AnnualRate
	item.FeeRule = req.FeeRule
	item.RiskControlRule = req.RiskControlRule
	item.RequiredDocuments = entity.StringList(req.RequiredDocuments)
	item.SupportedScenarios = entity.StringList(req.SupportedScenarios)
	item.Description = req.Description
	item.Status = entity.OfferStatusPending
	item.AuditRemark = ""
	item.ApprovedAt = nil
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("更新金融服务失败"))
		return
	}
	response.Success(c, financeToInfo(&item))
}

func (h *Handler) DeleteFinance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item entity.FinanceService
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND enterprise_id = ?", id, claims.EnterpriseID).First(&item).Error; err != nil {
		response.NotFound(c, "金融服务不存在或无权下线")
		return
	}
	item.Status = entity.OfferStatusOffline
	item.UpdatedBy = claims.UserID
	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("下线金融服务失败"))
		return
	}
	response.Success(c, financeToInfo(&item))
}

func (h *Handler) AuditWarehouse(c *gin.Context) {
	var item entity.WarehouseService
	if !h.auditOffer(c, &item) {
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, warehouseToInfo(&item))
}

func (h *Handler) AuditLogistics(c *gin.Context) {
	var item entity.LogisticsService
	if !h.auditOffer(c, &item) {
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, logisticsToInfo(&item))
}

func (h *Handler) AuditFinance(c *gin.Context) {
	var item entity.FinanceService
	if !h.auditOffer(c, &item) {
		return
	}
	invalidateServiceOfferCache(c.Request.Context())
	response.Success(c, financeToInfo(&item))
}

func (h *Handler) GetWarehouse(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	key := "service:warehouse:detail:" + cache.Hash(id)
	var cached dto.WarehouseServiceInfo
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	var item entity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "仓储资源不存在")
		return
	}
	if item.Status != entity.OfferStatusApproved {
		response.NotFound(c, "仓储资源不存在")
		return
	}
	result := warehouseToInfo(&item)
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferDetailTTL)
	response.Success(c, result)
}

func (h *Handler) auditOffer(c *gin.Context, item any) bool {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return false
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务资源ID")
		return false
	}
	var req dto.AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return false
	}
	status := entity.OfferStatus(req.Status)
	remark := strings.TrimSpace(req.AuditRemark)
	if remark == "" {
		remark = strings.TrimSpace(req.AuditRemarkCamel)
	}
	if err := h.db.WithContext(c.Request.Context()).First(item, id).Error; err != nil {
		response.NotFound(c, "服务资源不存在")
		return false
	}
	setOfferAudit(item, status, remark, claims.UserID)
	if err := h.db.WithContext(c.Request.Context()).Save(item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("审核服务资源失败"))
		return false
	}
	return true
}

func setOfferAudit(item any, status entity.OfferStatus, remark string, operatorID uint64) {
	var approvedAt *time.Time
	if status == entity.OfferStatusApproved {
		now := time.Now()
		approvedAt = &now
	}
	switch target := item.(type) {
	case *entity.WarehouseService:
		target.Status = status
		target.AuditRemark = remark
		target.UpdatedBy = operatorID
		target.ApprovedAt = approvedAt
	case *entity.LogisticsService:
		target.Status = status
		target.AuditRemark = remark
		target.UpdatedBy = operatorID
		target.ApprovedAt = approvedAt
	case *entity.FinanceService:
		target.Status = status
		target.AuditRemark = remark
		target.UpdatedBy = operatorID
		target.ApprovedAt = approvedAt
	}
}

func (h *Handler) GetLogistics(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	key := "service:logistics:detail:" + cache.Hash(id)
	var cached dto.LogisticsServiceInfo
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	var item entity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "物流资源不存在")
		return
	}
	if item.Status != entity.OfferStatusApproved {
		response.NotFound(c, "物流资源不存在")
		return
	}
	result := logisticsToInfo(&item)
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferDetailTTL)
	response.Success(c, result)
}

func (h *Handler) GetFinance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	key := "service:finance:detail:" + cache.Hash(id)
	var cached dto.FinanceServiceInfo
	if cache.GetJSON(c.Request.Context(), key, &cached) {
		response.Success(c, cached)
		return
	}
	var item entity.FinanceService
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "金融服务不存在")
		return
	}
	if item.Status != entity.OfferStatusApproved {
		response.NotFound(c, "金融服务不存在")
		return
	}
	result := financeToInfo(&item)
	cache.SetJSON(c.Request.Context(), key, result, cache.ServiceOfferDetailTTL)
	response.Success(c, result)
}

func invalidateServiceOfferCache(ctx context.Context) {
	cache.DeleteByPattern(ctx, "service:warehouse:*")
	cache.DeleteByPattern(ctx, "service:logistics:*")
	cache.DeleteByPattern(ctx, "service:finance:*")
	cache.DeleteByPattern(ctx, "ai:match-logistics:*")
	cache.Delete(ctx, "ai:approved-logistics-services")
}

func applyCommonQuery(query *gorm.DB, req dto.ListRequest, searchColumns ...string) *gorm.DB {
	if req.Keyword != "" {
		clauses := make([]string, 0, len(searchColumns)+1)
		args := make([]any, 0, len(searchColumns)+1)
		for _, column := range searchColumns {
			clauses = append(clauses, column+" LIKE ?")
			args = append(args, "%"+req.Keyword+"%")
		}
		clauses = append(clauses, "description LIKE ?")
		args = append(args, "%"+req.Keyword+"%")
		query = query.Where("("+strings.Join(clauses, " OR ")+")", args...)
	}
	if req.Province != "" {
		query = query.Where("province = ?", req.Province)
	}
	if req.City != "" {
		query = query.Where("city = ?", req.City)
	}
	if req.EnterpriseID > 0 {
		query = query.Where("enterprise_id = ?", req.EnterpriseID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	return query
}

func warehouseToInfo(item *entity.WarehouseService) dto.WarehouseServiceInfo {
	return dto.WarehouseServiceInfo{
		ID: item.ID, EnterpriseID: item.EnterpriseID, WarehouseName: item.WarehouseName,
		Province: item.Province, City: item.City, District: item.District, Address: item.Address,
		Capacity: item.Capacity, AvailableCapacity: item.AvailableCapacity, ServiceTypes: []string(item.ServiceTypes),
		PriceRule: item.PriceRule, LoadingCapacity: item.LoadingCapacity, BusinessHours: item.BusinessHours,
		PledgeSupervision: item.PledgeSupervision, OutboundVoucherReq: item.OutboundVoucherReq,
		Images: []string(item.Images), Certificates: []string(item.Certificates), Description: item.Description,
		Status: int(item.Status), AuditRemark: item.AuditRemark,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func logisticsToInfo(item *entity.LogisticsService) dto.LogisticsServiceInfo {
	return dto.LogisticsServiceInfo{
		ID: item.ID, EnterpriseID: item.EnterpriseID, Name: item.Name, RouteStart: item.RouteStart,
		RouteEnd: item.RouteEnd, ServiceScope: item.ServiceScope, VehicleType: item.VehicleType,
		LoadCapacity: item.LoadCapacity, PriceRule: item.PriceRule, QuoteRules: map[string]any(item.QuoteRules),
		FleetName: item.FleetName, VehicleCerts: []string(item.VehicleCerts), DriverCerts: []string(item.DriverCerts),
		Images: []string(item.Images), Certificates: []string(item.Certificates), Description: item.Description,
		Status: int(item.Status), AuditRemark: item.AuditRemark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func financeToInfo(item *entity.FinanceService) dto.FinanceServiceInfo {
	return dto.FinanceServiceInfo{
		ID: item.ID, EnterpriseID: item.EnterpriseID, Name: item.Name, FinanceType: item.FinanceType,
		MinAmount: item.MinAmount, MaxAmount: item.MaxAmount, TermDays: item.TermDays,
		AnnualRate: item.AnnualRate, FeeRule: item.FeeRule, RiskControlRule: item.RiskControlRule,
		RequiredDocuments: []string(item.RequiredDocuments), SupportedScenarios: []string(item.SupportedScenarios),
		Description: item.Description, Status: int(item.Status), AuditRemark: item.AuditRemark,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
