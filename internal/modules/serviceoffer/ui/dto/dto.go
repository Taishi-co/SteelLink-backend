package dto

import "time"

type WarehouseServiceRequest struct {
	WarehouseName     string   `json:"warehouse_name" binding:"required"`
	Province          string   `json:"province"`
	City              string   `json:"city"`
	District          string   `json:"district"`
	Address           string   `json:"address" binding:"required"`
	Capacity          float64  `json:"capacity"`
	AvailableCapacity float64  `json:"available_capacity"`
	ServiceTypes      []string `json:"service_types"`
	PriceRule         string   `json:"price_rule"`
	Images            []string `json:"images"`
	Certificates      []string `json:"certificates"`
	Description       string   `json:"description"`
}

type LogisticsServiceRequest struct {
	Name         string   `json:"name" binding:"required"`
	RouteStart   string   `json:"route_start" binding:"required"`
	RouteEnd     string   `json:"route_end" binding:"required"`
	ServiceScope string   `json:"service_scope"`
	VehicleType  string   `json:"vehicle_type"`
	LoadCapacity float64  `json:"load_capacity"`
	PriceRule    string   `json:"price_rule"`
	Images       []string `json:"images"`
	Certificates []string `json:"certificates"`
	Description  string   `json:"description"`
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
	ID                uint64    `json:"id"`
	EnterpriseID      uint64    `json:"enterprise_id"`
	WarehouseName     string    `json:"warehouse_name"`
	Province          string    `json:"province"`
	City              string    `json:"city"`
	District          string    `json:"district"`
	Address           string    `json:"address"`
	Capacity          float64   `json:"capacity"`
	AvailableCapacity float64   `json:"available_capacity"`
	ServiceTypes      []string  `json:"service_types"`
	PriceRule         string    `json:"price_rule"`
	Images            []string  `json:"images"`
	Certificates      []string  `json:"certificates"`
	Description       string    `json:"description"`
	Status            int       `json:"status"`
	AuditRemark       string    `json:"audit_remark"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type LogisticsServiceInfo struct {
	ID           uint64    `json:"id"`
	EnterpriseID uint64    `json:"enterprise_id"`
	Name         string    `json:"name"`
	RouteStart   string    `json:"route_start"`
	RouteEnd     string    `json:"route_end"`
	ServiceScope string    `json:"service_scope"`
	VehicleType  string    `json:"vehicle_type"`
	LoadCapacity float64   `json:"load_capacity"`
	PriceRule    string    `json:"price_rule"`
	Images       []string  `json:"images"`
	Certificates []string  `json:"certificates"`
	Description  string    `json:"description"`
	Status       int       `json:"status"`
	AuditRemark  string    `json:"audit_remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
