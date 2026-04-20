package entity

import (
	"crypto/rand"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type ProductStatus string

const (
	ProductStatusDraft         ProductStatus = "draft"
	ProductStatusPendingReview ProductStatus = "pending_review"
	ProductStatusApproved      ProductStatus = "approved"
	ProductStatusRejected      ProductStatus = "rejected"
	ProductStatusOffline       ProductStatus = "offline"
)

type Product struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SKUID             string         `gorm:"column:sku_id;type:varchar(40);uniqueIndex;comment:商品SKU" json:"sku_id"`
	SellerID          uint64         `gorm:"index;comment:卖家用户ID" json:"seller_id"`
	SellerUID         string         `gorm:"type:varchar(32);index;comment:卖家UID" json:"seller_uid"`
	SellerName        string         `gorm:"type:varchar(100);comment:卖家名称" json:"seller_name"`
	Name              string         `gorm:"type:varchar(200);index;comment:商品名称" json:"name"`
	Category          string         `gorm:"type:varchar(100);index;comment:品类" json:"category"`
	Specs             string         `gorm:"type:varchar(100);comment:规格品牌" json:"specs"`
	Price             float64        `gorm:"type:decimal(12,2);comment:单价" json:"price"`
	Unit              string         `gorm:"type:varchar(20);default:元/吨;comment:单位" json:"unit"`
	Stock             int64          `gorm:"comment:库存" json:"stock"`
	VisibleStock      int64          `gorm:"comment:可售库存" json:"visible_stock"`
	Sales             int64          `gorm:"comment:销量" json:"sales"`
	Description       string         `gorm:"type:text;comment:商品描述" json:"description"`
	ImageURLs         StringList     `gorm:"type:json;comment:商品主图" json:"image_urls"`
	QualificationURLs StringList     `gorm:"type:json;comment:审核资料" json:"qualification_urls"`
	Status            ProductStatus  `gorm:"type:varchar(32);index;comment:状态" json:"status"`
	AuditRemark       string         `gorm:"type:varchar(255);comment:审核备注" json:"audit_remark"`
	SubmittedAt       *time.Time     `json:"submitted_at"`
	ApprovedAt        *time.Time     `json:"approved_at"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.SKUID == "" {
		p.SKUID = "SKU" + time.Now().Format("20060102150405") + randomDigits(6)
	}
	if p.Unit == "" {
		p.Unit = "元/吨"
	}
	if p.Status == "" {
		p.Status = ProductStatusPendingReview
	}
	p.VisibleStock = p.Stock
	return nil
}

type ProductAuditLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID    uint64    `gorm:"index;comment:商品ID" json:"product_id"`
	ProductSKUID string    `gorm:"column:product_sku_id;type:varchar(40);index;comment:商品SKU" json:"product_sku_id"`
	Action       string    `gorm:"type:varchar(50);comment:动作" json:"action"`
	Status       string    `gorm:"type:varchar(32);comment:状态" json:"status"`
	OperatorID   uint64    `gorm:"index;comment:操作人ID" json:"operator_id"`
	OperatorName string    `gorm:"type:varchar(100);comment:操作人名称" json:"operator_name"`
	Remark       string    `gorm:"type:varchar(255);comment:备注" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductAuditLog) TableName() string {
	return "product_audit_logs"
}

func randomDigits(length int) string {
	digits := "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
