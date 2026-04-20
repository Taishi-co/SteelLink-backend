package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// Create 创建角色
	Create(ctx context.Context, role *entity.Role) error

	// Update 更新角色
	Update(ctx context.Context, role *entity.Role) error

	// Delete 软删除角色
	Delete(ctx context.Context, id uint64) error

	// FindByID 根据ID查询角色
	FindByID(ctx context.Context, id uint64) (*entity.Role, error)

	// FindByCode 根据角色编码查询
	FindByCode(ctx context.Context, code string) (*entity.Role, error)

	// List 查询角色列表
	List(ctx context.Context, query *RoleQuery) ([]*entity.Role, int64, error)

	// FindByUserID 查询用户的角色列表
	FindByUserID(ctx context.Context, userID uint64) ([]*entity.Role, error)

	// AssignRole 给用户分配角色
	AssignRole(ctx context.Context, userRole *entity.UserRole) error

	// RemoveRole 移除用户角色
	RemoveRole(ctx context.Context, userID, roleID uint64) error

	// FindUserPermissions 查询用户的所有权限
	FindUserPermissions(ctx context.Context, userID uint64) ([]string, error)
}

// RoleQuery 角色查询条件
type RoleQuery struct {
	Page     int
	PageSize int
	Code     string
	Name     string
}
