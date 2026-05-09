package entity

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// ProductStatus 商品状态
type ProductStatus int

const (
	ProductStatusDraft    ProductStatus = 0 // 草稿
	ProductStatusPending  ProductStatus = 1 // 待审核
	ProductStatusOnSale   ProductStatus = 2 // 已上架
	ProductStatusOffSale  ProductStatus = 3 // 已下架
	ProductStatusRejected ProductStatus = 4 // 审核拒绝
)

type GoodsType string

const (
	GoodsTypeSpot    GoodsType = "spot"
	GoodsTypeFutures GoodsType = "futures"
)

// ProductSpec 商品规格JSON结构
type ProductSpec struct {
	Material   string `json:"material"`   // 材质
	Spec       string `json:"spec"`       // 规格
	Thickness  string `json:"thickness"`  // 厚度
	Width      string `json:"width"`      // 宽度
	Length     string `json:"length"`     // 长度
	Surface    string `json:"surface"`    // 表面处理
	Edge       string `json:"edge"`       // 边部
	Processing string `json:"processing"` // 加工方式
}

// Value 实现driver.Valuer接口
func (s ProductSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现sql.Scanner接口
func (s *ProductSpec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// WarehouseInfo 仓库信息
type WarehouseInfo struct {
	WarehouseID           uint64 `json:"warehouse_id"`            // 仓库ID
	WarehouseServiceID    uint64 `json:"warehouse_service_id"`    // 仓储资源ID
	WarehouseEnterpriseID uint64 `json:"warehouse_enterprise_id"` // 仓储方企业ID
	WarehouseName         string `json:"warehouse_name"`          // 仓库名称
	WarehouseCode         string `json:"warehouse_code"`          // 仓库代码
	Address               string `json:"address"`                 // 仓库地址
	Province              string `json:"province"`                // 省份
	City                  string `json:"city"`                    // 城市
	District              string `json:"district"`                // 区县
}

// Value 实现driver.Valuer接口
func (w WarehouseInfo) Value() (driver.Value, error) {
	return json.Marshal(w)
}

// Scan 实现sql.Scanner接口
func (w *WarehouseInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, w)
}

// StringList JSON字符串数组
type StringList []string

// Value 实现driver.Valuer接口
func (s StringList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现sql.Scanner接口
func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if len(bytes) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Remark 备注字段JSON结构
type Remark map[string]interface{}

// Value 实现driver.Valuer接口
func (r Remark) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

// Scan 实现sql.Scanner接口
func (r *Remark) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

// Product 商品实体
type Product struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductNo             string         `gorm:"type:varchar(32);uniqueIndex;comment:商品编号" json:"product_no"`
	Name                  string         `gorm:"type:varchar(200);comment:商品名称" json:"name"`
	GoodsType             GoodsType      `gorm:"type:varchar(20);default:spot;index;comment:货品类型 spot现货 futures期货" json:"goods_type"`
	Category              string         `gorm:"type:varchar(50);comment:商品分类" json:"category"`
	SubCategory           string         `gorm:"type:varchar(50);comment:商品子分类" json:"sub_category"`
	SteelMill             string         `gorm:"type:varchar(100);comment:钢厂" json:"steel_mill"`
	Spec                  *ProductSpec   `gorm:"type:json;comment:商品规格" json:"spec"`
	WarehouseServiceID    uint64         `gorm:"index;comment:仓储资源ID" json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64         `gorm:"index;comment:仓储方企业ID" json:"warehouse_enterprise_id"`
	WarehouseInfo         *WarehouseInfo `gorm:"type:json;comment:仓库信息" json:"warehouse_info"`
	BundleNo              string         `gorm:"type:varchar(100);uniqueIndex;comment:捆包号" json:"bundle_no"`
	Quantity              float64        `gorm:"type:decimal(10,2);comment:库存数量(吨)" json:"quantity"`
	AvailableQty          float64        `gorm:"type:decimal(10,2);comment:可用数量(吨)" json:"available_qty"`
	LockedQty             float64        `gorm:"type:decimal(10,2);default:0;comment:锁定数量(吨)" json:"locked_qty"`
	UnitPrice             float64        `gorm:"type:decimal(10,2);comment:单价(元/吨)" json:"unit_price"`
	MinOrderQty           float64        `gorm:"type:decimal(10,2);comment:最小起订量(吨)" json:"min_order_qty"`
	Currency              string         `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	TaxIncluded           bool           `gorm:"default:true;comment:是否含税" json:"tax_included"`
	TaxRate               float64        `gorm:"type:decimal(5,2);default:13.00;comment:税率(%)" json:"tax_rate"`
	Status                ProductStatus  `gorm:"type:tinyint;default:0;comment:商品状态" json:"status"`
	EnterpriseID          uint64         `gorm:"index;comment:所属企业ID" json:"enterprise_id"`
	AuditRemark           string         `gorm:"type:varchar(255);comment:审核备注" json:"audit_remark"`
	ApprovedAt            *time.Time     `gorm:"comment:审核通过时间" json:"approved_at"`
	QualityCert           string         `gorm:"type:varchar(500);comment:质保书URL" json:"quality_cert"`
	Images                StringList     `gorm:"type:json;comment:商品图片" json:"images"`
	Description           string         `gorm:"type:text;comment:商品描述" json:"description"`
	DeliveryDays          int            `gorm:"default:3;comment:交货周期(天)" json:"delivery_days"`
	QuoteValidity         int            `gorm:"default:72;comment:报价有效期(小时)" json:"quote_validity"`
	Remark                Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy             uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt             time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy             uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// IsOnSale 检查商品是否在售
func (p *Product) IsOnSale() bool {
	if p.AvailableQty <= 0 {
		return false
	}
	if p.Status == ProductStatusOnSale {
		return true
	}
	return p.ApprovedAt != nil && p.Status != ProductStatusOffSale && p.Status != ProductStatusRejected
}

// CanOrder 检查是否可以下单
func (p *Product) CanOrder(qty float64) bool {
	return p.IsOnSale() && qty >= p.MinOrderQty && qty <= p.AvailableQty
}

// LockStock 锁定库存
func (p *Product) LockStock(qty float64) bool {
	if p.AvailableQty >= qty {
		p.AvailableQty -= qty
		p.LockedQty += qty
		return true
	}
	return false
}

// UnlockStock 解锁库存
func (p *Product) UnlockStock(qty float64) bool {
	if p.LockedQty >= qty {
		p.AvailableQty += qty
		p.LockedQty -= qty
		return true
	}
	return false
}

// DeductStock 扣减库存
func (p *Product) DeductStock(qty float64) bool {
	if p.LockedQty >= qty {
		p.LockedQty -= qty
		p.Quantity -= qty
		return true
	}
	return false
}

// ReturnStock 退回已扣减库存（取消/异常时回退）
func (p *Product) ReturnStock(qty float64) bool {
	p.Quantity += qty
	p.AvailableQty += qty
	return true
}

// BeforeCreate GORM钩子：创建前生成商品编号
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ProductNo == "" {
		p.ProductNo = generateProductNo()
	}
	if p.AvailableQty == 0 {
		p.AvailableQty = p.Quantity
	}
	return nil
}

// generateProductNo 生成商品编号
func generateProductNo() string {
	return "P" + time.Now().Format("20060102150405") + randomDigits(4)
}

// randomDigits 生成指定长度的数字字符串
func randomDigits(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
