package handler

import (
	"strconv"

	favoriteService "github.com/Taishi-co/SteelLink-backend/internal/modules/favorite"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	service *favoriteService.Service
}

func NewFavoriteHandler() *FavoriteHandler {
	return &FavoriteHandler{service: favoriteService.NewService()}
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	item, err := h.service.Add(c.Request.Context(), claims.UserID, claims.EnterpriseID, req.ProductID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, item)
}

func (h *FavoriteHandler) List(c *gin.Context) {
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

func (h *FavoriteHandler) Count(c *gin.Context) {
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

func (h *FavoriteHandler) Check(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}
	result, svcErr := h.service.Check(c.Request.Context(), claims.UserID, productID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的收藏ID")
		return
	}
	if svcErr := h.service.Remove(c.Request.Context(), claims.UserID, id); svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "已取消收藏"})
}

func (h *FavoriteHandler) RemoveByProduct(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}
	if svcErr := h.service.RemoveByProduct(c.Request.Context(), claims.UserID, productID); svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "已取消收藏"})
}
