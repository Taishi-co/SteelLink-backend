package dto

type AddCartItemRequest struct {
	ProductID uint64  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

type CartItemInfo struct {
	ID          uint64  `json:"id"`
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Spec        string  `json:"spec"`
	Price       float64 `json:"price"`
	Quantity    float64 `json:"quantity"`
	SellerID    uint64  `json:"seller_id"`
	SellerName  string  `json:"seller_name"`
	Unit        string  `json:"unit"`
	SubTotal    float64 `json:"sub_total"`
}

type CartListResponse struct {
	List  []*CartItemInfo `json:"list"`
	Total int             `json:"total"`
}

type CartCountResponse struct {
	Count float64 `json:"count"`
}
