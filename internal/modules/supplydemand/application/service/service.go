package service

import (
	"context"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/ui/dto"
	userEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

type SupplyDemandService struct {
	postRepo       repository.PostRepository
	userRepo       userRepository.UserRepository
	enterpriseRepo userRepository.EnterpriseRepository
}

func NewSupplyDemandService() *SupplyDemandService {
	return &SupplyDemandService{
		postRepo:       persistence.NewPostRepository(),
		userRepo:       userPersistence.NewUserRepository(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
	}
}

func (s *SupplyDemandService) Create(ctx context.Context, req *dto.CreatePostRequest, userID, enterpriseID uint64) (*dto.PostInfo, error) {
	postType := entity.PostType(req.Type)
	if !entity.IsValidPostType(postType) {
		return nil, ecode.ErrBadRequest.WithMessage("供求信息类型不支持")
	}
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("获取用户信息失败")
	}
	if user == nil {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if !user.IsIdentityVerified() {
		return nil, ecode.ErrForbidden.WithMessage("完成个人实名认证后才能发布供求信息")
	}

	publisherEnterpriseID, err := s.resolveApprovedEnterpriseID(ctx, user, enterpriseID)
	if err != nil {
		return nil, err
	}

	post := &entity.SupplyDemandPost{
		Type:              postType,
		Title:             req.Title,
		Category:          req.Category,
		Material:          req.Material,
		Spec:              req.Spec,
		Quantity:          req.Quantity,
		Unit:              req.Unit,
		TargetPrice:       req.TargetPrice,
		Budget:            req.Budget,
		Province:          req.Province,
		City:              req.City,
		District:          req.District,
		DeliveryAddress:   req.DeliveryAddress,
		Description:       req.Description,
		Tags:              entity.StringList(req.Tags),
		ContactName:       req.ContactName,
		ContactPhone:      req.ContactPhone,
		FinancingRequired: req.FinancingRequired,
		Status:            entity.PostStatusPendingReview,
		EnterpriseID:      publisherEnterpriseID,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建供求信息失败")
	}
	return s.toPostInfo(ctx, post)
}

func (s *SupplyDemandService) List(ctx context.Context, req *dto.PostListRequest, publicOnly bool) (*dto.PostListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var status *entity.PostStatus
	if req.Status != nil {
		value := entity.PostStatus(*req.Status)
		if !entity.IsValidPostStatus(value) {
			return nil, ecode.ErrBadRequest.WithMessage("供求信息状态不支持")
		}
		status = &value
	}

	postType := entity.PostType(req.Type)
	if postType != "" && !entity.IsValidPostType(postType) {
		return nil, ecode.ErrBadRequest.WithMessage("供求信息类型不支持")
	}

	posts, total, err := s.postRepo.List(ctx, &repository.PostQuery{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Type:         postType,
		Status:       status,
		Keyword:      req.Keyword,
		Category:     req.Category,
		Province:     req.Province,
		City:         req.City,
		EnterpriseID: req.EnterpriseID,
		PublicOnly:   publicOnly,
		SortBy:       req.SortBy,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}

	return &dto.PostListResponse{
		List:     s.toPostInfoList(ctx, posts),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *SupplyDemandService) ListMine(ctx context.Context, req *dto.PostListRequest, userID uint64) (*dto.PostListResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	postType := entity.PostType(req.Type)
	if postType != "" && !entity.IsValidPostType(postType) {
		return nil, ecode.ErrBadRequest.WithMessage("供求信息类型不支持")
	}

	var status *entity.PostStatus
	if req.Status != nil {
		value := entity.PostStatus(*req.Status)
		if !entity.IsValidPostStatus(value) {
			return nil, ecode.ErrBadRequest.WithMessage("供求信息状态不支持")
		}
		status = &value
	}

	posts, total, err := s.postRepo.List(ctx, &repository.PostQuery{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Type:       postType,
		Status:     status,
		Keyword:    req.Keyword,
		Category:   req.Category,
		Province:   req.Province,
		City:       req.City,
		CreatedBy:  userID,
		PublicOnly: false,
		SortBy:     req.SortBy,
	})
	if err != nil {
		return nil, ecode.ErrDatabase
	}

	return &dto.PostListResponse{
		List:     s.toPostInfoList(ctx, posts),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *SupplyDemandService) GetByID(ctx context.Context, id uint64, publicOnly bool) (*dto.PostInfo, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if post == nil || (publicOnly && !post.IsVisible()) {
		return nil, ecode.ErrNotFound.WithMessage("供求信息不存在")
	}
	return s.toPostInfo(ctx, post)
}

func (s *SupplyDemandService) Stats(ctx context.Context) (*dto.PostStatsResponse, error) {
	stats, err := s.postRepo.Stats(ctx)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	return &dto.PostStatsResponse{
		PublishedCount: stats.PublishedCount,
		MatchedCount:   stats.MatchedCount,
		MatchedAmount:  stats.MatchedAmount,
		TodayNewCount:  stats.TodayNewCount,
	}, nil
}

func (s *SupplyDemandService) UpdateStatus(ctx context.Context, id uint64, status int, userID uint64) error {
	if userID == 0 {
		return ecode.ErrUnauthorized.WithMessage("未登录")
	}
	value := entity.PostStatus(status)
	if !entity.IsValidPostStatus(value) {
		return ecode.ErrBadRequest.WithMessage("供求信息状态不支持")
	}
	if value != entity.PostStatusClosed {
		return ecode.ErrBadRequest.WithMessage("用户仅可关闭自己的供求信息，发布可见状态需平台审核")
	}

	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if post == nil {
		return ecode.ErrNotFound.WithMessage("供求信息不存在")
	}
	if post.CreatedBy != userID {
		return ecode.ErrForbidden.WithMessage("只能更新自己的供求信息")
	}
	if err := s.postRepo.UpdateStatus(ctx, id, value, userID); err != nil {
		return ecode.ErrDatabase.WithMessage("更新供求信息状态失败")
	}
	return nil
}

func (s *SupplyDemandService) toPostInfoList(ctx context.Context, posts []*entity.SupplyDemandPost) []*dto.PostInfo {
	list := make([]*dto.PostInfo, 0, len(posts))
	enterpriseNameCache := make(map[uint64]string)
	for _, post := range posts {
		list = append(list, s.toPostInfoWithEnterpriseCache(ctx, post, enterpriseNameCache))
	}
	return list
}

func (s *SupplyDemandService) toPostInfo(ctx context.Context, post *entity.SupplyDemandPost) (*dto.PostInfo, error) {
	if post == nil {
		return nil, nil
	}
	return s.toPostInfoWithEnterpriseCache(ctx, post, make(map[uint64]string)), nil
}

func (s *SupplyDemandService) toPostInfoWithEnterpriseCache(ctx context.Context, post *entity.SupplyDemandPost, enterpriseNameCache map[uint64]string) *dto.PostInfo {
	publisherType := "personal"
	publisherName := "实名个人"

	if post.EnterpriseID > 0 {
		if _, ok := enterpriseNameCache[post.EnterpriseID]; !ok {
			enterpriseNameCache[post.EnterpriseID] = ""
			enterprise, err := s.enterpriseRepo.FindByID(ctx, post.EnterpriseID)
			if err == nil && enterprise != nil && enterprise.IsApproved() {
				enterpriseNameCache[post.EnterpriseID] = maskEnterpriseName(enterprise.Name)
			}
		}
		if enterpriseNameCache[post.EnterpriseID] != "" {
			publisherType = "enterprise"
			publisherName = enterpriseNameCache[post.EnterpriseID]
		}
	}

	return &dto.PostInfo{
		ID:                post.ID,
		PostNo:            post.PostNo,
		Type:              string(post.Type),
		Title:             post.Title,
		Category:          post.Category,
		Material:          post.Material,
		Spec:              post.Spec,
		Quantity:          post.Quantity,
		Unit:              post.Unit,
		TargetPrice:       post.TargetPrice,
		Budget:            post.Budget,
		Currency:          post.Currency,
		Province:          post.Province,
		City:              post.City,
		District:          post.District,
		DeliveryAddress:   post.DeliveryAddress,
		Description:       post.Description,
		Tags:              []string(post.Tags),
		ContactName:       post.ContactName,
		ContactPhone:      post.ContactPhone,
		FinancingRequired: post.FinancingRequired,
		MatchCount:        post.MatchCount,
		MatchAmount:       post.MatchAmount,
		Status:            int(post.Status),
		AuditRemark:       post.AuditRemark,
		AuditorID:         post.AuditorID,
		AuditedAt:         post.AuditedAt,
		EnterpriseID:      post.EnterpriseID,
		PublisherType:     publisherType,
		PublisherName:     publisherName,
		CreatedBy:         post.CreatedBy,
		CreatedAt:         post.CreatedAt,
		UpdatedAt:         post.UpdatedAt,
	}
}

func (s *SupplyDemandService) resolveApprovedEnterpriseID(ctx context.Context, user *userEntity.User, claimsEnterpriseID uint64) (uint64, error) {
	if claimsEnterpriseID > 0 {
		enterprise, err := s.enterpriseRepo.FindByID(ctx, claimsEnterpriseID)
		if err != nil {
			return 0, ecode.ErrDatabase.WithMessage("获取企业信息失败")
		}
		if enterprise != nil && enterprise.IsApproved() && enterprise.EnterpriseCode == user.EnterpriseCode {
			return enterprise.ID, nil
		}
	}

	if user.EnterpriseCode == "" {
		return 0, nil
	}
	enterprise, err := s.enterpriseRepo.FindByEnterpriseCode(ctx, user.EnterpriseCode)
	if err != nil {
		return 0, ecode.ErrDatabase.WithMessage("获取企业信息失败")
	}
	if enterprise != nil && enterprise.IsApproved() {
		return enterprise.ID, nil
	}
	return 0, nil
}

func maskEnterpriseName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "认证企业"
	}
	if len(runes) <= 2 {
		return string(runes[:1]) + "*"
	}
	if len(runes) <= 4 {
		return string(runes[:1]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1:])
	}
	return string(runes[:2]) + strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-2:])
}
