package dto

import "time"

type ListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Category string `form:"category"`
	Scope    string `form:"scope"`
}

type BiddingStats struct {
	ActiveCount          int64   `json:"active_count"`
	SuccessCount         int64   `json:"success_count"`
	TotalTransaction     float64 `json:"total_transaction"`
	ParticipatingCompany int64   `json:"participating_company"`
}

type BiddingActivityItem struct {
	ID              uint64             `json:"id"`
	BiddingNo       string             `json:"bidding_no"`
	Title           string             `json:"title"`
	Category        string             `json:"category"`
	Quantity        float64            `json:"quantity"`
	Unit            string             `json:"unit"`
	StartPrice      float64            `json:"start_price"`
	CurrentPrice    float64            `json:"current_price"`
	DealAmount      float64            `json:"deal_amount"`
	PriceChangeRate float64            `json:"price_change_rate"`
	Participants    int                `json:"participants"`
	Progress        int                `json:"progress"`
	Status          string             `json:"status"`
	BuyerID         uint64             `json:"buyer_id"`
	BuyerName       string             `json:"buyer_name"`
	DeliveryRegion  string             `json:"delivery_region"`
	QuoteCount      int                `json:"quote_count"`
	AwardQuoteID    uint64             `json:"award_quote_id"`
	AwardSupplierID uint64             `json:"award_supplier_id"`
	OrderID         uint64             `json:"order_id"`
	SupplierName    string             `json:"supplier_name"`
	SavingRate      float64            `json:"saving_rate"`
	StartedAt       time.Time          `json:"started_at"`
	EndedAt         time.Time          `json:"ended_at"`
	ClosedAt        *time.Time         `json:"closed_at,omitempty"`
	Quotes          []BiddingQuoteItem `json:"quotes,omitempty"`
}

type BiddingListResponse struct {
	List     []BiddingActivityItem `json:"list"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type BiddingQuoteItem struct {
	ID           uint64    `json:"id"`
	BiddingID    uint64    `json:"bidding_id"`
	SupplierID   uint64    `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	UnitPrice    float64   `json:"unit_price"`
	Quantity     float64   `json:"quantity"`
	DeliveryDays int       `json:"delivery_days"`
	Remark       string    `json:"remark"`
	Round        int       `json:"round"`
	IsAwarded    bool      `json:"is_awarded"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateBiddingRequest struct {
	Title          string  `json:"title" binding:"required"`
	Category       string  `json:"category" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	Unit           string  `json:"unit"`
	StartPrice     float64 `json:"start_price" binding:"required,gt=0"`
	DeliveryRegion string  `json:"delivery_region" binding:"required"`
	EndedAt        string  `json:"ended_at" binding:"required"`
}

type SubmitBiddingQuoteRequest struct {
	UnitPrice    float64 `json:"unit_price" binding:"required,gt=0"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
	DeliveryDays int     `json:"delivery_days" binding:"required,gt=0"`
	Remark       string  `json:"remark"`
}

type AwardBiddingRequest struct {
	QuoteID uint64 `json:"quote_id" binding:"required"`
}

type CreateBiddingOrderResponse struct {
	Bidding BiddingActivityItem `json:"bidding"`
	OrderID uint64              `json:"order_id"`
	OrderNo string              `json:"order_no"`
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
	ID                 uint64                       `json:"id"`
	ProcurementNo      string                       `json:"procurement_no"`
	Title              string                       `json:"title"`
	LeaderName         string                       `json:"leader_name"`
	Status             string                       `json:"status"`
	ProductID          uint64                       `json:"product_id"`
	ProductNo          string                       `json:"product_no"`
	ProductName        string                       `json:"product_name"`
	SellerEnterpriseID uint64                       `json:"seller_enterprise_id"`
	SellerName         string                       `json:"seller_name"`
	GroupPrice         float64                      `json:"group_price"`
	MinGroupQuantity   float64                      `json:"min_group_quantity"`
	MaxGroupQuantity   float64                      `json:"max_group_quantity"`
	Category           string                       `json:"category"`
	Target             string                       `json:"target"`
	TargetQuantity     float64                      `json:"target_quantity"`
	Unit               string                       `json:"unit"`
	BudgetMin          float64                      `json:"budget_min"`
	BudgetMax          float64                      `json:"budget_max"`
	BudgetDisplay      string                       `json:"budget_display"`
	Stage              string                       `json:"stage"`
	Progress           int                          `json:"progress"`
	Region             string                       `json:"region"`
	OpenDate           string                       `json:"open_date"`
	ParticipatingCount int                          `json:"participating_count"`
	DemandQuantity     float64                      `json:"demand_quantity"`
	QualifiedSuppliers int                          `json:"qualified_suppliers"`
	QuoteCount         int                          `json:"quote_count"`
	ContractCount      int                          `json:"contract_count"`
	DeliveryDays       int                          `json:"delivery_days"`
	QuoteMode          string                       `json:"quote_mode"`
	EvaluationMode     string                       `json:"evaluation_mode"`
	Documents          []string                     `json:"documents"`
	AdminRemark        string                       `json:"admin_remark"`
	SellerRemark       string                       `json:"seller_remark"`
	Source             string                       `json:"source"`
	StartedAt          *time.Time                   `json:"started_at,omitempty"`
	EndedAt            *time.Time                   `json:"ended_at,omitempty"`
	SellerRespondedAt  *time.Time                   `json:"seller_responded_at,omitempty"`
	ClosedAt           *time.Time                   `json:"closed_at,omitempty"`
	MasterOrderID      uint64                       `json:"master_order_id"`
	MasterOrderNo      string                       `json:"master_order_no"`
	Participants       []ProcurementParticipantItem `json:"participants,omitempty"`
	Logs               []ProcurementLogItem         `json:"logs,omitempty"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}

type ProcurementParticipantItem struct {
	ID                        uint64     `json:"id"`
	ActivityID                uint64     `json:"activity_id"`
	BuyerEnterpriseID         uint64     `json:"buyer_enterprise_id"`
	BuyerUserID               uint64     `json:"buyer_user_id"`
	BuyerName                 string     `json:"buyer_name"`
	Quantity                  float64    `json:"quantity"`
	Unit                      string     `json:"unit"`
	UnitPrice                 float64    `json:"unit_price"`
	TotalAmount               float64    `json:"total_amount"`
	ContactName               string     `json:"contact_name"`
	ContactPhone              string     `json:"contact_phone"`
	DeliveryAddressSource     string     `json:"delivery_address_source"`
	DestWarehouseServiceID    uint64     `json:"dest_warehouse_service_id"`
	DestWarehouseEnterpriseID uint64     `json:"dest_warehouse_enterprise_id"`
	DestWarehouseName         string     `json:"dest_warehouse_name"`
	DestWarehouseProvince     string     `json:"dest_warehouse_province"`
	DestWarehouseCity         string     `json:"dest_warehouse_city"`
	DestWarehouseDistrict     string     `json:"dest_warehouse_district"`
	DestWarehouseAddress      string     `json:"dest_warehouse_address"`
	DeliveryAddress           string     `json:"delivery_address"`
	DeliveryMethod            string     `json:"delivery_method"`
	PaymentMethod             string     `json:"payment_method"`
	DeliveryNote              string     `json:"delivery_note"`
	AgreementSigned           bool       `json:"agreement_signed"`
	AgreementSignedAt         *time.Time `json:"agreement_signed_at,omitempty"`
	Status                    string     `json:"status"`
	ChildOrderID              uint64     `json:"child_order_id"`
	ChildOrderNo              string     `json:"child_order_no"`
	Remark                    string     `json:"remark"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
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

type JoinProcurementRequest struct {
	Quantity               float64 `json:"quantity" binding:"required,gt=0"`
	ContactName            string  `json:"contact_name" binding:"required"`
	ContactPhone           string  `json:"contact_phone" binding:"required"`
	DeliveryAddressSource  string  `json:"delivery_address_source"`
	DestWarehouseServiceID uint64  `json:"dest_warehouse_service_id"`
	DeliveryAddress        string  `json:"delivery_address"`
	DeliveryMethod         string  `json:"delivery_method"`
	PaymentMethod          string  `json:"payment_method"`
	DeliveryNote           string  `json:"delivery_note"`
	AgreementSigned        bool    `json:"agreement_signed" binding:"required"`
	Remark                 string  `json:"remark"`
}

type RespondProcurementRequest struct {
	Accepted         bool    `json:"accepted"`
	GroupPrice       float64 `json:"group_price"`
	MinGroupQuantity float64 `json:"min_group_quantity"`
	MaxGroupQuantity float64 `json:"max_group_quantity"`
	EndedAt          string  `json:"ended_at"`
	DeliveryDays     int     `json:"delivery_days"`
	Remark           string  `json:"remark"`
}

type CreateProcurementRequestRequest struct {
	Source                 string  `json:"source"`
	SourceProductID        uint64  `json:"source_product_id"`
	AIProcurementID        uint64  `json:"ai_procurement_id"`
	Category               string  `json:"category" binding:"required"`
	SubCategory            string  `json:"sub_category"`
	ProductName            string  `json:"product_name" binding:"required"`
	Material               string  `json:"material"`
	Spec                   string  `json:"spec"`
	Quantity               float64 `json:"quantity" binding:"required,gt=0"`
	Unit                   string  `json:"unit"`
	TargetPrice            float64 `json:"target_price"`
	ContactName            string  `json:"contact_name" binding:"required"`
	ContactPhone           string  `json:"contact_phone" binding:"required"`
	DeliveryAddressSource  string  `json:"delivery_address_source"`
	DestWarehouseServiceID uint64  `json:"dest_warehouse_service_id"`
	DeliveryAddress        string  `json:"delivery_address"`
	DeliveryMethod         string  `json:"delivery_method"`
	PaymentMethod          string  `json:"payment_method"`
	DeliveryNote           string  `json:"delivery_note"`
	Remark                 string  `json:"remark"`
}

type ProcurementRequestItem struct {
	ID                        uint64    `json:"id"`
	RequestNo                 string    `json:"request_no"`
	Source                    string    `json:"source"`
	SourceProductID           uint64    `json:"source_product_id"`
	AIProcurementID           uint64    `json:"ai_procurement_id"`
	Category                  string    `json:"category"`
	SubCategory               string    `json:"sub_category"`
	ProductName               string    `json:"product_name"`
	Material                  string    `json:"material"`
	Spec                      string    `json:"spec"`
	Quantity                  float64   `json:"quantity"`
	Unit                      string    `json:"unit"`
	TargetPrice               float64   `json:"target_price"`
	ContactName               string    `json:"contact_name"`
	ContactPhone              string    `json:"contact_phone"`
	DeliveryAddressSource     string    `json:"delivery_address_source"`
	DestWarehouseServiceID    uint64    `json:"dest_warehouse_service_id"`
	DestWarehouseEnterpriseID uint64    `json:"dest_warehouse_enterprise_id"`
	DestWarehouseName         string    `json:"dest_warehouse_name"`
	DestWarehouseProvince     string    `json:"dest_warehouse_province"`
	DestWarehouseCity         string    `json:"dest_warehouse_city"`
	DestWarehouseDistrict     string    `json:"dest_warehouse_district"`
	DestWarehouseAddress      string    `json:"dest_warehouse_address"`
	DeliveryAddress           string    `json:"delivery_address"`
	DeliveryMethod            string    `json:"delivery_method"`
	PaymentMethod             string    `json:"payment_method"`
	DeliveryNote              string    `json:"delivery_note"`
	BuyerEnterpriseID         uint64    `json:"buyer_enterprise_id"`
	BuyerUserID               uint64    `json:"buyer_user_id"`
	BuyerName                 string    `json:"buyer_name"`
	Status                    string    `json:"status"`
	ActivityID                uint64    `json:"activity_id"`
	Remark                    string    `json:"remark"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type ProcurementRequestListResponse struct {
	List     []ProcurementRequestItem `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

type ProcurementInvitationItem struct {
	ID                uint64                   `json:"id"`
	InvitationNo      string                   `json:"invitation_no"`
	ActivityID        uint64                   `json:"activity_id"`
	RequestID         uint64                   `json:"request_id"`
	BuyerEnterpriseID uint64                   `json:"buyer_enterprise_id"`
	BuyerUserID       uint64                   `json:"buyer_user_id"`
	BuyerName         string                   `json:"buyer_name"`
	Status            string                   `json:"status"`
	ParticipantID     uint64                   `json:"participant_id"`
	Remark            string                   `json:"remark"`
	ExpiresAt         *time.Time               `json:"expires_at,omitempty"`
	AcceptedAt        *time.Time               `json:"accepted_at,omitempty"`
	DeclinedAt        *time.Time               `json:"declined_at,omitempty"`
	Activity          *ProcurementActivityItem `json:"activity,omitempty"`
	Request           *ProcurementRequestItem  `json:"request,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

type ProcurementInvitationListResponse struct {
	List     []ProcurementInvitationItem `json:"list"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"page_size"`
}
