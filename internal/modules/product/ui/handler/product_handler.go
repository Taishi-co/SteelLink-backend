package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

// ProductHandler 商品处理器
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler 创建商品处理器
func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		productService: service.NewProductService(),
	}
}

// Create 创建商品
func (h *ProductHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.productService.Create(c.Request.Context(), &req, userID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// GetByID 获取商品详情
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var result *dto.ProductInfo
	var svcErr error
	if middleware.GetClaims(c) == nil {
		result, svcErr = h.productService.GetPublicByID(c.Request.Context(), id)
	} else {
		result, svcErr = h.productService.GetByID(c.Request.Context(), id)
	}
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// List 获取商品列表
func (h *ProductHandler) List(c *gin.Context) {
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 游客态默认只返回已上架商品，确保商城展示和审核状态联动。
	if middleware.GetClaims(c) == nil && req.Status == nil {
		defaultStatus := 2
		req.Status = &defaultStatus
	}

	result, err := h.productService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// UpdateStatus 更新商品状态
func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var req dto.UpdateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.productService.UpdateStatus(c.Request.Context(), id, req.Status, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "商品状态更新成功"})
}
