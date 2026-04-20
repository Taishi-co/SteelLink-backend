package dto

import "time"

type UserInfo struct {
	ID             uint64    `json:"id"`
	UID            string    `json:"uid"`
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

type UpdateUserRequest struct {
	QQ    string `json:"qq"`
	Email string `json:"email" binding:"omitempty,email"`
}

type UpdatePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required,len=11"`
	Code     string `json:"code" binding:"required,len=6"`
}

type JoinEnterpriseRequest struct {
	EnterpriseCode string `json:"enterprise_code" binding:"required"`
}

type LeaveEnterpriseRequest struct{}

type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Phone    string `form:"phone"`
	Email    string `form:"email"`
	Role     string `form:"role"`
	Status   *int   `form:"status"`
	Keyword  string `form:"keyword"`
}

type UserListResponse struct {
	List     []*UserInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func ToUserInfo(user interface{}) *UserInfo {
	return nil
}
