package entity

import (
	"time"

	"gorm.io/gorm"
)

type WarehouseMovementType string

const (
	WarehouseMovementInbound  WarehouseMovementType = "inbound"
	WarehouseMovementOutbound WarehouseMovementType = "outbound"
)

type WarehouseMovementStatus string

const (
	WarehouseMovementPending    WarehouseMovementStatus = "pending"
	WarehouseMovementProcessing WarehouseMovementStatus = "processing"
	WarehouseMovementCompleted  WarehouseMovementStatus = "completed"
	WarehouseMovementCancelled  WarehouseMovementStatus = "cancelled"
)

type LogisticsOrderStatus string

const (
	LogisticsOrderPending   LogisticsOrderStatus = "pending"
	LogisticsOrderAccepted  LogisticsOrderStatus = "accepted"
	LogisticsOrderPickedUp  LogisticsOrderStatus = "picked_up"
	LogisticsOrderInTransit LogisticsOrderStatus = "in_transit"
	LogisticsOrderDelivered LogisticsOrderStatus = "delivered"
	LogisticsOrderCompleted LogisticsOrderStatus = "completed"
	LogisticsOrderCancelled LogisticsOrderStatus = "cancelled"
	LogisticsOrderException LogisticsOrderStatus = "exception"
)

type GuaranteeStatus string

const (
	GuaranteeStatusActive     GuaranteeStatus = "active"
	GuaranteeStatusProcessing GuaranteeStatus = "processing"
	GuaranteeStatusResolved   GuaranteeStatus = "resolved"
	GuaranteeStatusRejected   GuaranteeStatus = "rejected"
)

type WarehouseInventoryMovement struct {
	ID                    uint64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	MovementNo            string                  `gorm:"type:varchar(64);uniqueIndex;not null;comment:出入库记录号" json:"movement_no"`
	WarehouseServiceID    uint64                  `gorm:"index;not null;comment:仓储资源ID" json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64                  `gorm:"index;not null;comment:仓储企业ID" json:"warehouse_enterprise_id"`
	RequestEnterpriseID   uint64                  `gorm:"index;comment:委托企业ID" json:"request_enterprise_id"`
	ProductID             uint64                  `gorm:"index;comment:商品ID" json:"product_id"`
	ProductName           string                  `gorm:"type:varchar(200);comment:商品名称" json:"product_name"`
	BundleNo              string                  `gorm:"type:varchar(100);index;comment:捆包号" json:"bundle_no"`
	OrderID               uint64                  `gorm:"index;comment:关联订单ID" json:"order_id"`
	OrderNo               string                  `gorm:"type:varchar(64);index;comment:关联订单号" json:"order_no"`
	MovementType          WarehouseMovementType   `gorm:"type:varchar(32);index;not null;comment:inbound/outbound" json:"movement_type"`
	Quantity              float64                 `gorm:"type:decimal(12,2);not null;comment:数量吨" json:"quantity"`
	Status                WarehouseMovementStatus `gorm:"type:varchar(32);index;not null;default:pending;comment:状态" json:"status"`
	ScheduledAt           string                  `gorm:"type:varchar(64);comment:计划作业时间" json:"scheduled_at"`
	CompletedAt           *time.Time              `gorm:"comment:完成时间" json:"completed_at"`
	Remark                string                  `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedBy             uint64                  `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy             uint64                  `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt             time.Time               `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time               `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (WarehouseInventoryMovement) TableName() string {
	return "warehouse_inventory_movements"
}

type LogisticsOrder struct {
	ID                  uint64               `gorm:"primaryKey;autoIncrement" json:"id"`
	LogisticsNo         string               `gorm:"type:varchar(64);uniqueIndex;not null;comment:物流委托单号" json:"logistics_no"`
	OrderID             uint64               `gorm:"index;comment:交易订单ID" json:"order_id"`
	OrderNo             string               `gorm:"type:varchar(64);index;comment:交易订单号" json:"order_no"`
	BuyerEnterpriseID   uint64               `gorm:"index;not null;comment:发单企业ID" json:"buyer_enterprise_id"`
	LogisticsServiceID  uint64               `gorm:"index;not null;comment:物流服务ID" json:"logistics_service_id"`
	CarrierEnterpriseID uint64               `gorm:"index;comment:承运企业ID" json:"carrier_enterprise_id"`
	WarehouseServiceID  uint64               `gorm:"index;comment:提货仓库服务ID" json:"warehouse_service_id"`
	Origin              string               `gorm:"type:varchar(255);not null;comment:提货点" json:"origin"`
	Destination         string               `gorm:"type:varchar(255);not null;comment:目的地" json:"destination"`
	CargoName           string               `gorm:"type:varchar(200);comment:货物名称" json:"cargo_name"`
	Weight              float64              `gorm:"type:decimal(12,2);not null;comment:重量吨" json:"weight"`
	VehicleType         string               `gorm:"type:varchar(100);comment:车型" json:"vehicle_type"`
	EstimatedPrice      float64              `gorm:"type:decimal(12,2);comment:估算价格" json:"estimated_price"`
	FinalPrice          float64              `gorm:"type:decimal(12,2);comment:最终价格" json:"final_price"`
	Status              LogisticsOrderStatus `gorm:"type:varchar(32);index;not null;default:pending;comment:状态" json:"status"`
	DriverName          string               `gorm:"type:varchar(100);comment:司机姓名" json:"driver_name"`
	DriverPhone         string               `gorm:"type:varchar(50);comment:司机电话" json:"driver_phone"`
	PlateNumber         string               `gorm:"type:varchar(50);comment:车牌号" json:"plate_number"`
	GuaranteeType       string               `gorm:"type:varchar(100);comment:保障方式" json:"guarantee_type"`
	AcceptedBy          uint64               `gorm:"comment:接单人" json:"accepted_by"`
	AcceptedAt          *time.Time           `gorm:"comment:接单时间" json:"accepted_at"`
	CreatedBy           uint64               `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy           uint64               `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt           time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt       `gorm:"index" json:"-"`
}

func (LogisticsOrder) TableName() string {
	return "logistics_orders"
}

type LogisticsTrackingEvent struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	LogisticsOrderID uint64    `gorm:"index;not null;comment:物流委托ID" json:"logistics_order_id"`
	EventType        string    `gorm:"type:varchar(64);index;not null;comment:节点类型" json:"event_type"`
	Location         string    `gorm:"type:varchar(255);comment:位置" json:"location"`
	Description      string    `gorm:"type:varchar(500);comment:描述" json:"description"`
	OccurredAt       time.Time `gorm:"index;comment:发生时间" json:"occurred_at"`
	CreatedBy        uint64    `gorm:"comment:创建者" json:"created_by"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (LogisticsTrackingEvent) TableName() string {
	return "logistics_tracking_events"
}

type FulfillmentGuarantee struct {
	ID               uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	GuaranteeNo      string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:保障单号" json:"guarantee_no"`
	OrderID          uint64          `gorm:"index;comment:订单ID" json:"order_id"`
	LogisticsOrderID uint64          `gorm:"index;comment:物流委托ID" json:"logistics_order_id"`
	EnterpriseID     uint64          `gorm:"index;not null;comment:申请企业ID" json:"enterprise_id"`
	GuaranteeType    string          `gorm:"type:varchar(100);not null;comment:保障类型" json:"guarantee_type"`
	Status           GuaranteeStatus `gorm:"type:varchar(32);index;not null;default:active;comment:状态" json:"status"`
	CoverageAmount   float64         `gorm:"type:decimal(12,2);comment:保障金额" json:"coverage_amount"`
	Description      string          `gorm:"type:varchar(500);comment:说明" json:"description"`
	Resolution       string          `gorm:"type:varchar(500);comment:处理结果" json:"resolution"`
	CreatedBy        uint64          `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy        uint64          `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (FulfillmentGuarantee) TableName() string {
	return "fulfillment_guarantees"
}
