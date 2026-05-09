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
	if err := db.Table(entity.TradeCapabilityPermissionTable).AutoMigrate(&entity.EnterpriseServicePermission{}); err != nil {
		return err
	}
	if err := db.Table(entity.ProviderServicePermissionTable).AutoMigrate(&entity.EnterpriseServicePermission{}); err != nil {
		return err
	}

	if err := dropLegacySupplyRoleColumns(db); err != nil {
		return err
	}

	if err := initRoles(db); err != nil {
		return err
	}
	if err := migrateLegacyServicePermissions(db); err != nil {
		return err
	}
	if err := seedLegacyEnterpriseServicePermissions(db); err != nil {
		return err
	}
	return seedFullPermissionTestAccounts(db)
}

func dropLegacySupplyRoleColumns(db *gorm.DB) error {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enterprise_service_permissions' AND COLUMN_NAME = 'supply_role'").Scan(&count)
	if count > 0 {
		if err := db.Exec("ALTER TABLE enterprise_service_permissions DROP COLUMN supply_role").Error; err != nil {
			return err
		}
	}

	db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'enterprise_service_applications' AND COLUMN_NAME = 'supply_role'").Scan(&count)
	if count > 0 {
		if err := db.Exec("ALTER TABLE enterprise_service_applications DROP COLUMN supply_role").Error; err != nil {
			return err
		}
	}
	return nil
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
			if err := db.Table(entity.CapabilityPermissionTable(service)).
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
				Status:         entity.ServicePermissionApproved,
				ApprovedAt:     &now,
				CreatedBy:      enterprise.CreatedBy,
				UpdatedBy:      enterprise.UpdatedBy,
			}
			if err := db.Table(entity.CapabilityPermissionTable(service)).Create(&permission).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateLegacyServicePermissions(db *gorm.DB) error {
	var permissions []entity.EnterpriseServicePermission
	if err := db.Table(entity.LegacyServicePermissionTable).Where("deleted_at IS NULL").Find(&permissions).Error; err != nil {
		return err
	}
	for _, permission := range permissions {
		if !entity.ValidPlatformService(string(permission.Service)) {
			continue
		}
		table := entity.CapabilityPermissionTable(permission.Service)
		var count int64
		if err := db.Table(table).
			Where("enterprise_id = ? AND service = ?", permission.EnterpriseID, permission.Service).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Table(table).Create(&permission).Error; err != nil {
			return err
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

func seedFullPermissionTestAccounts(db *gorm.DB) error {
	testPhones := []string{"15127439999"}
	allServices := []entity.PlatformService{
		entity.ServiceSales,
		entity.ServiceProcurement,
		entity.ServiceLogistics,
		entity.ServiceWarehouse,
		entity.ServiceFinance,
	}

	for _, phone := range testPhones {
		var user entity.User
		if err := db.Where("phone = ? AND deleted_at IS NULL", phone).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return err
		}
		if user.EnterpriseCode == "" {
			continue
		}

		var enterprise entity.Enterprise
		if err := db.Where("enterprise_code = ? AND deleted_at IS NULL", user.EnterpriseCode).First(&enterprise).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return err
		}
		if enterprise.Status != entity.EnterpriseStatusApproved {
			continue
		}

		if user.Role != entity.TypeEnterprise || user.IdentityStatus != entity.IdentityVerified {
			if err := db.Model(&entity.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]interface{}{
					"role":            entity.TypeEnterprise,
					"identity_status": entity.IdentityVerified,
					"updated_by":      user.ID,
				}).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		for _, service := range allServices {
			table := entity.CapabilityPermissionTable(service)
			var permission entity.EnterpriseServicePermission
			err := db.Table(table).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&permission).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == gorm.ErrRecordNotFound {
				permission = entity.EnterpriseServicePermission{
					EnterpriseID:   enterprise.ID,
					EnterpriseCode: enterprise.EnterpriseCode,
					Service:        service,
					Status:         entity.ServicePermissionApproved,
					ApprovedAt:     &now,
					CreatedBy:      user.ID,
					UpdatedBy:      user.ID,
				}
				if err := db.Table(table).Create(&permission).Error; err != nil {
					return err
				}
				continue
			}

			if permission.Status != entity.ServicePermissionApproved {
				permission.Status = entity.ServicePermissionApproved
				permission.AuditReason = ""
				permission.ApprovedAt = &now
				permission.UpdatedBy = user.ID
				if err := db.Table(table).Save(&permission).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
