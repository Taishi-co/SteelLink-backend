package handler

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ProcurementHandler struct {
	svc *service.ProcurementService
}

func NewProcurementHandler() *ProcurementHandler {
	return &ProcurementHandler{svc: service.NewProcurementService()}
}

func (h *ProcurementHandler) Parse(c *gin.Context) {
	var req dto.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Parse(c.Request.Context(), req.Prompt)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) ConfirmInstruction(c *gin.Context) {
	var req dto.ConfirmInstructionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.ConfirmInstruction(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) MatchProducts(c *gin.Context) {
	var req dto.MatchProductsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.MatchProducts(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) MatchLogisticsWarehouse(c *gin.Context) {
	var req dto.MatchLogisticsWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.MatchLogisticsWarehouse(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) GenerateOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.GenerateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.GenerateOrder(c.Request.Context(), &req, claims.UserID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.CreateProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Create(c.Request.Context(), &req, claims.UserID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) ListMine(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.ProcurementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.ListMine(c.Request.Context(), &req, claims.UserID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) GetByID(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的AI采购单ID")
		return
	}
	result, svcErr := h.svc.GetByID(c.Request.Context(), id, claims.UserID, claims.EnterpriseID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) SubmitReview(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的AI采购单ID")
		return
	}
	var req dto.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.svc.SubmitReviewWithOptions(c.Request.Context(), id, claims.UserID, claims.EnterpriseID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) MatchFinance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的AI采购单ID")
		return
	}
	var req dto.MatchFinanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.svc.MatchFinance(c.Request.Context(), id, claims.UserID, claims.EnterpriseID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) SyncVectors(c *gin.Context) {
	if !validateVectorSyncToken(c) {
		response.Unauthorized(c, "无效的向量同步令牌")
		return
	}
	var req dto.VectorSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.SyncVectors(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *ProcurementHandler) VectorSyncStatus(c *gin.Context) {
	if !validateVectorSyncToken(c) {
		response.Unauthorized(c, "无效的向量同步令牌")
		return
	}
	result := h.svc.GetVectorSyncStatus(c.Request.Context(), c.Query("target"))
	response.Success(c, result)
}

func validateVectorSyncToken(c *gin.Context) bool {
	expected := os.Getenv("AI_PROCUREMENT_VECTOR_SYNC_TOKEN")
	if expected == "" {
		expected = os.Getenv("AI_PROCUREMENT_CONFIG_SYNC_TOKEN")
	}
	if expected == "" {
		return false
	}
	token := strings.TrimSpace(c.GetHeader("X-AI-Procurement-Sync-Token"))
	if token == "" {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	return token != "" && token == expected
}
