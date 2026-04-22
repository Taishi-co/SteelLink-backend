package service

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// EnterpriseService 企业服务
type EnterpriseService struct {
	enterpriseRepo repository.EnterpriseRepository
}

// NewEnterpriseService 创建企业服务
func NewEnterpriseService() *EnterpriseService {
	return &EnterpriseService{
		enterpriseRepo: persistence.NewEnterpriseRepository(),
	}
}

// Create 创建企业申请
func (s *EnterpriseService) Create(ctx context.Context, req *dto.CreateEnterpriseRequest, operatorID uint64) (*dto.EnterpriseInfo, error) {
	existing, err := s.enterpriseRepo.FindByCreditCode(ctx, req.CreditCode)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if existing != nil {
		return nil, ecode.ErrEnterpriseExists
	}

	enterprise := &entity.Enterprise{
		Name:           req.Name,
		CreditCode:     req.CreditCode,
		EnterpriseRole: req.EnterpriseRole,
		Status:         entity.EnterpriseStatusPending,
		VerificationInfo: &entity.VerificationInfo{
			ContactName:  req.ContactName,
			ContactPhone: req.ContactPhone,
			Address:      req.Address,
		},
		CreatedBy: operatorID,
		UpdatedBy: operatorID,
	}

	if err := s.enterpriseRepo.Create(ctx, enterprise); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建企业申请失败")
	}

	return toEnterpriseInfo(enterprise), nil
}

// Verify 提交企业认证材料
func (s *EnterpriseService) Verify(ctx context.Context, id uint64, req *dto.VerifyEnterpriseRequest, operatorID uint64) (*dto.EnterpriseInfo, error) {
	enterprise, err := s.enterpriseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if enterprise == nil {
		return nil, ecode.ErrEnterpriseNotFound
	}

	enterprise.VerificationInfo = &entity.VerificationInfo{
		LegalPerson:     req.LegalPerson,
		LegalIdCard:     req.LegalIdCard,
		BusinessLicense: req.BusinessLicense,
		BankAccount:     req.BankAccount,
		BankName:        req.BankName,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		Address:         req.Address,
	}
	enterprise.Status = entity.EnterpriseStatusPending
	enterprise.UpdatedBy = operatorID

	if err := s.enterpriseRepo.Update(ctx, enterprise); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("提交企业认证失败")
	}

	return toEnterpriseInfo(enterprise), nil
}

// GetByID 查询企业详情
func (s *EnterpriseService) GetByID(ctx context.Context, id uint64) (*dto.EnterpriseInfo, error) {
	enterprise, err := s.enterpriseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if enterprise == nil {
		return nil, ecode.ErrEnterpriseNotFound
	}

	return toEnterpriseInfo(enterprise), nil
}

// List 查询企业列表
func (s *EnterpriseService) List(ctx context.Context, req *dto.EnterpriseListRequest) (*dto.EnterpriseListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	enterprises, total, err := s.enterpriseRepo.List(ctx, &repository.EnterpriseQuery{
		Page:     req.Page,
		PageSize: req.PageSize,
		Name:     req.Name,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}

	list := make([]*dto.EnterpriseInfo, 0, len(enterprises))
	for _, enterprise := range enterprises {
		list = append(list, toEnterpriseInfo(enterprise))
	}

	return &dto.EnterpriseListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Audit 审核企业
func (s *EnterpriseService) Audit(ctx context.Context, id uint64, req *dto.AuditEnterpriseRequest, operatorID uint64) error {
	enterprise, err := s.enterpriseRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if enterprise == nil {
		return ecode.ErrEnterpriseNotFound
	}

	status := entity.EnterpriseStatus(req.Status)
	if err := s.enterpriseRepo.UpdateStatus(ctx, id, status, operatorID); err != nil {
		return ecode.ErrDatabase.WithMessage("审核企业失败")
	}

	if req.Status == int(entity.EnterpriseStatusRejected) {
		if enterprise.Remark == nil {
			enterprise.Remark = entity.Remark{}
		}
		enterprise.Remark["audit_reason"] = req.Reason
		enterprise.UpdatedBy = operatorID
		if err := s.enterpriseRepo.Update(ctx, enterprise); err != nil {
			return ecode.ErrDatabase.WithMessage("保存审核意见失败")
		}
	}

	return nil
}

func toEnterpriseInfo(enterprise *entity.Enterprise) *dto.EnterpriseInfo {
	if enterprise == nil {
		return nil
	}

	var verificationInfo *dto.VerificationInfo
	if enterprise.VerificationInfo != nil {
		verificationInfo = &dto.VerificationInfo{
			LegalPerson:     enterprise.VerificationInfo.LegalPerson,
			LegalIdCard:     enterprise.VerificationInfo.LegalIdCard,
			BusinessLicense: enterprise.VerificationInfo.BusinessLicense,
			BankAccount:     enterprise.VerificationInfo.BankAccount,
			BankName:        enterprise.VerificationInfo.BankName,
			ContactName:     enterprise.VerificationInfo.ContactName,
			ContactPhone:    enterprise.VerificationInfo.ContactPhone,
			Address:         enterprise.VerificationInfo.Address,
		}
	}

	return &dto.EnterpriseInfo{
		ID:               enterprise.ID,
		EnterpriseCode:   enterprise.EnterpriseCode,
		Name:             enterprise.Name,
		CreditCode:       enterprise.CreditCode,
		EnterpriseRole:   enterprise.EnterpriseRole,
		Status:           int(enterprise.Status),
		VerificationInfo: verificationInfo,
		CreatedAt:        enterprise.CreatedAt,
		UpdatedAt:        enterprise.UpdatedAt,
	}
}
