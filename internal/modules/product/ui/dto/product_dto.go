package dto

import "time"

type ProductSpec struct {
	Material   string `json:"material"`
	Spec       string `json:"spec"`
	Thickness  string `json:"thickness,omitempty"`
	Width      string `json:"width,omitempty"`
	Length     string `json:"length,omitempty"`
	Surface    string `json:"surface,omitempty"`
	Edge       string `json:"edge,omitempty"`
	Processing string `json:"processing,omitempty"`
}

type WarehouseInfo struct {
	WarehouseID   uint64 `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	WarehouseCode string `json:"warehouse_code"`
	Address       string `json:"address"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
}

type CreateProductRequest struct {
	Name          string        `json:"name" binding:"required,max=200"`
	Category      string        `json:"category" binding:"required,max=50"`
	SubCategory   string        `json:"sub_category" binding:"required,max=50"`
	SteelMill     string        `json:"steel_mill" binding:"required,max=100"`
	Spec          ProductSpec   `json:"spec" binding:"required"`
	WarehouseInfo WarehouseInfo `json:"warehouse_info" binding:"required"`
	BundleNo      string        `json:"bundle_no" binding:"required,max=100"`
	Quantity      float64       `json:"quantity" binding:"required,gt=0"`
	UnitPrice     float64       `json:"unit_price" binding:"required,gt=0"`
	MinOrderQty   float64       `json:"min_order_qty" binding:"required,gt=0"`
	Description   string        `json:"description"`
	DeliveryDays  int           `json:"delivery_days"`
	QuoteValidity int           `json:"quote_validity"`
}

type UpdateProductRequest struct {
	Name          string        `json:"name"`
	Category      string        `json:"category"`
	SubCategory   string        `json:"sub_category"`
	SteelMill     string        `json:"steel_mill"`
	Spec          *ProductSpec  `json:"spec"`
	WarehouseInfo *WarehouseInfo `json:"warehouse_info"`
	Quantity      float64       `json:"quantity"`
	UnitPrice     float64       `json:"unit_price"`
	MinOrderQty   float64       `json:"min_order_qty"`
	Description   string        `json:"description"`
	DeliveryDays  int           `json:"delivery_days"`
	QuoteValidity int           `json:"quote_validity"`
}

type ProductInfo struct {
	ID            uint64        `json:"id"`
	ProductNo     string        `json:"product_no"`
	Name          string        `json:"name"`
	Category      string        `json:"category"`
	SubCategory   string        `json:"sub_category"`
	SteelMill     string        `json:"steel_mill"`
	Spec          ProductSpec   `json:"spec"`
	WarehouseInfo WarehouseInfo `json:"warehouse_info"`
	BundleNo      string        `json:"bundle_no"`
	Quantity      float64       `json:"quantity"`
	AvailableQty  float64       `json:"available_qty"`
	LockedQty     float64       `json:"locked_qty"`
	UnitPrice     float64       `json:"unit_price"`
	MinOrderQty   float64       `json:"min_order_qty"`
	Currency      string        `json:"currency"`
	TaxIncluded   bool          `json:"tax_included"`
	TaxRate       float64       `json:"tax_rate"`
	Status        int           `json:"status"`
	EnterpriseID  uint64        `json:"enterprise_id"`
	AuditRemark   string        `json:"audit_remark"`
	ApprovedAt    *time.Time    `json:"approved_at,omitempty"`
	QualityCert   string        `json:"quality_cert"`
	Images        []string      `json:"images"`
	Description   string        `json:"description"`
	DeliveryDays  int           `json:"delivery_days"`
	QuoteValidity int           `json:"quote_validity"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type ProductListRequest struct {
	Page         int     `form:"page" binding:"omitempty,min=1"`
	PageSize     int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Category     string  `form:"category"`
	SubCategory  string  `form:"sub_category"`
	Keyword      string  `form:"keyword"`
	SteelMill    string  `form:"steel_mill"`
	Status       *int    `form:"status"`
	EnterpriseID uint64  `form:"enterprise_id"`
	MinUnitPrice float64 `form:"min_unit_price"`
	MaxUnitPrice float64 `form:"max_unit_price"`
	MinAvailable float64 `form:"min_available"`
	Province     string  `form:"province"`
	City         string  `form:"city"`
	SortBy       string  `form:"sort_by"`
}

type ProductListResponse struct {
	List     []*ProductInfo `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type UpdateProductStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2 3 4"`
}
