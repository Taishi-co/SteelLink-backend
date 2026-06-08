package handler

import (
	"strconv"

	marketservice "github.com/Taishi-co/SteelLink-backend/internal/modules/market/application/service"
	marketdto "github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	service *marketservice.MarketService
}

func NewMarketHandler() *MarketHandler {
	return &MarketHandler{service: marketservice.NewMarketService()}
}

func (h *MarketHandler) GetBiddingStats(c *gin.Context) {
	stats, err := h.service.GetBiddingStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *MarketHandler) ListBiddings(c *gin.Context) {
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	var enterpriseID uint64
	if claims := middleware.GetClaims(c); claims != nil {
		enterpriseID = claims.EnterpriseID
	}
	result, err := h.service.ListBiddings(c.Request.Context(), &req, enterpriseID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) GetBidding(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var enterpriseID uint64
	if claims := middleware.GetClaims(c); claims != nil {
		enterpriseID = claims.EnterpriseID
	}
	result, err := h.service.GetBidding(c.Request.Context(), id, enterpriseID)
	if err != nil {
		if marketservice.IsNotFound(err) {
			response.NotFound(c, "bidding activity not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) CreateBiddingOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.CreateBiddingOrder(c.Request.Context(), claims.EnterpriseID, claims.UserID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) CreateBidding(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.CreateBiddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.CreateBidding(c.Request.Context(), claims.UserID, claims.EnterpriseID, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) SubmitBiddingQuote(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.SubmitBiddingQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.SubmitBiddingQuote(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) AwardBidding(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.AwardBiddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.AwardBidding(c.Request.Context(), claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) GetProcurementStats(c *gin.Context) {
	stats, err := h.service.GetProcurementStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *MarketHandler) ListProcurements(c *gin.Context) {
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListProcurements(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) GetProcurement(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetProcurement(c.Request.Context(), id)
	if err != nil {
		if marketservice.IsNotFound(err) {
			response.NotFound(c, "procurement activity not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	if result.Status == "pending_seller" || result.Status == "seller_rejected" {
		response.NotFound(c, "procurement activity not found")
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListSellerProcurements(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListSellerProcurements(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListMyJoinedProcurements(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListMyJoinedProcurements(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) RespondProcurement(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.RespondProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.RespondProcurement(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) JoinProcurement(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.JoinProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.JoinProcurement(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) CreateProcurementRequest(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.CreateProcurementRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.CreateProcurementRequest(c.Request.Context(), claims.UserID, claims.EnterpriseID, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListMyProcurementRequests(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListMyProcurementRequests(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListMyProcurementInvitations(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListMyProcurementInvitations(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) AcceptProcurementInvitation(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.JoinProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.AcceptProcurementInvitation(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListProcurementProjects(c *gin.Context) {
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	var enterpriseID uint64
	if claims := middleware.GetClaims(c); claims != nil {
		enterpriseID = claims.EnterpriseID
	}
	result, err := h.service.ListProcurementProjects(c.Request.Context(), &req, enterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListMyProcurementProjects(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListMyProcurementProjects(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) ListSellerProcurementProjectsV2(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req marketdto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}
	result, err := h.service.ListSellerProcurementProjectsV2(c.Request.Context(), &req, claims.EnterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) GetProcurementProject(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var enterpriseID uint64
	if claims := middleware.GetClaims(c); claims != nil {
		enterpriseID = claims.EnterpriseID
	}
	result, err := h.service.GetProcurementProject(c.Request.Context(), id, enterpriseID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) CreateProcurementProjectEnrollment(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.CreateProcurementProjectEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.CreateProcurementProjectEnrollment(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *MarketHandler) SubmitProcurementProjectQuote(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req marketdto.CreateProcurementProjectQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.service.SubmitProcurementProjectQuote(c.Request.Context(), claims.UserID, claims.EnterpriseID, id, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, result)
}

func parseUintParam(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func writeServiceError(c *gin.Context, err error) {
	if e, ok := err.(*ecode.Error); ok {
		response.Error(c, e)
		return
	}
	response.InternalError(c, err.Error())
}
