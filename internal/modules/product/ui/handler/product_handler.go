package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/product/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
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

// ParseBatchImport 解析卖家上传的商品台账或截图，返回待确认草稿
func (h *ProductHandler) ParseBatchImport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要导入的台账或截图")
		return
	}
	result, svcErr := h.productService.ParseProductImport(c.Request.Context(), file, c.PostForm("hint"))
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

type productImportStreamPayload struct {
	Stage      string `json:"stage"`
	Message    string `json:"message"`
	ElapsedSec int    `json:"elapsed_sec,omitempty"`
	Data       any    `json:"data,omitempty"`
}

// ParseBatchImportStream 以 SSE 方式解析卖家上传的商品台账或截图，避免长时间 AI 识别触发前端/网关超时
func (h *ProductHandler) ParseBatchImportStream(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要导入的台账或截图")
		return
	}
	hint := c.PostForm("hint")

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if !writeProductImportStreamEvent(c, "progress", productImportStreamPayload{
		Stage:   "received",
		Message: "已收到文件，正在准备解析",
	}) {
		return
	}

	type parseResult struct {
		data *dto.ProductImportParseResponse
		err  error
	}
	requestCtx := c.Request.Context()
	resultCh := make(chan parseResult, 1)
	go func() {
		result, svcErr := h.productService.ParseProductImport(requestCtx, file, hint)
		resultCh <- parseResult{data: result, err: svcErr}
	}()

	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	if !writeProductImportStreamEvent(c, "progress", productImportStreamPayload{
		Stage:   "recognizing",
		Message: "AI 正在识别商品字段，请稍候",
	}) {
		return
	}

	for {
		select {
		case <-requestCtx.Done():
			return
		case result := <-resultCh:
			if result.err != nil {
				message := result.err.Error()
				if svcErr, ok := result.err.(*ecode.Error); ok {
					message = svcErr.Message
				}
				writeProductImportStreamEvent(c, "error", productImportStreamPayload{
					Stage:      "failed",
					Message:    message,
					ElapsedSec: int(time.Since(start).Seconds()),
				})
				return
			}
			writeProductImportStreamEvent(c, "done", productImportStreamPayload{
				Stage:      "completed",
				Message:    "识别完成，请确认商品草稿",
				ElapsedSec: int(time.Since(start).Seconds()),
				Data:       result.data,
			})
			return
		case <-ticker.C:
			if !writeProductImportStreamEvent(c, "progress", productImportStreamPayload{
				Stage:      "recognizing",
				Message:    "AI 仍在识别和整理商品草稿",
				ElapsedSec: int(time.Since(start).Seconds()),
			}) {
				return
			}
		}
	}
}

func writeProductImportStreamEvent(c *gin.Context, event string, payload productImportStreamPayload) bool {
	data, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return false
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return true
}

// ConfirmBatchImport 确认批量导入草稿并创建商品
func (h *ProductHandler) ConfirmBatchImport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.ConfirmProductImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.productService.ConfirmProductImport(c.Request.Context(), &req, userID, claims.EnterpriseID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// Update 更新卖家自有商品，提交后重新进入平台审核
func (h *ProductHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, svcErr := h.productService.Update(c.Request.Context(), id, &req, userID, claims.EnterpriseID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
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

// Delete 删除卖家自有商品
func (h *ProductHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if svcErr := h.productService.Delete(c.Request.Context(), id, userID, claims.EnterpriseID); svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "商品删除成功"})
}

// UpdateDisplay 更新商品展示/审核补充资料
func (h *ProductHandler) UpdateDisplay(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var req dto.UpdateProductDisplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, svcErr := h.productService.UpdateDisplay(c.Request.Context(), id, &req, userID, claims.EnterpriseID)
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

	// 前台商城默认只返回已上架商品；卖家/管理场景可通过 enterprise_id 或 status 查询完整状态。
	if req.EnterpriseID == 0 && req.Status == nil {
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

// ListCategories 获取商品品类列表
func (h *ProductHandler) ListCategories(c *gin.Context) {
	var req dto.ProductCategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.productService.ListCategories(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

// ListGrades 获取商品牌号列表
func (h *ProductHandler) ListGrades(c *gin.Context) {
	var req dto.ProductGradeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.productService.ListGrades(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

// ListSteelMills 获取商品钢厂/产地列表
func (h *ProductHandler) ListSteelMills(c *gin.Context) {
	var req dto.ProductSteelMillListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.productService.ListSteelMills(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

// CreateCategory 创建商品品类
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.UpsertProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.productService.CreateCategory(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

// UpdateCategory 更新商品品类
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的品类ID")
		return
	}
	var req dto.UpsertProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.productService.UpdateCategory(c.Request.Context(), id, &req, userID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

// DeleteCategory 删除商品品类
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的品类ID")
		return
	}
	if err := h.productService.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "商品品类删除成功"})
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

	auditRemark := strings.TrimSpace(req.AuditRemark)
	if auditRemark == "" {
		auditRemark = strings.TrimSpace(req.AuditRemarkCamel)
	}
	if err := h.productService.UpdateStatus(c.Request.Context(), id, req.Status, auditRemark, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "商品状态更新成功"})
}
