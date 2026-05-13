package dto

import "time"

type BindProductWarehouseRequest struct {
	WarehouseServiceID uint64 `json:"warehouse_service_id" binding:"required"`
}

type CreateWarehouseMovementRequest struct {
	WarehouseServiceID uint64  `json:"warehouse_service_id" binding:"required"`
	ProductID          uint64  `json:"product_id"`
	ProductName        string  `json:"product_name"`
	BundleNo           string  `json:"bundle_no"`
	OrderID            uint64  `json:"order_id"`
	OrderNo            string  `json:"order_no"`
	MovementType       string  `json:"movement_type" binding:"required,oneof=inbound outbound"`
	Quantity           float64 `json:"quantity" binding:"required,gt=0"`
	Status             string  `json:"status"`
	ScheduledAt        string  `json:"scheduled_at"`
	Remark             string  `json:"remark"`
}

type CreateWarehouseStorageRequest struct {
	WarehouseServiceID uint64  `json:"warehouse_service_id" binding:"required"`
	ProductID          uint64  `json:"product_id"`
	ProductName        string  `json:"product_name" binding:"required"`
	BundleNo           string  `json:"bundle_no"`
	OrderID            uint64  `json:"order_id"`
	OrderNo            string  `json:"order_no"`
	Quantity           float64 `json:"quantity" binding:"required,gt=0"`
	ScheduledAt        string  `json:"scheduled_at"`
	ContactName        string  `json:"contact_name"`
	ContactPhone       string  `json:"contact_phone"`
	StorageRequirement string  `json:"storage_requirement"`
	Remark             string  `json:"remark"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Remark string `json:"remark"`
}

type EstimateLogisticsRequest struct {
	LogisticsServiceID uint64  `json:"logistics_service_id" binding:"required"`
	WarehouseServiceID uint64  `json:"warehouse_service_id"`
	Origin             string  `json:"origin"`
	Destination        string  `json:"destination" binding:"required"`
	Weight             float64 `json:"weight" binding:"required,gt=0"`
}

type EstimateLogisticsResponse struct {
	LogisticsServiceID uint64  `json:"logistics_service_id"`
	WarehouseServiceID uint64  `json:"warehouse_service_id"`
	Origin             string  `json:"origin"`
	Destination        string  `json:"destination"`
	Weight             float64 `json:"weight"`
	UnitPrice          float64 `json:"unit_price"`
	EstimatedPrice     float64 `json:"estimated_price"`
	PriceRule          string  `json:"price_rule"`
}

type CreateLogisticsOrderRequest struct {
	OrderID            uint64  `json:"order_id"`
	OrderNo            string  `json:"order_no"`
	LogisticsServiceID uint64  `json:"logistics_service_id" binding:"required"`
	WarehouseServiceID uint64  `json:"warehouse_service_id"`
	Origin             string  `json:"origin" binding:"required"`
	Destination        string  `json:"destination" binding:"required"`
	CargoName          string  `json:"cargo_name"`
	Weight             float64 `json:"weight" binding:"required,gt=0"`
	VehicleType        string  `json:"vehicle_type"`
	GuaranteeType      string  `json:"guarantee_type"`
}

type AcceptLogisticsOrderRequest struct {
	DriverName  string  `json:"driver_name"`
	DriverPhone string  `json:"driver_phone"`
	PlateNumber string  `json:"plate_number"`
	FinalPrice  float64 `json:"final_price"`
}

type CreateTrackingEventRequest struct {
	EventType   string `json:"event_type" binding:"required"`
	Location    string `json:"location"`
	Description string `json:"description"`
	OccurredAt  string `json:"occurred_at"`
}

type CreateGuaranteeRequest struct {
	OrderID          uint64  `json:"order_id"`
	LogisticsOrderID uint64  `json:"logistics_order_id"`
	GuaranteeType    string  `json:"guarantee_type" binding:"required"`
	CoverageAmount   float64 `json:"coverage_amount"`
	Description      string  `json:"description"`
}

type UpdateGuaranteeStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	Resolution string `json:"resolution"`
}

type ListRequest struct {
	Page               int    `form:"page"`
	PageSize           int    `form:"page_size"`
	OrderID            uint64 `form:"order_id"`
	WarehouseServiceID uint64 `form:"warehouse_service_id"`
	LogisticsServiceID uint64 `form:"logistics_service_id"`
	Status             string `form:"status"`
	Role               string `form:"role"`
}

type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type WarehouseMovementInfo struct {
	ID                    uint64     `json:"id"`
	MovementNo            string     `json:"movement_no"`
	WarehouseServiceID    uint64     `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64     `json:"warehouse_enterprise_id"`
	RequestEnterpriseID   uint64     `json:"request_enterprise_id"`
	ProductID             uint64     `json:"product_id"`
	ProductName           string     `json:"product_name"`
	BundleNo              string     `json:"bundle_no"`
	OrderID               uint64     `json:"order_id"`
	OrderNo               string     `json:"order_no"`
	MovementType          string     `json:"movement_type"`
	Quantity              float64    `json:"quantity"`
	Status                string     `json:"status"`
	ScheduledAt           string     `json:"scheduled_at"`
	CompletedAt           *time.Time `json:"completed_at"`
	Remark                string     `json:"remark"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type LogisticsOrderInfo struct {
	ID                      uint64              `json:"id"`
	LogisticsNo             string              `json:"logistics_no"`
	OrderID                 uint64              `json:"order_id"`
	OrderNo                 string              `json:"order_no"`
	BuyerEnterpriseID       uint64              `json:"buyer_enterprise_id"`
	LogisticsServiceID      uint64              `json:"logistics_service_id"`
	CarrierEnterpriseID     uint64              `json:"carrier_enterprise_id"`
	WarehouseServiceID      uint64              `json:"warehouse_service_id"`
	Origin                  string              `json:"origin"`
	Destination             string              `json:"destination"`
	CargoName               string              `json:"cargo_name"`
	Weight                  float64             `json:"weight"`
	VehicleType             string              `json:"vehicle_type"`
	EstimatedPrice          float64             `json:"estimated_price"`
	FinalPrice              float64             `json:"final_price"`
	Status                  string              `json:"status"`
	CurrentResponsibleParty string              `json:"current_responsible_party"`
	HandoverNote            string              `json:"handover_note"`
	DriverName              string              `json:"driver_name"`
	DriverPhone             string              `json:"driver_phone"`
	PlateNumber             string              `json:"plate_number"`
	GuaranteeType           string              `json:"guarantee_type"`
	AcceptedAt              *time.Time          `json:"accepted_at"`
	PickupAppointmentAt     *time.Time          `json:"pickup_appointment_at,omitempty"`
	ArrivalSignedInAt       *time.Time          `json:"arrival_signed_in_at,omitempty"`
	OutboundConfirmedAt     *time.Time          `json:"outbound_confirmed_at,omitempty"`
	LoadingConfirmedAt      *time.Time          `json:"loading_confirmed_at,omitempty"`
	ArrivalConfirmedAt      *time.Time          `json:"arrival_confirmed_at,omitempty"`
	ReceiptSignedAt         *time.Time          `json:"receipt_signed_at,omitempty"`
	InboundCompletedAt      *time.Time          `json:"inbound_completed_at,omitempty"`
	CreatedAt               time.Time           `json:"created_at"`
	UpdatedAt               time.Time           `json:"updated_at"`
	TrackingEvents          []TrackingEventInfo `json:"tracking_events,omitempty"`
}

type TrackingEventInfo struct {
	ID               uint64    `json:"id"`
	LogisticsOrderID uint64    `json:"logistics_order_id"`
	EventType        string    `json:"event_type"`
	Location         string    `json:"location"`
	Description      string    `json:"description"`
	OccurredAt       time.Time `json:"occurred_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type GuaranteeInfo struct {
	ID               uint64    `json:"id"`
	GuaranteeNo      string    `json:"guarantee_no"`
	OrderID          uint64    `json:"order_id"`
	LogisticsOrderID uint64    `json:"logistics_order_id"`
	EnterpriseID     uint64    `json:"enterprise_id"`
	GuaranteeType    string    `json:"guarantee_type"`
	Status           string    `json:"status"`
	CoverageAmount   float64   `json:"coverage_amount"`
	Description      string    `json:"description"`
	Resolution       string    `json:"resolution"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FulfillmentSummary struct {
	OrderID        uint64                  `json:"order_id"`
	WarehouseMoves []WarehouseMovementInfo `json:"warehouse_movements"`
	Logistics      []LogisticsOrderInfo    `json:"logistics_orders"`
	Guarantees     []GuaranteeInfo         `json:"guarantees"`
}

type UploadFulfillmentDocRequest struct {
	DocType      string   `json:"doc_type" binding:"required"`
	FileURL      string   `json:"file_url" binding:"required"`
	FileHash     string   `json:"file_hash"`
	Remark       string   `json:"remark"`
	GPSLatitude  *float64 `json:"gps_latitude"`
	GPSLongitude *float64 `json:"gps_longitude"`
	CapturedAt   string   `json:"captured_at"`
}

type SealConfirmRequest struct {
	FileURL string `json:"file_url" binding:"required"`
	Remark  string `json:"remark"`
}

type InspectionRequest struct {
	Result   string `json:"result" binding:"required"`
	FileURL  string `json:"file_url"`
	SealOK   bool   `json:"seal_ok"`
	Remark   string `json:"remark"`
}

type FulfillmentDocumentInfo struct {
	ID               uint64   `json:"id"`
	LogisticsOrderID uint64   `json:"logistics_order_id"`
	OrderID          uint64   `json:"order_id"`
	DocType          string   `json:"doc_type"`
	FileURL          string   `json:"file_url"`
	Remark           string   `json:"remark"`
	GPSLatitude      *float64 `json:"gps_latitude,omitempty"`
	GPSLongitude     *float64 `json:"gps_longitude,omitempty"`
	CapturedAt       string   `json:"captured_at,omitempty"`
	UploadedBy       uint64   `json:"uploaded_by"`
	EnterpriseID     uint64   `json:"enterprise_id"`
	UploadedRole     string   `json:"uploaded_role"`
	CreatedAt        string   `json:"created_at"`
}
