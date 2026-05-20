package dto

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"omitempty,email"`
	QQ       string `json:"qq" binding:"omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 手机号或邮箱
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type SwitchSessionRequest struct {
	Target string `json:"target" binding:"required,oneof=buyer seller logistics warehouse finance"`
}

type SwitchSessionResponse struct {
	SwitchToken string `json:"switch_token"`
}

type ConsumeSwitchSessionRequest struct {
	SwitchToken string `json:"switch_token" binding:"required"`
	Target      string `json:"target" binding:"required,oneof=buyer seller logistics warehouse finance"`
}

// SendSmsRequest 发送验证码请求
type SendSmsRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Type  string `json:"type" binding:"required,oneof=register login reset"` // register, login, reset
}

// SmsLoginRequest 短信登录请求
type SmsLoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Phone       string `json:"phone" binding:"required,len=11"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=20"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}
