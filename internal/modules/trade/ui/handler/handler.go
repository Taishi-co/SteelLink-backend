package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	tradeservice "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/application/service"
	tradedto "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

type TradeHandler struct {
	service *tradeservice.TradeService
}

func NewTradeHandler() *TradeHandler {
	return &TradeHandler{service: tradeservice.NewTradeService()}
}

func (h *TradeHandler) ListProducts(c *gin.Context) {
	var req tradedto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	result, err := h.service.ListPublicProducts(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *TradeHandler) GetProduct(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	product, err := h.service.GetPublicProduct(c.Request.Context(), id)
	if err != nil {
		if tradeservice.IsNotFound(err) {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, product)
}

func (h *TradeHandler) ListSellerProducts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req tradedto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	result, err := h.service.ListSellerProducts(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *TradeHandler) CreateSellerProduct(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req tradedto.UpsertProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	product, err := h.service.SaveSellerProduct(c.Request.Context(), userID, &req, nil)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, product)
}

func (h *TradeHandler) UpdateSellerProduct(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req tradedto.UpsertProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	product, err := h.service.SaveSellerProduct(c.Request.Context(), userID, &req, &id)
	if err != nil {
		if tradeservice.IsNotFound(err) {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, product)
}

func (h *TradeHandler) ListAdminProducts(c *gin.Context) {
	var req tradedto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	result, err := h.service.ListAdminProducts(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *TradeHandler) AuditProduct(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req tradedto.AuditProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	product, err := h.service.AuditProduct(c.Request.Context(), userID, id, &req)
	if err != nil {
		if tradeservice.IsNotFound(err) {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, product)
}

func (h *TradeHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req tradedto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		if tradeservice.IsNotFound(err) {
			response.NotFound(c, "product not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *TradeHandler) ListOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req tradedto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query")
		return
	}

	result, err := h.service.ListOrders(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *TradeHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), userID, id)
	if err != nil {
		if tradeservice.IsNotFound(err) {
			response.NotFound(c, "order not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, order)
}

func parseUintParam(c *gin.Context, name string) (uint64, bool) {
	value := c.Param(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return 0, false
	}
	return id, true
}
