package dto

import "time"

type CreatePostRequest struct {
	Type              string   `json:"type" binding:"required,oneof=supply demand warehouse logistics"`
	Title             string   `json:"title" binding:"required,max=200"`
	Category          string   `json:"category" binding:"required,max=50"`
	Material          string   `json:"material" binding:"max=100"`
	Spec              string   `json:"spec" binding:"max=100"`
	Quantity          float64  `json:"quantity" binding:"omitempty,gte=0"`
	Unit              string   `json:"unit" binding:"max=20"`
	TargetPrice       float64  `json:"target_price" binding:"omitempty,gte=0"`
	Budget            float64  `json:"budget" binding:"omitempty,gte=0"`
	Province          string   `json:"province" binding:"max=50"`
	City              string   `json:"city" binding:"max=50"`
	District          string   `json:"district" binding:"max=50"`
	DeliveryAddress   string   `json:"delivery_address" binding:"max=255"`
	Description       string   `json:"description"`
	Tags              []string `json:"tags"`
	ContactName       string   `json:"contact_name" binding:"max=50"`
	ContactPhone      string   `json:"contact_phone" binding:"max=30"`
	FinancingRequired bool     `json:"financing_required"`
}

type PostListRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type         string `form:"type" binding:"omitempty,oneof=supply demand warehouse logistics"`
	Status       *int   `form:"status" binding:"omitempty,oneof=0 1 2 3 4 5"`
	Keyword      string `form:"keyword"`
	Category     string `form:"category"`
	Province     string `form:"province"`
	City         string `form:"city"`
	EnterpriseID uint64 `form:"enterprise_id"`
	SortBy       string `form:"sort_by"`
}

type UpdatePostStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=3"`
}

type PostInfo struct {
	ID                uint64     `json:"id"`
	PostNo            string     `json:"post_no"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Category          string     `json:"category"`
	Material          string     `json:"material"`
	Spec              string     `json:"spec"`
	Quantity          float64    `json:"quantity"`
	Unit              string     `json:"unit"`
	TargetPrice       float64    `json:"target_price"`
	Budget            float64    `json:"budget"`
	Currency          string     `json:"currency"`
	Province          string     `json:"province"`
	City              string     `json:"city"`
	District          string     `json:"district"`
	DeliveryAddress   string     `json:"delivery_address"`
	Description       string     `json:"description"`
	Tags              []string   `json:"tags"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	FinancingRequired bool       `json:"financing_required"`
	MatchCount        int        `json:"match_count"`
	MatchAmount       float64    `json:"match_amount"`
	Status            int        `json:"status"`
	AuditRemark       string     `json:"audit_remark"`
	AuditorID         uint64     `json:"auditor_id"`
	AuditedAt         *time.Time `json:"audited_at"`
	EnterpriseID      uint64     `json:"enterprise_id"`
	PublisherType     string     `json:"publisher_type"`
	PublisherName     string     `json:"publisher_name"`
	CreatedBy         uint64     `json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type PostListResponse struct {
	List     []*PostInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type PostStatsResponse struct {
	PublishedCount int64   `json:"published_count"`
	MatchedCount   int64   `json:"matched_count"`
	MatchedAmount  float64 `json:"matched_amount"`
	TodayNewCount  int64   `json:"today_new_count"`
}
