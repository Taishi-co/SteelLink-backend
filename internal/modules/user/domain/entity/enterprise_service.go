package entity

import (
	"time"

	"gorm.io/gorm"
)

type SupplyChainRole string

const (
	SupplyRoleSupplier  SupplyChainRole = "supplier"
	SupplyRoleBuyer     SupplyChainRole = "buyer"
	SupplyRoleLogistics SupplyChainRole = "logistics"
	SupplyRoleWarehouse SupplyChainRole = "warehouse"
	SupplyRoleFund      SupplyChainRole = "fund"
)

type PlatformService string

const (
	ServiceSales       PlatformService = "sales"
	ServiceProcurement PlatformService = "procurement"
	ServiceLogistics   PlatformService = "logistics"
	ServiceWarehouse   PlatformService = "warehouse"
	ServiceFinance     PlatformService = "finance"
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
	SupplyRole     SupplyChainRole         `gorm:"type:varchar(32);not null;comment:供应链角色" json:"supply_role"`
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
	return "enterprise_service_permissions"
}

type EnterpriseServiceApplication struct {
	ID             uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseID   uint64            `gorm:"not null;index;comment:企业ID" json:"enterprise_id"`
	EnterpriseCode string            `gorm:"type:varchar(32);not null;index;comment:企业代码" json:"enterprise_code"`
	Service        PlatformService   `gorm:"type:varchar(32);not null;index;comment:平台服务" json:"service"`
	SupplyRole     SupplyChainRole   `gorm:"type:varchar(32);not null;comment:供应链角色" json:"supply_role"`
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

func ValidSupplyChainRole(role string) bool {
	switch SupplyChainRole(role) {
	case SupplyRoleSupplier, SupplyRoleBuyer, SupplyRoleLogistics, SupplyRoleWarehouse, SupplyRoleFund:
		return true
	}
	return false
}

func ValidPlatformService(service string) bool {
	switch PlatformService(service) {
	case ServiceSales, ServiceProcurement, ServiceLogistics, ServiceWarehouse, ServiceFinance:
		return true
	}
	return false
}

func ExpectedSupplyRole(service PlatformService) SupplyChainRole {
	switch service {
	case ServiceSales:
		return SupplyRoleSupplier
	case ServiceProcurement:
		return SupplyRoleBuyer
	case ServiceLogistics:
		return SupplyRoleLogistics
	case ServiceWarehouse:
		return SupplyRoleWarehouse
	case ServiceFinance:
		return SupplyRoleFund
	default:
		return ""
	}
}
