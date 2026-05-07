package service

import (
	"context"
	"errors"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

// EnterpriseService 企业服务
type EnterpriseService struct {
	enterpriseRepo repository.EnterpriseRepository
	userRepo       repository.UserRepository
}

func NewEnterpriseService() *EnterpriseService {
	return &EnterpriseService{
		enterpriseRepo: persistence.NewEnterpriseRepository(),
		userRepo:       persistence.NewUserRepository(),
	}
}

// Create 创建企业申请
func (s *EnterpriseService) Create(ctx context.Context, req *dto.CreateEnterpriseRequest, operatorID uint64) (*dto.EnterpriseInfo, error) {
	user, err := s.userRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	if !user.IsIdentityVerified() {
		return nil, ecode.ErrBadRequest.WithMessage("请先完成实名认证后再创建企业")
	}
	if user.EnterpriseCode != "" {
		return nil, ecode.ErrBadRequest.WithMessage("您已绑定企业，不可重复创建。1个实名个人仅可绑定1家企业")
	}
	if !entity.ValidEnterpriseRole(req.EnterpriseRole) {
		return nil, ecode.ErrBadRequest.WithMessage("企业角色无效，仅支持 buyer/seller/logistics/warehouse/fund")
	}

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

	primaryService := enterpriseRoleToService(req.EnterpriseRole)

	db := database.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(enterprise).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("创建企业申请失败")
		}

		user.EnterpriseCode = enterprise.EnterpriseCode
		user.UpdatedBy = operatorID
		if err := tx.Save(user).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("更新用户企业绑定失败")
		}

		if err := ensureInitialServiceApplication(tx, enterprise, primaryService, req.Documents, operatorID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if coded, ok := err.(*ecode.Error); ok {
			return nil, coded
		}
		return nil, ecode.ErrDatabase.WithMessage("创建企业申请失败")
	}

	return toEnterpriseInfo(enterprise), nil
}

func enterpriseRoleToService(role string) entity.PlatformService {
	switch role {
	case "buyer":
		return entity.ServiceProcurement
	case "seller":
		return entity.ServiceSales
	case "logistics":
		return entity.ServiceLogistics
	case "warehouse":
		return entity.ServiceWarehouse
	case "fund":
		return entity.ServiceFinance
	default:
		return entity.ServiceProcurement
	}
}

func ensureInitialServiceApplication(tx *gorm.DB, enterprise *entity.Enterprise, service entity.PlatformService, documents []dto.ApplicationDocumentRequest, operatorID uint64) error {
	var application entity.EnterpriseServiceApplication
	err := tx.Where("enterprise_id = ? AND service = ? AND status = ?", enterprise.ID, service, entity.ApplicationPending).First(&application).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		application = entity.EnterpriseServiceApplication{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			Status:         entity.ApplicationPending,
			AuditReason:    "",
			SubmittedBy:    operatorID,
		}
		if err := tx.Create(&application).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("创建初始服务申请失败")
		}
	} else if err != nil {
		return ecode.ErrDatabase.WithMessage("查询初始服务申请失败")
	} else {
		application.AuditReason = ""
		application.SubmittedBy = operatorID
		if err := tx.Save(&application).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("更新初始服务申请失败")
		}
		if err := tx.Where("application_id = ?", application.ID).Delete(&entity.EnterpriseApplicationDocument{}).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("清理旧服务资质文档失败")
		}
	}

	docs := make([]entity.EnterpriseApplicationDocument, 0, len(documents))
	for _, doc := range documents {
		docs = append(docs, entity.EnterpriseApplicationDocument{
			ApplicationID: application.ID,
			DocumentType:  doc.DocumentType,
			FileURL:       doc.FileURL,
		})
	}
	if len(docs) > 0 {
		if err := tx.Create(&docs).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("创建服务资质文档失败")
		}
	}

	permissionTable := entity.CapabilityPermissionTable(service)
	var permission entity.EnterpriseServicePermission
	err = tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		permission = entity.EnterpriseServicePermission{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			Status:         entity.ServicePermissionPending,
			AuditReason:    "",
			CreatedBy:      operatorID,
			UpdatedBy:      operatorID,
		}
		if err := tx.Table(permissionTable).Create(&permission).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("创建初始服务权限失败")
		}
		return nil
	}
	if err != nil {
		return ecode.ErrDatabase.WithMessage("查询初始服务权限失败")
	}

	permission.Status = entity.ServicePermissionPending
	permission.AuditReason = ""
	permission.ApprovedAt = nil
	permission.UpdatedBy = operatorID
	if err := tx.Table(permissionTable).Save(&permission).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("更新初始服务权限失败")
	}
	return nil
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

	primaryService := enterpriseRoleToService(enterprise.EnterpriseRole)
	db := database.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(enterprise).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("提交企业认证失败")
		}
		if len(req.Documents) > 0 {
			return ensureInitialServiceApplication(tx, enterprise, primaryService, req.Documents, operatorID)
		}
		return nil
	})
	if err != nil {
		if coded, ok := err.(*ecode.Error); ok {
			return nil, coded
		}
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
	if status != entity.EnterpriseStatusApproved && status != entity.EnterpriseStatusRejected {
		return ecode.ErrBadRequest.WithMessage("企业审核状态无效")
	}

	db := database.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status":     status,
			"updated_by": operatorID,
		}
		if status == entity.EnterpriseStatusRejected {
			if enterprise.Remark == nil {
				enterprise.Remark = entity.Remark{}
			}
			enterprise.Remark["audit_reason"] = req.Reason
			updates["remark"] = enterprise.Remark
		}
		if err := tx.Model(&entity.Enterprise{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("审核企业失败")
		}

		userRole := entity.TypeEnterprise
		permissionStatus := entity.ServicePermissionApproved
		applicationStatus := entity.ApplicationApproved
		auditReason := ""
		if status == entity.EnterpriseStatusRejected {
			userRole = entity.TypePersonal
			permissionStatus = entity.ServicePermissionRejected
			applicationStatus = entity.ApplicationRejected
			auditReason = req.Reason
		}
		if err := tx.Model(&entity.User{}).
			Where("enterprise_code = ? AND deleted_at IS NULL", enterprise.EnterpriseCode).
			Updates(map[string]interface{}{"role": userRole, "updated_by": operatorID}).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("更新用户角色失败")
		}

		primaryService := enterpriseRoleToService(enterprise.EnterpriseRole)
		return auditInitialService(tx, enterprise, primaryService, permissionStatus, applicationStatus, auditReason, operatorID)
	})
	if err != nil {
		if coded, ok := err.(*ecode.Error); ok {
			return coded
		}
		return ecode.ErrDatabase.WithMessage("审核企业失败")
	}

	return nil
}

func auditInitialService(tx *gorm.DB, enterprise *entity.Enterprise, service entity.PlatformService, permissionStatus entity.ServicePermissionStatus, applicationStatus entity.ApplicationStatus, reason string, operatorID uint64) error {
	now := time.Now()
	var approvedAt *time.Time
	if permissionStatus == entity.ServicePermissionApproved {
		approvedAt = &now
	}

	permissionTable := entity.CapabilityPermissionTable(service)
	var permission entity.EnterpriseServicePermission
	err := tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		permission = entity.EnterpriseServicePermission{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			CreatedBy:      operatorID,
		}
	} else if err != nil {
		return ecode.ErrDatabase.WithMessage("查询初始服务权限失败")
	}

	permission.Status = permissionStatus
	permission.AuditReason = reason
	permission.ApprovedAt = approvedAt
	permission.UpdatedBy = operatorID
	if permission.ID == 0 {
		if err := tx.Table(permissionTable).Create(&permission).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("保存初始服务权限失败")
		}
	} else if err := tx.Table(permissionTable).Save(&permission).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("保存初始服务权限失败")
	}

	result := tx.Model(&entity.EnterpriseServiceApplication{}).
		Where("enterprise_id = ? AND service = ? AND status = ?", enterprise.ID, service, entity.ApplicationPending).
		Updates(map[string]interface{}{
			"status":       applicationStatus,
			"audit_reason": reason,
			"audited_by":   operatorID,
			"audited_at":   now,
		})
	if result.Error != nil {
		return ecode.ErrDatabase.WithMessage("保存初始服务申请审核状态失败")
	}
	if result.RowsAffected == 0 {
		application := entity.EnterpriseServiceApplication{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			Status:         applicationStatus,
			AuditReason:    reason,
			SubmittedBy:    enterprise.CreatedBy,
			AuditedBy:      operatorID,
			AuditedAt:      &now,
		}
		if err := tx.Create(&application).Error; err != nil {
			return ecode.ErrDatabase.WithMessage("创建初始服务申请审核记录失败")
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
