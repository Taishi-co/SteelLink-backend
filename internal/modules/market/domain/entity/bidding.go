package entity

import (
	"time"

	"gorm.io/gorm"
)

type BiddingStatus string

const (
	BiddingStatusDraft        BiddingStatus = "draft"
	BiddingStatusBidding      BiddingStatus = "bidding"
	BiddingStatusPendingAward BiddingStatus = "pending_award"
	BiddingStatusAwarded      BiddingStatus = "awarded"
	BiddingStatusOrderCreated BiddingStatus = "order_created"
	BiddingStatusFailed       BiddingStatus = "failed"
	BiddingStatusClosed       BiddingStatus = "closed"

	// Legacy statuses are kept for data compatibility with early market seed data.
	BiddingStatusOngoing BiddingStatus = "ongoing"
	BiddingStatusSprint  BiddingStatus = "sprint"
	BiddingStatusSuccess BiddingStatus = "success"
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
	BuyerID         uint64         `gorm:"index;comment:采购方企业ID" json:"buyer_id"`
	BuyerName       string         `gorm:"type:varchar(120);comment:采购方" json:"buyer_name"`
	DeliveryRegion  string         `gorm:"type:varchar(160);comment:交付区域" json:"delivery_region"`
	QuoteCount      int            `gorm:"comment:报价次数" json:"quote_count"`
	AwardQuoteID    uint64         `gorm:"index;comment:中标报价ID" json:"award_quote_id"`
	AwardSupplierID uint64         `gorm:"index;comment:中标供应商企业ID" json:"award_supplier_id"`
	OrderID         uint64         `gorm:"index;comment:生成订单ID" json:"order_id"`
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

type BiddingQuote struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	BiddingID    uint64         `gorm:"index;comment:竞价活动ID" json:"bidding_id"`
	SupplierID   uint64         `gorm:"index;comment:供应商企业ID" json:"supplier_id"`
	SupplierName string         `gorm:"type:varchar(120);comment:供应商名称" json:"supplier_name"`
	UnitPrice    float64        `gorm:"type:decimal(14,2);comment:报价单价" json:"unit_price"`
	Quantity     float64        `gorm:"type:decimal(14,2);comment:可供数量" json:"quantity"`
	DeliveryDays int            `gorm:"comment:交付周期天数" json:"delivery_days"`
	Remark       string         `gorm:"type:varchar(500);comment:报价备注" json:"remark"`
	Round        int            `gorm:"comment:报价轮次" json:"round"`
	IsAwarded    bool           `gorm:"index;comment:是否中标" json:"is_awarded"`
	CreatedBy    uint64         `gorm:"index;comment:创建人" json:"created_by"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BiddingQuote) TableName() string {
	return "bidding_quotes"
}
