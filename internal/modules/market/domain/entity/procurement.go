package entity

import (
	"time"

	"gorm.io/gorm"
)

type ProcurementStatus string

const (
	ProcurementStatusPreparing ProcurementStatus = "preparing"
	ProcurementStatusSignup    ProcurementStatus = "signup"
	ProcurementStatusBidding   ProcurementStatus = "bidding"
	ProcurementStatusAwarded   ProcurementStatus = "awarded"
	ProcurementStatusClosed    ProcurementStatus = "closed"

	ProcurementStatusPendingSeller  ProcurementStatus = "pending_seller"
	ProcurementStatusActive         ProcurementStatus = "active"
	ProcurementStatusSellerRejected ProcurementStatus = "seller_rejected"
	ProcurementStatusEnded          ProcurementStatus = "ended"
	ProcurementStatusOrderCreated   ProcurementStatus = "order_created"
)

type ProcurementActivity struct {
	ID                 uint64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcurementNo      string                   `gorm:"type:varchar(40);uniqueIndex;comment:集采编号" json:"procurement_no"`
	Title              string                   `gorm:"type:varchar(200);index;comment:集采标题" json:"title"`
	LeaderName         string                   `gorm:"type:varchar(120);comment:牵头方" json:"leader_name"`
	Status             ProcurementStatus        `gorm:"type:varchar(32);index;comment:状态" json:"status"`
	ProductID          uint64                   `gorm:"index;comment:关联商品ID" json:"product_id"`
	ProductNo          string                   `gorm:"type:varchar(64);index;comment:商品编号" json:"product_no"`
	ProductName        string                   `gorm:"type:varchar(200);comment:商品名称" json:"product_name"`
	SellerEnterpriseID uint64                   `gorm:"index;comment:卖家企业ID" json:"seller_enterprise_id"`
	SellerName         string                   `gorm:"type:varchar(200);comment:卖家企业名称" json:"seller_name"`
	GroupPrice         float64                  `gorm:"type:decimal(16,2);comment:集采单价" json:"group_price"`
	MinGroupQuantity   float64                  `gorm:"type:decimal(14,2);comment:最小成团量" json:"min_group_quantity"`
	MaxGroupQuantity   float64                  `gorm:"type:decimal(14,2);comment:最大可参团量" json:"max_group_quantity"`
	Category           string                   `gorm:"type:varchar(80);index;comment:品类" json:"category"`
	Target             string                   `gorm:"type:varchar(120);comment:采购标的" json:"target"`
	TargetQuantity     float64                  `gorm:"type:decimal(14,2);comment:目标数量" json:"target_quantity"`
	Unit               string                   `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	BudgetMin          float64                  `gorm:"type:decimal(16,2);comment:预算下限" json:"budget_min"`
	BudgetMax          float64                  `gorm:"type:decimal(16,2);comment:预算上限" json:"budget_max"`
	BudgetDisplay      string                   `gorm:"type:varchar(60);comment:预算展示" json:"budget_display"`
	Stage              string                   `gorm:"type:varchar(60);comment:当前阶段" json:"stage"`
	Progress           int                      `gorm:"comment:进度百分比" json:"progress"`
	Region             string                   `gorm:"type:varchar(120);comment:交付区域" json:"region"`
	OpenDate           string                   `gorm:"type:varchar(40);comment:开标日期" json:"open_date"`
	ParticipatingCount int                      `gorm:"comment:参采企业" json:"participating_count"`
	DemandQuantity     float64                  `gorm:"type:decimal(14,2);comment:归集需求" json:"demand_quantity"`
	QualifiedSuppliers int                      `gorm:"comment:入围供应商" json:"qualified_suppliers"`
	QuoteCount         int                      `gorm:"comment:收到报价" json:"quote_count"`
	ContractCount      int                      `gorm:"comment:已签合同" json:"contract_count"`
	DeliveryDays       int                      `gorm:"comment:交货周期天数" json:"delivery_days"`
	QuoteMode          string                   `gorm:"type:varchar(40);comment:报价方式" json:"quote_mode"`
	EvaluationMode     string                   `gorm:"type:varchar(40);comment:评审标准" json:"evaluation_mode"`
	Documents          StringList               `gorm:"type:json;comment:资质文件" json:"documents"`
	AdminRemark        string                   `gorm:"type:varchar(500);comment:管理备注" json:"admin_remark"`
	SellerRemark       string                   `gorm:"type:varchar(500);comment:卖家备注" json:"seller_remark"`
	Source             string                   `gorm:"type:varchar(40);index;comment:来源 admin/ai_request" json:"source"`
	StartedAt          *time.Time               `gorm:"index;comment:开始时间" json:"started_at"`
	EndedAt            *time.Time               `gorm:"index;comment:截止时间" json:"ended_at"`
	SellerRespondedAt  *time.Time               `gorm:"comment:卖家响应时间" json:"seller_responded_at"`
	ClosedAt           *time.Time               `gorm:"comment:结束时间" json:"closed_at"`
	MasterOrderID      uint64                   `gorm:"index;comment:母订单ID" json:"master_order_id"`
	MasterOrderNo      string                   `gorm:"type:varchar(64);index;comment:母订单号" json:"master_order_no"`
	CreatedBy          uint64                   `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy          uint64                   `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt          time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time                `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt           `gorm:"index" json:"-"`
	Logs               []ProcurementLog         `gorm:"foreignKey:ActivityID;references:ID" json:"logs"`
	Participants       []ProcurementParticipant `gorm:"foreignKey:ActivityID;references:ID" json:"participants"`
}

func (ProcurementActivity) TableName() string {
	return "procurement_activities"
}

type ProcurementLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID uint64    `gorm:"index;comment:集采活动ID" json:"activity_id"`
	Content    string    `gorm:"type:varchar(255);comment:日志内容" json:"content"`
	Color      string    `gorm:"type:varchar(20);default:gray;comment:展示色" json:"color"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProcurementLog) TableName() string {
	return "procurement_logs"
}

type ProcurementParticipantStatus string

const (
	ProcurementParticipantJoined       ProcurementParticipantStatus = "joined"
	ProcurementParticipantCancelled    ProcurementParticipantStatus = "cancelled"
	ProcurementParticipantOrderCreated ProcurementParticipantStatus = "order_created"
)

type ProcurementParticipant struct {
	ID                        uint64                       `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID                uint64                       `gorm:"index;not null;comment:集采活动ID" json:"activity_id"`
	BuyerEnterpriseID         uint64                       `gorm:"index;not null;comment:买家企业ID" json:"buyer_enterprise_id"`
	BuyerUserID               uint64                       `gorm:"index;comment:买家用户ID" json:"buyer_user_id"`
	BuyerName                 string                       `gorm:"type:varchar(200);comment:买家企业名称" json:"buyer_name"`
	Quantity                  float64                      `gorm:"type:decimal(14,2);comment:参团数量" json:"quantity"`
	Unit                      string                       `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	UnitPrice                 float64                      `gorm:"type:decimal(16,2);comment:集采单价" json:"unit_price"`
	TotalAmount               float64                      `gorm:"type:decimal(16,2);comment:参团金额" json:"total_amount"`
	ContactName               string                       `gorm:"type:varchar(80);comment:收货联系人" json:"contact_name"`
	ContactPhone              string                       `gorm:"type:varchar(40);comment:收货联系电话" json:"contact_phone"`
	DeliveryAddressSource     string                       `gorm:"type:varchar(32);comment:收货地址来源 manual/warehouse" json:"delivery_address_source"`
	DestWarehouseServiceID    uint64                       `gorm:"index;comment:目的地仓储资源ID" json:"dest_warehouse_service_id"`
	DestWarehouseEnterpriseID uint64                       `gorm:"index;comment:目的地仓储方企业ID" json:"dest_warehouse_enterprise_id"`
	DestWarehouseName         string                       `gorm:"type:varchar(200);comment:目的地仓库名称" json:"dest_warehouse_name"`
	DestWarehouseProvince     string                       `gorm:"type:varchar(80);comment:目的地仓库省份" json:"dest_warehouse_province"`
	DestWarehouseCity         string                       `gorm:"type:varchar(80);comment:目的地仓库城市" json:"dest_warehouse_city"`
	DestWarehouseDistrict     string                       `gorm:"type:varchar(80);comment:目的地仓库区县" json:"dest_warehouse_district"`
	DestWarehouseAddress      string                       `gorm:"type:varchar(500);comment:目的地仓库地址" json:"dest_warehouse_address"`
	DeliveryAddress           string                       `gorm:"type:varchar(500);comment:收货地址" json:"delivery_address"`
	DeliveryMethod            string                       `gorm:"type:varchar(60);comment:配送方式" json:"delivery_method"`
	PaymentMethod             string                       `gorm:"type:varchar(60);comment:付款方式" json:"payment_method"`
	DeliveryNote              string                       `gorm:"type:varchar(500);comment:配送备注" json:"delivery_note"`
	AgreementSigned           bool                         `gorm:"default:false;comment:是否已签集采协议" json:"agreement_signed"`
	AgreementSignedAt         *time.Time                   `gorm:"comment:协议签署时间" json:"agreement_signed_at"`
	Status                    ProcurementParticipantStatus `gorm:"type:varchar(32);index;default:joined;comment:状态" json:"status"`
	ChildOrderID              uint64                       `gorm:"index;comment:子订单ID" json:"child_order_id"`
	ChildOrderNo              string                       `gorm:"type:varchar(64);index;comment:子订单号" json:"child_order_no"`
	Remark                    string                       `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt                 time.Time                    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time                    `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                 gorm.DeletedAt               `gorm:"index" json:"-"`
}

func (ProcurementParticipant) TableName() string {
	return "procurement_participants"
}

type ProcurementRequestStatus string

const (
	ProcurementRequestPending  ProcurementRequestStatus = "pending"
	ProcurementRequestSelected ProcurementRequestStatus = "selected"
	ProcurementRequestInvited  ProcurementRequestStatus = "invited"
	ProcurementRequestAccepted ProcurementRequestStatus = "accepted"
	ProcurementRequestRejected ProcurementRequestStatus = "rejected"
	ProcurementRequestExpired  ProcurementRequestStatus = "expired"
)

type ProcurementRequest struct {
	ID                        uint64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestNo                 string                   `gorm:"type:varchar(40);uniqueIndex;comment:集采申请编号" json:"request_no"`
	Source                    string                   `gorm:"type:varchar(40);index;comment:来源 product_detail/ai/manual" json:"source"`
	SourceProductID           uint64                   `gorm:"index;comment:来源商品ID" json:"source_product_id"`
	AIProcurementID           uint64                   `gorm:"index;comment:AI询价记录ID" json:"ai_procurement_id"`
	Category                  string                   `gorm:"type:varchar(80);index;comment:品类" json:"category"`
	SubCategory               string                   `gorm:"type:varchar(80);index;comment:子品类" json:"sub_category"`
	ProductName               string                   `gorm:"type:varchar(200);comment:采购物品" json:"product_name"`
	Material                  string                   `gorm:"type:varchar(120);comment:材质" json:"material"`
	Spec                      string                   `gorm:"type:varchar(200);comment:规格" json:"spec"`
	Quantity                  float64                  `gorm:"type:decimal(14,2);comment:需求数量" json:"quantity"`
	Unit                      string                   `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	TargetPrice               float64                  `gorm:"type:decimal(16,2);comment:期望价格" json:"target_price"`
	ContactName               string                   `gorm:"type:varchar(80);comment:收货联系人" json:"contact_name"`
	ContactPhone              string                   `gorm:"type:varchar(40);comment:收货联系电话" json:"contact_phone"`
	DeliveryAddressSource     string                   `gorm:"type:varchar(32);comment:收货地址来源 manual/warehouse" json:"delivery_address_source"`
	DestWarehouseServiceID    uint64                   `gorm:"index;comment:目的地仓储资源ID" json:"dest_warehouse_service_id"`
	DestWarehouseEnterpriseID uint64                   `gorm:"index;comment:目的地仓储方企业ID" json:"dest_warehouse_enterprise_id"`
	DestWarehouseName         string                   `gorm:"type:varchar(200);comment:目的地仓库名称" json:"dest_warehouse_name"`
	DestWarehouseProvince     string                   `gorm:"type:varchar(80);comment:目的地仓库省份" json:"dest_warehouse_province"`
	DestWarehouseCity         string                   `gorm:"type:varchar(80);comment:目的地仓库城市" json:"dest_warehouse_city"`
	DestWarehouseDistrict     string                   `gorm:"type:varchar(80);comment:目的地仓库区县" json:"dest_warehouse_district"`
	DestWarehouseAddress      string                   `gorm:"type:varchar(500);comment:目的地仓库地址" json:"dest_warehouse_address"`
	DeliveryAddress           string                   `gorm:"type:varchar(500);comment:收货地址" json:"delivery_address"`
	DeliveryMethod            string                   `gorm:"type:varchar(60);comment:配送方式" json:"delivery_method"`
	PaymentMethod             string                   `gorm:"type:varchar(60);comment:付款方式" json:"payment_method"`
	DeliveryNote              string                   `gorm:"type:varchar(500);comment:配送备注" json:"delivery_note"`
	BuyerEnterpriseID         uint64                   `gorm:"index;comment:申请企业ID" json:"buyer_enterprise_id"`
	BuyerUserID               uint64                   `gorm:"index;comment:申请用户ID" json:"buyer_user_id"`
	BuyerName                 string                   `gorm:"type:varchar(200);comment:申请企业名称" json:"buyer_name"`
	Status                    ProcurementRequestStatus `gorm:"type:varchar(32);index;default:pending;comment:状态" json:"status"`
	ActivityID                uint64                   `gorm:"index;comment:关联集采活动ID" json:"activity_id"`
	Remark                    string                   `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt                 time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time                `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                 gorm.DeletedAt           `gorm:"index" json:"-"`
}

func (ProcurementRequest) TableName() string {
	return "procurement_requests"
}

type ProcurementActivityRequest struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID uint64    `gorm:"uniqueIndex:idx_procurement_activity_request;comment:集采活动ID" json:"activity_id"`
	RequestID  uint64    `gorm:"uniqueIndex:idx_procurement_activity_request;comment:买家申请ID" json:"request_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProcurementActivityRequest) TableName() string {
	return "procurement_activity_requests"
}

type ProcurementInvitationStatus string

const (
	ProcurementInvitationPending  ProcurementInvitationStatus = "pending"
	ProcurementInvitationAccepted ProcurementInvitationStatus = "accepted"
	ProcurementInvitationDeclined ProcurementInvitationStatus = "declined"
	ProcurementInvitationExpired  ProcurementInvitationStatus = "expired"
)

type ProcurementInvitation struct {
	ID                uint64                      `gorm:"primaryKey;autoIncrement" json:"id"`
	InvitationNo      string                      `gorm:"type:varchar(40);uniqueIndex;comment:集采邀请编号" json:"invitation_no"`
	ActivityID        uint64                      `gorm:"index;comment:集采活动ID" json:"activity_id"`
	RequestID         uint64                      `gorm:"index;comment:买家申请ID" json:"request_id"`
	BuyerEnterpriseID uint64                      `gorm:"index;comment:邀请企业ID" json:"buyer_enterprise_id"`
	BuyerUserID       uint64                      `gorm:"index;comment:邀请用户ID" json:"buyer_user_id"`
	BuyerName         string                      `gorm:"type:varchar(200);comment:邀请企业名称" json:"buyer_name"`
	Status            ProcurementInvitationStatus `gorm:"type:varchar(32);index;default:pending;comment:状态" json:"status"`
	ParticipantID     uint64                      `gorm:"index;comment:参团记录ID" json:"participant_id"`
	Remark            string                      `gorm:"type:varchar(500);comment:备注" json:"remark"`
	ExpiresAt         *time.Time                  `gorm:"index;comment:过期时间" json:"expires_at"`
	AcceptedAt        *time.Time                  `gorm:"comment:接受时间" json:"accepted_at"`
	DeclinedAt        *time.Time                  `gorm:"comment:拒绝时间" json:"declined_at"`
	CreatedAt         time.Time                   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time                   `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt              `gorm:"index" json:"-"`
	Activity          ProcurementActivity         `gorm:"foreignKey:ActivityID;references:ID" json:"activity"`
	Request           ProcurementRequest          `gorm:"foreignKey:RequestID;references:ID" json:"request"`
}

func (ProcurementInvitation) TableName() string {
	return "procurement_invitations"
}
