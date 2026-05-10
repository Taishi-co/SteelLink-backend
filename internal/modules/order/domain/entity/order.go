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

	OrderStatusPendingWarehouseConfirm  OrderStatus = 12 // 待仓库确认
	OrderStatusPendingSellerConfirm     OrderStatus = 13 // 待卖方确认
	OrderStatusPendingBuyerConfirm      OrderStatus = 14 // 待买方最终确认
	OrderStatusLogisticsBidding         OrderStatus = 15 // 物流竞价中
	OrderStatusPendingLogisticsConfirm  OrderStatus = 16 // 待确认物流
	OrderStatusShipping                 OrderStatus = 17 // 运输中
	OrderStatusReceived                 OrderStatus = 18 // 已收货
	OrderStatusReconciled               OrderStatus = 19 // 已对账
	OrderStatusArchived                 OrderStatus = 20 // 已归档
	OrderStatusFinancingReview          OrderStatus = 21 // 托盘融资审批中
	OrderStatusPendingFinancingContract OrderStatus = 22 // 待融资合同
	OrderStatusPendingWarehouseReselect OrderStatus = 23 // 待重新选仓
	OrderStatusPendingLogisticsRebid    OrderStatus = 24 // 待重新发起物流
)

var validStatusTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusDraft:                    {OrderStatusPendingWarehouseConfirm, OrderStatusPendingConfirm, OrderStatusCancelled},
	OrderStatusPendingConfirm:           {OrderStatusPendingContract, OrderStatusCancelled},
	OrderStatusPendingContract:          {OrderStatusPendingPayment, OrderStatusCancelled},
	OrderStatusPendingPayment:           {OrderStatusPaid, OrderStatusFinancingReview, OrderStatusCancelled},
	OrderStatusPaid:                     {OrderStatusPendingDelivery, OrderStatusLogisticsBidding, OrderStatusPendingLogisticsRebid},
	OrderStatusPendingDelivery:          {OrderStatusDelivering, OrderStatusLogisticsBidding, OrderStatusPendingLogisticsRebid},
	OrderStatusDelivering:               {OrderStatusDelivered, OrderStatusReceived},
	OrderStatusDelivered:                {OrderStatusReceived},
	OrderStatusCompleted:                {OrderStatusArchived},
	OrderStatusCancelled:                {},
	OrderStatusFrozen:                   {OrderStatusCancelled},
	OrderStatusException:                {OrderStatusCancelled},
	OrderStatusPendingWarehouseConfirm:  {OrderStatusPendingSellerConfirm, OrderStatusPendingWarehouseReselect, OrderStatusCancelled},
	OrderStatusPendingWarehouseReselect: {OrderStatusPendingWarehouseConfirm, OrderStatusCancelled},
	OrderStatusPendingSellerConfirm:     {OrderStatusPendingBuyerConfirm, OrderStatusCancelled},
	OrderStatusPendingBuyerConfirm:      {OrderStatusPendingContract, OrderStatusCancelled},
	OrderStatusLogisticsBidding:         {OrderStatusPendingLogisticsConfirm, OrderStatusPendingLogisticsRebid},
	OrderStatusPendingLogisticsConfirm:  {OrderStatusShipping, OrderStatusPendingDelivery},
	OrderStatusShipping:                 {OrderStatusDelivered, OrderStatusReceived},
	OrderStatusReceived:                 {OrderStatusReconciled},
	OrderStatusReconciled:               {OrderStatusArchived},
	OrderStatusArchived:                 {},
	OrderStatusFinancingReview:          {OrderStatusPendingFinancingContract, OrderStatusPendingPayment, OrderStatusCancelled},
	OrderStatusPendingFinancingContract: {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPendingLogisticsRebid:    {OrderStatusLogisticsBidding, OrderStatusCancelled},
}

func ValidateStatusTransition(from, to OrderStatus) bool {
	if to == OrderStatusCancelled || to == OrderStatusFrozen || to == OrderStatusException {
		return true
	}
	allowed, ok := validStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

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
	DriverName            string `json:"driver_name"`             // 司机姓名
	DriverIdCard          string `json:"driver_id_card"`          // 司机身份证
	DriverPhone           string `json:"driver_phone"`            // 司机电话
	PlateNumber           string `json:"plate_number"`            // 车牌号
	PickupCode            string `json:"pickup_code"`             // 提货码
	WarehouseID           uint64 `json:"warehouse_id"`            // 仓库ID
	WarehouseServiceID    uint64 `json:"warehouse_service_id"`    // 仓储资源ID
	WarehouseEnterpriseID uint64 `json:"warehouse_enterprise_id"` // 仓储方企业ID
	WarehouseName         string `json:"warehouse_name"`          // 仓库名称
	WarehouseCode         string `json:"warehouse_code"`          // 仓库代码
	WarehouseProvince     string `json:"warehouse_province"`      // 仓库省份
	WarehouseCity         string `json:"warehouse_city"`          // 仓库城市
	WarehouseDistrict     string `json:"warehouse_district"`      // 仓库区县
	WarehouseAddress      string `json:"warehouse_address"`       // 仓库地址
	Address               string `json:"address"`                 // 收货地址
	EstimatedTime         string `json:"estimated_time"`          // 预计送达时间
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
	UploadedBy    uint64    `json:"uploaded_by"`    // 上传人
	UploadedAt    time.Time `json:"uploaded_at"`    // 上传时间
	CapturedAt    string    `json:"captured_at"`    // 凭证水印时间
	GPSLatitude   *float64  `json:"gps_latitude"`   // 凭证GPS纬度
	GPSLongitude  *float64  `json:"gps_longitude"`  // 凭证GPS经度
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

// VoucherEvidence 凭证留痕信息
type VoucherEvidence struct {
	VoucherURL   string    `json:"voucher_url"`   // 凭证URL
	UploadedBy   uint64    `json:"uploaded_by"`   // 上传人
	UploadedAt   time.Time `json:"uploaded_at"`   // 上传时间
	CapturedAt   string    `json:"captured_at"`   // 凭证水印时间
	GPSLatitude  *float64  `json:"gps_latitude"`  // 凭证GPS纬度
	GPSLongitude *float64  `json:"gps_longitude"` // 凭证GPS经度
}

type OrderEventPayload map[string]interface{}

func (p OrderEventPayload) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *OrderEventPayload) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// Value 实现driver.Valuer接口
func (v VoucherEvidence) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan 实现sql.Scanner接口
func (v *VoucherEvidence) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, v)
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
	ID                 uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo            string           `gorm:"type:varchar(32);uniqueIndex;comment:订单编号" json:"order_no"`
	BuyerID            uint64           `gorm:"index;comment:买家企业ID" json:"buyer_id"`
	SellerEnterpriseID uint64           `gorm:"column:seller_enterprise_id;index;comment:卖家企业ID兼容字段" json:"seller_enterprise_id"`
	BuyerName          string           `gorm:"type:varchar(200);comment:买家企业名称" json:"buyer_name"`
	SellerID           uint64           `gorm:"index;comment:卖家企业ID" json:"seller_id"`
	SellerName         string           `gorm:"type:varchar(200);comment:卖家企业名称" json:"seller_name"`
	Items              OrderItems       `gorm:"type:json;comment:订单项" json:"items"`
	TotalQuantity      float64          `gorm:"type:decimal(10,2);comment:总数量(吨)" json:"total_quantity"`
	TotalAmount        float64          `gorm:"type:decimal(12,2);comment:总金额(元)" json:"total_amount"`
	TaxTotalAmount     float64          `gorm:"type:decimal(12,2);comment:含税总金额(元)" json:"tax_total_amount"`
	Currency           string           `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	Status             OrderStatus      `gorm:"type:tinyint;default:3;comment:订单状态" json:"status"`
	ContractInfo       *ContractInfo    `gorm:"type:json;comment:合同信息" json:"contract_info"`
	PaymentInfo        *PaymentInfo     `gorm:"type:json;comment:支付信息" json:"payment_info"`
	DeliveryInfo       *DeliveryInfo    `gorm:"type:json;comment:配送信息" json:"delivery_info"`
	DeliveryVoucher    string           `gorm:"type:varchar(500);comment:配送凭证URL" json:"delivery_voucher"`
	ReceiptVoucher     string           `gorm:"type:varchar(500);comment:签收凭证URL" json:"receipt_voucher"`
	ReceiptInfo        *VoucherEvidence `gorm:"type:json;comment:签收凭证留痕信息" json:"receipt_info"`
	ReceiptAt          *time.Time       `gorm:"comment:签收时间" json:"receipt_at"`
	ExpiresAt          *time.Time       `gorm:"index;comment:过期时间" json:"expires_at"`
	FinanceID          uint64           `gorm:"index;comment:关联融资申请ID" json:"finance_id"`
	Remark             Remark           `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy          uint64           `gorm:"comment:创建者" json:"created_by"`
	CreatedAt          time.Time        `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy          uint64           `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt          time.Time        `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

type OrderEvent struct {
	ID                uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID           uint64            `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo           string            `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	EventType         string            `gorm:"type:varchar(64);not null;index;comment:事件类型" json:"event_type"`
	ActorUserID       uint64            `gorm:"index;comment:操作者用户ID" json:"actor_user_id"`
	ActorEnterpriseID uint64            `gorm:"index;comment:操作者企业ID" json:"actor_enterprise_id"`
	ActorService      string            `gorm:"type:varchar(32);comment:操作者服务身份" json:"actor_service"`
	FromStatus        OrderStatus       `gorm:"type:tinyint;comment:变更前状态" json:"from_status"`
	ToStatus          OrderStatus       `gorm:"type:tinyint;comment:变更后状态" json:"to_status"`
	Payload           OrderEventPayload `gorm:"type:json;comment:事件载荷" json:"payload"`
	CreatedAt         time.Time         `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (OrderEvent) TableName() string {
	return "order_events"
}

type LogisticsBiddingStatus string

const (
	LogisticsBiddingStatusOpen     LogisticsBiddingStatus = "open"
	LogisticsBiddingStatusSelected LogisticsBiddingStatus = "selected"
	LogisticsBiddingStatusClosed   LogisticsBiddingStatus = "closed"
)

type LogisticsBidding struct {
	ID                     uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID                uint64                 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo                string                 `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	SellerEnterpriseID     uint64                 `gorm:"index;comment:卖方企业ID" json:"seller_enterprise_id"`
	WarehouseEnterpriseID  uint64                 `gorm:"index;comment:仓储方企业ID" json:"warehouse_enterprise_id"`
	PickupAddress          string                 `gorm:"type:varchar(500);comment:提货地址" json:"pickup_address"`
	DestinationAddress     string                 `gorm:"type:varchar(500);comment:目的地地址" json:"destination_address"`
	Requirement            string                 `gorm:"type:varchar(1000);comment:承运要求" json:"requirement"`
	FreightPayer           string                 `gorm:"type:varchar(16);default:buyer;comment:运费承担方 buyer/seller" json:"freight_payer"`
	Status                 LogisticsBiddingStatus `gorm:"type:varchar(32);index;comment:状态" json:"status"`
	SelectedQuoteID        uint64                 `gorm:"index;comment:选中报价ID" json:"selected_quote_id"`
	BuyerConfirmedQuoteID  uint64                 `gorm:"comment:买方确认报价ID" json:"buyer_confirmed_quote_id"`
	SellerConfirmedQuoteID uint64                 `gorm:"comment:卖方确认报价ID" json:"seller_confirmed_quote_id"`
	CreatedBy              uint64                 `gorm:"comment:创建者" json:"created_by"`
	CreatedAt              time.Time              `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt              time.Time              `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (LogisticsBidding) TableName() string {
	return "logistics_biddings"
}

type LogisticsBidQuote struct {
	ID                    uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	BiddingID             uint64    `gorm:"index;not null;comment:竞价ID" json:"bidding_id"`
	OrderID               uint64    `gorm:"index;not null;comment:订单ID" json:"order_id"`
	LogisticsEnterpriseID uint64    `gorm:"index;comment:物流企业ID" json:"logistics_enterprise_id"`
	LogisticsName         string    `gorm:"type:varchar(200);comment:物流企业名称" json:"logistics_name"`
	VehicleInfo           string    `gorm:"type:varchar(500);comment:车辆信息" json:"vehicle_info"`
	DriverInfo            string    `gorm:"type:varchar(500);comment:司机信息" json:"driver_info"`
	Amount                float64   `gorm:"type:decimal(12,2);comment:报价金额" json:"amount"`
	EstimatedPickupAt     string    `gorm:"type:varchar(64);comment:预计提货时间" json:"estimated_pickup_at"`
	EstimatedDeliveryAt   string    `gorm:"type:varchar(64);comment:预计送达时间" json:"estimated_delivery_at"`
	Remark                string    `gorm:"type:varchar(1000);comment:备注" json:"remark"`
	CreatedBy             uint64    `gorm:"comment:创建者" json:"created_by"`
	CreatedAt             time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (LogisticsBidQuote) TableName() string {
	return "logistics_bid_quotes"
}

type ShipmentStatus string

const (
	ShipmentStatusPendingPickup ShipmentStatus = "pending_pickup"
	ShipmentStatusPickedUp      ShipmentStatus = "picked_up"
	ShipmentStatusDelivered     ShipmentStatus = "delivered"
	ShipmentStatusSigned        ShipmentStatus = "signed"
)

type ShipmentTask struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID               uint64         `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo               string         `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	LogisticsEnterpriseID uint64         `gorm:"index;comment:物流企业ID" json:"logistics_enterprise_id"`
	LogisticsQuoteID      uint64         `gorm:"index;comment:物流报价ID" json:"logistics_quote_id"`
	Status                ShipmentStatus `gorm:"type:varchar(32);index;comment:运输状态" json:"status"`
	PickupVoucher         string         `gorm:"type:varchar(500);comment:提货凭证" json:"pickup_voucher"`
	DeliveryVoucher       string         `gorm:"type:varchar(500);comment:送达凭证" json:"delivery_voucher"`
	SignatureVoucher      string         `gorm:"type:varchar(500);comment:签收回单" json:"signature_voucher"`
	PickedUpAt            *time.Time     `gorm:"comment:提货时间" json:"picked_up_at"`
	DeliveredAt           *time.Time     `gorm:"comment:送达时间" json:"delivered_at"`
	SignedAt              *time.Time     `gorm:"comment:签收回传时间" json:"signed_at"`
	CreatedAt             time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ShipmentTask) TableName() string {
	return "shipment_tasks"
}

type FinancingRequestStatus string

const (
	FinancingRequestStatusSubmitted          FinancingRequestStatus = "submitted"
	FinancingRequestStatusRiskApproved       FinancingRequestStatus = "risk_approved"
	FinancingRequestStatusRiskRejected       FinancingRequestStatus = "risk_rejected"
	FinancingRequestStatusBuyerConfirmed     FinancingRequestStatus = "buyer_confirmed"
	FinancingRequestStatusContractGenerated  FinancingRequestStatus = "contract_generated"
	FinancingRequestStatusBuyerSignedFeePaid FinancingRequestStatus = "buyer_signed_fee_paid"
	FinancingRequestStatusDisbursed          FinancingRequestStatus = "disbursed"
	FinancingRequestStatusPledged            FinancingRequestStatus = "pledged"
	FinancingRequestStatusReleased           FinancingRequestStatus = "released"
)

type OrderFinancingRequest struct {
	ID                  uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID             uint64                 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo             string                 `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	BuyerEnterpriseID   uint64                 `gorm:"index;comment:买方企业ID" json:"buyer_enterprise_id"`
	SellerEnterpriseID  uint64                 `gorm:"index;comment:卖方企业ID" json:"seller_enterprise_id"`
	FinanceServiceID    uint64                 `gorm:"index;comment:金融服务产品ID" json:"finance_service_id"`
	FundingEnterpriseID uint64                 `gorm:"index;comment:资金方企业ID" json:"funding_enterprise_id"`
	RequestedAmount     float64                `gorm:"type:decimal(14,2);comment:申请融资金额" json:"requested_amount"`
	ApprovedAmount      float64                `gorm:"type:decimal(14,2);comment:授信金额" json:"approved_amount"`
	TermDays            int                    `gorm:"comment:账期天数" json:"term_days"`
	AnnualRate          float64                `gorm:"type:decimal(8,4);comment:年化费率" json:"annual_rate"`
	ServiceFee          float64                `gorm:"type:decimal(12,2);comment:手续费" json:"service_fee"`
	Status              FinancingRequestStatus `gorm:"type:varchar(40);index;comment:融资状态" json:"status"`
	RiskResult          string                 `gorm:"type:varchar(32);comment:风控结果" json:"risk_result"`
	RiskRemark          string                 `gorm:"type:varchar(1000);comment:风控备注" json:"risk_remark"`
	BuyerConfirmedAt    *time.Time             `gorm:"comment:买方确认金额和定价时间" json:"buyer_confirmed_at"`
	DebtContractNo      string                 `gorm:"type:varchar(64);comment:债务合同编号" json:"debt_contract_no"`
	DebtContractURL     string                 `gorm:"type:varchar(500);comment:债务合同文件" json:"debt_contract_url"`
	DebtContractHash    string                 `gorm:"type:varchar(128);comment:债务合同哈希" json:"debt_contract_hash"`
	BuyerSignedAt       *time.Time             `gorm:"comment:买方签署时间" json:"buyer_signed_at"`
	FeeVoucherURL       string                 `gorm:"type:varchar(500);comment:手续费支付凭证" json:"fee_voucher_url"`
	FeePaidAt           *time.Time             `gorm:"comment:手续费支付时间" json:"fee_paid_at"`
	FundingCommission   float64                `gorm:"type:decimal(12,2);comment:资金方平台分佣金额" json:"funding_commission"`
	CommissionVoucher   string                 `gorm:"type:varchar(500);comment:资金方分佣凭证" json:"commission_voucher"`
	CommissionPaidAt    *time.Time             `gorm:"comment:资金方分佣支付时间" json:"commission_paid_at"`
	DisbursementVoucher string                 `gorm:"type:varchar(500);comment:放款凭证" json:"disbursement_voucher"`
	DisbursedAt         *time.Time             `gorm:"comment:放款时间" json:"disbursed_at"`
	PledgeVoucherURL    string                 `gorm:"type:varchar(500);comment:货权质押凭证" json:"pledge_voucher_url"`
	PledgeRemark        string                 `gorm:"type:varchar(1000);comment:质押备注" json:"pledge_remark"`
	PledgedAt           *time.Time             `gorm:"comment:质押记录时间" json:"pledged_at"`
	ReleaseCondition    string                 `gorm:"type:varchar(1000);comment:货权释放条件" json:"release_condition"`
	ReleaseReason       string                 `gorm:"type:varchar(1000);comment:货权释放原因" json:"release_reason"`
	ReleasedAt          *time.Time             `gorm:"comment:释放时间" json:"released_at"`
	CreatedBy           uint64                 `gorm:"comment:创建者" json:"created_by"`
	CreatedAt           time.Time              `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy           uint64                 `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt           time.Time              `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (OrderFinancingRequest) TableName() string {
	return "order_financing_requests"
}

type PledgeReleaseEvent struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FinancingRequestID uint64    `gorm:"index;not null;comment:融资申请ID" json:"financing_request_id"`
	OrderID            uint64    `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo            string    `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	EventType          string    `gorm:"type:varchar(64);index;comment:事件类型" json:"event_type"`
	Condition          string    `gorm:"type:varchar(1000);comment:释放条件" json:"condition"`
	Reason             string    `gorm:"type:varchar(1000);comment:释放原因" json:"reason"`
	VoucherURL         string    `gorm:"type:varchar(500);comment:凭证" json:"voucher_url"`
	ActorUserID        uint64    `gorm:"index;comment:操作者用户ID" json:"actor_user_id"`
	ActorEnterpriseID  uint64    `gorm:"index;comment:操作者企业ID" json:"actor_enterprise_id"`
	CreatedAt          time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (PledgeReleaseEvent) TableName() string {
	return "pledge_release_events"
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
