package handler

import (
	"strconv"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/ui/dto"
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
		EnterpriseID:      claims.EnterpriseID,
		WarehouseName:     req.WarehouseName,
		Province:          req.Province,
		City:              req.City,
		District:          req.District,
		Address:           req.Address,
		Capacity:          req.Capacity,
		AvailableCapacity: req.AvailableCapacity,
		ServiceTypes:      entity.StringList(req.ServiceTypes),
		PriceRule:         req.PriceRule,
		Images:            entity.StringList(req.Images),
		Certificates:      entity.StringList(req.Certificates),
		Description:       req.Description,
		Status:            entity.OfferStatusPending,
		CreatedBy:         claims.UserID,
		UpdatedBy:         claims.UserID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, ecode.ErrDatabase.WithMessage("创建仓储资源失败"))
		return
	}
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
	response.Success(c, dto.PageResponse[dto.WarehouseServiceInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
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
	query := h.db.WithContext(c.Request.Context()).Model(&entity.LogisticsService{})
	query = applyCommonQuery(query, req, "name")
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
	response.Success(c, dto.PageResponse[dto.LogisticsServiceInfo]{List: list, Total: total, Page: req.Page, PageSize: req.PageSize})
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

func (h *Handler) GetWarehouse(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item entity.WarehouseService
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "仓储资源不存在")
		return
	}
	response.Success(c, warehouseToInfo(&item))
}

func (h *Handler) GetLogistics(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item entity.LogisticsService
	if err := h.db.WithContext(c.Request.Context()).First(&item, id).Error; err != nil {
		response.NotFound(c, "物流资源不存在")
		return
	}
	response.Success(c, logisticsToInfo(&item))
}

func applyCommonQuery(query *gorm.DB, req dto.ListRequest, nameColumn string) *gorm.DB {
	if req.Keyword != "" {
		query = query.Where(nameColumn+" LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
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
		PriceRule: item.PriceRule, Images: []string(item.Images), Certificates: []string(item.Certificates),
		Description: item.Description, Status: int(item.Status), AuditRemark: item.AuditRemark,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func logisticsToInfo(item *entity.LogisticsService) dto.LogisticsServiceInfo {
	return dto.LogisticsServiceInfo{
		ID: item.ID, EnterpriseID: item.EnterpriseID, Name: item.Name, RouteStart: item.RouteStart,
		RouteEnd: item.RouteEnd, ServiceScope: item.ServiceScope, VehicleType: item.VehicleType,
		LoadCapacity: item.LoadCapacity, PriceRule: item.PriceRule, Images: []string(item.Images),
		Certificates: []string(item.Certificates), Description: item.Description, Status: int(item.Status),
		AuditRemark: item.AuditRemark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
