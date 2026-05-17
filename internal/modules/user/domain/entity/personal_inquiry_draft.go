package entity

import (
	"time"

	"gorm.io/gorm"
)

type PersonalInquiryDraft struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	DraftNo         string         `gorm:"type:varchar(32);uniqueIndex;comment:个人询价草稿编号" json:"draft_no"`
	UserID          uint64         `gorm:"index;not null;comment:个人用户ID" json:"user_id"`
	Title           string         `gorm:"type:varchar(200);not null;comment:草稿标题" json:"title"`
	ProductName     string         `gorm:"type:varchar(200);comment:商品名称" json:"product_name"`
	Category        string         `gorm:"type:varchar(100);comment:品类" json:"category"`
	Material        string         `gorm:"type:varchar(100);comment:材质" json:"material"`
	Spec            string         `gorm:"type:varchar(100);comment:规格" json:"spec"`
	Quantity        float64        `gorm:"type:decimal(12,2);comment:数量" json:"quantity"`
	Unit            string         `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	TargetPrice     float64        `gorm:"type:decimal(12,2);comment:目标单价" json:"target_price"`
	DeliveryAddress string         `gorm:"type:varchar(255);comment:收货地址" json:"delivery_address"`
	Requirement     string         `gorm:"type:varchar(1000);comment:采购要求" json:"requirement"`
	Status          string         `gorm:"type:varchar(32);index;default:draft;comment:状态 draft/archived" json:"status"`
	CreatedBy       uint64         `gorm:"comment:创建者" json:"created_by"`
	UpdatedBy       uint64         `gorm:"comment:修改者" json:"updated_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (PersonalInquiryDraft) TableName() string {
	return "personal_inquiry_drafts"
}

func (d *PersonalInquiryDraft) BeforeCreate(tx *gorm.DB) error {
	if d.DraftNo == "" {
		d.DraftNo = "PID" + time.Now().Format("20060102150405") + randomDigits(4)
	}
	if d.Unit == "" {
		d.Unit = "吨"
	}
	if d.Status == "" {
		d.Status = "draft"
	}
	return nil
}
