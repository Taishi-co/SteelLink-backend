package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/sms"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/redis/go-redis/v9"
)

// AuthService 认证服务
type AuthService struct {
	userRepo       repository.UserRepository
	enterpriseRepo repository.EnterpriseRepository
	redis          *redis.Client
	smsProvider    sms.Provider
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:       persistence.NewUserRepository(),
		enterpriseRepo: persistence.NewEnterpriseRepository(),
		redis:          database.GetRedis(),
		smsProvider:    sms.NewProvider(),
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
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), 0, "", "")
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
	enterpriseID, enterpriseCode, enterpriseRole, err := s.resolveEnterpriseInfo(ctx, user)
	if err != nil {
		return nil, err
	}

	serviceCapabilities, err := listApprovedServiceCapabilities(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	// 生成Token
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID, enterpriseCode, enterpriseRole, serviceCapabilities)
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
	if cache.Exists(ctx, refreshTokenBlacklistKey(req.RefreshToken)) {
		return nil, ecode.ErrTokenInvalid
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

	enterpriseID, enterpriseCode, enterpriseRole, err := s.resolveEnterpriseInfo(ctx, user)
	if err != nil {
		return nil, err
	}
	serviceCapabilities, err := listApprovedServiceCapabilities(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	// 生成新Token
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID, enterpriseCode, enterpriseRole, serviceCapabilities)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("生成Token失败")
	}
	if claims.ExpiresAt != nil {
		if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
			cache.SetString(ctx, refreshTokenBlacklistKey(req.RefreshToken), "1", ttl)
		}
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
	code := resolveSmsCode()

	// 存储验证码（5分钟有效）
	codeKey := fmt.Sprintf("sms:code:%s:%s", req.Type, req.Phone)
	if err := s.redis.Set(ctx, codeKey, code, 5*time.Minute).Err(); err != nil {
		return ecode.ErrRedis.WithMessage("存储验证码失败")
	}

	// 设置发送频率限制（1分钟）
	if err := s.redis.Set(ctx, key, "1", time.Minute).Err(); err != nil {
	}

	if err := s.smsProvider.SendCode(req.Phone, code, req.Type); err != nil {
		return ecode.ErrBadRequest.WithMessage("发送验证码失败: " + err.Error())
	}

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

	enterpriseID, enterpriseCode, enterpriseRole, err := s.resolveEnterpriseInfo(ctx, user)
	if err != nil {
		return nil, err
	}
	serviceCapabilities, err := listApprovedServiceCapabilities(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}

	// 生成Token
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	tokenPair, err := auth.GenerateToken(user.ID, user.Phone, email, string(user.Role), enterpriseID, enterpriseCode, enterpriseRole, serviceCapabilities)
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
	cache.Delete(ctx, authSnapshotCacheKey(user.ID))

	return nil
}

func (s *AuthService) resolveEnterpriseID(ctx context.Context, user *entity.User) (uint64, error) {
	id, _, _, err := s.resolveEnterpriseInfo(ctx, user)
	return id, err
}

func (s *AuthService) resolveEnterpriseInfo(ctx context.Context, user *entity.User) (uint64, string, string, error) {
	if user == nil || user.EnterpriseCode == "" {
		return 0, "", "", nil
	}

	enterprise, err := s.enterpriseRepo.FindByEnterpriseCode(ctx, user.EnterpriseCode)
	if err != nil {
		return 0, "", "", ecode.ErrDatabase.WithMessage("查询企业信息失败")
	}
	if enterprise == nil {
		return 0, "", "", ecode.ErrEnterpriseNotFound
	}

	if enterprise.IsApproved() && user.Role == entity.TypePersonal {
		_ = s.userRepo.UpdateRoleByEnterpriseCode(ctx, enterprise.EnterpriseCode, entity.TypeEnterprise, user.ID)
		user.Role = entity.TypeEnterprise
	}

	return enterprise.ID, enterprise.EnterpriseCode, enterprise.EnterpriseRole, nil
}

func refreshTokenBlacklistKey(token string) string {
	return "auth:refresh-token:blacklist:" + cache.HashString(token)
}

func authSnapshotCacheKey(userID uint64) string {
	return "auth:user-snapshot:" + cache.Hash(userID)
}

func listApprovedServiceCapabilities(ctx context.Context, enterpriseID uint64) ([]string, error) {
	if enterpriseID == 0 {
		return nil, nil
	}
	db := database.GetDB()
	if db == nil {
		return nil, ecode.ErrDatabase.WithMessage("数据库未初始化")
	}

	capabilities := make([]string, 0, 5)
	for table := range map[string]struct{}{
		entity.TradeCapabilityPermissionTable: {},
		entity.ProviderServicePermissionTable: {},
	} {
		var rows []struct {
			Service string
		}
		if err := db.WithContext(ctx).
			Table(table).
			Select("service").
			Where("enterprise_id = ? AND status = ? AND deleted_at IS NULL", enterpriseID, entity.ServicePermissionApproved).
			Find(&rows).Error; err != nil {
			return nil, ecode.ErrDatabase.WithMessage("查询企业能力失败")
		}
		for _, row := range rows {
			capabilities = append(capabilities, row.Service)
		}
	}
	return capabilities, nil
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

func resolveSmsCode() string {
	if code := fixedDevSmsCode(); code != "" {
		return code
	}
	return generateSmsCode()
}

func fixedDevSmsCode() string {
	provider := strings.TrimSpace(strings.ToLower(os.Getenv("SMS_PROVIDER")))
	if provider != "" && provider != "dev" {
		return ""
	}

	code := strings.TrimSpace(os.Getenv("SMS_DEV_CODE"))
	if len(code) != 6 {
		return ""
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return code
}

// toUserInfo 转换用户信息
func toUserInfo(user *entity.User) *dto.UserInfo {
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	info := &dto.UserInfo{
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
	if user.EnterpriseCode != "" {
		db := database.GetDB()
		if db != nil {
			var enterprise entity.Enterprise
			if err := db.Where("enterprise_code = ? AND deleted_at IS NULL", user.EnterpriseCode).First(&enterprise).Error; err == nil {
				info.EnterpriseRole = enterprise.EnterpriseRole
				if capabilities, err := listApprovedServiceCapabilities(context.Background(), enterprise.ID); err == nil {
					info.ServiceCapabilities = capabilities
				}
			}
		}
	}
	return info
}
