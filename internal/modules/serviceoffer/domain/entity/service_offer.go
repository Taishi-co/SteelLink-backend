package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type OfferStatus int

const (
	OfferStatusPending  OfferStatus = 1
	OfferStatusApproved OfferStatus = 2
	OfferStatusRejected OfferStatus = 3
	OfferStatusOffline  OfferStatus = 4
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

type WarehouseService struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID       uint64         `gorm:"index;not null;comment:企业ID" json:"enterprise_id"`
	WarehouseName      string         `gorm:"type:varchar(200);not null;comment:仓库名称" json:"warehouse_name"`
	Province           string         `gorm:"type:varchar(50);comment:省份" json:"province"`
	City               string         `gorm:"type:varchar(50);comment:城市" json:"city"`
	District           string         `gorm:"type:varchar(50);comment:区县" json:"district"`
	Address            string         `gorm:"type:varchar(255);comment:地址" json:"address"`
	Capacity           float64        `gorm:"type:decimal(12,2);comment:总库容" json:"capacity"`
	AvailableCapacity  float64        `gorm:"type:decimal(12,2);comment:可用库容" json:"available_capacity"`
	ServiceTypes       StringList     `gorm:"type:json;comment:服务类型" json:"service_types"`
	PriceRule          string         `gorm:"type:varchar(500);comment:价格规则" json:"price_rule"`
	LoadingCapacity    string         `gorm:"type:varchar(255);comment:装卸能力" json:"loading_capacity"`
	BusinessHours      string         `gorm:"type:varchar(255);comment:营业时间" json:"business_hours"`
	PledgeSupervision  bool           `gorm:"default:false;comment:是否支持质押监管" json:"pledge_supervision"`
	OutboundVoucherReq string         `gorm:"type:varchar(500);comment:出库凭证要求" json:"outbound_voucher_req"`
	Images             StringList     `gorm:"type:json;comment:展示图片" json:"images"`
	Certificates       StringList     `gorm:"type:json;comment:资质附件" json:"certificates"`
	Description        string         `gorm:"type:text;comment:描述" json:"description"`
	Status             OfferStatus    `gorm:"type:tinyint;default:1;index;comment:状态" json:"status"`
	AuditRemark        string         `gorm:"type:varchar(500);comment:审核备注" json:"audit_remark"`
	ApprovedAt         *time.Time     `gorm:"comment:审核通过时间" json:"approved_at"`
	CreatedBy          uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt          time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy          uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (WarehouseService) TableName() string {
	return "warehouse_services"
}

type LogisticsService struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID uint64         `gorm:"index;not null;comment:企业ID" json:"enterprise_id"`
	Name         string         `gorm:"type:varchar(200);not null;comment:服务名称" json:"name"`
	RouteStart   string         `gorm:"type:varchar(100);comment:起始地" json:"route_start"`
	RouteEnd     string         `gorm:"type:varchar(100);comment:目的地" json:"route_end"`
	ServiceScope string         `gorm:"type:varchar(255);comment:服务范围" json:"service_scope"`
	VehicleType  string         `gorm:"type:varchar(100);comment:车型" json:"vehicle_type"`
	LoadCapacity float64        `gorm:"type:decimal(12,2);comment:载重能力" json:"load_capacity"`
	PriceRule    string         `gorm:"type:varchar(500);comment:价格规则" json:"price_rule"`
	QuoteRules   JSONMap        `gorm:"type:json;comment:结构化报价规则" json:"quote_rules"`
	FleetName    string         `gorm:"type:varchar(200);comment:车队名称" json:"fleet_name"`
	VehicleCerts StringList     `gorm:"type:json;comment:车辆资质" json:"vehicle_certs"`
	DriverCerts  StringList     `gorm:"type:json;comment:司机资质" json:"driver_certs"`
	Images       StringList     `gorm:"type:json;comment:展示图片" json:"images"`
	Certificates StringList     `gorm:"type:json;comment:资质附件" json:"certificates"`
	Description  string         `gorm:"type:text;comment:描述" json:"description"`
	Status       OfferStatus    `gorm:"type:tinyint;default:1;index;comment:状态" json:"status"`
	AuditRemark  string         `gorm:"type:varchar(500);comment:审核备注" json:"audit_remark"`
	ApprovedAt   *time.Time     `gorm:"comment:审核通过时间" json:"approved_at"`
	CreatedBy    uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy    uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (LogisticsService) TableName() string {
	return "logistics_services"
}

type FinanceService struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID       uint64         `gorm:"index;not null;comment:企业ID" json:"enterprise_id"`
	Name               string         `gorm:"type:varchar(200);not null;comment:金融产品名称" json:"name"`
	FinanceType        string         `gorm:"type:varchar(64);comment:金融服务类型" json:"finance_type"`
	MinAmount          float64        `gorm:"type:decimal(14,2);comment:最低融资金额" json:"min_amount"`
	MaxAmount          float64        `gorm:"type:decimal(14,2);comment:最高融资金额" json:"max_amount"`
	TermDays           int            `gorm:"comment:账期天数" json:"term_days"`
	AnnualRate         float64        `gorm:"type:decimal(8,4);comment:年化费率" json:"annual_rate"`
	FeeRule            string         `gorm:"type:varchar(500);comment:手续费规则" json:"fee_rule"`
	RiskControlRule    string         `gorm:"type:text;comment:风控要求" json:"risk_control_rule"`
	RequiredDocuments  StringList     `gorm:"type:json;comment:申请材料" json:"required_documents"`
	SupportedScenarios StringList     `gorm:"type:json;comment:适用场景" json:"supported_scenarios"`
	Description        string         `gorm:"type:text;comment:描述" json:"description"`
	Status             OfferStatus    `gorm:"type:tinyint;default:1;index;comment:状态" json:"status"`
	AuditRemark        string         `gorm:"type:varchar(500);comment:审核备注" json:"audit_remark"`
	ApprovedAt         *time.Time     `gorm:"comment:审核通过时间" json:"approved_at"`
	CreatedBy          uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt          time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy          uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (FinanceService) TableName() string {
	return "finance_services"
}
