package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// InquiryStatus 询价单状态
type InquiryStatus int

const (
	InquiryStatusDraft     InquiryStatus = 0 // 草稿
	InquiryStatusPending   InquiryStatus = 1 // 待报价
	InquiryStatusQuoted    InquiryStatus = 2 // 已报价
	InquiryStatusAccepted  InquiryStatus = 3 // 已接受
	InquiryStatusRejected  InquiryStatus = 4 // 已拒绝
	InquiryStatusExpired   InquiryStatus = 5 // 已过期
	InquiryStatusCancelled InquiryStatus = 6 // 已取消
)

// InquiryItem 询价单项
type InquiryItem struct {
	ProductID   uint64  `json:"product_id"`   // 商品ID（如果有）
	ProductName string  `json:"product_name"` // 商品名称
	Category    string  `json:"category"`     // 分类
	SteelMill   string  `json:"steel_mill"`   // 钢厂
	Material    string  `json:"material"`     // 材质
	Spec        string  `json:"spec"`         // 规格
	Quantity    float64 `json:"quantity"`     // 数量(吨)
	TargetPrice float64 `json:"target_price"` // 目标单价(元/吨)
	QuotePrice  float64 `json:"quote_price"`  // 报价单价(元/吨)
}

// Value 实现driver.Valuer接口
func (i InquiryItem) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan 实现sql.Scanner接口
func (i *InquiryItem) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, i)
}

// InquiryItems 询价单项列表
type InquiryItems []InquiryItem

// Value 实现driver.Valuer接口
func (items InquiryItems) Value() (driver.Value, error) {
	return json.Marshal(items)
}

// Scan 实现sql.Scanner接口
func (items *InquiryItems) Scan(value interface{}) error {
	if value == nil {
		*items = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, items)
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

// Inquiry 询价单实体
type Inquiry struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	InquiryNo      string         `gorm:"type:varchar(32);uniqueIndex;comment:询价单编号" json:"inquiry_no"`
	BuyerID        uint64         `gorm:"index;comment:买家企业ID" json:"buyer_id"`
	BuyerName      string         `gorm:"type:varchar(200);comment:买家企业名称" json:"buyer_name"`
	SellerID       uint64         `gorm:"index;comment:卖家企业ID" json:"seller_id"`
	SellerName     string         `gorm:"type:varchar(200);comment:卖家企业名称" json:"seller_name"`
	Items          InquiryItems   `gorm:"type:json;comment:询价单项" json:"items"`
	TotalQuantity  float64        `gorm:"type:decimal(10,2);comment:总数量(吨)" json:"total_quantity"`
	TotalTargetAmt float64        `gorm:"type:decimal(12,2);comment:目标总金额(元)" json:"total_target_amt"`
	TotalQuoteAmt  float64        `gorm:"type:decimal(12,2);comment:报价总金额(元)" json:"total_quote_amt"`
	Currency       string         `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	Status         InquiryStatus  `gorm:"type:tinyint;default:0;comment:询价单状态" json:"status"`
	ValidHours     int            `gorm:"default:72;comment:报价有效期(小时)" json:"valid_hours"`
	QuoteAt        *time.Time     `gorm:"comment:报价时间" json:"quote_at"`
	AcceptedAt     *time.Time     `gorm:"comment:接受时间" json:"accepted_at"`
	ExpiresAt      *time.Time     `gorm:"index;comment:过期时间" json:"expires_at"`
	RejectReason   string         `gorm:"type:varchar(500);comment:拒绝原因" json:"reject_reason"`
	Remark         Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy      uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy      uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Inquiry) TableName() string {
	return "inquiries"
}

// CanQuote 检查是否可以报价
func (i *Inquiry) CanQuote() bool {
	return i.Status == InquiryStatusPending && time.Now().Before(*i.ExpiresAt)
}

// CanAccept 检查是否可以接受报价
func (i *Inquiry) CanAccept() bool {
	return i.Status == InquiryStatusQuoted && time.Now().Before(*i.ExpiresAt)
}

// CanCancel 检查是否可以取消
func (i *Inquiry) CanCancel() bool {
	return i.Status == InquiryStatusDraft || 
		   i.Status == InquiryStatusPending
}

// IsExpired 检查是否过期
func (i *Inquiry) IsExpired() bool {
	return i.ExpiresAt != nil && time.Now().After(*i.ExpiresAt)
}

// BeforeCreate GORM钩子：创建前生成询价单编号
func (i *Inquiry) BeforeCreate(tx *gorm.DB) error {
	if i.InquiryNo == "" {
		i.InquiryNo = generateInquiryNo()
	}
	// 计算总数量和总目标金额
	var totalQty float64
	var totalTargetAmt float64
	for _, item := range i.Items {
		totalQty += item.Quantity
		totalTargetAmt += item.Quantity * item.TargetPrice
	}
	i.TotalQuantity = totalQty
	i.TotalTargetAmt = totalTargetAmt
	
	// 设置过期时间
	expiresAt := time.Now().Add(time.Duration(i.ValidHours) * time.Hour)
	i.ExpiresAt = &expiresAt
	
	return nil
}

// generateInquiryNo 生成询价单编号
func generateInquiryNo() string {
	return "IQ" + time.Now().Format("20060102150405") + randomDigits(4)
}

// randomDigits 生成指定长度的数字字符串
func randomDigits(length int) string {
	digits := "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = digits[time.Now().Nanosecond()%10]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(result)
}
