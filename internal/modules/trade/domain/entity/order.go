package entity

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusShipped        OrderStatus = "shipped"
	OrderStatusCompleted      OrderStatus = "completed"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       string         `gorm:"column:order_id;type:varchar(40);uniqueIndex;comment:订单号" json:"order_id"`
	BuyerID       uint64         `gorm:"index;comment:买家ID" json:"buyer_id"`
	BuyerUID      string         `gorm:"type:varchar(32);index;comment:买家UID" json:"buyer_uid"`
	SellerID      uint64         `gorm:"index;comment:卖家ID" json:"seller_id"`
	SellerUID     string         `gorm:"type:varchar(32);index;comment:卖家UID" json:"seller_uid"`
	ProductID     uint64         `gorm:"index;comment:商品ID" json:"product_id"`
	ProductSKUID  string         `gorm:"column:product_sku_id;type:varchar(40);index;comment:商品SKU" json:"product_sku_id"`
	Status        OrderStatus    `gorm:"type:varchar(32);index;comment:订单状态" json:"status"`
	TotalAmount   float64        `gorm:"type:decimal(14,2);comment:订单金额" json:"total_amount"`
	TotalQuantity int64          `gorm:"comment:总吨数" json:"total_quantity"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Items         []OrderItem    `gorm:"foreignKey:OrderRefID;references:ID" json:"items"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderID == "" {
		o.OrderID = "ORD" + time.Now().Format("20060102150405") + randomDigits(6)
	}
	if o.Status == "" {
		o.Status = OrderStatusPendingPayment
	}
	return nil
}

type OrderItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderRefID      uint64    `gorm:"column:order_ref_id;index;comment:订单自增ID" json:"order_ref_id"`
	ProductID       uint64    `gorm:"index;comment:商品ID" json:"product_id"`
	ProductSKUID    string    `gorm:"column:product_sku_id;type:varchar(40);index;comment:商品SKU" json:"product_sku_id"`
	ProductName     string    `gorm:"type:varchar(200);comment:商品名快照" json:"product_name"`
	ProductCategory string    `gorm:"type:varchar(100);comment:品类快照" json:"product_category"`
	Specs           string    `gorm:"type:varchar(100);comment:规格快照" json:"specs"`
	Price           float64   `gorm:"type:decimal(12,2);comment:下单单价" json:"price"`
	Quantity        int64     `gorm:"comment:下单吨数" json:"quantity"`
	Amount          float64   `gorm:"type:decimal(14,2);comment:小计" json:"amount"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
