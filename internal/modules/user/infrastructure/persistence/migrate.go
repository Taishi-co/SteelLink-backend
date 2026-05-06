package persistence

import (
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

func AutoMigrate() error {
	db := database.GetDB()

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Enterprise{},
		&entity.Role{},
		&entity.UserRole{},
		&entity.EnterpriseServicePermission{},
		&entity.EnterpriseServiceApplication{},
		&entity.EnterpriseApplicationDocument{},
	); err != nil {
		return err
	}

	if err := initRoles(db); err != nil {
		return err
	}
	return seedLegacyEnterpriseServicePermissions(db)
}

func initRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{
			Code:        "guest",
			Name:        "游客",
			Description: "未登录用户",
			Permissions: []string{"product:read"},
		},
		{
			Code:        "personal",
			Name:        "个人用户",
			Description: "已注册个人用户",
			Permissions: []string{"product:read", "order:create", "order:read"},
		},
		{
			Code:        "enterprise",
			Name:        "企业用户",
			Description: "已加入企业的用户",
			Permissions: []string{"product:read", "product:create", "order:create", "order:read", "order:update"},
		},
		{
			Code:        "operator",
			Name:        "运营",
			Description: "平台运营人员",
			Permissions: []string{"*"},
		},
		{
			Code:        "super_admin",
			Name:        "超级管理员",
			Description: "系统超级管理员",
			Permissions: []string{"*"},
		},
	}

	for _, role := range roles {
		var existing entity.Role
		if err := db.Where("code = ?", role.Code).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedLegacyEnterpriseServicePermissions(db *gorm.DB) error {
	type legacyEnterprise struct {
		ID             uint64
		EnterpriseCode string
		EnterpriseRole string
		CreatedBy      uint64
		UpdatedBy      uint64
	}

	var enterprises []legacyEnterprise
	if err := db.Table("enterprises").
		Select("id, enterprise_code, enterprise_role, created_by, updated_by").
		Where("deleted_at IS NULL AND status = ? AND COALESCE(enterprise_role, '') <> ''", entity.EnterpriseStatusApproved).
		Find(&enterprises).Error; err != nil {
		return err
	}

	for _, enterprise := range enterprises {
		services := legacyServicesForEnterpriseRole(enterprise.EnterpriseRole)
		for _, service := range services {
			var count int64
			if err := db.Model(&entity.EnterpriseServicePermission{}).
				Where("enterprise_id = ? AND service = ?", enterprise.ID, service).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			now := time.Now()
			permission := entity.EnterpriseServicePermission{
				EnterpriseID:   enterprise.ID,
				EnterpriseCode: enterprise.EnterpriseCode,
				Service:        service,
				SupplyRole:     entity.ExpectedSupplyRole(service),
				Status:         entity.ServicePermissionApproved,
				ApprovedAt:     &now,
				CreatedBy:      enterprise.CreatedBy,
				UpdatedBy:      enterprise.UpdatedBy,
			}
			if err := db.Create(&permission).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func legacyServicesForEnterpriseRole(role string) []entity.PlatformService {
	switch entity.EnterpriseRole(role) {
	case entity.EnterpriseRoleSeller:
		return []entity.PlatformService{entity.ServiceSales}
	case entity.EnterpriseRoleBuyer:
		return []entity.PlatformService{entity.ServiceProcurement}
	case entity.EnterpriseRoleLogistics:
		return []entity.PlatformService{entity.ServiceLogistics}
	case entity.EnterpriseRoleWarehouse:
		return []entity.PlatformService{entity.ServiceWarehouse}
	case entity.EnterpriseRoleFund:
		return []entity.PlatformService{entity.ServiceFinance}
	default:
		return nil
	}
}
