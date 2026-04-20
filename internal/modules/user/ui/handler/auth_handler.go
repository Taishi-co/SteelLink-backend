package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/application/service"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 通过手机号注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// Login 用户登录
// @Summary 用户登录
// @Description 通过手机号/邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（客户端清除Token即可）
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端清除Token即可
	// 如果需要服务端失效，可以将Token加入黑名单
	response.Success(c, nil)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 使用RefreshToken获取新的AccessToken
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// SendSms 发送验证码
// @Summary 发送短信验证码
// @Description 发送短信验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.SendSmsRequest true "发送验证码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/sms/send [post]
func (h *AuthHandler) SendSms(c *gin.Context) {
	var req dto.SendSmsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.authService.SendSms(c.Request.Context(), &req); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "验证码已发送"})
}

// SmsLogin 短信登录
// @Summary 短信验证码登录
// @Description 通过手机号和验证码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.SmsLoginRequest true "短信登录请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/sms/login [post]
func (h *AuthHandler) SmsLogin(c *gin.Context) {
	var req dto.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.SmsLogin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, result)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 通过手机验证码重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/password/reset [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 用户修改密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/password/change [put]
// @Security BearerAuth
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userService := service.NewUserService()
	if err := userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}
