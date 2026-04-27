package dto

import "time"

type ListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
}

type BiddingStats struct {
	ActiveCount          int64   `json:"active_count"`
	SuccessCount         int64   `json:"success_count"`
	TotalTransaction     float64 `json:"total_transaction"`
	ParticipatingCompany int64   `json:"participating_company"`
}

type BiddingActivityItem struct {
	ID              uint64     `json:"id"`
	BiddingNo       string     `json:"bidding_no"`
	Title           string     `json:"title"`
	Category        string     `json:"category"`
	Quantity        float64    `json:"quantity"`
	Unit            string     `json:"unit"`
	StartPrice      float64    `json:"start_price"`
	CurrentPrice    float64    `json:"current_price"`
	DealAmount      float64    `json:"deal_amount"`
	PriceChangeRate float64    `json:"price_change_rate"`
	Participants    int        `json:"participants"`
	Progress        int        `json:"progress"`
	Status          string     `json:"status"`
	BuyerName       string     `json:"buyer_name"`
	SupplierName    string     `json:"supplier_name"`
	SavingRate      float64    `json:"saving_rate"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         time.Time  `json:"ended_at"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
}

type BiddingListResponse struct {
	List     []BiddingActivityItem `json:"list"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type ProcurementStats struct {
	ActivityCount        int64   `json:"activity_count"`
	PreparingCount       int64   `json:"preparing_count"`
	SuccessCount         int64   `json:"success_count"`
	TotalTransaction     float64 `json:"total_transaction"`
	ParticipatingCompany int64   `json:"participating_company"`
	DemandQuantity       float64 `json:"demand_quantity"`
}

type ProcurementActivityItem struct {
	ID                 uint64               `json:"id"`
	ProcurementNo      string               `json:"procurement_no"`
	Title              string               `json:"title"`
	LeaderName         string               `json:"leader_name"`
	Status             string               `json:"status"`
	Category           string               `json:"category"`
	Target             string               `json:"target"`
	TargetQuantity     float64              `json:"target_quantity"`
	Unit               string               `json:"unit"`
	BudgetMin          float64              `json:"budget_min"`
	BudgetMax          float64              `json:"budget_max"`
	BudgetDisplay      string               `json:"budget_display"`
	Stage              string               `json:"stage"`
	Progress           int                  `json:"progress"`
	Region             string               `json:"region"`
	OpenDate           string               `json:"open_date"`
	ParticipatingCount int                  `json:"participating_count"`
	DemandQuantity     float64              `json:"demand_quantity"`
	QualifiedSuppliers int                  `json:"qualified_suppliers"`
	QuoteCount         int                  `json:"quote_count"`
	ContractCount      int                  `json:"contract_count"`
	DeliveryDays       int                  `json:"delivery_days"`
	QuoteMode          string               `json:"quote_mode"`
	EvaluationMode     string               `json:"evaluation_mode"`
	Documents          []string             `json:"documents"`
	Logs               []ProcurementLogItem `json:"logs,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

type ProcurementLogItem struct {
	ID        uint64    `json:"id"`
	Content   string    `json:"content"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type ProcurementListResponse struct {
	List     []ProcurementActivityItem `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}
