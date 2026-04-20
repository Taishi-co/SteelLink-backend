package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *entity.User) error

	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error

	// Delete 软删除用户
	Delete(ctx context.Context, id uint64) error

	// FindByID 根据ID查询用户
	FindByID(ctx context.Context, id uint64) (*entity.User, error)

	// FindByUserNo 根据用户编号查询
	FindByUserNo(ctx context.Context, userNo string) (*entity.User, error)

	// FindByPhone 根据手机号查询
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)

	// FindByEmail 根据邮箱查询
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByPersonalCode 根据个人代码查询
	FindByPersonalCode(ctx context.Context, personalCode string) (*entity.User, error)

	// List 查询用户列表
	List(ctx context.Context, query *UserQuery) ([]*entity.User, int64, error)

	// ExistsByPhone 检查手机号是否存在
	ExistsByPhone(ctx context.Context, phone string) (bool, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// UpdateStatus 更新用户状态
	UpdateStatus(ctx context.Context, id uint64, status entity.UserStatus, updatedBy uint64) error
}

// UserQuery 用户查询条件
type UserQuery struct {
	Page           int
	PageSize       int
	Phone          string
	Email          string
	Role           string
	Status         *int
	EnterpriseCode string
	Keyword        string // 关键词搜索（手机号/邮箱/用户编号）
}
