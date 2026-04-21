package entity

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPendingContract     OrderStatus = "pending_contract"
	OrderStatusPendingPayment      OrderStatus = "pending_payment"
	OrderStatusPendingPaymentAudit OrderStatus = "pending_payment_audit"
	OrderStatusPendingPickup       OrderStatus = "pending_pickup"
	OrderStatusCompleted           OrderStatus = "completed"
	OrderStatusCancelled           OrderStatus = "cancelled"
)

type OrderContractStatus string

const (
	OrderContractStatusSubmitted OrderContractStatus = "submitted"
)

type PaymentReceiptStatus string

const (
	PaymentReceiptStatusPending  PaymentReceiptStatus = "pending"
	PaymentReceiptStatusApproved PaymentReceiptStatus = "approved"
	PaymentReceiptStatusRejected PaymentReceiptStatus = "rejected"
)

type PickupCodeStatus string

const (
	PickupCodeStatusActive   PickupCodeStatus = "active"
	PickupCodeStatusVerified PickupCodeStatus = "verified"
)

type Order struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        string          `gorm:"column:order_id;type:varchar(40);uniqueIndex;comment:订单号" json:"order_id"`
	BuyerID        uint64          `gorm:"index;comment:买家ID" json:"buyer_id"`
	BuyerUID       string          `gorm:"type:varchar(32);index;comment:买家UID" json:"buyer_uid"`
	SellerID       uint64          `gorm:"index;comment:卖家ID" json:"seller_id"`
	SellerUID      string          `gorm:"type:varchar(32);index;comment:卖家UID" json:"seller_uid"`
	ProductID      uint64          `gorm:"index;comment:商品ID" json:"product_id"`
	ProductSKUID   string          `gorm:"column:product_sku_id;type:varchar(40);index;comment:商品SKU" json:"product_sku_id"`
	Status         OrderStatus     `gorm:"type:varchar(32);index;comment:订单状态" json:"status"`
	TotalAmount    float64         `gorm:"type:decimal(14,2);comment:订单金额" json:"total_amount"`
	TotalQuantity  int64           `gorm:"comment:总吨数" json:"total_quantity"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
	Items          []OrderItem     `gorm:"foreignKey:OrderRefID;references:ID" json:"items"`
	Contract       *OrderContract  `gorm:"foreignKey:OrderRefID;references:ID" json:"contract,omitempty"`
	PaymentReceipt *PaymentReceipt `gorm:"foreignKey:OrderRefID;references:ID" json:"payment_receipt,omitempty"`
	PickupCode     *PickupCode     `gorm:"foreignKey:OrderRefID;references:ID" json:"pickup_code,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderID == "" {
		o.OrderID = "ORD" + time.Now().Format("20060102150405") + randomDigits(6)
	}
	if o.Status == "" {
		o.Status = OrderStatusPendingContract
	}
	return nil
}

type OrderItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderRefID      uint64    `gorm:"column:order_ref_id;index;comment:订单自增ID" json:"order_ref_id"`
	ProductID       uint64    `gorm:"index;comment:商品ID" json:"product_id"`
	ProductSKUID    string    `gorm:"column:product_sku_id;type:varchar(40);index;comment:商品SKU" json:"product_sku_id"`
	ProductName     string    `gorm:"type:varchar(200);comment:商品名快照" json:"product_name"`
	ProductCategory string    `gorm:"type:varchar(100);comment:品类快照" json:"product_category"`
	Specs           string    `gorm:"type:varchar(100);comment:规格快照" json:"specs"`
	Price           float64   `gorm:"type:decimal(12,2);comment:下单单价" json:"price"`
	Quantity        int64     `gorm:"comment:下单吨数" json:"quantity"`
	Amount          float64   `gorm:"type:decimal(14,2);comment:小计" json:"amount"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type OrderContract struct {
	ID           uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderRefID   uint64              `gorm:"column:order_ref_id;uniqueIndex;comment:订单自增ID" json:"order_ref_id"`
	OrderID      string              `gorm:"column:order_id;type:varchar(40);index;comment:订单号" json:"order_id"`
	BuyerID      uint64              `gorm:"index;comment:买家ID" json:"buyer_id"`
	ContractURL  string              `gorm:"type:varchar(500);comment:合同地址" json:"contract_url"`
	ContractHash string              `gorm:"type:varchar(128);comment:合同哈希" json:"contract_hash"`
	Status       OrderContractStatus `gorm:"type:varchar(32);comment:合同状态" json:"status"`
	SubmittedBy  uint64              `gorm:"comment:提交人ID" json:"submitted_by"`
	SubmittedAt  time.Time           `gorm:"comment:提交时间" json:"submitted_at"`
	CreatedAt    time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderContract) TableName() string {
	return "order_contracts"
}

type PaymentReceipt struct {
	ID          uint64               `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderRefID  uint64               `gorm:"column:order_ref_id;uniqueIndex;comment:订单自增ID" json:"order_ref_id"`
	OrderID     string               `gorm:"column:order_id;type:varchar(40);index;comment:订单号" json:"order_id"`
	BuyerID     uint64               `gorm:"index;comment:买家ID" json:"buyer_id"`
	ReceiptURL  string               `gorm:"type:varchar(500);comment:回单地址" json:"receipt_url"`
	VoucherNo   string               `gorm:"type:varchar(64);comment:支付凭证号" json:"voucher_no"`
	Amount      float64              `gorm:"type:decimal(14,2);comment:回单金额" json:"amount"`
	Status      PaymentReceiptStatus `gorm:"type:varchar(32);index;comment:回单状态" json:"status"`
	AuditRemark string               `gorm:"type:varchar(255);comment:审核备注" json:"audit_remark"`
	SubmittedBy uint64               `gorm:"comment:提交人ID" json:"submitted_by"`
	SubmittedAt time.Time            `gorm:"comment:提交时间" json:"submitted_at"`
	AuditorID   *uint64              `gorm:"comment:审核人ID" json:"auditor_id"`
	AuditedAt   *time.Time           `gorm:"comment:审核时间" json:"audited_at"`
	CreatedAt   time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PaymentReceipt) TableName() string {
	return "payment_receipts"
}

type PickupCode struct {
	ID          uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderRefID  uint64           `gorm:"column:order_ref_id;uniqueIndex;comment:订单自增ID" json:"order_ref_id"`
	OrderID     string           `gorm:"column:order_id;type:varchar(40);index;comment:订单号" json:"order_id"`
	Code        string           `gorm:"type:varchar(32);uniqueIndex;comment:提货码" json:"code"`
	Status      PickupCodeStatus `gorm:"type:varchar(32);index;comment:提货码状态" json:"status"`
	GeneratedBy uint64           `gorm:"comment:生成人ID" json:"generated_by"`
	GeneratedAt time.Time        `gorm:"comment:生成时间" json:"generated_at"`
	VerifiedBy  *uint64          `gorm:"comment:核销人ID" json:"verified_by"`
	VerifiedAt  *time.Time       `gorm:"comment:核销时间" json:"verified_at"`
	CreatedAt   time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PickupCode) TableName() string {
	return "pickup_codes"
}
