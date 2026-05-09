package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProductCategoryDisabled = 0
	ProductCategoryEnabled  = 1
)

// ProductCategory 商品品类主数据。Name 是商品表中保存的业务品类值，DisplayName 用于前台展示。
type ProductCategory struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID    uint64         `gorm:"default:0;index;comment:父级品类ID" json:"parent_id"`
	Name        string         `gorm:"type:varchar(80);not null;uniqueIndex;comment:业务品类名" json:"name"`
	DisplayName string         `gorm:"type:varchar(80);not null;comment:展示名" json:"display_name"`
	Code        string         `gorm:"type:varchar(80);index;comment:品类编码" json:"code"`
	Icon        string         `gorm:"type:varchar(64);comment:导航图标" json:"icon"`
	ImageURL    string         `gorm:"type:varchar(512);comment:导航图片URL" json:"image_url"`
	Sort        int            `gorm:"default:0;index;comment:排序" json:"sort"`
	Status      int            `gorm:"default:1;index;comment:状态 1启用 0停用" json:"status"`
	CreatedBy   uint64         `gorm:"comment:创建人ID" json:"created_by"`
	UpdatedBy   uint64         `gorm:"comment:更新人ID" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductCategory) TableName() string {
	return "product_categories"
}
