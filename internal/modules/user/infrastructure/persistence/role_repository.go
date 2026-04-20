package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.RoleRepository = (*roleRepository)(nil)

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository() repository.RoleRepository {
	return &roleRepository{db: database.GetDB()}
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete 软删除角色
func (r *roleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

// FindByID 根据ID查询角色
func (r *roleRepository) FindByID(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// FindByCode 根据角色编码查询
func (r *roleRepository) FindByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// List 查询角色列表
func (r *roleRepository) List(ctx context.Context, query *repository.RoleQuery) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Role{})

	// 构建查询条件
	if query.Code != "" {
		db = db.Where("code = ?", query.Code)
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if query.Page <= 0 {
		offset = 0
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	if err := db.Offset(offset).Limit(query.PageSize).Order("id ASC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// FindByUserID 查询用户的角色列表
func (r *roleRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

// AssignRole 给用户分配角色
func (r *roleRepository) AssignRole(ctx context.Context, userRole *entity.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// RemoveRole 移除用户角色
func (r *roleRepository) RemoveRole(ctx context.Context, userID, roleID uint64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&entity.UserRole{}).Error
}

// FindUserPermissions 查询用户的所有权限
func (r *roleRepository) FindUserPermissions(ctx context.Context, userID uint64) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("roles").
		Select("DISTINCT roles.permissions").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.deleted_at IS NULL", userID).
		Scan(&permissions).Error
	return permissions, err
}
