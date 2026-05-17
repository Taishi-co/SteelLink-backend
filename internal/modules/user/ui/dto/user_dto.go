package dto

import "time"

type UserInfo struct {
	ID                  uint64    `json:"id"`
	UID                 string    `json:"uid"`
	UserNo              string    `json:"user_no"`
	PlatformID          string    `json:"platform_id"`
	PersonalCode        string    `json:"personal_code"`
	EnterpriseCode      string    `json:"enterprise_code"`
	EnterpriseRole      string    `json:"enterprise_role"`
	ServiceCapabilities []string  `json:"service_capabilities"`
	Role                string    `json:"role"`
	Phone               string    `json:"phone"`
	QQ                  string    `json:"qq"`
	Email               string    `json:"email"`
	RealName            string    `json:"real_name"`
	IDCard              string    `json:"id_card"`
	IdentityStatus      int       `json:"identity_status"`
	IdentityReject      string    `json:"identity_reject"`
	Status              int       `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type SubmitIdentityRequest struct {
	RealName       string `json:"real_name" binding:"required"`
	IDCard         string `json:"id_card" binding:"required,len=18"`
	IDCardFrontURL string `json:"id_card_front_url" binding:"required,url"`
	IDCardBackURL  string `json:"id_card_back_url" binding:"required,url"`
}

type IdentityStatusResponse struct {
	IdentityStatus int    `json:"identity_status"`
	RealName       string `json:"real_name"`
	IDCard         string `json:"id_card"`
	IdentityReject string `json:"identity_reject"`
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

type CreateSubAccountRequest struct {
	Phone       string   `json:"phone" binding:"required,len=11"`
	Password    string   `json:"password" binding:"required,min=6"`
	RealName    string   `json:"real_name" binding:"required"`
	QQ          string   `json:"qq"`
	Email       string   `json:"email" binding:"omitempty,email"`
	Permissions []string `json:"permissions"`
}

type UpdateSubAccountRequest struct {
	RealName    string   `json:"real_name"`
	QQ          string   `json:"qq"`
	Email       string   `json:"email" binding:"omitempty,email"`
	Permissions []string `json:"permissions"`
	Status      *int     `json:"status"`
}

type SubAccountInfo struct {
	ID          uint64    `json:"id"`
	UserNo      string    `json:"user_no"`
	Phone       string    `json:"phone"`
	RealName    string    `json:"real_name"`
	Email       string    `json:"email"`
	QQ          string    `json:"qq"`
	Permissions []string  `json:"permissions"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type DashboardStats struct {
	InTransitAmount       float64 `json:"in_transit_amount"`
	PendingPaymentCount   int64   `json:"pending_payment_count"`
	MonthlyPurchaseAmount float64 `json:"monthly_purchase_amount"`
	StockQuantity         float64 `json:"stock_quantity"`
}

type RecentOrder struct {
	ID          uint64  `json:"id"`
	OrderNo     string  `json:"order_no"`
	Status      int     `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt   string  `json:"created_at"`
	ProductName string  `json:"product_name"`
}

type DashboardResponse struct {
	Stats        DashboardStats `json:"stats"`
	RecentOrders []RecentOrder  `json:"recent_orders"`
}

type UpsertPersonalInquiryDraftRequest struct {
	Title           string  `json:"title" binding:"required"`
	ProductName     string  `json:"product_name"`
	Category        string  `json:"category"`
	Material        string  `json:"material"`
	Spec            string  `json:"spec"`
	Quantity        float64 `json:"quantity"`
	Unit            string  `json:"unit"`
	TargetPrice     float64 `json:"target_price"`
	DeliveryAddress string  `json:"delivery_address"`
	Requirement     string  `json:"requirement"`
}

type PersonalInquiryDraftInfo struct {
	ID              uint64    `json:"id"`
	DraftNo         string    `json:"draft_no"`
	UserID          uint64    `json:"user_id"`
	Title           string    `json:"title"`
	ProductName     string    `json:"product_name"`
	Category        string    `json:"category"`
	Material        string    `json:"material"`
	Spec            string    `json:"spec"`
	Quantity        float64   `json:"quantity"`
	Unit            string    `json:"unit"`
	TargetPrice     float64   `json:"target_price"`
	DeliveryAddress string    `json:"delivery_address"`
	Requirement     string    `json:"requirement"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PersonalInquiryDraftListResponse struct {
	List []PersonalInquiryDraftInfo `json:"list"`
}

func ToUserInfo(user interface{}) *UserInfo {
	return nil
}
