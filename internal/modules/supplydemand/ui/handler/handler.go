package handler

import (
	"strconv"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SupplyDemandHandler struct {
	svc *service.SupplyDemandService
}

func NewSupplyDemandHandler() *SupplyDemandHandler {
	return &SupplyDemandHandler{svc: service.NewSupplyDemandService()}
}

func (h *SupplyDemandHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.Create(c.Request.Context(), &req, userID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *SupplyDemandHandler) List(c *gin.Context) {
	var req dto.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.List(c.Request.Context(), &req, true)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *SupplyDemandHandler) ListMine(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.ListMine(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *SupplyDemandHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的供求信息ID")
		return
	}

	result, svcErr := h.svc.GetByID(c.Request.Context(), id, true)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *SupplyDemandHandler) Stats(c *gin.Context) {
	result, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *SupplyDemandHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的供求信息ID")
		return
	}

	var req dto.UpdatePostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "供求信息状态更新成功"})
}
