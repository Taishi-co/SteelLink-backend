package dto

import "time"

type ProductListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Status   string `form:"status"`
}

type UpsertProductRequest struct {
	Name              string   `json:"name" binding:"required"`
	Category          string   `json:"category" binding:"required"`
	Specs             string   `json:"specs" binding:"required"`
	Price             float64  `json:"price" binding:"required,gt=0"`
	Stock             int64    `json:"stock" binding:"required,gt=0"`
	Description       string   `json:"description" binding:"required"`
	ImageURLs         []string `json:"image_urls" binding:"required,min=1"`
	QualificationURLs []string `json:"qualification_urls" binding:"required,min=1"`
}

type AuditProductRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Remark string `json:"remark"`
}

type CreateOrderRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,gt=0"`
}

type SubmitContractRequest struct {
	ContractURL  string `json:"contract_url" binding:"required,url"`
	ContractHash string `json:"contract_hash" binding:"required"`
}

type SubmitPaymentReceiptRequest struct {
	ReceiptURL string  `json:"receipt_url" binding:"required,url"`
	VoucherNo  string  `json:"voucher_no" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

type AuditPaymentRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Remark string `json:"remark"`
}

type VerifyPickupCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

type ProductAuditLogItem struct {
	ID           uint64    `json:"id"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductItem struct {
	ID                uint64                `json:"id"`
	SKUID             string                `json:"sku_id"`
	SellerID          uint64                `json:"seller_id"`
	SellerUID         string                `json:"seller_uid"`
	SellerName        string                `json:"seller_name"`
	Name              string                `json:"name"`
	Category          string                `json:"category"`
	Specs             string                `json:"specs"`
	Price             float64               `json:"price"`
	Unit              string                `json:"unit"`
	Stock             int64                 `json:"stock"`
	VisibleStock      int64                 `json:"visible_stock"`
	Sales             int64                 `json:"sales"`
	Description       string                `json:"description"`
	ImageURLs         []string              `json:"image_urls"`
	QualificationURLs []string              `json:"qualification_urls,omitempty"`
	Status            string                `json:"status"`
	AuditRemark       string                `json:"audit_remark,omitempty"`
	SubmittedAt       *time.Time            `json:"submitted_at,omitempty"`
	ApprovedAt        *time.Time            `json:"approved_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	AuditLogs         []ProductAuditLogItem `json:"audit_logs,omitempty"`
}

type ProductListResponse struct {
	List     []ProductItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type OrderItem struct {
	ID              uint64  `json:"id"`
	ProductID       uint64  `json:"product_id"`
	ProductSKUID    string  `json:"product_sku_id"`
	ProductName     string  `json:"product_name"`
	ProductCategory string  `json:"product_category"`
	Specs           string  `json:"specs"`
	Price           float64 `json:"price"`
	Quantity        int64   `json:"quantity"`
	Amount          float64 `json:"amount"`
}

type OrderInfo struct {
	ID             uint64              `json:"id"`
	OrderID        string              `json:"order_id"`
	BuyerID        uint64              `json:"buyer_id"`
	BuyerUID       string              `json:"buyer_uid"`
	SellerID       uint64              `json:"seller_id"`
	SellerUID      string              `json:"seller_uid"`
	ProductID      uint64              `json:"product_id"`
	ProductSKUID   string              `json:"product_sku_id"`
	Status         string              `json:"status"`
	TotalAmount    float64             `json:"total_amount"`
	TotalQuantity  int64               `json:"total_quantity"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Items          []OrderItem         `json:"items"`
	Contract       *OrderContractInfo  `json:"contract,omitempty"`
	PaymentReceipt *PaymentReceiptInfo `json:"payment_receipt,omitempty"`
	PickupCode     *PickupCodeInfo     `json:"pickup_code,omitempty"`
}

type OrderListResponse struct {
	List     []OrderInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type OrderListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

type OrderContractInfo struct {
	ID           uint64    `json:"id"`
	ContractURL  string    `json:"contract_url"`
	ContractHash string    `json:"contract_hash"`
	Status       string    `json:"status"`
	SubmittedBy  uint64    `json:"submitted_by"`
	SubmittedAt  time.Time `json:"submitted_at"`
	SignedBy     *uint64   `json:"signed_by,omitempty"`
	SignedAt     *time.Time `json:"signed_at,omitempty"`
}

type PaymentReceiptInfo struct {
	ID          uint64     `json:"id"`
	ReceiptURL  string     `json:"receipt_url"`
	VoucherNo   string     `json:"voucher_no"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	AuditRemark string     `json:"audit_remark,omitempty"`
	SubmittedBy uint64     `json:"submitted_by"`
	SubmittedAt time.Time  `json:"submitted_at"`
	AuditorID   *uint64    `json:"auditor_id,omitempty"`
	AuditedAt   *time.Time `json:"audited_at,omitempty"`
}

type PickupCodeInfo struct {
	ID          uint64     `json:"id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	GeneratedBy uint64     `json:"generated_by"`
	GeneratedAt time.Time  `json:"generated_at"`
	VerifiedBy  *uint64    `json:"verified_by,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
}
