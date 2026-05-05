package entity

import (
	"time"

	"gorm.io/gorm"
)

type CartItem struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64         `gorm:"index:idx_cart_owner_product,priority:1;comment:用户ID" json:"user_id"`
	EnterpriseID uint64         `gorm:"index;comment:企业ID" json:"enterprise_id"`
	ProductID    uint64         `gorm:"index:idx_cart_owner_product,priority:2;comment:商品ID" json:"product_id"`
	Quantity     float64        `gorm:"type:decimal(12,2);comment:采购数量" json:"quantity"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (CartItem) TableName() string {
	return "cart_items"
}
