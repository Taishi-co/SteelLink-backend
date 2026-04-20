package repository

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
)

// EnterpriseRepository 企业仓储接口
type EnterpriseRepository interface {
	// Create 创建企业
	Create(ctx context.Context, enterprise *entity.Enterprise) error

	// Update 更新企业
	Update(ctx context.Context, enterprise *entity.Enterprise) error

	// Delete 软删除企业
	Delete(ctx context.Context, id uint64) error

	// FindByID 根据ID查询企业
	FindByID(ctx context.Context, id uint64) (*entity.Enterprise, error)

	// FindByEnterpriseCode 根据企业代码查询
	FindByEnterpriseCode(ctx context.Context, code string) (*entity.Enterprise, error)

	// FindByCreditCode 根据统一社会信用代码查询
	FindByCreditCode(ctx context.Context, creditCode string) (*entity.Enterprise, error)

	// List 查询企业列表
	List(ctx context.Context, query *EnterpriseQuery) ([]*entity.Enterprise, int64, error)

	// UpdateStatus 更新企业状态
	UpdateStatus(ctx context.Context, id uint64, status entity.EnterpriseStatus, updatedBy uint64) error
}

// EnterpriseQuery 企业查询条件
type EnterpriseQuery struct {
	Page     int
	PageSize int
	Name     string
	Status   *int
	Keyword  string
}
