package entity

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type PostType string

const (
	PostTypeSupply    PostType = "supply"
	PostTypeDemand    PostType = "demand"
	PostTypeWarehouse PostType = "warehouse"
	PostTypeLogistics PostType = "logistics"
)

type PostStatus int

const (
	PostStatusDraft     PostStatus = 0
	PostStatusPublished PostStatus = 1
	PostStatusMatched   PostStatus = 2
	PostStatusClosed    PostStatus = 3
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if len(bytes) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(bytes, s)
}

type SupplyDemandPost struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PostNo            string         `gorm:"type:varchar(32);uniqueIndex;comment:供求信息编号" json:"post_no"`
	Type              PostType       `gorm:"type:varchar(32);index;comment:类型(supply/demand/warehouse/logistics)" json:"type"`
	Title             string         `gorm:"type:varchar(200);comment:标题" json:"title"`
	Category          string         `gorm:"type:varchar(50);index;comment:品类" json:"category"`
	Material          string         `gorm:"type:varchar(100);comment:材质" json:"material"`
	Spec              string         `gorm:"type:varchar(100);comment:规格" json:"spec"`
	Quantity          float64        `gorm:"type:decimal(12,2);comment:数量" json:"quantity"`
	Unit              string         `gorm:"type:varchar(20);comment:单位" json:"unit"`
	TargetPrice       float64        `gorm:"type:decimal(12,2);comment:目标单价/服务单价" json:"target_price"`
	Budget            float64        `gorm:"type:decimal(14,2);comment:预算/报价金额" json:"budget"`
	Currency          string         `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	Province          string         `gorm:"type:varchar(50);index;comment:省份" json:"province"`
	City              string         `gorm:"type:varchar(50);index;comment:城市" json:"city"`
	District          string         `gorm:"type:varchar(50);comment:区县" json:"district"`
	DeliveryAddress   string         `gorm:"type:varchar(255);comment:交付/服务地址" json:"delivery_address"`
	Description       string         `gorm:"type:text;comment:描述" json:"description"`
	Tags              StringList     `gorm:"type:json;comment:标签" json:"tags"`
	ContactName       string         `gorm:"type:varchar(50);comment:联系人" json:"contact_name"`
	ContactPhone      string         `gorm:"type:varchar(30);comment:联系电话" json:"contact_phone"`
	FinancingRequired bool           `gorm:"default:false;comment:是否需要融资" json:"financing_required"`
	MatchCount        int            `gorm:"default:0;comment:意向/撮合数量" json:"match_count"`
	MatchAmount       float64        `gorm:"type:decimal(14,2);default:0;comment:撮合金额" json:"match_amount"`
	Status            PostStatus     `gorm:"type:tinyint;default:1;index;comment:状态" json:"status"`
	EnterpriseID      uint64         `gorm:"index;comment:所属企业ID" json:"enterprise_id"`
	CreatedBy         uint64         `gorm:"index;comment:创建者" json:"created_by"`
	CreatedAt         time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy         uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (SupplyDemandPost) TableName() string {
	return "supply_demand_posts"
}

func (p *SupplyDemandPost) BeforeCreate(tx *gorm.DB) error {
	if p.PostNo == "" {
		p.PostNo = generatePostNo()
	}
	if p.Currency == "" {
		p.Currency = "CNY"
	}
	if p.Unit == "" {
		p.Unit = "吨"
	}
	if p.Status == 0 {
		p.Status = PostStatusPublished
	}
	return nil
}

func (p *SupplyDemandPost) IsVisible() bool {
	return p.Status == PostStatusPublished || p.Status == PostStatusMatched
}

func IsValidPostType(value PostType) bool {
	switch value {
	case PostTypeSupply, PostTypeDemand, PostTypeWarehouse, PostTypeLogistics:
		return true
	default:
		return false
	}
}

func IsValidPostStatus(value PostStatus) bool {
	switch value {
	case PostStatusDraft, PostStatusPublished, PostStatusMatched, PostStatusClosed:
		return true
	default:
		return false
	}
}

func generatePostNo() string {
	return "SD" + time.Now().Format("20060102150405") + randomDigits(4)
}

func randomDigits(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
