package dto

import "time"

type ProductSpec struct {
	Material        string  `json:"material"`
	Spec            string  `json:"spec"`
	Thickness       string  `json:"thickness,omitempty"`
	Width           string  `json:"width,omitempty"`
	Length          string  `json:"length,omitempty"`
	HeightSpecType  string  `json:"height_spec_type,omitempty"`
	HeightSpecValue float64 `json:"height_spec_value,omitempty"`
	HeightSpecUnit  string  `json:"height_spec_unit,omitempty"`
	WidthSpecType   string  `json:"width_spec_type,omitempty"`
	WidthSpecValue  float64 `json:"width_spec_value,omitempty"`
	WidthSpecUnit   string  `json:"width_spec_unit,omitempty"`
	LengthType      string  `json:"length_type,omitempty"`
	LengthValue     float64 `json:"length_value,omitempty"`
	LengthUnit      string  `json:"length_unit,omitempty"`
	Surface         string  `json:"surface,omitempty"`
	Edge            string  `json:"edge,omitempty"`
	Processing      string  `json:"processing,omitempty"`
}

type WarehouseInfo struct {
	WarehouseID           uint64 `json:"warehouse_id"`
	WarehouseServiceID    uint64 `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64 `json:"warehouse_enterprise_id"`
	WarehouseName         string `json:"warehouse_name"`
	WarehouseCode         string `json:"warehouse_code"`
	Address               string `json:"address"`
	Province              string `json:"province"`
	City                  string `json:"city"`
	District              string `json:"district"`
}

type CreateProductRequest struct {
	Name               string        `json:"name" binding:"required,max=200"`
	GoodsType          string        `json:"goods_type" binding:"omitempty,oneof=spot futures"`
	CategoryID         uint64        `json:"category_id"`
	GradeID            uint64        `json:"grade_id"`
	SteelMillID        uint64        `json:"steel_mill_id"`
	Category           string        `json:"category" binding:"required,max=50"`
	SubCategory        string        `json:"sub_category" binding:"required,max=50"`
	SteelMill          string        `json:"steel_mill" binding:"required,max=100"`
	Spec               ProductSpec   `json:"spec" binding:"required"`
	WarehouseServiceID uint64        `json:"warehouse_service_id"`
	WarehouseInfo      WarehouseInfo `json:"warehouse_info" binding:"required"`
	BundleNo           string        `json:"bundle_no" binding:"required,max=100"`
	Quantity           float64       `json:"quantity" binding:"required,gt=0"`
	UnitPrice          float64       `json:"unit_price" binding:"required,gt=0"`
	MinOrderQty        float64       `json:"min_order_qty" binding:"required,gt=0"`
	Images             []string      `json:"images" binding:"required,min=1"`
	QualityCert        string        `json:"quality_cert"`
	TitleVoucherURL    string        `json:"title_voucher_url"`
	TitleVoucherRemark string        `json:"title_voucher_remark"`
	Description        string        `json:"description"`
	DeliveryDays       int           `json:"delivery_days"`
	QuoteValidity      int           `json:"quote_validity"`
}

type UpdateProductRequest struct {
	Name               string         `json:"name"`
	GoodsType          string         `json:"goods_type"`
	CategoryID         uint64         `json:"category_id"`
	GradeID            uint64         `json:"grade_id"`
	SteelMillID        uint64         `json:"steel_mill_id"`
	Category           string         `json:"category"`
	SubCategory        string         `json:"sub_category"`
	SteelMill          string         `json:"steel_mill"`
	Spec               *ProductSpec   `json:"spec"`
	WarehouseServiceID uint64         `json:"warehouse_service_id"`
	WarehouseInfo      *WarehouseInfo `json:"warehouse_info"`
	Quantity           float64        `json:"quantity"`
	UnitPrice          float64        `json:"unit_price"`
	MinOrderQty        float64        `json:"min_order_qty"`
	TitleVoucherURL    string         `json:"title_voucher_url"`
	TitleVoucherRemark string         `json:"title_voucher_remark"`
	Description        string         `json:"description"`
	DeliveryDays       int            `json:"delivery_days"`
	QuoteValidity      int            `json:"quote_validity"`
}

type UpdateProductDisplayRequest struct {
	Images        []string `json:"images" binding:"required,min=1"`
	QualityCert   string   `json:"quality_cert"`
	Description   string   `json:"description"`
	DeliveryDays  int      `json:"delivery_days"`
	QuoteValidity int      `json:"quote_validity"`
}

type ProductInfo struct {
	ID                    uint64        `json:"id"`
	ProductNo             string        `json:"product_no"`
	Name                  string        `json:"name"`
	GoodsType             string        `json:"goods_type"`
	CategoryID            uint64        `json:"category_id"`
	GradeID               uint64        `json:"grade_id"`
	SteelMillID           uint64        `json:"steel_mill_id"`
	Category              string        `json:"category"`
	SubCategory           string        `json:"sub_category"`
	SteelMill             string        `json:"steel_mill"`
	Spec                  ProductSpec   `json:"spec"`
	SpecRaw               string        `json:"spec_raw"`
	HeightSpecType        string        `json:"height_spec_type"`
	HeightSpecValue       float64       `json:"height_spec_value"`
	HeightSpecUnit        string        `json:"height_spec_unit"`
	WidthSpecType         string        `json:"width_spec_type"`
	WidthSpecValue        float64       `json:"width_spec_value"`
	WidthSpecUnit         string        `json:"width_spec_unit"`
	LengthType            string        `json:"length_type"`
	LengthValue           float64       `json:"length_value"`
	LengthUnit            string        `json:"length_unit"`
	WarehouseServiceID    uint64        `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64        `json:"warehouse_enterprise_id"`
	WarehouseInfo         WarehouseInfo `json:"warehouse_info"`
	BundleNo              string        `json:"bundle_no"`
	Quantity              float64       `json:"quantity"`
	AvailableQty          float64       `json:"available_qty"`
	LockedQty             float64       `json:"locked_qty"`
	UnitPrice             float64       `json:"unit_price"`
	MinOrderQty           float64       `json:"min_order_qty"`
	Currency              string        `json:"currency"`
	TaxIncluded           bool          `json:"tax_included"`
	TaxRate               float64       `json:"tax_rate"`
	Status                int           `json:"status"`
	EnterpriseID          uint64        `json:"enterprise_id"`
	AuditRemark           string        `json:"audit_remark"`
	ApprovedAt            *time.Time    `json:"approved_at,omitempty"`
	QualityCert           string        `json:"quality_cert"`
	TitleVoucherURL       string        `json:"title_voucher_url"`
	TitleVoucherRemark    string        `json:"title_voucher_remark"`
	Images                []string      `json:"images"`
	Description           string        `json:"description"`
	DeliveryDays          int           `json:"delivery_days"`
	QuoteValidity         int           `json:"quote_validity"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

type ProductListRequest struct {
	Page               int     `form:"page" binding:"omitempty,min=1"`
	PageSize           int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	CategoryID         uint64  `form:"category_id"`
	GradeID            uint64  `form:"grade_id"`
	SteelMillID        uint64  `form:"steel_mill_id"`
	Category           string  `form:"category"`
	SubCategory        string  `form:"sub_category"`
	Material           string  `form:"material"`
	Keyword            string  `form:"keyword"`
	SteelMill          string  `form:"steel_mill"`
	Status             *int    `form:"status"`
	EnterpriseID       uint64  `form:"enterprise_id"`
	MinUnitPrice       float64 `form:"min_unit_price"`
	MaxUnitPrice       float64 `form:"max_unit_price"`
	MinAvailable       float64 `form:"min_available"`
	Province           string  `form:"province"`
	City               string  `form:"city"`
	GoodsType          string  `form:"goods_type"`
	WarehouseServiceID uint64  `form:"warehouse_service_id"`
	HeightSpecType     string  `form:"height_spec_type"`
	MinHeightSpec      float64 `form:"min_height_spec"`
	MaxHeightSpec      float64 `form:"max_height_spec"`
	WidthSpecType      string  `form:"width_spec_type"`
	MinWidthSpec       float64 `form:"min_width_spec"`
	MaxWidthSpec       float64 `form:"max_width_spec"`
	LengthType         string  `form:"length_type"`
	MinLength          float64 `form:"min_length"`
	MaxLength          float64 `form:"max_length"`
	SortBy             string  `form:"sort_by"`
}

type ProductCategoryInfo struct {
	ID          uint64    `json:"id"`
	ParentID    uint64    `json:"parent_id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Code        string    `json:"code"`
	Icon        string    `json:"icon"`
	ImageURL    string    `json:"image_url"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductCategoryListRequest struct {
	IncludeDisabled bool `form:"include_disabled"`
}

type ProductGradeInfo struct {
	ID          uint64    `json:"id"`
	CategoryID  uint64    `json:"category_id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Code        string    `json:"code"`
	Aliases     []string  `json:"aliases"`
	Standard    string    `json:"standard"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductGradeListRequest struct {
	CategoryID      uint64 `form:"category_id"`
	Category        string `form:"category"`
	IncludeDisabled bool   `form:"include_disabled"`
}

type ProductSteelMillInfo struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Code        string    `json:"code"`
	Aliases     []string  `json:"aliases"`
	Letter      string    `json:"letter"`
	LogoURL     string    `json:"logo_url"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductSteelMillListRequest struct {
	IncludeDisabled bool `form:"include_disabled"`
}

type UpsertProductCategoryRequest struct {
	ParentID         uint64 `json:"parent_id"`
	Name             string `json:"name" binding:"required,max=80"`
	DisplayName      string `json:"display_name" binding:"omitempty,max=80"`
	DisplayNameCamel string `json:"displayName" binding:"omitempty,max=80"`
	Code             string `json:"code" binding:"omitempty,max=80"`
	Icon             string `json:"icon" binding:"omitempty,max=64"`
	ImageURL         string `json:"image_url" binding:"omitempty,max=512"`
	ImageURLCamel    string `json:"imageUrl" binding:"omitempty,max=512"`
	Sort             int    `json:"sort"`
	Status           int    `json:"status"`
}

type ProductListResponse struct {
	List     []*ProductInfo `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type UpdateProductStatusRequest struct {
	Status           int    `json:"status" binding:"required,oneof=1 2 3 4"`
	AuditRemark      string `json:"audit_remark" binding:"omitempty,max=255"`
	AuditRemarkCamel string `json:"auditRemark" binding:"omitempty,max=255"`
}

type ProductImportProduct struct {
	Name               string        `json:"name"`
	GoodsType          string        `json:"goods_type"`
	CategoryID         uint64        `json:"category_id"`
	GradeID            uint64        `json:"grade_id"`
	SteelMillID        uint64        `json:"steel_mill_id"`
	Category           string        `json:"category"`
	SubCategory        string        `json:"sub_category"`
	SteelMill          string        `json:"steel_mill"`
	Spec               ProductSpec   `json:"spec"`
	WarehouseServiceID uint64        `json:"warehouse_service_id"`
	WarehouseInfo      WarehouseInfo `json:"warehouse_info"`
	BundleNo           string        `json:"bundle_no"`
	Quantity           float64       `json:"quantity"`
	UnitPrice          float64       `json:"unit_price"`
	MinOrderQty        float64       `json:"min_order_qty"`
	Images             []string      `json:"images"`
	QualityCert        string        `json:"quality_cert"`
	Description        string        `json:"description"`
	DeliveryDays       int           `json:"delivery_days"`
	QuoteValidity      int           `json:"quote_validity"`
}

type ProductImportDraftItem struct {
	RowNo      int                  `json:"row_no"`
	Confidence float64              `json:"confidence"`
	Warnings   []string             `json:"warnings"`
	Product    ProductImportProduct `json:"product"`
}

type ProductImportParseResponse struct {
	SourceFilename string                   `json:"source_filename"`
	SourceType     string                   `json:"source_type"`
	UsedModel      bool                     `json:"used_model"`
	Items          []ProductImportDraftItem `json:"items"`
	Warnings       []string                 `json:"warnings"`
}

type ConfirmProductImportRequest struct {
	SubmitForReview bool                     `json:"submit_for_review"`
	Items           []ProductImportDraftItem `json:"items" binding:"required,min=1,max=200"`
}

type ProductImportConfirmItemResult struct {
	RowNo   int          `json:"row_no"`
	Status  string       `json:"status"`
	Error   string       `json:"error,omitempty"`
	Product *ProductInfo `json:"product,omitempty"`
}

type ProductImportConfirmResponse struct {
	SuccessCount int                              `json:"success_count"`
	FailedCount  int                              `json:"failed_count"`
	Items        []ProductImportConfirmItemResult `json:"items"`
}
