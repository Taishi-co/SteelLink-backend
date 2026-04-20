package persistence

import (
	"context"
	"errors"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

var _ repository.EnterpriseRepository = (*enterpriseRepository)(nil)

type enterpriseRepository struct {
	db *gorm.DB
}

// NewEnterpriseRepository 创建企业仓储实例
func NewEnterpriseRepository() repository.EnterpriseRepository {
	return &enterpriseRepository{db: database.GetDB()}
}

// Create 创建企业
func (r *enterpriseRepository) Create(ctx context.Context, enterprise *entity.Enterprise) error {
	return r.db.WithContext(ctx).Create(enterprise).Error
}

// Update 更新企业
func (r *enterpriseRepository) Update(ctx context.Context, enterprise *entity.Enterprise) error {
	return r.db.WithContext(ctx).Save(enterprise).Error
}

// Delete 软删除企业
func (r *enterpriseRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Enterprise{}, id).Error
}

// FindByID 根据ID查询企业
func (r *enterpriseRepository) FindByID(ctx context.Context, id uint64) (*entity.Enterprise, error) {
	var enterprise entity.Enterprise
	err := r.db.WithContext(ctx).First(&enterprise, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &enterprise, nil
}

// FindByEnterpriseCode 根据企业代码查询
func (r *enterpriseRepository) FindByEnterpriseCode(ctx context.Context, code string) (*entity.Enterprise, error) {
	var enterprise entity.Enterprise
	err := r.db.WithContext(ctx).Where("enterprise_code = ?", code).First(&enterprise).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &enterprise, nil
}

// FindByCreditCode 根据统一社会信用代码查询
func (r *enterpriseRepository) FindByCreditCode(ctx context.Context, creditCode string) (*entity.Enterprise, error) {
	var enterprise entity.Enterprise
	err := r.db.WithContext(ctx).Where("credit_code = ?", creditCode).First(&enterprise).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &enterprise, nil
}

// List 查询企业列表
func (r *enterpriseRepository) List(ctx context.Context, query *repository.EnterpriseQuery) ([]*entity.Enterprise, int64, error) {
	var enterprises []*entity.Enterprise
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Enterprise{})

	// 构建查询条件
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR credit_code LIKE ? OR enterprise_code LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
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

	if err := db.Offset(offset).Limit(query.PageSize).Order("id DESC").Find(&enterprises).Error; err != nil {
		return nil, 0, err
	}

	return enterprises, total, nil
}

// UpdateStatus 更新企业状态
func (r *enterpriseRepository) UpdateStatus(ctx context.Context, id uint64, status entity.EnterpriseStatus, updatedBy uint64) error {
	return r.db.WithContext(ctx).Model(&entity.Enterprise{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_by": updatedBy,
		}).Error
}
