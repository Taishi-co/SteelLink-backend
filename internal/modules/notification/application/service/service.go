package service

import (
	"context"
	"strings"
	"time"

	notificationEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
	userEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

type CreateNotificationInput struct {
	RecipientUserIDs       []uint64
	RecipientEnterpriseIDs []uint64
	RecipientRoles         []string
	SendToAll              bool
	SenderUserID           uint64
	Type                   notificationEntity.NotificationType
	Title                  string
	Content                string
	LinkURL                string
	Metadata               notificationEntity.NotificationMetadata
}

type ListQuery struct {
	UserID   uint64
	Status   *int
	Page     int
	PageSize int
}

func NewNotificationService() *NotificationService {
	return &NotificationService{db: database.GetDB()}
}

func (s *NotificationService) Create(ctx context.Context, input CreateNotificationInput) ([]notificationEntity.Notification, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return nil, ecode.ErrBadRequest.WithMessage("消息标题和内容不能为空")
	}
	if input.Type == "" {
		input.Type = notificationEntity.NotificationTypeSystem
	}

	recipients, err := s.resolveRecipients(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("没有可接收站内信的用户")
	}

	notifications := make([]notificationEntity.Notification, 0, len(recipients))
	for _, recipient := range recipients {
		notifications = append(notifications, notificationEntity.Notification{
			RecipientUserID:       recipient.UserID,
			RecipientEnterpriseID: recipient.EnterpriseID,
			SenderUserID:          input.SenderUserID,
			Type:                  input.Type,
			Title:                 title,
			Content:               content,
			LinkURL:               strings.TrimSpace(input.LinkURL),
			Status:                notificationEntity.NotificationStatusUnread,
			Metadata:              input.Metadata,
		})
	}
	if err := s.db.WithContext(ctx).Create(&notifications).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("创建站内信失败")
	}
	return notifications, nil
}

func (s *NotificationService) CreateForEnterprise(ctx context.Context, enterpriseID uint64, input CreateNotificationInput) error {
	if enterpriseID == 0 {
		return nil
	}
	input.RecipientEnterpriseIDs = []uint64{enterpriseID}
	_, err := s.Create(ctx, input)
	if err != nil {
		if e, ok := err.(*ecode.Error); ok && e.Code == ecode.ErrBadRequest.Code {
			return nil
		}
		return err
	}
	return nil
}

func (s *NotificationService) List(ctx context.Context, query ListQuery) ([]notificationEntity.Notification, int64, int64, error) {
	if query.UserID == 0 {
		return nil, 0, 0, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	db := s.db.WithContext(ctx).Model(&notificationEntity.Notification{}).Where("recipient_user_id = ?", query.UserID)
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, 0, ecode.ErrDatabase.WithMessage("查询站内信失败")
	}
	var unread int64
	if err := s.db.WithContext(ctx).Model(&notificationEntity.Notification{}).
		Where("recipient_user_id = ? AND status = ?", query.UserID, notificationEntity.NotificationStatusUnread).
		Count(&unread).Error; err != nil {
		return nil, 0, 0, ecode.ErrDatabase.WithMessage("查询未读站内信失败")
	}
	var list []notificationEntity.Notification
	if err := db.Order("id desc").Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&list).Error; err != nil {
		return nil, 0, 0, ecode.ErrDatabase.WithMessage("查询站内信失败")
	}
	return list, total, unread, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uint64) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&notificationEntity.Notification{}).
		Where("id = ? AND recipient_user_id = ?", id, userID).
		Updates(map[string]interface{}{"status": notificationEntity.NotificationStatusRead, "read_at": &now})
	if result.Error != nil {
		return ecode.ErrDatabase.WithMessage("标记站内信已读失败")
	}
	if result.RowsAffected == 0 {
		return ecode.ErrNotFound.WithMessage("站内信不存在")
	}
	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint64) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&notificationEntity.Notification{}).
		Where("recipient_user_id = ? AND status = ?", userID, notificationEntity.NotificationStatusUnread).
		Updates(map[string]interface{}{"status": notificationEntity.NotificationStatusRead, "read_at": &now}).Error; err != nil {
		return ecode.ErrDatabase.WithMessage("标记全部站内信已读失败")
	}
	return nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&notificationEntity.Notification{}).
		Where("recipient_user_id = ? AND status = ?", userID, notificationEntity.NotificationStatusUnread).
		Count(&count).Error; err != nil {
		return 0, ecode.ErrDatabase.WithMessage("查询未读站内信失败")
	}
	return count, nil
}

type recipient struct {
	UserID       uint64
	EnterpriseID uint64
}

func (s *NotificationService) resolveRecipients(ctx context.Context, input CreateNotificationInput) ([]recipient, error) {
	db := s.db.WithContext(ctx).Model(&userEntity.User{}).Where("status = ? AND deleted_at IS NULL", userEntity.StatusActive)
	if !input.SendToAll {
		conditions := make([]string, 0, 3)
		args := make([]interface{}, 0, 3)
		hasFilter := false
		if len(input.RecipientUserIDs) > 0 {
			conditions = append(conditions, "id IN ?")
			args = append(args, uniqueUint64(input.RecipientUserIDs))
			hasFilter = true
		}
		if len(input.RecipientEnterpriseIDs) > 0 {
			var enterprises []userEntity.Enterprise
			if err := s.db.WithContext(ctx).Select("enterprise_code").Where("id IN ?", uniqueUint64(input.RecipientEnterpriseIDs)).Find(&enterprises).Error; err != nil {
				return nil, ecode.ErrDatabase.WithMessage("查询站内信接收企业失败")
			}
			codes := make([]string, 0, len(enterprises))
			for _, enterprise := range enterprises {
				codes = append(codes, enterprise.EnterpriseCode)
			}
			if len(codes) > 0 {
				conditions = append(conditions, "enterprise_code IN ?")
				args = append(args, uniqueStrings(codes))
				hasFilter = true
			}
		}
		if len(input.RecipientRoles) > 0 {
			conditions = append(conditions, "role IN ?")
			args = append(args, uniqueStrings(input.RecipientRoles))
			hasFilter = true
		}
		if !hasFilter {
			return nil, nil
		}
		db = db.Where(strings.Join(conditions, " OR "), args...)
	}

	var rows []struct {
		ID             uint64
		EnterpriseCode string
	}
	if err := db.Select("id, enterprise_code").Find(&rows).Error; err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询站内信接收用户失败")
	}

	enterpriseIDs := map[string]uint64{}
	if len(rows) > 0 {
		var codes []string
		for _, row := range rows {
			if row.EnterpriseCode != "" {
				codes = append(codes, row.EnterpriseCode)
			}
		}
		if len(codes) > 0 {
			var enterprises []userEntity.Enterprise
			if err := s.db.WithContext(ctx).Select("id, enterprise_code").Where("enterprise_code IN ?", uniqueStrings(codes)).Find(&enterprises).Error; err != nil {
				return nil, ecode.ErrDatabase.WithMessage("查询站内信接收企业失败")
			}
			for _, enterprise := range enterprises {
				enterpriseIDs[enterprise.EnterpriseCode] = enterprise.ID
			}
		}
	}

	seen := map[uint64]struct{}{}
	recipients := make([]recipient, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.ID]; ok {
			continue
		}
		seen[row.ID] = struct{}{}
		recipients = append(recipients, recipient{UserID: row.ID, EnterpriseID: enterpriseIDs[row.EnterpriseCode]})
	}
	return recipients, nil
}

func uniqueUint64(values []uint64) []uint64 {
	seen := map[uint64]struct{}{}
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
