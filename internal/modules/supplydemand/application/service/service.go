package service

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

type SupplyDemandService struct {
	postRepo repository.PostRepository
}

func NewSupplyDemandService() *SupplyDemandService {
	return &SupplyDemandService{postRepo: persistence.NewPostRepository()}
}

func (s *SupplyDemandService) Create(ctx context.Context, req *dto.CreatePostRequest, userID, enterpriseID uint64) (*dto.PostInfo, error) {
	postType := entity.PostType(req.Type)
	if !entity.IsValidPostType(postType) {
		return nil, ecode.ErrBadRequest.WithMessage("供求信息类型不支持")
	}
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
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
		Status:            entity.PostStatusPublished,
		EnterpriseID:      enterpriseID,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建供求信息失败")
	}
	return toPostInfo(post), nil
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
		List:     toPostInfoList(posts),
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
		List:     toPostInfoList(posts),
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
	return toPostInfo(post), nil
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

func toPostInfoList(posts []*entity.SupplyDemandPost) []*dto.PostInfo {
	list := make([]*dto.PostInfo, 0, len(posts))
	for _, post := range posts {
		list = append(list, toPostInfo(post))
	}
	return list
}

func toPostInfo(post *entity.SupplyDemandPost) *dto.PostInfo {
	if post == nil {
		return nil
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
		EnterpriseID:      post.EnterpriseID,
		CreatedBy:         post.CreatedBy,
		CreatedAt:         post.CreatedAt,
		UpdatedAt:         post.UpdatedAt,
	}
}
