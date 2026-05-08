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
	result, svcErr := h.orderService.GetByID(c.Request.Context(), id, claims.EnterpriseID)
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
	var result []*dto.OrderInfo
	var err error
	switch role {
	case "seller":
		result, err = h.orderService.ListBySeller(c.Request.Context(), claims.EnterpriseID)
	case "buyer":
		result, err = h.orderService.ListByBuyer(c.Request.Context(), claims.EnterpriseID)
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
	result, svcErr := h.orderService.SignContract(c.Request.Context(), id, claims.EnterpriseID, claims.UserID)
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
	result, svcErr := h.orderService.ListEvents(c.Request.Context(), id, claims.EnterpriseID)
	if svcErr != nil {
		response.Error(c, svcErr.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": result})
}
