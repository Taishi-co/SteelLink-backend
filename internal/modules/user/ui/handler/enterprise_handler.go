package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

// EnterpriseHandler 企业处理器
type EnterpriseHandler struct {
	enterpriseService *service.EnterpriseService
}

// NewEnterpriseHandler 创建企业处理器
func NewEnterpriseHandler() *EnterpriseHandler {
	return &EnterpriseHandler{
		enterpriseService: service.NewEnterpriseService(),
	}
}

// Create 创建企业申请
func (h *EnterpriseHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.CreateEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.enterpriseService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// Verify 提交企业认证
func (h *EnterpriseHandler) Verify(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	id, ok := parseUintID(c)
	if !ok {
		return
	}

	var req dto.VerifyEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.enterpriseService.Verify(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// GetByID 获取企业详情
func (h *EnterpriseHandler) GetByID(c *gin.Context) {
	id, ok := parseUintID(c)
	if !ok {
		return
	}

	result, err := h.enterpriseService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// List 获取企业列表
func (h *EnterpriseHandler) List(c *gin.Context) {
	var req dto.EnterpriseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.enterpriseService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// Audit 审核企业
func (h *EnterpriseHandler) Audit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	id, ok := parseUintID(c)
	if !ok {
		return
	}

	var req dto.AuditEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.enterpriseService.Audit(c.Request.Context(), id, &req, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "审核成功"})
}

func parseUintID(c *gin.Context) (uint64, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return 0, false
	}

	return id, true
}
