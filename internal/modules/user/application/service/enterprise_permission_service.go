package service

import (
	"context"
	"errors"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

type EnterprisePermissionService struct {
	db *gorm.DB
}

func NewEnterprisePermissionService() *EnterprisePermissionService {
	return &EnterprisePermissionService{db: database.GetDB()}
}

func RequiredServiceDocuments() map[string][]string {
	return map[string][]string{
		string(entity.ServiceSales):       {"business_license", "legal_person_id", "bank_account", "authorization_letter"},
		string(entity.ServiceProcurement): {"business_license", "invoice_info", "authorization_letter"},
		string(entity.ServiceLogistics):   {"business_license", "transport_license", "carrier_certificate"},
		string(entity.ServiceWarehouse):   {"business_license", "warehouse_site_certificate", "safety_certificate"},
		string(entity.ServiceFinance):     {"finance_license_or_agreement", "risk_contact_info", "credit_documents"},
	}
}

func (s *EnterprisePermissionService) GetOverview(ctx context.Context, userID uint64) (*dto.EnterpriseServiceOverview, error) {
	user, enterprise, err := s.currentEnterprise(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || enterprise == nil {
		return nil, ecode.ErrEnterpriseNotFound
	}

	permissions, err := s.listPermissions(ctx, enterprise.ID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询企业服务权限失败")
	}

	return &dto.EnterpriseServiceOverview{
		Permissions:       permissions,
		RequiredDocuments: RequiredServiceDocuments(),
	}, nil
}

func (s *EnterprisePermissionService) Apply(ctx context.Context, userID uint64, req *dto.ApplyEnterpriseServiceRequest) (*dto.EnterpriseServiceApplicationInfo, error) {
	if !entity.ValidPlatformService(req.Service) {
		return nil, ecode.ErrBadRequest.WithMessage("平台服务无效")
	}
	if len(req.Documents) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("请上传审核材料")
	}

	_, enterprise, err := s.currentEnterprise(ctx, userID)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
		return nil, ecode.ErrEnterpriseNotFound
	}
	if !enterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("企业未通过认证，暂不可申请平台服务")
	}

	service := entity.PlatformService(req.Service)
	if !entity.ValidPlatformService(req.Service) {
		return nil, ecode.ErrBadRequest.WithMessage("平台服务无效")
	}

	var application entity.EnterpriseServiceApplication
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("enterprise_id = ? AND service = ? AND status = ?", enterprise.ID, service, entity.ApplicationPending).
			First(&application).Error; err == nil {
			return ecode.ErrBadRequest.WithMessage("该服务已有审核中的申请")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		permissionTable := entity.CapabilityPermissionTable(service)
		var existing entity.EnterpriseServicePermission
		if err := tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&existing).Error; err == nil {
			if existing.Status == entity.ServicePermissionApproved {
				return ecode.ErrBadRequest.WithMessage("该服务已开通，无需重复申请")
			}
			if existing.Status == entity.ServicePermissionPending {
				return ecode.ErrBadRequest.WithMessage("该服务已有审核中的申请")
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		application = entity.EnterpriseServiceApplication{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			Status:         entity.ApplicationPending,
			SubmittedBy:    userID,
		}
		if err := tx.Create(&application).Error; err != nil {
			return err
		}

		documents := make([]entity.EnterpriseApplicationDocument, 0, len(req.Documents))
		for _, doc := range req.Documents {
			documents = append(documents, entity.EnterpriseApplicationDocument{
				ApplicationID: application.ID,
				DocumentType:  doc.DocumentType,
				FileURL:       doc.FileURL,
			})
		}
		if err := tx.Create(&documents).Error; err != nil {
			return err
		}

		nowPermission := entity.EnterpriseServicePermission{
			EnterpriseID:   enterprise.ID,
			EnterpriseCode: enterprise.EnterpriseCode,
			Service:        service,
			Status:         entity.ServicePermissionPending,
			AuditReason:    "",
			CreatedBy:      userID,
			UpdatedBy:      userID,
		}
		if err := tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&existing).Error; err == nil {
			existing.Status = entity.ServicePermissionPending
			existing.AuditReason = ""
			existing.UpdatedBy = userID
			return tx.Table(permissionTable).Save(&existing).Error
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Table(permissionTable).Create(&nowPermission).Error
	})
	if err != nil {
		if coded, ok := err.(*ecode.Error); ok {
			return nil, coded
		}
		return nil, ecode.ErrDatabase.WithMessage("提交服务申请失败")
	}
	invalidateEnterprisePermissionCache(ctx, enterprise.ID)

	return s.GetApplicationByID(ctx, application.ID)
}

func (s *EnterprisePermissionService) ListApplications(ctx context.Context, req *dto.EnterpriseServiceApplicationListRequest) (*dto.EnterpriseServiceApplicationListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := s.db.WithContext(ctx).Model(&entity.EnterpriseServiceApplication{})
	if req.EnterpriseCode != "" {
		query = query.Where("enterprise_code = ?", req.EnterpriseCode)
	}
	if req.Service != "" {
		query = query.Where("service = ?", req.Service)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, ecode.ErrDatabase
	}

	var applications []entity.EnterpriseServiceApplication
	if err := query.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&applications).Error; err != nil {
		return nil, ecode.ErrDatabase
	}

	list := make([]*dto.EnterpriseServiceApplicationInfo, 0, len(applications))
	for _, application := range applications {
		info, err := s.applicationToInfo(ctx, &application)
		if err != nil {
			return nil, ecode.ErrDatabase
		}
		list = append(list, info)
	}
	return &dto.EnterpriseServiceApplicationListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *EnterprisePermissionService) GetApplicationByID(ctx context.Context, id uint64) (*dto.EnterpriseServiceApplicationInfo, error) {
	var application entity.EnterpriseServiceApplication
	if err := s.db.WithContext(ctx).First(&application, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrNotFound.WithMessage("服务申请不存在")
		}
		return nil, ecode.ErrDatabase
	}
	info, err := s.applicationToInfo(ctx, &application)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	return info, nil
}

func (s *EnterprisePermissionService) Audit(ctx context.Context, id uint64, req *dto.AuditEnterpriseServiceRequest, operatorID uint64) error {
	status := entity.ApplicationStatus(req.Status)
	if status != entity.ApplicationApproved && status != entity.ApplicationRejected {
		return ecode.ErrBadRequest.WithMessage("审核状态无效")
	}

	var enterpriseID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var application entity.EnterpriseServiceApplication
		if err := tx.First(&application, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrNotFound.WithMessage("服务申请不存在")
			}
			return ecode.ErrDatabase
		}
		enterpriseID = application.EnterpriseID

		now := time.Now()
		permissionStatus := entity.ServicePermissionRejected
		if req.Status == int(entity.ApplicationApproved) {
			permissionStatus = entity.ServicePermissionApproved
			application.Status = entity.ApplicationApproved
		} else {
			application.Status = entity.ApplicationRejected
		}
		application.AuditReason = req.Reason
		application.AuditedBy = operatorID
		application.AuditedAt = &now
		if err := tx.Save(&application).Error; err != nil {
			return ecode.ErrDatabase
		}

		permissionTable := entity.CapabilityPermissionTable(application.Service)
		var permission entity.EnterpriseServicePermission
		if err := tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", application.EnterpriseID, application.Service).First(&permission).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrDatabase
			}
			permission = entity.EnterpriseServicePermission{
				EnterpriseID:   application.EnterpriseID,
				EnterpriseCode: application.EnterpriseCode,
				Service:        application.Service,
				CreatedBy:      operatorID,
			}
		}
		permission.Status = permissionStatus
		permission.AuditReason = req.Reason
		permission.UpdatedBy = operatorID
		if permissionStatus == entity.ServicePermissionApproved {
			permission.ApprovedAt = &now
		}
		if permission.ID == 0 {
			if err := tx.Table(permissionTable).Create(&permission).Error; err != nil {
				return ecode.ErrDatabase
			}
		} else if err := tx.Table(permissionTable).Save(&permission).Error; err != nil {
			return ecode.ErrDatabase
		}
		return nil
	})
	if err != nil {
		return err
	}
	invalidateEnterprisePermissionCache(ctx, enterpriseID)
	return nil
}

func (s *EnterprisePermissionService) UpdatePermissionStatus(ctx context.Context, enterpriseID uint64, serviceName string, req *dto.UpdateEnterpriseServicePermissionRequest, operatorID uint64) (*dto.EnterpriseServicePermissionInfo, error) {
	if !entity.ValidPlatformService(serviceName) {
		return nil, ecode.ErrBadRequest.WithMessage("平台服务无效")
	}
	status := entity.ServicePermissionStatus(req.Status)
	if status != entity.ServicePermissionApproved && status != entity.ServicePermissionRejected && status != entity.ServicePermissionSuspended {
		return nil, ecode.ErrBadRequest.WithMessage("服务权限状态无效")
	}

	var enterprise entity.Enterprise
	if err := s.db.WithContext(ctx).First(&enterprise, enterpriseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrEnterpriseNotFound
		}
		return nil, ecode.ErrDatabase
	}

	service := entity.PlatformService(serviceName)
	permissionTable := entity.CapabilityPermissionTable(service)
	now := time.Now()
	var permission entity.EnterpriseServicePermission
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(permissionTable).Where("enterprise_id = ? AND service = ?", enterprise.ID, service).First(&permission).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrDatabase
			}
			permission = entity.EnterpriseServicePermission{
				EnterpriseID:   enterprise.ID,
				EnterpriseCode: enterprise.EnterpriseCode,
				Service:        service,
				CreatedBy:      operatorID,
			}
		}
		permission.Status = status
		permission.AuditReason = req.Reason
		permission.UpdatedBy = operatorID
		if status == entity.ServicePermissionApproved {
			permission.ApprovedAt = &now
		}
		if permission.ID == 0 {
			return tx.Table(permissionTable).Create(&permission).Error
		}
		return tx.Table(permissionTable).Save(&permission).Error
	})
	if err != nil {
		if coded, ok := err.(*ecode.Error); ok {
			return nil, coded
		}
		return nil, ecode.ErrDatabase.WithMessage("更新企业服务权限失败")
	}
	invalidateEnterprisePermissionCache(ctx, enterprise.ID)
	return permissionToInfo(&permission), nil
}

func (s *EnterprisePermissionService) currentEnterprise(ctx context.Context, userID uint64) (*entity.User, *entity.Enterprise, error) {
	var user entity.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ecode.ErrUserNotFound
		}
		return nil, nil, ecode.ErrDatabase
	}
	if user.EnterpriseCode == "" {
		return &user, nil, nil
	}

	var enterprise entity.Enterprise
	if err := s.db.WithContext(ctx).Where("enterprise_code = ?", user.EnterpriseCode).First(&enterprise).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &user, nil, nil
		}
		return nil, nil, ecode.ErrDatabase
	}
	return &user, &enterprise, nil
}

func (s *EnterprisePermissionService) listPermissions(ctx context.Context, enterpriseID uint64) ([]*dto.EnterpriseServicePermissionInfo, error) {
	var permissions []entity.EnterpriseServicePermission
	for _, table := range []string{entity.TradeCapabilityPermissionTable, entity.ProviderServicePermissionTable} {
		var rows []entity.EnterpriseServicePermission
		if err := s.db.WithContext(ctx).Table(table).Where("enterprise_id = ?", enterpriseID).Order("service ASC").Find(&rows).Error; err != nil {
			return nil, err
		}
		permissions = append(permissions, rows...)
	}
	result := make([]*dto.EnterpriseServicePermissionInfo, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, permissionToInfo(&permission))
	}
	return result, nil
}

func (s *EnterprisePermissionService) applicationToInfo(ctx context.Context, application *entity.EnterpriseServiceApplication) (*dto.EnterpriseServiceApplicationInfo, error) {
	var docs []entity.EnterpriseApplicationDocument
	if err := s.db.WithContext(ctx).Where("application_id = ?", application.ID).Order("id ASC").Find(&docs).Error; err != nil {
		return nil, err
	}
	docInfos := make([]*dto.EnterpriseApplicationDocumentInfo, 0, len(docs))
	for _, doc := range docs {
		docInfos = append(docInfos, &dto.EnterpriseApplicationDocumentInfo{
			ID:            doc.ID,
			ApplicationID: doc.ApplicationID,
			DocumentType:  doc.DocumentType,
			FileURL:       doc.FileURL,
			CreatedAt:     doc.CreatedAt,
		})
	}
	return &dto.EnterpriseServiceApplicationInfo{
		ID:             application.ID,
		EnterpriseID:   application.EnterpriseID,
		EnterpriseCode: application.EnterpriseCode,
		Service:        string(application.Service),
		CapabilityKind: entity.CapabilityKind(application.Service),
		EnterpriseRole: entity.ServiceToEnterpriseRole(application.Service),
		Status:         int(application.Status),
		AuditReason:    application.AuditReason,
		SubmittedBy:    application.SubmittedBy,
		AuditedBy:      application.AuditedBy,
		AuditedAt:      application.AuditedAt,
		Documents:      docInfos,
		CreatedAt:      application.CreatedAt,
		UpdatedAt:      application.UpdatedAt,
	}, nil
}

func permissionToInfo(permission *entity.EnterpriseServicePermission) *dto.EnterpriseServicePermissionInfo {
	return &dto.EnterpriseServicePermissionInfo{
		ID:             permission.ID,
		EnterpriseID:   permission.EnterpriseID,
		EnterpriseCode: permission.EnterpriseCode,
		Service:        string(permission.Service),
		CapabilityKind: entity.CapabilityKind(permission.Service),
		EnterpriseRole: entity.ServiceToEnterpriseRole(permission.Service),
		Status:         int(permission.Status),
		AuditReason:    permission.AuditReason,
		ApprovedAt:     permission.ApprovedAt,
		UpdatedAt:      permission.UpdatedAt,
	}
}

func invalidateEnterprisePermissionCache(ctx context.Context, enterpriseID uint64) {
	cache.DeleteByPattern(ctx, "enterprise:services:"+cache.Hash(enterpriseID)+":*")
}
