package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// AuthService 认证服务
type AuthService struct {
	userRepo       repository.UserRepository
	enterpriseRepo repository.EnterpriseRepository
	redis          *redis.Client
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:       persistence.NewUserRepository(),
		enterpriseRepo: persistence.NewEnterpriseRepository(),
		redis:          database.GetRedis(),
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	// 检查手机号是否已存在
	exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("检查手机号失败")
	}
	if exists {
		return nil, ecode.ErrPhoneExists
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, ecode.ErrDatabase.WithMessage("检查邮箱失败")
		}
		if exists {
			return nil, ecode.ErrEmailExists
		}
	}

	// 创建用户
	user := &entity.User{
		Phone:  req.Phone,
		QQ:     req.QQ,
		Role:   entity.TypePersonal, // 默认个人用户
		Status: entity.StatusActive,
	}
	// 设置邮箱（如果有）
	if req.Email != "" {
		user.Email = &req.Email
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		return nil, ecode.ErrInternal.WithMessage("密码加密失败")
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建用户失败")
	}

	// 生成Token
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), 0)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("生成Token失败")
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         toUserInfo(user),
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查找用户（支持手机号或邮箱登录）
	var user *entity.User
	var err error

	// 判断是手机号还是邮箱
	if isPhone(req.Account) {
		user, err = s.userRepo.FindByPhone(ctx, req.Account)
	} else {
		user, err = s.userRepo.FindByEmail(ctx, req.Account)
	}

	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询用户失败")
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, ecode.ErrUserDisabled
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return nil, ecode.ErrPasswordWrong
	}

	// 获取企业ID（如果已加入企业）
	enterpriseID, err := s.resolveEnterpriseID(ctx, user)
	if err != nil {
		return nil, err
	}

	// 生成Token
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("生成Token失败")
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         toUserInfo(user),
	}, nil
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	// 解析刷新Token
	claims, err := auth.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 查询用户
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, ecode.ErrUserDisabled
	}

	// 生成新Token
	enterpriseID, err := s.resolveEnterpriseID(ctx, user)
	if err != nil {
		return nil, err
	}

	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("生成Token失败")
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         toUserInfo(user),
	}, nil
}

// SendSms 发送短信验证码
func (s *AuthService) SendSms(ctx context.Context, req *dto.SendSmsRequest) error {
	// 检查发送频率限制
	key := fmt.Sprintf("sms:limit:%s", req.Phone)
	count := s.redis.Get(ctx, key).Val()
	if count != "" {
		return ecode.ErrBadRequest.WithMessage("发送太频繁，请稍后再试")
	}

	// 生成6位验证码
	code := generateSmsCode()

	// 存储验证码（5分钟有效）
	codeKey := fmt.Sprintf("sms:code:%s:%s", req.Type, req.Phone)
	if err := s.redis.Set(ctx, codeKey, code, 5*time.Minute).Err(); err != nil {
		return ecode.ErrRedis.WithMessage("存储验证码失败")
	}

	// 设置发送频率限制（1分钟）
	if err := s.redis.Set(ctx, key, "1", time.Minute).Err(); err != nil {
		// 忽略错误
	}

	// TODO: 调用短信服务发送验证码
	// 这里需要集成实际的短信服务

	return nil
}

// VerifySms 验证短信验证码
func (s *AuthService) VerifySms(ctx context.Context, phone, smsType, code string) error {
	key := fmt.Sprintf("sms:code:%s:%s", smsType, phone)
	storedCode := s.redis.Get(ctx, key).Val()
	if storedCode == "" {
		return ecode.ErrSmsCodeExpired
	}
	if storedCode != code {
		return ecode.ErrSmsCodeInvalid
	}
	// 验证成功后删除验证码
	s.redis.Del(ctx, key)
	return nil
}

// SmsLogin 短信登录
func (s *AuthService) SmsLogin(ctx context.Context, req *dto.SmsLoginRequest) (*dto.LoginResponse, error) {
	// 验证短信验证码
	if err := s.VerifySms(ctx, req.Phone, "login", req.Code); err != nil {
		return nil, err
	}

	// 查找用户
	user, err := s.userRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, ecode.ErrUserDisabled
	}

	// 生成Token
	enterpriseID, err := s.resolveEnterpriseID(ctx, user)
	if err != nil {
		return nil, err
	}

	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("生成Token失败")
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         toUserInfo(user),
	}, nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	// 验证短信验证码
	if err := s.VerifySms(ctx, req.Phone, "reset", req.Code); err != nil {
		return err
	}

	// 查找用户
	user, err := s.userRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return ecode.ErrInternal.WithMessage("密码加密失败")
	}

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("更新密码失败")
	}

	return nil
}

func (s *AuthService) resolveEnterpriseID(ctx context.Context, user *entity.User) (uint64, error) {
	if user == nil || user.EnterpriseCode == "" {
		return 0, nil
	}

	enterprise, err := s.enterpriseRepo.FindByEnterpriseCode(ctx, user.EnterpriseCode)
	if err != nil {
		return 0, ecode.ErrDatabase.WithMessage("查询企业信息失败")
	}
	if enterprise == nil {
		return 0, ecode.ErrEnterpriseNotFound
	}

	return enterprise.ID, nil
}

// 辅助函数

// isPhone 判断是否为手机号
func isPhone(s string) bool {
	if len(s) != 11 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// generateSmsCode 生成6位验证码
func generateSmsCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}

// toUserInfo 转换用户信息
func toUserInfo(user *entity.User) *dto.UserInfo {
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	return &dto.UserInfo{
		ID:             user.ID,
		UID:            user.UID,
		UserNo:         user.UserNo,
		PlatformID:     user.PlatformID,
		PersonalCode:   user.PersonalCode,
		EnterpriseCode: user.EnterpriseCode,
		Role:           string(user.Role),
		Phone:          user.Phone,
		QQ:             user.QQ,
		Email:          email,
		RealName:       user.RealName,
		IDCard:         user.IDCard,
		IdentityStatus: int(user.IdentityStatus),
		IdentityReject: user.IdentityReject,
		Status:         int(user.Status),
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}
