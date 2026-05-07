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

type ProcurementStatus string

const (
	ProcurementStatusDraft           ProcurementStatus = "draft"
	ProcurementStatusParsed          ProcurementStatus = "parsed"
	ProcurementStatusMatched         ProcurementStatus = "matched"
	ProcurementStatusOrderGenerated  ProcurementStatus = "order_generated"
	ProcurementStatusFinancingReady  ProcurementStatus = "financing_ready"
	ProcurementStatusReviewSubmitted ProcurementStatus = "review_submitted"
)

type WorkflowStage string

const (
	StageInstructionParsing         WorkflowStage = "instruction_parsing"
	StageProductMatching            WorkflowStage = "product_matching"
	StageLogisticsWarehouseMatching WorkflowStage = "logistics_warehouse_matching"
	StageOrderGeneration            WorkflowStage = "order_generation"
	StageFinancingMatching          WorkflowStage = "financing_matching"
	StageReviewSubmission           WorkflowStage = "review_submission"
)

type ParsedInstruction struct {
	ProductName           string   `json:"product_name"`
	Category              string   `json:"category"`
	Quantity              float64  `json:"quantity"`
	Unit                  string   `json:"unit"`
	Usage                 string   `json:"usage"`
	FinancingRequired     bool     `json:"financing_required"`
	Keywords              []string `json:"keywords"`
	Confidence            float64  `json:"confidence"`
	DeliveryAddress       string   `json:"delivery_address,omitempty"`
	DeliveryAddressSource string   `json:"delivery_address_source,omitempty"`
	DeliveryWarehouseID   uint64   `json:"delivery_warehouse_id,omitempty"`
	DeliveryWarehouseName string   `json:"delivery_warehouse_name,omitempty"`
}

func (p ParsedInstruction) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ParsedInstruction) Scan(value interface{}) error {
	return scanJSON(value, p)
}

type MatchedProduct struct {
	ProductID            uint64         `json:"product_id"`
	ProductNo            string         `json:"product_no"`
	ProductName          string         `json:"product_name"`
	Category             string         `json:"category"`
	SteelMill            string         `json:"steel_mill"`
	Quantity             float64        `json:"quantity"`
	AvailableQty         float64        `json:"available_qty"`
	UnitPrice            float64        `json:"unit_price"`
	TotalAmount          float64        `json:"total_amount"`
	Score                int            `json:"score"`
	MatchReason          string         `json:"match_reason"`
	EnterpriseID         uint64         `json:"enterprise_id"`
	Spec                 map[string]any `json:"spec,omitempty"`
	Warehouse            map[string]any `json:"warehouse,omitempty"`
	BundleNo             string         `json:"bundle_no,omitempty"`
	DeliveryDays         int            `json:"delivery_days,omitempty"`
	QuoteValidity        int            `json:"quote_validity,omitempty"`
	QualityCert          string         `json:"quality_cert,omitempty"`
	Description          string         `json:"description,omitempty"`
	OriginWarehouseName  string         `json:"origin_warehouse_name,omitempty"`
	OriginAddress        string         `json:"origin_address,omitempty"`
	DestinationAddress   string         `json:"destination_address,omitempty"`
	LogisticsServiceID   uint64         `json:"logistics_service_id,omitempty"`
	LogisticsServiceName string         `json:"logistics_service_name,omitempty"`
	LogisticsRoute       string         `json:"logistics_route,omitempty"`
	FreightEstimate      float64        `json:"freight_estimate,omitempty"`
	LandedTotal          float64        `json:"landed_total,omitempty"`
	UnitLandedPrice      float64        `json:"unit_landed_price,omitempty"`
	Selected             bool           `json:"selected,omitempty"`
}

type MatchedProducts []MatchedProduct

func (m MatchedProducts) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MatchedProducts) Scan(value interface{}) error {
	return scanJSON(value, m)
}

type OrderDraft struct {
	OrderID         uint64         `json:"order_id"`
	OrderNo         string         `json:"order_no"`
	Orders          OrderSummaries `json:"orders,omitempty"`
	ProductID       uint64         `json:"product_id"`
	ProductName     string         `json:"product_name"`
	Quantity        float64        `json:"quantity"`
	Unit            string         `json:"unit"`
	EstimatedAmt    float64        `json:"estimated_amount"`
	PaymentMethod   string         `json:"payment_method"`
	DeliveryNote    string         `json:"delivery_note"`
	ContactName     string         `json:"contact_name,omitempty"`
	ContactPhone    string         `json:"contact_phone,omitempty"`
	WarehouseName   string         `json:"warehouse_name,omitempty"`
	DeliveryAddress string         `json:"delivery_address,omitempty"`
	DeliveryMethod  string         `json:"delivery_method,omitempty"`
	Remark          string         `json:"remark,omitempty"`
	OrderDetail     map[string]any `json:"order_detail,omitempty"`
}

func (o OrderDraft) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OrderDraft) Scan(value interface{}) error {
	return scanJSON(value, o)
}

type OrderSummary struct {
	OrderID       uint64  `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	SellerID      uint64  `json:"seller_id"`
	SellerName    string  `json:"seller_name"`
	Quantity      float64 `json:"quantity"`
	TotalAmount   float64 `json:"total_amount"`
	ProductCount  int     `json:"product_count"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	DeliveryNote  string  `json:"delivery_note,omitempty"`
}

type OrderSummaries []OrderSummary

func (o OrderSummaries) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OrderSummaries) Scan(value interface{}) error {
	return scanJSON(value, o)
}

type FinancePlan struct {
	Provider    string  `json:"provider"`
	ProductName string  `json:"product_name"`
	Amount      float64 `json:"amount"`
	Rate        float64 `json:"rate"`
	TermDays    int     `json:"term_days"`
	Matched     bool    `json:"matched"`
	Reason      string  `json:"reason"`
}

type FinancePlans []FinancePlan

func (f FinancePlans) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *FinancePlans) Scan(value interface{}) error {
	return scanJSON(value, f)
}

type AuditItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Message  string `json:"message"`
}

type AuditItems []AuditItem

func (a AuditItems) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *AuditItems) Scan(value interface{}) error {
	return scanJSON(value, a)
}

type AIProcurement struct {
	ID                uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcurementNo     string            `gorm:"type:varchar(32);uniqueIndex;comment:AI采购单号" json:"procurement_no"`
	Prompt            string            `gorm:"type:text;comment:采购对话指令" json:"prompt"`
	ProductName       string            `gorm:"type:varchar(200);index;comment:解析商品名称" json:"product_name"`
	Category          string            `gorm:"type:varchar(80);index;comment:解析品类" json:"category"`
	Quantity          float64           `gorm:"type:decimal(12,2);comment:采购数量" json:"quantity"`
	Unit              string            `gorm:"type:varchar(20);comment:采购单位" json:"unit"`
	Usage             string            `gorm:"type:varchar(255);comment:采购用途" json:"usage"`
	FinancingRequired bool              `gorm:"default:false;comment:是否需要订单融资" json:"financing_required"`
	CurrentStage      WorkflowStage     `gorm:"type:varchar(50);index;comment:当前阶段" json:"current_stage"`
	Status            ProcurementStatus `gorm:"type:varchar(50);index;comment:状态" json:"status"`
	ParsedInstruction ParsedInstruction `gorm:"type:json;comment:指令解析结果" json:"parsed_instruction"`
	MatchedProducts   MatchedProducts   `gorm:"type:json;comment:商品匹配结果" json:"matched_products"`
	OrderDraft        OrderDraft        `gorm:"type:json;comment:订单草稿" json:"order_draft"`
	FinancePlans      FinancePlans      `gorm:"type:json;comment:融资匹配方案" json:"finance_plans"`
	AuditItems        AuditItems        `gorm:"type:json;comment:审核提交检查项" json:"audit_items"`
	HistorySaved      bool              `gorm:"default:true;index;comment:是否保存到历史AI采购" json:"history_saved"`
	EnterpriseID      uint64            `gorm:"index;comment:所属企业ID" json:"enterprise_id"`
	CreatedBy         uint64            `gorm:"index;comment:创建者" json:"created_by"`
	CreatedAt         time.Time         `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy         uint64            `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	ReviewSubmittedAt *time.Time        `gorm:"comment:审核提交时间" json:"review_submitted_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index;comment:删除时间" json:"-"`
}

func (AIProcurement) TableName() string {
	return "ai_procurements"
}

func (p *AIProcurement) BeforeCreate(tx *gorm.DB) error {
	if p.ProcurementNo == "" {
		p.ProcurementNo = generateProcurementNo()
	}
	if p.Unit == "" {
		p.Unit = "套"
	}
	if p.CurrentStage == "" {
		p.CurrentStage = StageInstructionParsing
	}
	if p.Status == "" {
		p.Status = ProcurementStatusParsed
	}
	return nil
}

func WorkflowStages() []WorkflowStage {
	return []WorkflowStage{
		StageInstructionParsing,
		StageProductMatching,
		StageLogisticsWarehouseMatching,
		StageOrderGeneration,
		StageFinancingMatching,
		StageReviewSubmission,
	}
}

func generateProcurementNo() string {
	return "AIP" + time.Now().Format("20060102150405") + randomDigits(4)
}

func randomDigits(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}

func scanJSON(value interface{}, target interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, target)
}
