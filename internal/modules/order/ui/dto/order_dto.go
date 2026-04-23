package dto

import "time"

type CreateOrderRequest struct {
	ProductID uint64  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type UploadPaymentVoucherRequest struct {
	VoucherURL    string `json:"voucher_url" binding:"required"`
	TransactionNo string `json:"transaction_no"`
	PayerAccount  string `json:"payer_account"`
	PayeeAccount  string `json:"payee_account"`
}

type SetDeliveryInfoRequest struct {
	DriverName    string `json:"driver_name" binding:"required"`
	DriverPhone   string `json:"driver_phone" binding:"required"`
	DriverIDCard  string `json:"driver_id_card"`
	PlateNumber   string `json:"plate_number" binding:"required"`
	WarehouseID   uint64 `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name" binding:"required"`
	Address       string `json:"address" binding:"required"`
	EstimatedTime string `json:"estimated_time"`
}

type ConfirmReceiptRequest struct {
	ReceiptVoucher string `json:"receipt_voucher" binding:"required"`
}

type OrderItemInfo struct {
	ProductID   uint64  `json:"product_id"`
	ProductNo   string  `json:"product_no"`
	ProductName string  `json:"product_name"`
	SteelMill   string  `json:"steel_mill"`
	Category    string  `json:"category"`
	BundleNo    string  `json:"bundle_no"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalAmount float64 `json:"total_amount"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
}

type OrderInfo struct {
	ID             uint64          `json:"id"`
	OrderNo        string          `json:"order_no"`
	BuyerID        uint64          `json:"buyer_id"`
	BuyerName      string          `json:"buyer_name"`
	SellerID       uint64          `json:"seller_id"`
	SellerName     string          `json:"seller_name"`
	Items          []OrderItemInfo `json:"items"`
	TotalQuantity  float64         `json:"total_quantity"`
	TotalAmount    float64         `json:"total_amount"`
	TaxTotalAmount float64         `json:"tax_total_amount"`
	Currency       string          `json:"currency"`
	Status         int             `json:"status"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	PaymentInfo    *PaymentInfo    `json:"payment_info,omitempty"`
	DeliveryInfo   *DeliveryInfo   `json:"delivery_info,omitempty"`
	ReceiptVoucher string          `json:"receipt_voucher,omitempty"`
	ReceiptAt      *time.Time      `json:"receipt_at,omitempty"`
}

type PaymentInfo struct {
	PaymentMethod string    `json:"payment_method"`
	PaymentTime   time.Time `json:"payment_time"`
	TransactionNo string    `json:"transaction_no"`
	PayeeAccount  string    `json:"payee_account"`
	PayerAccount  string    `json:"payer_account"`
	VoucherURL    string    `json:"voucher_url"`
}

type DeliveryInfo struct {
	DriverName    string `json:"driver_name"`
	DriverPhone   string `json:"driver_phone"`
	DriverIDCard  string `json:"driver_id_card"`
	PlateNumber   string `json:"plate_number"`
	PickupCode    string `json:"pickup_code"`
	WarehouseID   uint64 `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	Address       string `json:"address"`
	EstimatedTime string `json:"estimated_time"`
}
