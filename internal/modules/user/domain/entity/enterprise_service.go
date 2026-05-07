package entity

import (
	"time"

	"gorm.io/gorm"
)

type PlatformService string

const (
	ServiceSales       PlatformService = "sales"
	ServiceProcurement PlatformService = "procurement"
	ServiceLogistics   PlatformService = "logistics"
	ServiceWarehouse   PlatformService = "warehouse"
	ServiceFinance     PlatformService = "finance"
)

const (
	TradeCapabilityPermissionTable = "enterprise_trade_capability_permissions"
	ProviderServicePermissionTable = "enterprise_provider_service_permissions"
	LegacyServicePermissionTable   = "enterprise_service_permissions"
)

type ServicePermissionStatus int

const (
	ServicePermissionNone      ServicePermissionStatus = 0
	ServicePermissionPending   ServicePermissionStatus = 1
	ServicePermissionApproved  ServicePermissionStatus = 2
	ServicePermissionRejected  ServicePermissionStatus = 3
	ServicePermissionSuspended ServicePermissionStatus = 4
)

type ApplicationStatus int

const (
	ApplicationPending  ApplicationStatus = 1
	ApplicationApproved ApplicationStatus = 2
	ApplicationRejected ApplicationStatus = 3
)

type EnterpriseServicePermission struct {
	ID             uint64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID   uint64                  `gorm:"not null;index;uniqueIndex:uk_enterprise_service;comment:企业ID" json:"enterprise_id"`
	EnterpriseCode string                  `gorm:"type:varchar(32);not null;index;comment:企业代码" json:"enterprise_code"`
	Service        PlatformService         `gorm:"type:varchar(32);not null;uniqueIndex:uk_enterprise_service;comment:平台服务" json:"service"`
	Status         ServicePermissionStatus `gorm:"type:tinyint;default:1;index;comment:状态 1审核中2通过3驳回4暂停" json:"status"`
	AuditReason    string                  `gorm:"type:varchar(500);default:;comment:审核/暂停原因" json:"audit_reason"`
	ApprovedAt     *time.Time              `gorm:"comment:通过时间" json:"approved_at"`
	CreatedBy      uint64                  `gorm:"comment:创建者" json:"created_by"`
	CreatedAt      time.Time               `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy      uint64                  `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt      time.Time               `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt          `gorm:"index;comment:删除时间" json:"-"`
}

func (EnterpriseServicePermission) TableName() string {
	return LegacyServicePermissionTable
}

type EnterpriseServiceApplication struct {
	ID             uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID   uint64            `gorm:"not null;index;comment:企业ID" json:"enterprise_id"`
	EnterpriseCode string            `gorm:"type:varchar(32);not null;index;comment:企业代码" json:"enterprise_code"`
	Service        PlatformService   `gorm:"type:varchar(32);not null;index;comment:平台服务" json:"service"`
	Status         ApplicationStatus `gorm:"type:tinyint;default:1;index;comment:状态 1审核中2通过3驳回" json:"status"`
	AuditReason    string            `gorm:"type:varchar(500);default:;comment:审核原因" json:"audit_reason"`
	SubmittedBy    uint64            `gorm:"not null;index;comment:提交人" json:"submitted_by"`
	AuditedBy      uint64            `gorm:"comment:审核人" json:"audited_by"`
	AuditedAt      *time.Time        `gorm:"comment:审核时间" json:"audited_at"`
	CreatedAt      time.Time         `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index;comment:删除时间" json:"-"`
}

func (EnterpriseServiceApplication) TableName() string {
	return "enterprise_service_applications"
}

type EnterpriseApplicationDocument struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ApplicationID uint64         `gorm:"not null;index;comment:申请ID" json:"application_id"`
	DocumentType  string         `gorm:"type:varchar(64);not null;comment:材料类型" json:"document_type"`
	FileURL       string         `gorm:"type:varchar(500);not null;comment:文件URL" json:"file_url"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (EnterpriseApplicationDocument) TableName() string {
	return "enterprise_application_documents"
}

func ValidPlatformService(service string) bool {
	switch PlatformService(service) {
	case ServiceSales, ServiceProcurement, ServiceLogistics, ServiceWarehouse, ServiceFinance:
		return true
	}
	return false
}

func IsTradeCapability(service PlatformService) bool {
	return service == ServiceSales || service == ServiceProcurement
}

func IsProviderService(service PlatformService) bool {
	return service == ServiceLogistics || service == ServiceWarehouse || service == ServiceFinance
}

func CapabilityKind(service PlatformService) string {
	if IsTradeCapability(service) {
		return "trade"
	}
	if IsProviderService(service) {
		return "provider"
	}
	return ""
}

func CapabilityPermissionTable(service PlatformService) string {
	if IsTradeCapability(service) {
		return TradeCapabilityPermissionTable
	}
	return ProviderServicePermissionTable
}

func PermissionTablesForServices(services []string) map[string][]string {
	tables := map[string][]string{}
	for _, service := range services {
		if !ValidPlatformService(service) {
			continue
		}
		table := CapabilityPermissionTable(PlatformService(service))
		tables[table] = append(tables[table], service)
	}
	return tables
}

func ServiceToEnterpriseRole(service PlatformService) string {
	switch service {
	case ServiceSales:
		return "seller"
	case ServiceProcurement:
		return "buyer"
	case ServiceLogistics:
		return "logistics"
	case ServiceWarehouse:
		return "warehouse"
	case ServiceFinance:
		return "fund"
	default:
		return ""
	}
}
