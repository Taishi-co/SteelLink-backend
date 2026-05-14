package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProductTaxonomyDisabled = 0
	ProductTaxonomyEnabled  = 1
)

// ProductGrade 商品牌号主数据。商品仍保存文本快照，Grade 用于标准化检索和批量导入映射。
type ProductGrade struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID  uint64         `gorm:"index:idx_product_grade_category_name,unique;index;comment:品种ID" json:"category_id"`
	Name        string         `gorm:"type:varchar(80);not null;index:idx_product_grade_category_name,unique;comment:牌号名称" json:"name"`
	DisplayName string         `gorm:"type:varchar(80);not null;comment:展示名" json:"display_name"`
	Code        string         `gorm:"type:varchar(80);index;comment:牌号编码" json:"code"`
	Aliases     StringList     `gorm:"type:json;comment:牌号别名" json:"aliases"`
	Standard    string         `gorm:"type:varchar(80);comment:执行标准" json:"standard"`
	Sort        int            `gorm:"default:0;index;comment:排序" json:"sort"`
	Status      int            `gorm:"default:1;index;comment:状态 1启用 0停用" json:"status"`
	CreatedBy   uint64         `gorm:"comment:创建人ID" json:"created_by"`
	UpdatedBy   uint64         `gorm:"comment:更新人ID" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductGrade) TableName() string {
	return "product_grades"
}

// ProductSteelMill 商品钢厂/产地主数据。商品仍保存 steel_mill 文本快照。
type ProductSteelMill struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex;comment:钢厂名称" json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null;comment:展示名" json:"display_name"`
	Code        string         `gorm:"type:varchar(80);index;comment:钢厂编码" json:"code"`
	Aliases     StringList     `gorm:"type:json;comment:钢厂别名" json:"aliases"`
	Letter      string         `gorm:"type:varchar(100);index;comment:拼音/字母索引" json:"letter"`
	LogoURL     string         `gorm:"type:varchar(512);comment:Logo URL" json:"logo_url"`
	Sort        int            `gorm:"default:0;index;comment:排序" json:"sort"`
	Status      int            `gorm:"default:1;index;comment:状态 1启用 0停用" json:"status"`
	CreatedBy   uint64         `gorm:"comment:创建人ID" json:"created_by"`
	UpdatedBy   uint64         `gorm:"comment:更新人ID" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSteelMill) TableName() string {
	return "product_steel_mills"
}
