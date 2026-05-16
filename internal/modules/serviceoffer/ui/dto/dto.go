package dto

import "time"

type WarehouseServiceRequest struct {
	WarehouseName      string   `json:"warehouse_name" binding:"required"`
	Province           string   `json:"province"`
	City               string   `json:"city"`
	District           string   `json:"district"`
	Address            string   `json:"address" binding:"required"`
	Capacity           float64  `json:"capacity"`
	AvailableCapacity  float64  `json:"available_capacity"`
	ServiceTypes       []string `json:"service_types"`
	PriceRule          string   `json:"price_rule"`
	LoadingCapacity    string   `json:"loading_capacity"`
	BusinessHours      string   `json:"business_hours"`
	PledgeSupervision  bool     `json:"pledge_supervision"`
	OutboundVoucherReq string   `json:"outbound_voucher_req"`
	Images             []string `json:"images"`
	Certificates       []string `json:"certificates"`
	Description        string   `json:"description"`
}

type LogisticsServiceRequest struct {
	Name          string         `json:"name" binding:"required"`
	RouteStart    string         `json:"route_start" binding:"required"`
	RouteEnd      string         `json:"route_end" binding:"required"`
	ServiceScope  string         `json:"service_scope"`
	TransportMode string         `json:"transport_mode"`
	TransitCycle  string         `json:"transit_cycle"`
	UnitPrice     float64        `json:"unit_price"`
	ContactInfo   string         `json:"contact_info"`
	VehicleType   string         `json:"vehicle_type"`
	LoadCapacity  float64        `json:"load_capacity"`
	PriceRule     string         `json:"price_rule"`
	QuoteRules    map[string]any `json:"quote_rules"`
	FleetName     string         `json:"fleet_name"`
	VehicleCerts  []string       `json:"vehicle_certs"`
	DriverCerts   []string       `json:"driver_certs"`
	Images        []string       `json:"images"`
	Certificates  []string       `json:"certificates"`
	Description   string         `json:"description"`
}

type FinanceServiceRequest struct {
	Name               string   `json:"name" binding:"required"`
	FinanceType        string   `json:"finance_type"`
	MinAmount          float64  `json:"min_amount"`
	MaxAmount          float64  `json:"max_amount"`
	TermDays           int      `json:"term_days"`
	AnnualRate         float64  `json:"annual_rate"`
	FeeRule            string   `json:"fee_rule"`
	RiskControlRule    string   `json:"risk_control_rule"`
	RequiredDocuments  []string `json:"required_documents"`
	SupportedScenarios []string `json:"supported_scenarios"`
	Description        string   `json:"description"`
}

type AuditRequest struct {
	Status           int    `json:"status" binding:"required,oneof=2 3 4"`
	AuditRemark      string `json:"audit_remark" binding:"omitempty,max=500"`
	AuditRemarkCamel string `json:"auditRemark" binding:"omitempty,max=500"`
}

type ListRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Keyword      string `form:"keyword"`
	Province     string `form:"province"`
	City         string `form:"city"`
	EnterpriseID uint64 `form:"enterprise_id"`
	Status       *int   `form:"status"`
}

type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type WarehouseServiceInfo struct {
	ID                 uint64    `json:"id"`
	EnterpriseID       uint64    `json:"enterprise_id"`
	WarehouseName      string    `json:"warehouse_name"`
	Province           string    `json:"province"`
	City               string    `json:"city"`
	District           string    `json:"district"`
	Address            string    `json:"address"`
	Capacity           float64   `json:"capacity"`
	AvailableCapacity  float64   `json:"available_capacity"`
	ServiceTypes       []string  `json:"service_types"`
	PriceRule          string    `json:"price_rule"`
	LoadingCapacity    string    `json:"loading_capacity"`
	BusinessHours      string    `json:"business_hours"`
	PledgeSupervision  bool      `json:"pledge_supervision"`
	FinanceCertified   bool      `json:"finance_certified"`
	OutboundVoucherReq string    `json:"outbound_voucher_req"`
	Images             []string  `json:"images"`
	Certificates       []string  `json:"certificates"`
	Description        string    `json:"description"`
	Status             int       `json:"status"`
	AuditRemark        string    `json:"audit_remark"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type LogisticsServiceInfo struct {
	ID            uint64         `json:"id"`
	EnterpriseID  uint64         `json:"enterprise_id"`
	Name          string         `json:"name"`
	RouteStart    string         `json:"route_start"`
	RouteEnd      string         `json:"route_end"`
	ServiceScope  string         `json:"service_scope"`
	TransportMode string         `json:"transport_mode"`
	TransitCycle  string         `json:"transit_cycle"`
	UnitPrice     float64        `json:"unit_price"`
	ContactInfo   string         `json:"contact_info"`
	VehicleType   string         `json:"vehicle_type"`
	LoadCapacity  float64        `json:"load_capacity"`
	PriceRule     string         `json:"price_rule"`
	QuoteRules    map[string]any `json:"quote_rules"`
	FleetName     string         `json:"fleet_name"`
	VehicleCerts  []string       `json:"vehicle_certs"`
	DriverCerts   []string       `json:"driver_certs"`
	Images        []string       `json:"images"`
	Certificates  []string       `json:"certificates"`
	Description   string         `json:"description"`
	Status        int            `json:"status"`
	AuditRemark   string         `json:"audit_remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type FinanceServiceInfo struct {
	ID                     uint64    `json:"id"`
	EnterpriseID           uint64    `json:"enterprise_id"`
	Name                   string    `json:"name"`
	FinanceType            string    `json:"finance_type"`
	MinAmount              float64   `json:"min_amount"`
	MaxAmount              float64   `json:"max_amount"`
	TermDays               int       `json:"term_days"`
	AnnualRate             float64   `json:"annual_rate"`
	FeeRule                string    `json:"fee_rule"`
	RiskControlRule        string    `json:"risk_control_rule"`
	RequiredDocuments      []string  `json:"required_documents"`
	SupportedScenarios     []string  `json:"supported_scenarios"`
	Description            string    `json:"description"`
	Status                 int       `json:"status"`
	AuditRemark            string    `json:"audit_remark"`
	AuthorizedWarehouseIDs []uint64  `json:"authorized_warehouse_ids,omitempty"`
	AuthorizedLogisticsIDs []uint64  `json:"authorized_logistics_ids,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type AuthorizeWarehouseRequest struct {
	WarehouseServiceID uint64 `json:"warehouse_service_id" binding:"required"`
	AuthorizationType  string `json:"authorization_type"`
}

type AuthorizeLogisticsRequest struct {
	LogisticsServiceID uint64 `json:"logistics_service_id" binding:"required"`
}

type AuthorizedWarehouseInfo struct {
	ID                    uint64    `json:"id"`
	FinanceServiceID      uint64    `json:"finance_service_id"`
	WarehouseServiceID    uint64    `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64    `json:"warehouse_enterprise_id"`
	AuthorizationType     string    `json:"authorization_type"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}

type AuthorizedLogisticsInfo struct {
	ID                    uint64    `json:"id"`
	FinanceServiceID      uint64    `json:"finance_service_id"`
	LogisticsServiceID    uint64    `json:"logistics_service_id"`
	LogisticsEnterpriseID uint64    `json:"logistics_enterprise_id"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}
