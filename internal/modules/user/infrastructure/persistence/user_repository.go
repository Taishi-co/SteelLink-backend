package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.UserRepository = (*userRepository)(nil)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() repository.UserRepository {
	return &userRepository{db: database.GetDB()}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 软删除用户
func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

// FindByID 根据ID查询用户
func (r *userRepository) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUserNo 根据用户编号查询
func (r *userRepository) FindByUserNo(ctx context.Context, userNo string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("user_no = ?", userNo).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据手机号查询
func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查询
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPersonalCode 根据个人代码查询
func (r *userRepository) FindByPersonalCode(ctx context.Context, personalCode string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("personal_code = ?", personalCode).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// List 查询用户列表
func (r *userRepository) List(ctx context.Context, query *repository.UserQuery) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.User{})

	// 构建查询条件
	if query.Phone != "" {
		db = db.Where("phone = ?", query.Phone)
	}
	if query.Email != "" {
		db = db.Where("email = ?", query.Email)
	}
	if query.Role != "" {
		db = db.Where("role = ?", query.Role)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.EnterpriseCode != "" {
		db = db.Where("enterprise_code = ?", query.EnterpriseCode)
	}
	if query.Keyword != "" {
		db = db.Where("phone LIKE ? OR email LIKE ? OR user_no LIKE ? OR uid LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
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

	if err := db.Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ExistsByPhone 检查手机号是否存在
func (r *userRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, id uint64, status entity.UserStatus, updatedBy uint64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_by": updatedBy,
		}).Error
}

func (r *userRepository) CountSubAccounts(ctx context.Context, parentID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("parent_id = ? AND deleted_at IS NULL", parentID).Count(&count).Error
	return count, err
}

func (r *userRepository) FindSubAccounts(ctx context.Context, parentID uint64) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).Where("parent_id = ? AND deleted_at IS NULL", parentID).Find(&users).Error
	return users, err
}

func (r *userRepository) FindByEnterpriseCode(ctx context.Context, enterpriseCode string) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).Where("enterprise_code = ? AND deleted_at IS NULL", enterpriseCode).Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateRoleByEnterpriseCode(ctx context.Context, enterpriseCode string, role entity.UserType, updatedBy uint64) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("enterprise_code = ? AND deleted_at IS NULL", enterpriseCode).
		Updates(map[string]interface{}{"role": role, "updated_by": updatedBy}).Error
}
