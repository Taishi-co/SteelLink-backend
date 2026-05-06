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
	ID                  uint64              `json:"id"`
	LogisticsNo         string              `json:"logistics_no"`
	OrderID             uint64              `json:"order_id"`
	OrderNo             string              `json:"order_no"`
	BuyerEnterpriseID   uint64              `json:"buyer_enterprise_id"`
	LogisticsServiceID  uint64              `json:"logistics_service_id"`
	CarrierEnterpriseID uint64              `json:"carrier_enterprise_id"`
	WarehouseServiceID  uint64              `json:"warehouse_service_id"`
	Origin              string              `json:"origin"`
	Destination         string              `json:"destination"`
	CargoName           string              `json:"cargo_name"`
	Weight              float64             `json:"weight"`
	VehicleType         string              `json:"vehicle_type"`
	EstimatedPrice      float64             `json:"estimated_price"`
	FinalPrice          float64             `json:"final_price"`
	Status              string              `json:"status"`
	DriverName          string              `json:"driver_name"`
	DriverPhone         string              `json:"driver_phone"`
	PlateNumber         string              `json:"plate_number"`
	GuaranteeType       string              `json:"guarantee_type"`
	AcceptedAt          *time.Time          `json:"accepted_at"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	TrackingEvents      []TrackingEventInfo `json:"tracking_events,omitempty"`
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
