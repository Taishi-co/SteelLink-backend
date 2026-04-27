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
)

type ProcurementActivity struct {
	ID                 uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcurementNo      string            `gorm:"type:varchar(40);uniqueIndex;comment:集采编号" json:"procurement_no"`
	Title              string            `gorm:"type:varchar(200);index;comment:集采标题" json:"title"`
	LeaderName         string            `gorm:"type:varchar(120);comment:牵头方" json:"leader_name"`
	Status             ProcurementStatus `gorm:"type:varchar(32);index;comment:状态" json:"status"`
	Category           string            `gorm:"type:varchar(80);index;comment:品类" json:"category"`
	Target             string            `gorm:"type:varchar(120);comment:采购标的" json:"target"`
	TargetQuantity     float64           `gorm:"type:decimal(14,2);comment:目标数量" json:"target_quantity"`
	Unit               string            `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	BudgetMin          float64           `gorm:"type:decimal(16,2);comment:预算下限" json:"budget_min"`
	BudgetMax          float64           `gorm:"type:decimal(16,2);comment:预算上限" json:"budget_max"`
	BudgetDisplay      string            `gorm:"type:varchar(60);comment:预算展示" json:"budget_display"`
	Stage              string            `gorm:"type:varchar(60);comment:当前阶段" json:"stage"`
	Progress           int               `gorm:"comment:进度百分比" json:"progress"`
	Region             string            `gorm:"type:varchar(120);comment:交付区域" json:"region"`
	OpenDate           string            `gorm:"type:varchar(40);comment:开标日期" json:"open_date"`
	ParticipatingCount int               `gorm:"comment:参采企业" json:"participating_count"`
	DemandQuantity     float64           `gorm:"type:decimal(14,2);comment:归集需求" json:"demand_quantity"`
	QualifiedSuppliers int               `gorm:"comment:入围供应商" json:"qualified_suppliers"`
	QuoteCount         int               `gorm:"comment:收到报价" json:"quote_count"`
	ContractCount      int               `gorm:"comment:已签合同" json:"contract_count"`
	DeliveryDays       int               `gorm:"comment:交货周期天数" json:"delivery_days"`
	QuoteMode          string            `gorm:"type:varchar(40);comment:报价方式" json:"quote_mode"`
	EvaluationMode     string            `gorm:"type:varchar(40);comment:评审标准" json:"evaluation_mode"`
	Documents          StringList        `gorm:"type:json;comment:资质文件" json:"documents"`
	CreatedAt          time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `gorm:"index" json:"-"`
	Logs               []ProcurementLog  `gorm:"foreignKey:ActivityID;references:ID" json:"logs"`
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
