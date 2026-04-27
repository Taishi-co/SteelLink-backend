package entity

import (
	"time"

	"gorm.io/gorm"
)

type BiddingStatus string

const (
	BiddingStatusOngoing BiddingStatus = "ongoing"
	BiddingStatusSprint  BiddingStatus = "sprint"
	BiddingStatusSuccess BiddingStatus = "success"
	BiddingStatusClosed  BiddingStatus = "closed"
)

type BiddingActivity struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	BiddingNo       string         `gorm:"type:varchar(40);uniqueIndex;comment:竞价编号" json:"bidding_no"`
	Title           string         `gorm:"type:varchar(200);index;comment:竞价标题" json:"title"`
	Category        string         `gorm:"type:varchar(80);index;comment:品类" json:"category"`
	Quantity        float64        `gorm:"type:decimal(14,2);comment:数量" json:"quantity"`
	Unit            string         `gorm:"type:varchar(20);default:吨;comment:单位" json:"unit"`
	StartPrice      float64        `gorm:"type:decimal(14,2);comment:起拍价" json:"start_price"`
	CurrentPrice    float64        `gorm:"type:decimal(14,2);comment:当前价" json:"current_price"`
	DealAmount      float64        `gorm:"type:decimal(16,2);comment:成交金额" json:"deal_amount"`
	PriceChangeRate float64        `gorm:"type:decimal(8,2);comment:价格变化百分比" json:"price_change_rate"`
	Participants    int            `gorm:"comment:参与企业数" json:"participants"`
	Progress        int            `gorm:"comment:进度百分比" json:"progress"`
	Status          BiddingStatus  `gorm:"type:varchar(32);index;comment:状态" json:"status"`
	BuyerName       string         `gorm:"type:varchar(120);comment:采购方" json:"buyer_name"`
	SupplierName    string         `gorm:"type:varchar(120);comment:中标方" json:"supplier_name"`
	SavingRate      float64        `gorm:"type:decimal(8,2);comment:节省比例" json:"saving_rate"`
	StartedAt       time.Time      `gorm:"comment:开始时间" json:"started_at"`
	EndedAt         time.Time      `gorm:"comment:结束时间" json:"ended_at"`
	ClosedAt        *time.Time     `gorm:"comment:成交时间" json:"closed_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BiddingActivity) TableName() string {
	return "bidding_activities"
}
