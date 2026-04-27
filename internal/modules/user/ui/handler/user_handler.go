package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /api/v1/users/me [get]
// @Security BearerAuth
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, user)
}

// UpdateCurrentUser 更新当前用户信息
// @Summary 更新当前用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body dto.UpdateUserRequest true "更新请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users/me [put]
// @Security BearerAuth
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.Update(c.Request.Context(), userID, &req, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// GetUserByID 根据ID获取用户信息
// @Summary 获取用户信息
// @Description 根据用户ID获取用户信息
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /api/v1/users/{id} [get]
// @Security BearerAuth
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, user)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（管理员）
// @Tags 用户
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param phone query string false "手机号"
// @Param email query string false "邮箱"
// @Param role query string false "角色"
// @Param status query int false "状态"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=dto.UserListResponse}
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) GetDashboard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.userService.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) CreateSubAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.CreateSubAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.CreateSubAccount(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) ListSubAccounts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.userService.ListSubAccounts(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新用户状态（管理员）
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body map[string]int true "状态请求 {\"status\": 1}"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/status [put]
// @Security BearerAuth
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	if operatorID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1 2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.UpdateStatus(c.Request.Context(), id, req.Status, operatorID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

// JoinEnterprise 加入企业
// @Summary 加入企业
// @Description 用户加入企业
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body dto.JoinEnterpriseRequest true "加入企业请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users/me/enterprise/join [post]
// @Security BearerAuth
func (h *UserHandler) JoinEnterprise(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.JoinEnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.JoinEnterprise(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "加入企业成功"})
}

// LeaveEnterprise 离开企业
// @Summary 离开企业
// @Description 用户离开当前企业
// @Tags 用户
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/users/me/enterprise/leave [post]
// @Security BearerAuth
func (h *UserHandler) LeaveEnterprise(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.userService.LeaveEnterprise(c.Request.Context(), userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "已离开企业"})
}

func (h *UserHandler) SubmitIdentity(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.SubmitIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.SubmitIdentityVerification(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "实名认证已提交，请等待审核"})
}

func (h *UserHandler) GetIdentityStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.userService.GetIdentityStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}
