package entity

import "time"

type Favorite struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"uniqueIndex:idx_favorites_user_product;index;not null;comment:用户ID" json:"user_id"`
	EnterpriseID uint64    `gorm:"index;default:0;comment:企业ID" json:"enterprise_id"`
	ProductID    uint64    `gorm:"uniqueIndex:idx_favorites_user_product;index;not null;comment:商品ID" json:"product_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
