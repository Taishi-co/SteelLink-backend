package dto

import "time"

type AddFavoriteRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
}

type FavoriteInfo struct {
	ID          uint64    `json:"id"`
	ProductID   uint64    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Category    string    `json:"category"`
	Spec        string    `json:"spec"`
	Price       float64   `json:"price"`
	Image       string    `json:"image"`
	SellerID    uint64    `json:"seller_id"`
	SellerName  string    `json:"seller_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type FavoriteListResponse struct {
	List  []*FavoriteInfo `json:"list"`
	Total int             `json:"total"`
}

type FavoriteCountResponse struct {
	Count int64 `json:"count"`
}

type FavoriteCheckResponse struct {
	IsFavorite bool `json:"is_favorite"`
}
