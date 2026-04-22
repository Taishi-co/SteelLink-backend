package persistence

import (
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
	); err != nil {
		return err
	}

	return initRoles(db)
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
