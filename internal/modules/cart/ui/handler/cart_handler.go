package handler

import (
	"strconv"

	cartService "github.com/Taishi-co/SteelLink-backend/internal/modules/cart"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	service *cartService.Service
}

func NewCartHandler() *CartHandler {
	return &CartHandler{service: cartService.NewService()}
}

func (h *CartHandler) Add(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item, err := h.service.AddOrIncrease(c.Request.Context(), claims.UserID, claims.EnterpriseID, req.ProductID, req.Quantity)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, item)
}

func (h *CartHandler) List(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	result, err := h.service.List(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *CartHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的购物车商品ID")
		return
	}
	var req dto.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item, svcErr := h.service.Update(c.Request.Context(), claims.UserID, id, req.Quantity)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, item)
}

func (h *CartHandler) Remove(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的购物车商品ID")
		return
	}
	if svcErr := h.service.Remove(c.Request.Context(), claims.UserID, id); svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "已移除"})
}

func (h *CartHandler) Clear(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	if err := h.service.Clear(c.Request.Context(), claims.UserID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "已清空"})
}

func (h *CartHandler) Count(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	result, err := h.service.Count(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}
