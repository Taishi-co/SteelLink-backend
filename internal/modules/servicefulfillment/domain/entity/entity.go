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
	LogisticsOrderTaskCreated       LogisticsOrderStatus = "task_created"
	LogisticsOrderPickupScheduled   LogisticsOrderStatus = "pickup_scheduled"
	LogisticsOrderArrivalSignedIn   LogisticsOrderStatus = "arrival_signed_in"
	LogisticsOrderOutboundConfirmed LogisticsOrderStatus = "outbound_confirmed"
	LogisticsOrderLoadingConfirmed  LogisticsOrderStatus = "loading_confirmed"
	LogisticsOrderInTransit         LogisticsOrderStatus = "in_transit"
	LogisticsOrderArrivalConfirmed  LogisticsOrderStatus = "arrival_confirmed"
	LogisticsOrderReceiptSigned     LogisticsOrderStatus = "receipt_signed"
	LogisticsOrderInboundCompleted  LogisticsOrderStatus = "inbound_completed"
	LogisticsOrderCompleted         LogisticsOrderStatus = "completed"
	LogisticsOrderCancelled         LogisticsOrderStatus = "cancelled"
	LogisticsOrderException         LogisticsOrderStatus = "exception"

	// Legacy statuses kept for backward compatibility with old rows and UI.
	LogisticsOrderPending   LogisticsOrderStatus = "pending"
	LogisticsOrderAccepted  LogisticsOrderStatus = "accepted"
	LogisticsOrderPickedUp  LogisticsOrderStatus = "picked_up"
	LogisticsOrderDelivered LogisticsOrderStatus = "delivered"
)

type ResponsibilityParty string

const (
	ResponsibilityPartyUnknown   ResponsibilityParty = "unknown"
	ResponsibilityPartyBuyer     ResponsibilityParty = "buyer"
	ResponsibilityPartySeller    ResponsibilityParty = "seller"
	ResponsibilityPartyWarehouse ResponsibilityParty = "warehouse"
	ResponsibilityPartyLogistics ResponsibilityParty = "logistics"
)

func NormalizeLogisticsOrderStatus(status LogisticsOrderStatus) LogisticsOrderStatus {
	switch status {
	case LogisticsOrderPending:
		return LogisticsOrderTaskCreated
	case LogisticsOrderAccepted:
		return LogisticsOrderPickupScheduled
	case LogisticsOrderPickedUp:
		return LogisticsOrderInTransit
	case LogisticsOrderDelivered:
		return LogisticsOrderArrivalConfirmed
	default:
		return status
	}
}

func ResponsibilityForLogisticsStatus(status LogisticsOrderStatus) ResponsibilityParty {
	switch NormalizeLogisticsOrderStatus(status) {
	case LogisticsOrderTaskCreated, LogisticsOrderLoadingConfirmed, LogisticsOrderInTransit, LogisticsOrderReceiptSigned, LogisticsOrderCompleted:
		return ResponsibilityPartyLogistics
	case LogisticsOrderPickupScheduled, LogisticsOrderArrivalSignedIn, LogisticsOrderOutboundConfirmed, LogisticsOrderArrivalConfirmed:
		return ResponsibilityPartyWarehouse
	case LogisticsOrderInboundCompleted:
		return ResponsibilityPartyBuyer
	default:
		return ResponsibilityPartyUnknown
	}
}

func CanTransitionLogisticsOrderStatus(from, to LogisticsOrderStatus) bool {
	from = NormalizeLogisticsOrderStatus(from)
	to = NormalizeLogisticsOrderStatus(to)
	if from == to {
		return true
	}
	if to == LogisticsOrderCancelled || to == LogisticsOrderException {
		return true
	}
	allowed := map[LogisticsOrderStatus][]LogisticsOrderStatus{
		LogisticsOrderTaskCreated:       {LogisticsOrderPickupScheduled},
		LogisticsOrderPickupScheduled:   {LogisticsOrderArrivalSignedIn},
		LogisticsOrderArrivalSignedIn:   {LogisticsOrderOutboundConfirmed},
		LogisticsOrderOutboundConfirmed: {LogisticsOrderLoadingConfirmed},
		LogisticsOrderLoadingConfirmed:  {LogisticsOrderInTransit},
		LogisticsOrderInTransit:         {LogisticsOrderArrivalConfirmed},
		LogisticsOrderArrivalConfirmed:  {LogisticsOrderInboundCompleted},
		LogisticsOrderInboundCompleted:  {LogisticsOrderReceiptSigned},
		LogisticsOrderReceiptSigned:     {LogisticsOrderCompleted},
		LogisticsOrderCompleted:         {},
	}
	nextStatuses, ok := allowed[from]
	if !ok {
		return false
	}
	for _, next := range nextStatuses {
		if next == to {
			return true
		}
	}
	return false
}

type GuaranteeStatus string

const (
	GuaranteeStatusActive     GuaranteeStatus = "active"
	GuaranteeStatusProcessing GuaranteeStatus = "processing"
	GuaranteeStatusResolved   GuaranteeStatus = "resolved"
	GuaranteeStatusRejected   GuaranteeStatus = "rejected"
)

type WarehouseReceiptPledgeStatus string

const (
	WarehouseReceiptPledgeSubmitted       WarehouseReceiptPledgeStatus = "submitted"
	WarehouseReceiptPledgeRiskApproved    WarehouseReceiptPledgeStatus = "risk_approved"
	WarehouseReceiptPledgeRiskRejected    WarehouseReceiptPledgeStatus = "risk_rejected"
	WarehouseReceiptPledgeAgreementSigned WarehouseReceiptPledgeStatus = "agreement_signed"
	WarehouseReceiptPledgeDisbursed       WarehouseReceiptPledgeStatus = "disbursed"
	WarehouseReceiptPledgeReleasing       WarehouseReceiptPledgeStatus = "releasing"
	WarehouseReceiptPledgeReleased        WarehouseReceiptPledgeStatus = "released"
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

type WarehouseInventorySnapshot struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SnapshotNo            string         `gorm:"type:varchar(64);uniqueIndex;not null;comment:库存快照编号" json:"snapshot_no"`
	WarehouseServiceID    uint64         `gorm:"index;not null;comment:仓储资源ID" json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64         `gorm:"index;not null;comment:仓储企业ID" json:"warehouse_enterprise_id"`
	ProductName           string         `gorm:"type:varchar(200);not null;comment:商品名称" json:"product_name"`
	BundleNo              string         `gorm:"type:varchar(100);index;comment:捆包号/批次号" json:"bundle_no"`
	BatchNo               string         `gorm:"type:varchar(100);index;comment:批次号" json:"batch_no"`
	LocationCode          string         `gorm:"type:varchar(100);index;comment:库位" json:"location_code"`
	Quantity              float64        `gorm:"type:decimal(12,2);not null;comment:当前库存数量" json:"quantity"`
	FrozenQuantity        float64        `gorm:"type:decimal(12,2);comment:质押/冻结数量" json:"frozen_quantity"`
	QualityStatus         string         `gorm:"type:varchar(32);index;default:unchecked;comment:质检状态 unchecked/passed/failed" json:"quality_status"`
	PledgeStatus          string         `gorm:"type:varchar(32);index;default:none;comment:质押状态 none/frozen/released" json:"pledge_status"`
	ExternalSystem        string         `gorm:"type:varchar(64);comment:外部系统 MRP/WMS" json:"external_system"`
	ExternalRefNo         string         `gorm:"type:varchar(128);index;comment:外部单号" json:"external_ref_no"`
	SyncStatus            string         `gorm:"type:varchar(32);index;default:manual;comment:同步状态 manual/synced/failed" json:"sync_status"`
	Remark                string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedBy             uint64         `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy             uint64         `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt             time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WarehouseInventorySnapshot) TableName() string {
	return "warehouse_inventory_snapshots"
}

type LogisticsOrder struct {
	ID                      uint64               `gorm:"primaryKey;autoIncrement" json:"id"`
	LogisticsNo             string               `gorm:"type:varchar(64);uniqueIndex;not null;comment:物流委托单号" json:"logistics_no"`
	OrderID                 uint64               `gorm:"index;comment:交易订单ID" json:"order_id"`
	OrderNo                 string               `gorm:"type:varchar(64);index;comment:交易订单号" json:"order_no"`
	BuyerEnterpriseID       uint64               `gorm:"index;not null;comment:发单企业ID" json:"buyer_enterprise_id"`
	LogisticsServiceID      uint64               `gorm:"index;not null;comment:物流服务ID" json:"logistics_service_id"`
	CarrierEnterpriseID     uint64               `gorm:"index;comment:承运企业ID" json:"carrier_enterprise_id"`
	WarehouseServiceID      uint64               `gorm:"index;comment:提货仓库服务ID" json:"warehouse_service_id"`
	Origin                  string               `gorm:"type:varchar(255);not null;comment:提货点" json:"origin"`
	Destination             string               `gorm:"type:varchar(255);not null;comment:目的地" json:"destination"`
	CargoName               string               `gorm:"type:varchar(200);comment:货物名称" json:"cargo_name"`
	Weight                  float64              `gorm:"type:decimal(12,2);not null;comment:重量吨" json:"weight"`
	VehicleType             string               `gorm:"type:varchar(100);comment:车型" json:"vehicle_type"`
	EstimatedPrice          float64              `gorm:"type:decimal(12,2);comment:估算价格" json:"estimated_price"`
	FinalPrice              float64              `gorm:"type:decimal(12,2);comment:最终价格" json:"final_price"`
	Status                  LogisticsOrderStatus `gorm:"type:varchar(32);index;not null;default:task_created;comment:状态" json:"status"`
	CurrentResponsibleParty ResponsibilityParty  `gorm:"type:varchar(32);index;not null;default:logistics;comment:当前责任主体" json:"current_responsible_party"`
	HandoverNote            string               `gorm:"type:varchar(500);comment:交接说明" json:"handover_note"`
	DriverName              string               `gorm:"type:varchar(100);comment:司机姓名" json:"driver_name"`
	DriverPhone             string               `gorm:"type:varchar(50);comment:司机电话" json:"driver_phone"`
	PlateNumber             string               `gorm:"type:varchar(50);comment:车牌号" json:"plate_number"`
	GuaranteeType           string               `gorm:"type:varchar(100);comment:保障方式" json:"guarantee_type"`
	AcceptedBy              uint64               `gorm:"comment:接单人" json:"accepted_by"`
	AcceptedAt              *time.Time           `gorm:"comment:接单时间" json:"accepted_at"`
	PickupAppointmentAt     *time.Time           `gorm:"comment:提货预约时间" json:"pickup_appointment_at"`
	ArrivalSignedInAt       *time.Time           `gorm:"comment:到场签到时间" json:"arrival_signed_in_at"`
	OutboundConfirmedAt     *time.Time           `gorm:"comment:出库确认时间" json:"outbound_confirmed_at"`
	LoadingConfirmedAt      *time.Time           `gorm:"comment:装货确认时间" json:"loading_confirmed_at"`
	ArrivalConfirmedAt      *time.Time           `gorm:"comment:到货确认时间" json:"arrival_confirmed_at"`
	ReceiptSignedAt         *time.Time           `gorm:"comment:验收签收时间" json:"receipt_signed_at"`
	InboundCompletedAt      *time.Time           `gorm:"comment:入库完成时间" json:"inbound_completed_at"`
	CreatedBy               uint64               `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy               uint64               `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt               time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt               gorm.DeletedAt       `gorm:"index" json:"-"`
}

func (LogisticsOrder) TableName() string {
	return "logistics_orders"
}

type LogisticsDriver struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID          uint64         `gorm:"index;not null;comment:物流企业ID" json:"enterprise_id"`
	Name                  string         `gorm:"type:varchar(100);not null;comment:司机姓名" json:"name"`
	Phone                 string         `gorm:"type:varchar(50);not null;comment:司机电话" json:"phone"`
	IDCard                string         `gorm:"type:varchar(50);comment:身份证号" json:"id_card"`
	LicenseNo             string         `gorm:"type:varchar(100);comment:驾驶证号" json:"license_no"`
	LicenseExpireAt       string         `gorm:"type:varchar(64);comment:驾驶证有效期" json:"license_expire_at"`
	QualificationNo       string         `gorm:"type:varchar(100);comment:从业资格证号" json:"qualification_no"`
	QualificationExpireAt string         `gorm:"type:varchar(64);comment:从业资格证有效期" json:"qualification_expire_at"`
	Status                string         `gorm:"type:varchar(32);index;default:active;comment:状态 active/inactive" json:"status"`
	Remark                string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedBy             uint64         `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy             uint64         `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt             time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LogisticsDriver) TableName() string {
	return "logistics_drivers"
}

type LogisticsVehicle struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID      uint64         `gorm:"index;not null;comment:物流企业ID" json:"enterprise_id"`
	PlateNumber       string         `gorm:"type:varchar(50);not null;index;comment:车牌号" json:"plate_number"`
	VehicleType       string         `gorm:"type:varchar(100);comment:车型" json:"vehicle_type"`
	LoadCapacity      float64        `gorm:"type:decimal(12,2);comment:载重能力" json:"load_capacity"`
	VehicleLicense    string         `gorm:"type:varchar(100);comment:行驶证号" json:"vehicle_license"`
	LicenseExpireAt   string         `gorm:"type:varchar(64);comment:行驶证有效期" json:"license_expire_at"`
	InsuranceNo       string         `gorm:"type:varchar(100);comment:保险单号" json:"insurance_no"`
	InsuranceExpireAt string         `gorm:"type:varchar(64);comment:保险有效期" json:"insurance_expire_at"`
	DriverID          uint64         `gorm:"index;comment:绑定司机ID" json:"driver_id"`
	Status            string         `gorm:"type:varchar(32);index;default:active;comment:状态 active/inactive" json:"status"`
	Remark            string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedBy         uint64         `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy         uint64         `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LogisticsVehicle) TableName() string {
	return "logistics_vehicles"
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

// FulfillmentDocType 履约单据类型
type FulfillmentDocType string

const (
	DocTypePickupInstruction FulfillmentDocType = "pickup_instruction"
	DocTypePickupOrder       FulfillmentDocType = "pickup_order"
	DocTypeOutboundOrder     FulfillmentDocType = "outbound_order"
	DocTypeSealPhoto         FulfillmentDocType = "seal_photo"
	DocTypeLoadingVoucher    FulfillmentDocType = "loading_voucher"
	DocTypeGPSTrack          FulfillmentDocType = "gps_track"
	DocTypeInspectionResult  FulfillmentDocType = "inspection_result"
	DocTypeReceiptOrder      FulfillmentDocType = "receipt_order"
	DocTypeInboundOrder      FulfillmentDocType = "inbound_order"
	DocTypeWaybill           FulfillmentDocType = "waybill"
	DocTypeDeliveryPhoto     FulfillmentDocType = "delivery_photo"
)

type FulfillmentDocument struct {
	ID               uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	LogisticsOrderID uint64             `gorm:"index;not null;comment:物流委托单ID" json:"logistics_order_id"`
	OrderID          uint64             `gorm:"index;comment:交易订单ID" json:"order_id"`
	DocType          FulfillmentDocType `gorm:"type:varchar(64);index;not null;comment:单据类型" json:"doc_type"`
	FileURL          string             `gorm:"type:varchar(500);not null;comment:文件地址" json:"file_url"`
	FileHash         string             `gorm:"type:varchar(128);comment:文件哈希" json:"file_hash"`
	Remark           string             `gorm:"type:varchar(500);comment:备注" json:"remark"`
	GPSLatitude      *float64           `gorm:"comment:GPS纬度" json:"gps_latitude"`
	GPSLongitude     *float64           `gorm:"comment:GPS经度" json:"gps_longitude"`
	CapturedAt       string             `gorm:"type:varchar(64);comment:拍摄时间" json:"captured_at"`
	UploadedBy       uint64             `gorm:"index;comment:上传人" json:"uploaded_by"`
	EnterpriseID     uint64             `gorm:"index;comment:上传企业" json:"enterprise_id"`
	UploadedRole     string             `gorm:"type:varchar(32);comment:上传角色" json:"uploaded_role"`
	CreatedAt        time.Time          `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (FulfillmentDocument) TableName() string {
	return "fulfillment_documents"
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

type WarehouseReceiptPledgeApplication struct {
	ID                      uint64                       `gorm:"primaryKey;autoIncrement" json:"id"`
	ApplicationNo           string                       `gorm:"type:varchar(64);uniqueIndex;not null;comment:仓单质押申请号" json:"application_no"`
	ApplicantEnterpriseID   uint64                       `gorm:"index;not null;comment:申请企业ID" json:"applicant_enterprise_id"`
	ApplicantUserID         uint64                       `gorm:"index;not null;comment:申请人ID" json:"applicant_user_id"`
	ApplicantRole           string                       `gorm:"type:varchar(32);index;comment:申请方角色 procurement/sales" json:"applicant_role"`
	FinanceServiceID        uint64                       `gorm:"index;not null;comment:资金产品ID" json:"finance_service_id"`
	FundingEnterpriseID     uint64                       `gorm:"index;not null;comment:资金方企业ID" json:"funding_enterprise_id"`
	WarehouseServiceID      uint64                       `gorm:"index;not null;comment:监管仓库服务ID" json:"warehouse_service_id"`
	WarehouseEnterpriseID   uint64                       `gorm:"index;not null;comment:仓储企业ID" json:"warehouse_enterprise_id"`
	WarehouseName           string                       `gorm:"type:varchar(200);comment:监管仓库名称" json:"warehouse_name"`
	ReceiptNo               string                       `gorm:"type:varchar(100);index;not null;comment:仓单编号" json:"receipt_no"`
	ReceiptVoucherURL       string                       `gorm:"type:varchar(500);not null;comment:仓单文件URL" json:"receipt_voucher_url"`
	GoodsName               string                       `gorm:"type:varchar(200);not null;comment:货物名称" json:"goods_name"`
	Quantity                float64                      `gorm:"type:decimal(12,2);not null;comment:数量吨" json:"quantity"`
	QualityGrade            string                       `gorm:"type:varchar(100);comment:品质/牌号" json:"quality_grade"`
	RequestedAmount         float64                      `gorm:"type:decimal(14,2);not null;comment:申请金额" json:"requested_amount"`
	ApprovedAmount          float64                      `gorm:"type:decimal(14,2);comment:审批金额" json:"approved_amount"`
	TermDays                int                          `gorm:"comment:账期天数" json:"term_days"`
	AnnualRate              float64                      `gorm:"type:decimal(8,4);comment:年化费率" json:"annual_rate"`
	RiskRemark              string                       `gorm:"type:varchar(1000);comment:风控备注" json:"risk_remark"`
	PledgeAgreementURL      string                       `gorm:"type:varchar(500);comment:仓单质押协议URL" json:"pledge_agreement_url"`
	SupervisionAgreementURL string                       `gorm:"type:varchar(500);comment:监管协议URL" json:"supervision_agreement_url"`
	DisbursementVoucherURL  string                       `gorm:"type:varchar(500);comment:放款凭证URL" json:"disbursement_voucher_url"`
	ReleaseVoucherURL       string                       `gorm:"type:varchar(500);comment:货权释放凭证URL" json:"release_voucher_url"`
	ReleaseRemark           string                       `gorm:"type:varchar(1000);comment:释放备注" json:"release_remark"`
	Status                  WarehouseReceiptPledgeStatus `gorm:"type:varchar(40);index;not null;default:submitted;comment:状态" json:"status"`
	SubmittedAt             *time.Time                   `gorm:"comment:提交时间" json:"submitted_at"`
	RiskReviewedAt          *time.Time                   `gorm:"comment:风控时间" json:"risk_reviewed_at"`
	AgreementSignedAt       *time.Time                   `gorm:"comment:协议签署时间" json:"agreement_signed_at"`
	DisbursedAt             *time.Time                   `gorm:"comment:放款时间" json:"disbursed_at"`
	ReleasedAt              *time.Time                   `gorm:"comment:释放时间" json:"released_at"`
	Remark                  string                       `gorm:"type:varchar(1000);comment:申请备注" json:"remark"`
	CreatedBy               uint64                       `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy               uint64                       `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt               time.Time                    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time                    `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt               gorm.DeletedAt               `gorm:"index" json:"-"`
}

func (WarehouseReceiptPledgeApplication) TableName() string {
	return "warehouse_receipt_pledge_applications"
}
