package handler

import (
	"strconv"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{orderService: service.NewOrderService()}
}

func (h *OrderHandler) PreviewWarehousePreData(c *gin.Context) {
	var req dto.DraftOrderPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.orderService.PreviewWarehousePreData(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ListProductPool(c *gin.Context) {
	result, err := h.orderService.ListProductPool(c.Request.Context())
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	if userID == 0 || claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.orderService.Create(c.Request.Context(), &req, userID, claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.GetByID(c.Request.Context(), id, claims.EnterpriseID, isOrderAdmin(claims.UserType))
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) List(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	role := c.Query("role")
	statusValue := 0
	var status *int
	if rawStatus := c.Query("status"); rawStatus != "" {
		parsed, parseErr := strconv.Atoi(rawStatus)
		if parseErr != nil {
			response.BadRequest(c, "无效的订单状态")
			return
		}
		statusValue = parsed
		status = &statusValue
	}
	if role == "" && isOrderAdmin(claims.UserType) {
		page := intQuery(c, "page", 1)
		pageSize := intQuery(c, "page_size", 0)
		if pageSize == 0 {
			pageSize = intQuery(c, "pageSize", 10)
		}
		result, err := h.orderService.ListAll(c.Request.Context(), c.Query("keyword"), status, page, pageSize)
		if err != nil {
			response.Error(c, err.(*ecode.Error))
			return
		}
		response.Success(c, result)
		return
	}
	var result []*dto.OrderInfo
	var err error
	switch role {
	case "seller":
		result, err = h.orderService.ListBySeller(c.Request.Context(), claims.EnterpriseID)
	case "buyer":
		result, err = h.orderService.ListByBuyer(c.Request.Context(), claims.EnterpriseID)
	case "warehouse":
		result, err = h.orderService.ListByWarehouse(c.Request.Context(), claims.EnterpriseID, status)
	case "logistics":
		result, err = h.orderService.ListByLogistics(c.Request.Context(), claims.EnterpriseID, status)
	default:
		result, err = h.orderService.ListByEnterprise(c.Request.Context(), claims.EnterpriseID)
	}
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{
		"list": result,
	})
}

func (h *OrderHandler) ListFundingFinancingRequests(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	result, err := h.orderService.ListFundingFinancingRequests(c.Request.Context(), claims.EnterpriseID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.Cancel(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ConfirmOrder(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) WarehouseConfirm(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.WarehouseConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.WarehouseConfirm(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ReselectWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.WarehouseReselectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ReselectWarehouse(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) SelectDestWarehouse(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.DestWarehouseSelectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.SelectDestWarehouse(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) SellerConfirm(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.SellerConfirmRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.SellerConfirm(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) BuyerConfirm(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.BuyerConfirmRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.BuyerConfirm(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) SignContract(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.SignContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, c.GetHeader("X-Trade-Role"))
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) AttachContract(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.AttachContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.AttachContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) UploadPaymentVoucher(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.UploadPaymentVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.UploadPaymentVoucher(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) DirectPayment(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.DirectPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.DirectPayment(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) CreateFinancingRequest(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.CreateFinancingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.CreateFinancingRequest(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) SubmitRiskResult(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.RiskResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.SubmitRiskResult(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) BuyerConfirmFinancing(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.BuyerFinancingConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.BuyerConfirmFinancing(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) GenerateDebtContract(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.DebtContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.GenerateDebtContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) BuyerSignDebtAndPayFee(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.BuyerDebtSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.BuyerSignDebtAndPayFee(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) DisburseFinancing(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.FinancingDisburseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.DisburseFinancing(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) RecordPledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.PledgeRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.RecordPledge(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ReleasePledge(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.PledgeReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ReleasePledge(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ConfirmPayment(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ConfirmPayment(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) SetDeliveryInfo(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.SetDeliveryInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.SetDeliveryInfo(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ConfirmOutbound(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.VerifyPickupCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ConfirmOutbound(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ConfirmReceipt(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.ConfirmReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ConfirmReceipt(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) CreateLogisticsBidding(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.CreateLogisticsBiddingRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.CreateLogisticsBidding(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ReportLogisticsNoCapacity(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.CreateLogisticsBiddingRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.ReportLogisticsNoCapacity(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) CreateLogisticsQuote(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.CreateLogisticsQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.CreateLogisticsQuote(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) CloseLogisticsBidding(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.CloseLogisticsBidding(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ListLogisticsQuotes(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ListLogisticsQuotes(c.Request.Context(), id, claims.EnterpriseID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}

func (h *OrderHandler) SelectLogisticsQuote(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.ConfirmLogisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.SelectLogisticsQuote(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ConfirmLogisticsQuote(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.ConfirmLogisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ConfirmLogisticsQuote(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ShipmentProgress(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.ShipmentProgressRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.UpdateShipmentProgress(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, c.Param("action"), &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) GetLogisticsContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.GetLogisticsContract(c.Request.Context(), id)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) BuyerSignLogisticsContract(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.SignLogisticsContractRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.BuyerSignLogisticsContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) LogisticsSignLogisticsContract(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.SignLogisticsContractRequest
	_ = c.ShouldBindJSON(&req)
	result, svcErr := h.orderService.LogisticsSignLogisticsContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) GetLogisticsSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.GetLogisticsSettlement(c.Request.Context(), id)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) PayLogisticsSettlement(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.PayLogisticsSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.PayLogisticsSettlement(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) PayLogisticsCommission(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.PayLogisticsSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.PayLogisticsCommission(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) Reconcile(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.Reconcile(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) Archive(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.Archive(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ListEvents(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ListEvents(c.Request.Context(), id, claims.EnterpriseID, isOrderAdmin(claims.UserType))
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}

func isOrderAdmin(userType string) bool {
	return userType == "operator" || userType == "super_admin"
}

func (h *OrderHandler) ReportBreach(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.ReportBreachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ReportBreach(c.Request.Context(), id, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ListBreachRecords(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ListBreachRecords(c.Request.Context(), id)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}

func (h *OrderHandler) MediateBreach(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	breachID, err := strconv.ParseUint(c.Param("breach_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的违约记录ID")
		return
	}
	var req dto.MediateBreachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.MediateBreach(c.Request.Context(), breachID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ResolveBreach(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	breachID, err := strconv.ParseUint(c.Param("breach_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的违约记录ID")
		return
	}
	var req dto.ResolveBreachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.ResolveBreach(c.Request.Context(), breachID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) FreezeOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.FreezeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.FreezeOrder(c.Request.Context(), id, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) UnfreezeOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req dto.UnfreezeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.UnfreezeOrder(c.Request.Context(), id, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) CreateLitigation(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	breachID, err := strconv.ParseUint(c.Param("breach_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的违约记录ID")
		return
	}
	var req dto.CreateLitigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.CreateLitigation(c.Request.Context(), breachID, claims.EnterpriseID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) UpdateLitigationStatus(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	litigationID, err := strconv.ParseUint(c.Param("litigation_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的诉讼记录ID")
		return
	}
	var req dto.UpdateLitigationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, svcErr := h.orderService.UpdateLitigationStatus(c.Request.Context(), litigationID, claims.UserID, &req)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, result)
}

func (h *OrderHandler) ListLitigationRecords(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	result, svcErr := h.orderService.ListLitigationRecords(c.Request.Context(), id)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}

func intQuery(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
