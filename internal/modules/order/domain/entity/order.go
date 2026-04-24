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

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusDraft           OrderStatus = 0  // 草稿
	OrderStatusPendingConfirm  OrderStatus = 1  // 待确认
	OrderStatusPendingContract OrderStatus = 2  // 待签合同
	OrderStatusPendingPayment  OrderStatus = 3  // 待支付
	OrderStatusPaid            OrderStatus = 4  // 已支付
	OrderStatusPendingDelivery OrderStatus = 5  // 待提货
	OrderStatusDelivering      OrderStatus = 6  // 配送中
	OrderStatusDelivered       OrderStatus = 7  // 已送达
	OrderStatusCompleted       OrderStatus = 8  // 已完成
	OrderStatusCancelled       OrderStatus = 9  // 已取消
	OrderStatusFrozen          OrderStatus = 10 // 已冻结
	OrderStatusException       OrderStatus = 11 // 异常
)

// OrderItem 订单项
type OrderItem struct {
	ProductID   uint64  `json:"product_id"`   // 商品ID
	ProductNo   string  `json:"product_no"`   // 商品编号
	ProductName string  `json:"product_name"` // 商品名称
	SteelMill   string  `json:"steel_mill"`   // 钢厂
	Category    string  `json:"category"`     // 分类
	BundleNo    string  `json:"bundle_no"`    // 捆包号
	Quantity    float64 `json:"quantity"`     // 数量(吨)
	UnitPrice   float64 `json:"unit_price"`   // 单价(元/吨)
	TotalAmount float64 `json:"total_amount"` // 小计金额
	TaxRate     float64 `json:"tax_rate"`     // 税率
	TaxAmount   float64 `json:"tax_amount"`   // 税额
}

// Value 实现driver.Valuer接口
func (i OrderItem) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan 实现sql.Scanner接口
func (i *OrderItem) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, i)
}

// OrderItems 订单项列表
type OrderItems []OrderItem

// Value 实现driver.Valuer接口
func (items OrderItems) Value() (driver.Value, error) {
	return json.Marshal(items)
}

// Scan 实现sql.Scanner接口
func (items *OrderItems) Scan(value interface{}) error {
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

// DeliveryInfo 配送信息
type DeliveryInfo struct {
	DriverName    string `json:"driver_name"`    // 司机姓名
	DriverIdCard  string `json:"driver_id_card"` // 司机身份证
	DriverPhone   string `json:"driver_phone"`   // 司机电话
	PlateNumber   string `json:"plate_number"`   // 车牌号
	PickupCode    string `json:"pickup_code"`    // 提货码
	WarehouseID   uint64 `json:"warehouse_id"`   // 仓库ID
	WarehouseName string `json:"warehouse_name"` // 仓库名称
	Address       string `json:"address"`        // 收货地址
	EstimatedTime string `json:"estimated_time"` // 预计送达时间
}

// Value 实现driver.Valuer接口
func (d DeliveryInfo) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan 实现sql.Scanner接口
func (d *DeliveryInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, d)
}

// PaymentInfo 支付信息
type PaymentInfo struct {
	PaymentMethod string    `json:"payment_method"` // 支付方式
	PaymentTime   time.Time `json:"payment_time"`   // 支付时间
	TransactionNo string    `json:"transaction_no"` // 交易流水号
	PayeeAccount  string    `json:"payee_account"`  // 收款人账户
	PayerAccount  string    `json:"payer_account"`  // 付款人账户
	VoucherURL    string    `json:"voucher_url"`    // 支付凭证URL
}

// Value 实现driver.Valuer接口
func (p PaymentInfo) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现sql.Scanner接口
func (p *PaymentInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// ContractInfo 合同信息
type ContractInfo struct {
	ContractNo   string    `json:"contract_no"`   // 合同编号
	ContractURL  string    `json:"contract_url"`  // 合同文件URL
	ContractHash string    `json:"contract_hash"` // 合同哈希值
	SignedAt     time.Time `json:"signed_at"`     // 签署时间
	BuyerSigned  bool      `json:"buyer_signed"`  // 买方已签
	SellerSigned bool      `json:"seller_signed"` // 卖方已签
}

// Value 实现driver.Valuer接口
func (c ContractInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现sql.Scanner接口
func (c *ContractInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
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

// Order 订单实体
type Order struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo            string         `gorm:"type:varchar(32);uniqueIndex;comment:订单编号" json:"order_no"`
	BuyerID            uint64         `gorm:"index;comment:买家企业ID" json:"buyer_id"`
	SellerEnterpriseID uint64         `gorm:"column:seller_enterprise_id;index;comment:卖家企业ID兼容字段" json:"seller_enterprise_id"`
	BuyerName          string         `gorm:"type:varchar(200);comment:买家企业名称" json:"buyer_name"`
	SellerID           uint64         `gorm:"index;comment:卖家企业ID" json:"seller_id"`
	SellerName         string         `gorm:"type:varchar(200);comment:卖家企业名称" json:"seller_name"`
	Items              OrderItems     `gorm:"type:json;comment:订单项" json:"items"`
	TotalQuantity      float64        `gorm:"type:decimal(10,2);comment:总数量(吨)" json:"total_quantity"`
	TotalAmount        float64        `gorm:"type:decimal(12,2);comment:总金额(元)" json:"total_amount"`
	TaxTotalAmount     float64        `gorm:"type:decimal(12,2);comment:含税总金额(元)" json:"tax_total_amount"`
	Currency           string         `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	Status             OrderStatus    `gorm:"type:tinyint;default:0;comment:订单状态" json:"status"`
	ContractInfo       *ContractInfo  `gorm:"type:json;comment:合同信息" json:"contract_info"`
	PaymentInfo        *PaymentInfo   `gorm:"type:json;comment:支付信息" json:"payment_info"`
	DeliveryInfo       *DeliveryInfo  `gorm:"type:json;comment:配送信息" json:"delivery_info"`
	DeliveryVoucher    string         `gorm:"type:varchar(500);comment:配送凭证URL" json:"delivery_voucher"`
	ReceiptVoucher     string         `gorm:"type:varchar(500);comment:签收凭证URL" json:"receipt_voucher"`
	ReceiptAt          *time.Time     `gorm:"comment:签收时间" json:"receipt_at"`
	ExpiresAt          *time.Time     `gorm:"index;comment:过期时间" json:"expires_at"`
	FinanceID          uint64         `gorm:"index;comment:关联融资申请ID" json:"finance_id"`
	Remark             Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy          uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt          time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy          uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// CanCancel 检查是否可以取消
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusDraft ||
		o.Status == OrderStatusPendingConfirm ||
		o.Status == OrderStatusPendingContract
}

// CanPay 检查是否可以支付
func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPendingPayment
}

// CanDeliver 检查是否可以提货
func (o *Order) CanDeliver() bool {
	return o.Status == OrderStatusPendingDelivery
}

// CanReceive 检查是否可以签收
func (o *Order) CanReceive() bool {
	return o.Status == OrderStatusDelivering || o.Status == OrderStatusDelivered
}

// BeforeCreate GORM钩子：创建前生成订单编号
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderNo == "" {
		o.OrderNo = generateOrderNo()
	}
	// 计算总数量和总金额
	var totalQty float64
	var totalAmt float64
	var taxAmt float64
	for _, item := range o.Items {
		totalQty += item.Quantity
		totalAmt += item.TotalAmount
		taxAmt += item.TaxAmount
	}
	o.TotalQuantity = totalQty
	o.TotalAmount = totalAmt
	o.TaxTotalAmount = totalAmt + taxAmt

	// 设置过期时间（24小时后）
	expiresAt := time.Now().Add(24 * time.Hour)
	o.ExpiresAt = &expiresAt

	return nil
}

// generateOrderNo 生成订单编号
func generateOrderNo() string {
	return "O" + time.Now().Format("20060102150405") + randomDigits(4)
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
