package service

import (
	"context"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// UserService 用户服务
type UserService struct {
	userRepo       repository.UserRepository
	enterpriseRepo repository.EnterpriseRepository
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		userRepo:       persistence.NewUserRepository(),
		enterpriseRepo: persistence.NewEnterpriseRepository(),
	}
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(ctx context.Context, id uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	return toUserInfo(user), nil
}

// GetByPhone 根据手机号获取用户
func (s *UserService) GetByPhone(ctx context.Context, phone string) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	return toUserInfo(user), nil
}

// Update 更新用户信息
func (s *UserService) Update(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	// 检查邮箱是否已被使用
	var currentEmail string
	if user.Email != nil {
		currentEmail = *user.Email
	}
	if req.Email != "" && req.Email != currentEmail {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return ecode.ErrDatabase
		}
		if exists {
			return ecode.ErrEmailExists
		}
		user.Email = &req.Email
	}

	// 更新QQ
	if req.QQ != "" {
		user.QQ = req.QQ
	}

	user.UpdatedBy = operatorID

	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("更新用户失败")
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, id uint64, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return ecode.ErrPasswordWrong
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return ecode.ErrInternal.WithMessage("密码加密失败")
	}

	user.UpdatedBy = id

	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("更新密码失败")
	}

	return nil
}

// JoinEnterprise 加入企业
func (s *UserService) JoinEnterprise(ctx context.Context, userID uint64, req *dto.JoinEnterpriseRequest) error {
	// 检查企业是否存在
	enterprise, err := s.enterpriseRepo.FindByEnterpriseCode(ctx, req.EnterpriseCode)
	if err != nil {
		return ecode.ErrDatabase
	}
	if enterprise == nil {
		return ecode.ErrEnterpriseNotFound
	}

	// 检查企业是否已认证
	if !enterprise.IsApproved() {
		return ecode.ErrBadRequest.WithMessage("企业未通过认证")
	}

	// 获取用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	// 更新企业代码
	user.EnterpriseCode = req.EnterpriseCode
	user.Role = entity.TypeEnterprise // 更新为企业用户
	user.UpdatedBy = userID

	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("加入企业失败")
	}

	return nil
}

// LeaveEnterprise 离开企业
func (s *UserService) LeaveEnterprise(ctx context.Context, userID uint64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	// 清空企业代码
	user.EnterpriseCode = ""
	user.Role = entity.TypePersonal // 恢复为个人用户
	user.UpdatedBy = userID

	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("离开企业失败")
	}

	return nil
}

// List 获取用户列表
func (s *UserService) List(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error) {
	// 设置默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := &repository.UserQuery{
		Page:           req.Page,
		PageSize:       req.PageSize,
		Phone:          req.Phone,
		Email:          req.Email,
		Role:           req.Role,
		Status:         req.Status,
		Keyword:        req.Keyword,
	}

	users, total, err := s.userRepo.List(ctx, query)
	if err != nil {
		return nil, ecode.ErrDatabase
	}

	list := make([]*dto.UserInfo, 0, len(users))
	for _, user := range users {
		list = append(list, toUserInfo(user))
	}

	return &dto.UserListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateStatus 更新用户状态（管理员操作）
func (s *UserService) UpdateStatus(ctx context.Context, id uint64, status int, operatorID uint64) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	if err := s.userRepo.UpdateStatus(ctx, id, entity.UserStatus(status), operatorID); err != nil {
		return ecode.ErrDatabase.WithMessage("更新状态失败")
	}

	return nil
}
