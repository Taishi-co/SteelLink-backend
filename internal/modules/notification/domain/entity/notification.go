package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeAdmin  NotificationType = "admin"
	NotificationTypeOrder  NotificationType = "order"
	NotificationTypeAudit  NotificationType = "audit"
	NotificationTypeSystem NotificationType = "system"
)

type NotificationStatus int

const (
	NotificationStatusUnread NotificationStatus = 0
	NotificationStatusRead   NotificationStatus = 1
)

type NotificationMetadata map[string]interface{}

func (m NotificationMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *NotificationMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

type Notification struct {
	ID                    uint64               `gorm:"primaryKey;autoIncrement" json:"id"`
	RecipientUserID       uint64               `gorm:"index;not null;comment:接收用户ID" json:"recipient_user_id"`
	RecipientEnterpriseID uint64               `gorm:"index;default:0;comment:接收企业ID" json:"recipient_enterprise_id"`
	SenderUserID          uint64               `gorm:"index;default:0;comment:发送用户ID，0表示系统" json:"sender_user_id"`
	Type                  NotificationType     `gorm:"type:varchar(32);index;not null;comment:消息类型" json:"type"`
	Title                 string               `gorm:"type:varchar(120);not null;comment:标题" json:"title"`
	Content               string               `gorm:"type:text;not null;comment:正文" json:"content"`
	LinkURL               string               `gorm:"type:varchar(500);default:;comment:站内跳转链接" json:"link_url"`
	Status                NotificationStatus   `gorm:"type:tinyint;default:0;index;comment:0未读1已读" json:"status"`
	ReadAt                *time.Time           `gorm:"comment:阅读时间" json:"read_at,omitempty"`
	Metadata              NotificationMetadata `gorm:"type:json;comment:扩展上下文" json:"metadata"`
	CreatedAt             time.Time            `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt             time.Time            `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt             gorm.DeletedAt       `gorm:"index;comment:删除时间" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
