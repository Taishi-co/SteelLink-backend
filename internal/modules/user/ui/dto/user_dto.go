package dto

import "time"

// UserInfo 用户信息响应
type UserInfo struct {
	ID             uint64    `json:"id"`
	UserNo         string    `json:"user_no"`
	PlatformID     string    `json:"platform_id"`
	PersonalCode   string    `json:"personal_code"`
	EnterpriseCode string    `json:"enterprise_code"`
	Role           string    `json:"role"`
	Phone          string    `json:"phone"`
	QQ             string    `json:"qq"`
	Email          string    `json:"email"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	QQ    string `json:"qq"`
	Email string `json:"email" binding:"omitempty,email"`
}

// UpdatePhoneRequest 更新手机号请求
type UpdatePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required,len=11"`
	Code     string `json:"code" binding:"required,len=6"`
}

// JoinEnterpriseRequest 加入企业请求
type JoinEnterpriseRequest struct {
	EnterpriseCode string `json:"enterprise_code" binding:"required"`
}

// LeaveEnterpriseRequest 离开企业请求
type LeaveEnterpriseRequest struct {
	// 空请求，用户离开当前企业
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Phone    string `form:"phone"`
	Email    string `form:"email"`
	Role     string `form:"role"`
	Status   *int   `form:"status"`
	Keyword  string `form:"keyword"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	List     []*UserInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ToUserInfo 将用户实体转换为DTO
func ToUserInfo(user interface{}) *UserInfo {
	// 由具体实现完成
	return nil
}
