package entity

import (
	"time"

	"gorm.io/gorm"
)

type ProcurementProjectStatus string

const (
	ProcurementProjectStatusCollecting   ProcurementProjectStatus = "collecting"
	ProcurementProjectStatusQuoting      ProcurementProjectStatus = "quoting"
	ProcurementProjectStatusAwarded      ProcurementProjectStatus = "awarded"
	ProcurementProjectStatusOrderCreated ProcurementProjectStatus = "order_created"
	ProcurementProjectStatusClosed       ProcurementProjectStatus = "closed"
)

type ProcurementProject struct {
	ID                   uint64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectNo            string                   `gorm:"type:varchar(40);uniqueIndex;comment:项目编号" json:"project_no"`
	Title                string                   `gorm:"type:varchar(200);index;comment:项目标题" json:"title"`
	Status               ProcurementProjectStatus `gorm:"type:varchar(32);index;comment:项目状态" json:"status"`
	Category             string                   `gorm:"type:varchar(80);index;comment:品类" json:"category"`
	SubCategory          string                   `gorm:"type:varchar(80);index;comment:子品类" json:"sub_category"`
	ProductName          string                   `gorm:"type:varchar(200);comment:采购物品" json:"product_name"`
	Material             string                   `gorm:"type:varchar(120);comment:材质" json:"material"`
	Spec                 string                   `gorm:"type:varchar(200);comment:规格" json:"spec"`
	TargetQuantity       float64                  `gorm:"type:decimal(14,2);comment:目标数量" json:"target_quantity"`
	DemandQuantity       float64                  `gorm:"type:decimal(14,2);comment:归集数量" json:"demand_quantity"`
	Unit                 string                   `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	TargetPrice          float64                  `gorm:"type:decimal(16,2);comment:目标单价" json:"target_price"`
	DeliveryAddress      string                   `gorm:"type:varchar(500);comment:交付地址" json:"delivery_address"`
	BuyerRequestCount    int                      `gorm:"comment:关联申请数" json:"buyer_request_count"`
	BuyerEnterpriseCount int                      `gorm:"comment:参与买家数" json:"buyer_enterprise_count"`
	InvitedSupplierCount int                      `gorm:"comment:邀请供应商数" json:"invited_supplier_count"`
	QuotedSupplierCount  int                      `gorm:"comment:有效报价供应商数" json:"quoted_supplier_count"`
	QuoteDeadline        *time.Time               `gorm:"index;comment:报价截止时间" json:"quote_deadline"`
	AwardedAt            *time.Time               `gorm:"comment:定标时间" json:"awarded_at"`
	OrderedAt            *time.Time               `gorm:"comment:成单时间" json:"ordered_at"`
	AwardQuoteID         uint64                   `gorm:"index;comment:中标报价ID" json:"award_quote_id"`
	AwardSupplierID      uint64                   `gorm:"index;comment:中标卖家企业ID" json:"award_supplier_id"`
	AwardSupplierName    string                   `gorm:"type:varchar(200);comment:中标卖家名称" json:"award_supplier_name"`
	AwardProductID       uint64                   `gorm:"index;comment:中标商品ID" json:"award_product_id"`
	AwardProductNo       string                   `gorm:"type:varchar(64);comment:中标商品编号" json:"award_product_no"`
	AwardProductName     string                   `gorm:"type:varchar(200);comment:中标商品名称" json:"award_product_name"`
	AwardUnitPrice       float64                  `gorm:"type:decimal(16,2);comment:中标单价" json:"award_unit_price"`
	AwardDeliveryDays    int                      `gorm:"comment:中标交期" json:"award_delivery_days"`
	OrderID              uint64                   `gorm:"index;comment:主订单ID" json:"order_id"`
	OrderNo              string                   `gorm:"type:varchar(64);index;comment:主订单号" json:"order_no"`
	AdminRemark          string                   `gorm:"type:varchar(500);comment:管理备注" json:"admin_remark"`
	CreatedBy            uint64                   `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy            uint64                   `gorm:"comment:更新者" json:"updated_by"`
	CreatedAt            time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time                `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt           `gorm:"index" json:"-"`
	Requests             []ProcurementRequest     `gorm:"many2many:procurement_project_requests;joinForeignKey:ProjectID;JoinReferences:RequestID" json:"requests"`
	Enrollments          []ProcurementEnrollment  `gorm:"foreignKey:ProjectID;references:ID" json:"enrollments"`
	Quotes               []ProcurementQuote       `gorm:"foreignKey:ProjectID;references:ID" json:"quotes"`
	Logs                 []ProcurementLog         `gorm:"foreignKey:ProjectID;references:ID" json:"logs"`
}

func (ProcurementProject) TableName() string {
	return "procurement_projects"
}

type ProcurementProjectRequest struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID uint64         `gorm:"uniqueIndex:idx_procurement_project_request;comment:项目ID" json:"project_id"`
	RequestID uint64         `gorm:"uniqueIndex:idx_procurement_project_request;comment:申请ID" json:"request_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProcurementProjectRequest) TableName() string {
	return "procurement_project_requests"
}

type ProcurementEnrollmentStatus string

const (
	ProcurementEnrollmentStatusInvited  ProcurementEnrollmentStatus = "invited"
	ProcurementEnrollmentStatusApproved ProcurementEnrollmentStatus = "approved"
	ProcurementEnrollmentStatusRejected ProcurementEnrollmentStatus = "rejected"
)

type ProcurementEnrollment struct {
	ID                 uint64                      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID          uint64                      `gorm:"index;not null;comment:项目ID" json:"project_id"`
	SellerEnterpriseID uint64                      `gorm:"index;not null;comment:卖家企业ID" json:"seller_enterprise_id"`
	SellerUserID       uint64                      `gorm:"index;comment:卖家用户ID" json:"seller_user_id"`
	SellerName         string                      `gorm:"type:varchar(200);comment:卖家企业名称" json:"seller_name"`
	ProductID          uint64                      `gorm:"index;comment:推荐报价商品ID" json:"product_id"`
	ProductNo          string                      `gorm:"type:varchar(64);comment:推荐商品编号" json:"product_no"`
	ProductName        string                      `gorm:"type:varchar(200);comment:推荐商品名称" json:"product_name"`
	Status             ProcurementEnrollmentStatus `gorm:"type:varchar(32);index;default:invited;comment:状态" json:"status"`
	Remark             string                      `gorm:"type:varchar(500);comment:备注" json:"remark"`
	ReviewedBy         uint64                      `gorm:"comment:审核人" json:"reviewed_by"`
	ReviewedAt         *time.Time                  `gorm:"comment:审核时间" json:"reviewed_at"`
	CreatedAt          time.Time                   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time                   `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt              `gorm:"index" json:"-"`
}

func (ProcurementEnrollment) TableName() string {
	return "procurement_supplier_enrollments"
}

type ProcurementQuoteStatus string

const (
	ProcurementQuoteStatusSubmitted ProcurementQuoteStatus = "submitted"
	ProcurementQuoteStatusReplaced  ProcurementQuoteStatus = "replaced"
	ProcurementQuoteStatusAwarded   ProcurementQuoteStatus = "awarded"
)

type ProcurementQuote struct {
	ID                 uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID          uint64                 `gorm:"index;not null;comment:项目ID" json:"project_id"`
	EnrollmentID       uint64                 `gorm:"index;not null;comment:报名ID" json:"enrollment_id"`
	SellerEnterpriseID uint64                 `gorm:"index;not null;comment:卖家企业ID" json:"seller_enterprise_id"`
	SellerUserID       uint64                 `gorm:"index;comment:卖家用户ID" json:"seller_user_id"`
	SellerName         string                 `gorm:"type:varchar(200);comment:卖家名称" json:"seller_name"`
	ProductID          uint64                 `gorm:"index;comment:商品ID" json:"product_id"`
	ProductNo          string                 `gorm:"type:varchar(64);comment:商品编号" json:"product_no"`
	ProductName        string                 `gorm:"type:varchar(200);comment:商品名称" json:"product_name"`
	UnitPrice          float64                `gorm:"type:decimal(16,2);comment:报价单价" json:"unit_price"`
	SupplyQuantity     float64                `gorm:"type:decimal(14,2);comment:可供数量" json:"supply_quantity"`
	Unit               string                 `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	DeliveryDays       int                    `gorm:"comment:交期天数" json:"delivery_days"`
	DeliveryWarehouse  string                 `gorm:"type:varchar(200);comment:发货仓库" json:"delivery_warehouse"`
	TaxIncluded        bool                   `gorm:"default:true;comment:是否含税" json:"tax_included"`
	TaxRate            float64                `gorm:"type:decimal(8,2);comment:税率" json:"tax_rate"`
	PaymentTerms       string                 `gorm:"type:varchar(200);comment:付款条件" json:"payment_terms"`
	Remark             string                 `gorm:"type:varchar(500);comment:备注" json:"remark"`
	IsLatest           bool                   `gorm:"index;default:true;comment:是否最新报价" json:"is_latest"`
	Status             ProcurementQuoteStatus `gorm:"type:varchar(32);index;default:submitted;comment:状态" json:"status"`
	CreatedAt          time.Time              `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (ProcurementQuote) TableName() string {
	return "procurement_supplier_quotes"
}

type ProcurementAward struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID          uint64         `gorm:"uniqueIndex;comment:项目ID" json:"project_id"`
	QuoteID            uint64         `gorm:"index;not null;comment:报价ID" json:"quote_id"`
	SellerEnterpriseID uint64         `gorm:"index;not null;comment:卖家企业ID" json:"seller_enterprise_id"`
	SellerName         string         `gorm:"type:varchar(200);comment:卖家名称" json:"seller_name"`
	ProductID          uint64         `gorm:"index;comment:商品ID" json:"product_id"`
	ProductNo          string         `gorm:"type:varchar(64);comment:商品编号" json:"product_no"`
	ProductName        string         `gorm:"type:varchar(200);comment:商品名称" json:"product_name"`
	UnitPrice          float64        `gorm:"type:decimal(16,2);comment:中标单价" json:"unit_price"`
	AwardQuantity      float64        `gorm:"type:decimal(14,2);comment:中标数量" json:"award_quantity"`
	DeliveryDays       int            `gorm:"comment:交期" json:"delivery_days"`
	Remark             string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedBy          uint64         `gorm:"comment:定标人" json:"created_by"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProcurementAward) TableName() string {
	return "procurement_project_awards"
}
