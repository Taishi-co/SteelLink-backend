package dto

type AddCartItemRequest struct {
	ProductID uint64  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

type CartItemInfo struct {
	ID            uint64             `json:"id"`
	ProductID     uint64             `json:"product_id"`
	ProductName   string             `json:"product_name"`
	ImageURL      string             `json:"image_url"`
	Spec          string             `json:"spec"`
	Price         float64            `json:"price"`
	Quantity      float64            `json:"quantity"`
	SellerID      uint64             `json:"seller_id"`
	SellerName    string             `json:"seller_name"`
	WarehouseInfo *CartWarehouseInfo `json:"warehouse_info,omitempty"`
	Unit          string             `json:"unit"`
	SubTotal      float64            `json:"sub_total"`
}

type CartWarehouseInfo struct {
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

type CartListResponse struct {
	List  []*CartItemInfo `json:"list"`
	Total int             `json:"total"`
}

type CartCountResponse struct {
	Count float64 `json:"count"`
}
