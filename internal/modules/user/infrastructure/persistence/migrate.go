package persistence

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	db := database.GetDB()

	// 迁移用户相关表
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Enterprise{},
		&entity.Role{},
		&entity.UserRole{},
	)
	if err != nil {
		return err
	}

	// 初始化默认角色
	if err := initRoles(); err != nil {
		return err
	}

	return nil
}

// initRoles 初始化默认角色
func initRoles() error {
	db := database.GetDB()

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
			Description: "已注册的个人用户",
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
		var existingRole entity.Role
		result := db.Where("code = ?", role.Code).First(&existingRole)
		if result.Error != nil {
			// 角色不存在，创建
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
