package handler

import (
	"strconv"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type EnterprisePermissionHandler struct {
	service *service.EnterprisePermissionService
}

func NewEnterprisePermissionHandler() *EnterprisePermissionHandler {
	return &EnterprisePermissionHandler{service: service.NewEnterprisePermissionService()}
}

func (h *EnterprisePermissionHandler) GetMyServices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.service.GetOverview(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *EnterprisePermissionHandler) Apply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.ApplyEnterpriseServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.service.Apply(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *EnterprisePermissionHandler) ListApplications(c *gin.Context) {
	var req dto.EnterpriseServiceApplicationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.service.ListApplications(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *EnterprisePermissionHandler) GetApplication(c *gin.Context) {
	id, ok := parseApplicationID(c)
	if !ok {
		return
	}

	result, err := h.service.GetApplicationByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *EnterprisePermissionHandler) Audit(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	if operatorID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	id, ok := parseApplicationID(c)
	if !ok {
		return
	}

	var req dto.AuditEnterpriseServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.service.Audit(c.Request.Context(), id, &req, operatorID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "审核完成"})
}

func (h *EnterprisePermissionHandler) UpdatePermissionStatus(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	if operatorID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	enterpriseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的企业ID")
		return
	}

	var req dto.UpdateEnterpriseServicePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, serviceErr := h.service.UpdatePermissionStatus(c.Request.Context(), enterpriseID, c.Param("service"), &req, operatorID)
	if serviceErr != nil {
		response.Error(c, serviceErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func parseApplicationID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("application_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的申请ID")
		return 0, false
	}
	return id, true
}
