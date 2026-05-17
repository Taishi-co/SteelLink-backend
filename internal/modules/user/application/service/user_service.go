package service

import (
	"context"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"golang.org/x/crypto/bcrypt"
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

	// 检查实名认证
	if !user.IsIdentityVerified() {
		return ecode.ErrBadRequest.WithMessage("请先完成实名认证后再加入企业")
	}

	// 检查1人1企限制
	if user.EnterpriseCode != "" {
		return ecode.ErrBadRequest.WithMessage("您已绑定企业，不可重复加入。1个实名个人仅可绑定1家企业")
	}

	// 更新企业代码
	user.EnterpriseCode = req.EnterpriseCode
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
		Page:     req.Page,
		PageSize: req.PageSize,
		Phone:    req.Phone,
		Email:    req.Email,
		Role:     req.Role,
		Status:   req.Status,
		Keyword:  req.Keyword,
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

func (s *UserService) ListPersonalInquiryDrafts(ctx context.Context, userID uint64) (*dto.PersonalInquiryDraftListResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	if !user.IsIdentityVerified() {
		return nil, ecode.ErrBadRequest.WithMessage("请先完成实名认证后再使用询价草稿")
	}

	var drafts []entity.PersonalInquiryDraft
	if err := database.GetDB().WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "draft").
		Order("updated_at DESC, id DESC").
		Find(&drafts).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询询价草稿失败")
	}

	list := make([]dto.PersonalInquiryDraftInfo, 0, len(drafts))
	for _, draft := range drafts {
		list = append(list, personalInquiryDraftToInfo(&draft))
	}
	return &dto.PersonalInquiryDraftListResponse{List: list}, nil
}

func (s *UserService) CreatePersonalInquiryDraft(ctx context.Context, userID uint64, req *dto.UpsertPersonalInquiryDraftRequest) (*dto.PersonalInquiryDraftInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	if !user.IsIdentityVerified() {
		return nil, ecode.ErrBadRequest.WithMessage("请先完成实名认证后再保存询价草稿")
	}

	draft := entity.PersonalInquiryDraft{
		UserID:          userID,
		Title:           req.Title,
		ProductName:     req.ProductName,
		Category:        req.Category,
		Material:        req.Material,
		Spec:            req.Spec,
		Quantity:        req.Quantity,
		Unit:            req.Unit,
		TargetPrice:     req.TargetPrice,
		DeliveryAddress: req.DeliveryAddress,
		Requirement:     req.Requirement,
		Status:          "draft",
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}
	if err := database.GetDB().WithContext(ctx).Create(&draft).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("保存询价草稿失败")
	}
	info := personalInquiryDraftToInfo(&draft)
	return &info, nil
}

func (s *UserService) UpdatePersonalInquiryDraft(ctx context.Context, userID, draftID uint64, req *dto.UpsertPersonalInquiryDraftRequest) (*dto.PersonalInquiryDraftInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	if !user.IsIdentityVerified() {
		return nil, ecode.ErrBadRequest.WithMessage("请先完成实名认证后再更新询价草稿")
	}

	var draft entity.PersonalInquiryDraft
	if err := database.GetDB().WithContext(ctx).Where("id = ? AND user_id = ? AND status = ?", draftID, userID, "draft").First(&draft).Error; err != nil {
		return nil, ecode.ErrNotFound.WithMessage("询价草稿不存在")
	}
	draft.Title = req.Title
	draft.ProductName = req.ProductName
	draft.Category = req.Category
	draft.Material = req.Material
	draft.Spec = req.Spec
	draft.Quantity = req.Quantity
	draft.Unit = req.Unit
	draft.TargetPrice = req.TargetPrice
	draft.DeliveryAddress = req.DeliveryAddress
	draft.Requirement = req.Requirement
	draft.UpdatedBy = userID
	if err := database.GetDB().WithContext(ctx).Save(&draft).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新询价草稿失败")
	}
	info := personalInquiryDraftToInfo(&draft)
	return &info, nil
}

func (s *UserService) DeletePersonalInquiryDraft(ctx context.Context, userID, draftID uint64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}
	if !user.IsIdentityVerified() {
		return ecode.ErrBadRequest.WithMessage("请先完成实名认证后再删除询价草稿")
	}
	result := database.GetDB().WithContext(ctx).
		Model(&entity.PersonalInquiryDraft{}).
		Where("id = ? AND user_id = ? AND status = ?", draftID, userID, "draft").
		Updates(map[string]interface{}{"status": "archived", "updated_by": userID})
	if result.Error != nil {
		return ecode.ErrDatabase.WithMessage("删除询价草稿失败")
	}
	if result.RowsAffected == 0 {
		return ecode.ErrNotFound.WithMessage("询价草稿不存在")
	}
	return nil
}

func personalInquiryDraftToInfo(draft *entity.PersonalInquiryDraft) dto.PersonalInquiryDraftInfo {
	return dto.PersonalInquiryDraftInfo{
		ID:              draft.ID,
		DraftNo:         draft.DraftNo,
		UserID:          draft.UserID,
		Title:           draft.Title,
		ProductName:     draft.ProductName,
		Category:        draft.Category,
		Material:        draft.Material,
		Spec:            draft.Spec,
		Quantity:        draft.Quantity,
		Unit:            draft.Unit,
		TargetPrice:     draft.TargetPrice,
		DeliveryAddress: draft.DeliveryAddress,
		Requirement:     draft.Requirement,
		Status:          draft.Status,
		CreatedAt:       draft.CreatedAt,
		UpdatedAt:       draft.UpdatedAt,
	}
}

func (s *UserService) SubmitIdentityVerification(ctx context.Context, userID uint64, req *dto.SubmitIdentityRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ecode.ErrDatabase
	}
	if user == nil {
		return ecode.ErrUserNotFound
	}
	if !user.CanSubmitIdentity() {
		return ecode.ErrBadRequest.WithMessage("当前实名状态不可提交，请等待审核完成")
	}
	user.RealName = req.RealName
	user.IDCard = req.IDCard
	user.IDCardFrontURL = req.IDCardFrontURL
	user.IDCardBackURL = req.IDCardBackURL
	user.IdentityStatus = entity.IdentityPending
	user.IdentityReject = ""
	user.UpdatedBy = userID
	if err := s.userRepo.Update(ctx, user); err != nil {
		return ecode.ErrDatabase.WithMessage("提交实名认证失败")
	}
	return nil
}

func (s *UserService) GetIdentityStatus(ctx context.Context, userID uint64) (*dto.IdentityStatusResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if user == nil {
		return nil, ecode.ErrUserNotFound
	}
	return &dto.IdentityStatusResponse{
		IdentityStatus: int(user.IdentityStatus),
		RealName:       user.RealName,
		IDCard:         user.IDCard,
		IdentityReject: user.IdentityReject,
	}, nil
}

func (s *UserService) CreateSubAccount(ctx context.Context, parentID uint64, req *dto.CreateSubAccountRequest) (*dto.SubAccountInfo, error) {
	parent, err := s.userRepo.FindByID(ctx, parentID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if parent == nil {
		return nil, ecode.ErrUserNotFound
	}
	if parent.IsSubAccount() {
		return nil, ecode.ErrBadRequest.WithMessage("子账号不可再创建子账号")
	}
	if parent.EnterpriseCode == "" {
		return nil, ecode.ErrBadRequest.WithMessage("主账号未加入企业，不可创建子账号")
	}

	count, err := s.userRepo.CountSubAccounts(ctx, parentID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if count >= entity.MaxSubAccounts {
		return nil, ecode.ErrBadRequest.WithMessage("子账号数量已达上限（最多5个）")
	}

	exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if exists {
		return nil, ecode.ErrPhoneExists
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ecode.ErrInternal.WithMessage("密码加密失败")
	}

	sub := &entity.User{
		Phone:          req.Phone,
		Password:       string(hashedPwd),
		RealName:       req.RealName,
		QQ:             req.QQ,
		EnterpriseCode: parent.EnterpriseCode,
		Role:           entity.TypeEnterprise,
		ParentID:       parentID,
		Permissions:    entity.StringList(req.Permissions),
		Status:         entity.StatusActive,
		CreatedBy:      parentID,
		UpdatedBy:      parentID,
	}
	if req.Email != "" {
		sub.Email = &req.Email
	}

	if err := s.userRepo.Create(ctx, sub); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建子账号失败")
	}

	return &dto.SubAccountInfo{
		ID:          sub.ID,
		UserNo:      sub.UserNo,
		Phone:       sub.Phone,
		RealName:    sub.RealName,
		Email:       req.Email,
		QQ:          sub.QQ,
		Permissions: []string(sub.Permissions),
		Status:      int(sub.Status),
		CreatedAt:   sub.CreatedAt,
	}, nil
}

func (s *UserService) ListSubAccounts(ctx context.Context, parentID uint64) ([]*dto.SubAccountInfo, error) {
	subs, err := s.userRepo.FindSubAccounts(ctx, parentID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	result := make([]*dto.SubAccountInfo, 0, len(subs))
	for _, sub := range subs {
		var email string
		if sub.Email != nil {
			email = *sub.Email
		}
		result = append(result, &dto.SubAccountInfo{
			ID:          sub.ID,
			UserNo:      sub.UserNo,
			Phone:       sub.Phone,
			RealName:    sub.RealName,
			Email:       email,
			QQ:          sub.QQ,
			Permissions: []string(sub.Permissions),
			Status:      int(sub.Status),
			CreatedAt:   sub.CreatedAt,
		})
	}
	return result, nil
}

func (s *UserService) UpdateSubAccount(ctx context.Context, parentID, subID uint64, req *dto.UpdateSubAccountRequest) (*dto.SubAccountInfo, error) {
	sub, err := s.userRepo.FindByID(ctx, subID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if sub == nil {
		return nil, ecode.ErrUserNotFound
	}
	if sub.ParentID != parentID {
		return nil, ecode.ErrBadRequest.WithMessage("无权操作此子账号")
	}

	if req.RealName != "" {
		sub.RealName = req.RealName
	}
	if req.QQ != "" {
		sub.QQ = req.QQ
	}
	if req.Email != "" {
		sub.Email = &req.Email
	}
	if req.Permissions != nil {
		sub.Permissions = entity.StringList(req.Permissions)
	}
	if req.Status != nil {
		sub.Status = entity.UserStatus(*req.Status)
	}
	sub.UpdatedBy = parentID

	if err := s.userRepo.Update(ctx, sub); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新子账号失败")
	}

	updated, _ := s.userRepo.FindByID(ctx, subID)
	if updated == nil {
		return nil, ecode.ErrUserNotFound
	}

	var email string
	if updated.Email != nil {
		email = *updated.Email
	}
	return &dto.SubAccountInfo{
		ID:          updated.ID,
		UserNo:      updated.UserNo,
		Phone:       updated.Phone,
		RealName:    updated.RealName,
		Email:       email,
		QQ:          updated.QQ,
		Permissions: []string(updated.Permissions),
		Status:      int(updated.Status),
		CreatedAt:   updated.CreatedAt,
	}, nil
}

func (s *UserService) GetDashboard(ctx context.Context, userID uint64) (*dto.DashboardResponse, error) {
	dashboardKey := cache.Key("user:dashboard", userID)
	var cached dto.DashboardResponse
	if cache.GetJSON(ctx, dashboardKey, &cached) {
		return &cached, nil
	}

	db := database.GetDB()
	stats := dto.DashboardStats{}

	var inTransitResult struct {
		Total float64
	}
	db.WithContext(ctx).Table("orders").
		Where("buyer_id = ? AND status = ?", userID, 6).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&inTransitResult)
	stats.InTransitAmount = inTransitResult.Total

	var pendingCount int64
	db.WithContext(ctx).Table("orders").
		Where("buyer_id = ? AND status = ?", userID, 3).
		Count(&pendingCount)
	stats.PendingPaymentCount = pendingCount

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthlyResult struct {
		Total float64
	}
	db.WithContext(ctx).Table("orders").
		Where("buyer_id = ? AND status NOT IN ? AND created_at >= ?", userID, []int{9, 0}, firstOfMonth).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&monthlyResult)
	stats.MonthlyPurchaseAmount = monthlyResult.Total

	var stockResult struct {
		Total float64
	}
	db.WithContext(ctx).Table("products").
		Where("status = ? AND available_qty > 0", 3).
		Select("COALESCE(SUM(available_qty), 0) as total").
		Scan(&stockResult)
	stats.StockQuantity = stockResult.Total

	type recentOrderRow struct {
		ID          uint64
		OrderNo     string
		Status      int
		TotalAmount float64
		CreatedAt   time.Time
		ProductName string
	}
	var rows []recentOrderRow
	db.WithContext(ctx).Table("orders").
		Where("buyer_id = ?", userID).
		Order("id desc").
		Limit(5).
		Select("id, order_no, status, total_amount, created_at, JSON_UNQUOTE(JSON_EXTRACT(JSON_EXTRACT(items, '$[0]'), '$.product_name')) as product_name").
		Scan(&rows)

	recentOrders := make([]dto.RecentOrder, 0, len(rows))
	for _, r := range rows {
		recentOrders = append(recentOrders, dto.RecentOrder{
			ID:          r.ID,
			OrderNo:     r.OrderNo,
			Status:      r.Status,
			TotalAmount: r.TotalAmount,
			CreatedAt:   r.CreatedAt.Format("2006-01-02"),
			ProductName: r.ProductName,
		})
	}

	result := &dto.DashboardResponse{
		Stats:        stats,
		RecentOrders: recentOrders,
	}
	cache.SetJSON(ctx, dashboardKey, result, cache.DashboardStatsTTL)
	return result, nil
}
