package dto

import "time"

type CreateOrderRequest struct {
	ProductID              uint64                   `json:"product_id"`
	Quantity               float64                  `json:"quantity"`
	Items                  []CreateOrderItemRequest `json:"items"`
	CartItemIDs            []uint64                 `json:"cart_item_ids"`
	Source                 string                   `json:"source"`
	RestoreCartOnCancel    bool                     `json:"restore_cart_on_cancel"`
	ContactName            string                   `json:"contact_name"`
	ContactPhone           string                   `json:"contact_phone"`
	WarehouseName          string                   `json:"warehouse_name"`
	DestWarehouseServiceID uint64                   `json:"dest_warehouse_service_id"`
	DeliveryAddress        string                   `json:"delivery_address"`
	DeliveryMethod         string                   `json:"delivery_method"`
	PaymentMethod          string                   `json:"payment_method"`
	DeliveryNote           string                   `json:"delivery_note"`
	Remark                 string                   `json:"remark"`
}

type CreateOrderItemRequest struct {
	ProductID uint64  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type WarehousePreDataResponse struct {
	WarehouseID           uint64 `json:"warehouse_id"`
	WarehouseServiceID    uint64 `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64 `json:"warehouse_enterprise_id"`
	WarehouseName         string `json:"warehouse_name"`
	WarehouseCode         string `json:"warehouse_code"`
	WarehouseProvince     string `json:"warehouse_province"`
	WarehouseCity         string `json:"warehouse_city"`
	WarehouseDistrict     string `json:"warehouse_district"`
	WarehouseAddress      string `json:"warehouse_address"`
	ProductID             uint64 `json:"product_id"`
	ProductName           string `json:"product_name"`
	SellerEnterpriseID    uint64 `json:"seller_enterprise_id"`
	SellerName            string `json:"seller_name"`
	CanWarehouseConfirm   bool   `json:"can_warehouse_confirm"`
}

type DraftOrderPreviewRequest struct {
	Items []CreateOrderItemRequest `json:"items" binding:"required,dive"`
}

type ProductPoolItem struct {
	ID                  uint64  `json:"id"`
	ProductNo           string  `json:"product_no"`
	Name                string  `json:"name"`
	Category            string  `json:"category"`
	SubCategory         string  `json:"sub_category"`
	SteelMill           string  `json:"steel_mill"`
	UnitPrice           float64 `json:"unit_price"`
	AvailableQuantity   float64 `json:"available_quantity"`
	HasWarehouseBinding bool    `json:"has_warehouse_binding"`
	WarehouseName       string  `json:"warehouse_name,omitempty"`
	WarehouseAddress    string  `json:"warehouse_address,omitempty"`
	SellerEnterpriseID  uint64  `json:"seller_enterprise_id"`
	SellerName          string  `json:"seller_name"`
}

type UploadPaymentVoucherRequest struct {
	VoucherURL    string   `json:"voucher_url" binding:"required"`
	TransactionNo string   `json:"transaction_no"`
	PayerAccount  string   `json:"payer_account"`
	PayeeAccount  string   `json:"payee_account"`
	CapturedAt    string   `json:"captured_at"`
	GPSLatitude   *float64 `json:"gps_latitude"`
	GPSLongitude  *float64 `json:"gps_longitude"`
}

type AttachContractRequest struct {
	ContractNo   string `json:"contract_no"`
	ContractURL  string `json:"contract_url" binding:"required"`
	ContractHash string `json:"contract_hash"`
}

type WarehouseConfirmRequest struct {
	Confirmed     bool   `json:"confirmed"`
	ConditionNote string `json:"condition_note"`
	VoucherURL    string `json:"voucher_url"`
}

type WarehouseReselectRequest struct {
	WarehouseServiceID uint64 `json:"warehouse_service_id" binding:"required"`
	ConfirmNote        string `json:"confirm_note"`
}

type DestWarehouseSelectRequest struct {
	DestWarehouseServiceID uint64 `json:"dest_warehouse_service_id" binding:"required"`
	DestConfirmNote        string `json:"dest_confirm_note"`
}

type SellerConfirmRequest struct {
	TermsNote string `json:"terms_note"`
}

type BuyerConfirmRequest struct {
	ConfirmNote string `json:"confirm_note"`
}

type DirectPaymentRequest struct {
	VoucherURL    string   `json:"voucher_url"`
	TransactionNo string   `json:"transaction_no"`
	PayerAccount  string   `json:"payer_account"`
	PayeeAccount  string   `json:"payee_account"`
	CapturedAt    string   `json:"captured_at"`
	GPSLatitude   *float64 `json:"gps_latitude"`
	GPSLongitude  *float64 `json:"gps_longitude"`
}

type CreateFinancingRequest struct {
	FinanceServiceID uint64  `json:"finance_service_id" binding:"required"`
	RequestedAmount  float64 `json:"requested_amount"`
}

type RiskResultRequest struct {
	Approved       bool    `json:"approved"`
	ApprovedAmount float64 `json:"approved_amount"`
	TermDays       int     `json:"term_days"`
	AnnualRate     float64 `json:"annual_rate"`
	ServiceFee     float64 `json:"service_fee"`
	RiskRemark     string  `json:"risk_remark"`
}

type BuyerFinancingConfirmRequest struct {
	Confirmed bool   `json:"confirmed"`
	Remark    string `json:"remark"`
}

type DebtContractRequest struct {
	DebtContractNo   string `json:"debt_contract_no"`
	DebtContractURL  string `json:"debt_contract_url" binding:"required"`
	DebtContractHash string `json:"debt_contract_hash"`
}

type BuyerDebtSignRequest struct {
	FeeVoucherURL string `json:"fee_voucher_url" binding:"required"`
	Remark        string `json:"remark"`
}

type FinancingDisburseRequest struct {
	DisbursementVoucher string  `json:"disbursement_voucher" binding:"required"`
	FundingCommission   float64 `json:"funding_commission"`
	CommissionVoucher   string  `json:"commission_voucher" binding:"required"`
	Remark              string  `json:"remark"`
}

type PledgeRecordRequest struct {
	PledgeVoucherURL string `json:"pledge_voucher_url" binding:"required"`
	ReleaseCondition string `json:"release_condition" binding:"required"`
	PledgeRemark     string `json:"pledge_remark"`
}

type PledgeReleaseRequest struct {
	Reason     string `json:"reason" binding:"required"`
	VoucherURL string `json:"voucher_url"`
}

type CreateLogisticsBiddingRequest struct {
	PickupAddress      string `json:"pickup_address"`
	DestinationAddress string `json:"destination_address"`
	Requirement        string `json:"requirement"`
	FreightPayer       string `json:"freight_payer" binding:"omitempty,oneof=buyer seller"`
	NoCapacityReason   string `json:"no_capacity_reason"`
}

type CreateLogisticsQuoteRequest struct {
	VehicleInfo         string  `json:"vehicle_info" binding:"required"`
	DriverInfo          string  `json:"driver_info" binding:"required"`
	Amount              float64 `json:"amount" binding:"required,gt=0"`
	EstimatedPickupAt   string  `json:"estimated_pickup_at"`
	EstimatedDeliveryAt string  `json:"estimated_delivery_at"`
	Remark              string  `json:"remark"`
}

type ConfirmLogisticsRequest struct {
	QuoteID uint64 `json:"quote_id" binding:"required"`
}

type ShipmentProgressRequest struct {
	VoucherURL string `json:"voucher_url"`
	Remark     string `json:"remark"`
}

type SignLogisticsContractRequest struct {
	ContractURL string `json:"contract_url"`
}

type PayLogisticsSettlementRequest struct {
	VoucherURL string `json:"voucher_url" binding:"required"`
}

type SetDeliveryInfoRequest struct {
	DriverName         string `json:"driver_name" binding:"required"`
	DriverPhone        string `json:"driver_phone" binding:"required"`
	DriverIDCard       string `json:"driver_id_card"`
	PlateNumber        string `json:"plate_number" binding:"required"`
	WarehouseID        uint64 `json:"warehouse_id"`
	WarehouseServiceID uint64 `json:"warehouse_service_id"`
	WarehouseName      string `json:"warehouse_name" binding:"required"`
	Address            string `json:"address" binding:"required"`
	EstimatedTime      string `json:"estimated_time"`
}

type ConfirmReceiptRequest struct {
	ReceiptVoucher string   `json:"receipt_voucher" binding:"required"`
	CapturedAt     string   `json:"captured_at"`
	GPSLatitude    *float64 `json:"gps_latitude"`
	GPSLongitude   *float64 `json:"gps_longitude"`
}

type VerifyPickupCodeRequest struct {
	PickupCode string `json:"pickup_code" binding:"required"`
}

type OrderItemInfo struct {
	ProductID   uint64  `json:"product_id"`
	ProductNo   string  `json:"product_no"`
	ProductName string  `json:"product_name"`
	SteelMill   string  `json:"steel_mill"`
	Category    string  `json:"category"`
	BundleNo    string  `json:"bundle_no"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalAmount float64 `json:"total_amount"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
}

type OrderInfo struct {
	ID               uint64                 `json:"id"`
	OrderNo          string                 `json:"order_no"`
	BuyerID          uint64                 `json:"buyer_id"`
	BuyerName        string                 `json:"buyer_name"`
	SellerID         uint64                 `json:"seller_id"`
	SellerName       string                 `json:"seller_name"`
	Items            []OrderItemInfo        `json:"items"`
	TotalQuantity    float64                `json:"total_quantity"`
	TotalAmount      float64                `json:"total_amount"`
	TaxTotalAmount   float64                `json:"tax_total_amount"`
	Currency         string                 `json:"currency"`
	Status           int                    `json:"status"`
	Expired          bool                   `json:"expired"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	ContractInfo     *ContractInfo          `json:"contract_info,omitempty"`
	PaymentInfo      *PaymentInfo           `json:"payment_info,omitempty"`
	DeliveryInfo     *DeliveryInfo          `json:"delivery_info,omitempty"`
	ReceiptVoucher   string                 `json:"receipt_voucher,omitempty"`
	ReceiptInfo      *VoucherEvidence       `json:"receipt_info,omitempty"`
	ReceiptAt        *time.Time             `json:"receipt_at,omitempty"`
	Timeline         []TimelineItem         `json:"timeline"`
	Events           []OrderEventInfo       `json:"events,omitempty"`
	Remark           map[string]interface{} `json:"remark,omitempty"`
	FinancingInfo    *FinancingRequestInfo  `json:"financing_info,omitempty"`
	LogisticsBidding *LogisticsBiddingInfo  `json:"logistics_bidding,omitempty"`
	LogisticsQuotes  []LogisticsQuoteInfo   `json:"logistics_quotes,omitempty"`
	ShipmentTask         *ShipmentTaskInfo         `json:"shipment_task,omitempty"`
	LogisticsContractDTO *LogisticsContractInfo    `json:"logistics_contract,omitempty"`
	LogisticsSettlement  *LogisticsSettlementInfo  `json:"logistics_settlement,omitempty"`
	PledgeEvents         []PledgeEventInfo         `json:"pledge_events,omitempty"`
}

type OrderListResponse struct {
	List     []*OrderInfo `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type FinancingRequestInfo struct {
	ID                  uint64     `json:"id"`
	OrderID             uint64     `json:"order_id"`
	OrderNo             string     `json:"order_no"`
	BuyerEnterpriseID   uint64     `json:"buyer_enterprise_id"`
	SellerEnterpriseID  uint64     `json:"seller_enterprise_id"`
	FinanceServiceID    uint64     `json:"finance_service_id"`
	FundingEnterpriseID uint64     `json:"funding_enterprise_id"`
	RequestedAmount     float64    `json:"requested_amount"`
	ApprovedAmount      float64    `json:"approved_amount"`
	TermDays            int        `json:"term_days"`
	AnnualRate          float64    `json:"annual_rate"`
	ServiceFee          float64    `json:"service_fee"`
	Status              string     `json:"status"`
	RiskResult          string     `json:"risk_result"`
	RiskRemark          string     `json:"risk_remark"`
	BuyerConfirmedAt    *time.Time `json:"buyer_confirmed_at,omitempty"`
	DebtContractNo      string     `json:"debt_contract_no"`
	DebtContractURL     string     `json:"debt_contract_url"`
	DebtContractHash    string     `json:"debt_contract_hash"`
	BuyerSignedAt       *time.Time `json:"buyer_signed_at,omitempty"`
	FeeVoucherURL       string     `json:"fee_voucher_url"`
	FeePaidAt           *time.Time `json:"fee_paid_at,omitempty"`
	FundingCommission   float64    `json:"funding_commission"`
	CommissionVoucher   string     `json:"commission_voucher"`
	CommissionPaidAt    *time.Time `json:"commission_paid_at,omitempty"`
	DisbursementVoucher string     `json:"disbursement_voucher"`
	DisbursedAt         *time.Time `json:"disbursed_at,omitempty"`
	PledgeVoucherURL    string     `json:"pledge_voucher_url"`
	PledgeRemark        string     `json:"pledge_remark"`
	PledgedAt           *time.Time `json:"pledged_at,omitempty"`
	ReleaseCondition    string     `json:"release_condition"`
	ReleaseReason       string     `json:"release_reason"`
	ReleasedAt          *time.Time `json:"released_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type ContractInfo struct {
	ContractNo   string     `json:"contract_no,omitempty"`
	ContractURL  string     `json:"contract_url,omitempty"`
	ContractHash string     `json:"contract_hash,omitempty"`
	SignStatus   string     `json:"sign_status"`
	BuyerSigned  bool       `json:"buyer_signed"`
	SellerSigned bool       `json:"seller_signed"`
	SignedAt     *time.Time `json:"signed_at,omitempty"`
}

type PaymentInfo struct {
	PaymentMethod string    `json:"payment_method"`
	PaymentTime   time.Time `json:"payment_time"`
	TransactionNo string    `json:"transaction_no"`
	PayeeAccount  string    `json:"payee_account"`
	PayerAccount  string    `json:"payer_account"`
	VoucherURL    string    `json:"voucher_url"`
	UploadedBy    uint64    `json:"uploaded_by"`
	UploadedAt    time.Time `json:"uploaded_at"`
	CapturedAt    string    `json:"captured_at,omitempty"`
	GPSLatitude   *float64  `json:"gps_latitude,omitempty"`
	GPSLongitude  *float64  `json:"gps_longitude,omitempty"`
}

type VoucherEvidence struct {
	VoucherURL   string    `json:"voucher_url"`
	UploadedBy   uint64    `json:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at"`
	CapturedAt   string    `json:"captured_at,omitempty"`
	GPSLatitude  *float64  `json:"gps_latitude,omitempty"`
	GPSLongitude *float64  `json:"gps_longitude,omitempty"`
}

type DeliveryInfo struct {
	DriverName            string `json:"driver_name"`
	DriverPhone           string `json:"driver_phone"`
	DriverIDCard          string `json:"driver_id_card"`
	PlateNumber           string `json:"plate_number"`
	PickupCode            string `json:"pickup_code"`
	WarehouseID           uint64 `json:"warehouse_id"`
	WarehouseServiceID    uint64 `json:"warehouse_service_id"`
	WarehouseEnterpriseID uint64 `json:"warehouse_enterprise_id"`
	WarehouseName         string `json:"warehouse_name"`
	WarehouseCode         string `json:"warehouse_code"`
	WarehouseProvince     string `json:"warehouse_province"`
	WarehouseCity         string `json:"warehouse_city"`
	WarehouseDistrict     string `json:"warehouse_district"`
	WarehouseAddress      string `json:"warehouse_address"`
	Address               string `json:"address"`
	EstimatedTime         string `json:"estimated_time"`
}

type TimelineItem struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Status      string     `json:"status"`
	Time        *time.Time `json:"time,omitempty"`
	Description string     `json:"description,omitempty"`
}

type OrderEventInfo struct {
	ID                uint64                 `json:"id"`
	EventType         string                 `json:"event_type"`
	ActorUserID       uint64                 `json:"actor_user_id"`
	ActorEnterpriseID uint64                 `json:"actor_enterprise_id"`
	ActorService      string                 `json:"actor_service"`
	FromStatus        int                    `json:"from_status"`
	ToStatus          int                    `json:"to_status"`
	Payload           map[string]interface{} `json:"payload,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

type LogisticsBiddingInfo struct {
	ID                     uint64    `json:"id"`
	OrderID                uint64    `json:"order_id"`
	OrderNo                string    `json:"order_no"`
	Status                 string    `json:"status"`
	FreightPayer           string    `json:"freight_payer"`
	InitiatorEnterpriseID  uint64    `json:"initiator_enterprise_id"`
	InitiatorRole          string    `json:"initiator_role"`
	ResponderEnterpriseID  uint64    `json:"responder_enterprise_id"`
	ResponderConfirmed     bool      `json:"responder_confirmed"`
	SelectedQuoteID        uint64    `json:"selected_quote_id"`
	BuyerConfirmedQuoteID  uint64    `json:"buyer_confirmed_quote_id"`
	SellerConfirmedQuoteID uint64    `json:"seller_confirmed_quote_id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type LogisticsQuoteInfo struct {
	ID                    uint64    `json:"id"`
	BiddingID             uint64    `json:"bidding_id"`
	OrderID               uint64    `json:"order_id"`
	LogisticsEnterpriseID uint64    `json:"logistics_enterprise_id"`
	LogisticsName         string    `json:"logistics_name"`
	VehicleInfo           string    `json:"vehicle_info"`
	DriverInfo            string    `json:"driver_info"`
	Amount                float64   `json:"amount"`
	EstimatedPickupAt     string    `json:"estimated_pickup_at"`
	EstimatedDeliveryAt   string    `json:"estimated_delivery_at"`
	Remark                string    `json:"remark"`
	CreatedAt             time.Time `json:"created_at"`
}

type ShipmentTaskInfo struct {
	ID                    uint64     `json:"id"`
	OrderID               uint64     `json:"order_id"`
	OrderNo               string     `json:"order_no"`
	LogisticsEnterpriseID uint64     `json:"logistics_enterprise_id"`
	LogisticsQuoteID      uint64     `json:"logistics_quote_id"`
	Status                string     `json:"status"`
	PickupVoucher         string     `json:"pickup_voucher"`
	DeliveryVoucher       string     `json:"delivery_voucher"`
	SignatureVoucher       string     `json:"signature_voucher"`
	PickedUpAt            *time.Time `json:"picked_up_at,omitempty"`
	DeliveredAt           *time.Time `json:"delivered_at,omitempty"`
	SignedAt              *time.Time `json:"signed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type LogisticsContractInfo struct {
	ID                    uint64     `json:"id"`
	ContractNo            string     `json:"contract_no"`
	OrderID               uint64     `json:"order_id"`
	BiddingID             uint64     `json:"bidding_id"`
	QuoteID               uint64     `json:"quote_id"`
	BuyerEnterpriseID     uint64     `json:"buyer_enterprise_id"`
	LogisticsEnterpriseID uint64     `json:"logistics_enterprise_id"`
	SellerEnterpriseID    uint64     `json:"seller_enterprise_id"`
	ContractURL           string     `json:"contract_url"`
	FreightAmount         float64    `json:"freight_amount"`
	CommissionRate        float64    `json:"commission_rate"`
	PickupWarehouseID     uint64     `json:"pickup_warehouse_id"`
	PickupWarehouseName   string     `json:"pickup_warehouse_name"`
	DestinationAddress    string     `json:"destination_address"`
	IsFinancingOrder      bool       `json:"is_financing_order"`
	Status                string     `json:"status"`
	BuyerSignedAt         *time.Time `json:"buyer_signed_at,omitempty"`
	LogisticsSignedAt     *time.Time `json:"logistics_signed_at,omitempty"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

type LogisticsSettlementInfo struct {
	ID                   uint64     `json:"id"`
	SettlementNo         string     `json:"settlement_no"`
	LogisticsContractID  uint64     `json:"logistics_contract_id"`
	OrderID              uint64     `json:"order_id"`
	PayerEnterpriseID    uint64     `json:"payer_enterprise_id"`
	PayeeEnterpriseID    uint64     `json:"payee_enterprise_id"`
	FreightAmount        float64    `json:"freight_amount"`
	PlatformCommission   float64    `json:"platform_commission"`
	NetPayable           float64    `json:"net_payable"`
	PaymentVoucherURL    string     `json:"payment_voucher_url"`
	CommissionVoucherURL string     `json:"commission_voucher_url"`
	Status               string     `json:"status"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CommissionPaidAt     *time.Time `json:"commission_paid_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

type PledgeEventInfo struct {
	ID                 uint64    `json:"id"`
	FinancingRequestID uint64    `json:"financing_request_id"`
	OrderID            uint64    `json:"order_id"`
	OrderNo            string    `json:"order_no"`
	EventType          string    `json:"event_type"`
	Condition          string    `json:"condition"`
	Reason             string    `json:"reason"`
	VoucherURL         string    `json:"voucher_url"`
	ActorUserID        uint64    `json:"actor_user_id"`
	ActorEnterpriseID  uint64    `json:"actor_enterprise_id"`
	CreatedAt          time.Time `json:"created_at"`
}

type ReportBreachRequest struct {
	BreachParty  string   `json:"breach_party" binding:"required,oneof=buyer seller"`
	BreachType   string   `json:"breach_type" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	EvidenceURLs []string `json:"evidence_urls"`
}

type BreachRecordInfo struct {
	ID                   uint64     `json:"id"`
	OrderID              uint64     `json:"order_id"`
	OrderNo              string     `json:"order_no"`
	ReporterEnterpriseID uint64     `json:"reporter_enterprise_id"`
	BreachParty          string     `json:"breach_party"`
	BreachEnterpriseID   uint64     `json:"breach_enterprise_id"`
	BreachType           string     `json:"breach_type"`
	Description          string     `json:"description"`
	EvidenceURLs         []string   `json:"evidence_urls"`
	Status               string     `json:"status"`
	MediationResult      string     `json:"mediation_result,omitempty"`
	MediatedBy           uint64     `json:"mediated_by,omitempty"`
	MediatedAt           *time.Time `json:"mediated_at,omitempty"`
	ResolutionResult     string     `json:"resolution_result,omitempty"`
	ResolvedBy           uint64     `json:"resolved_by,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type MediateBreachRequest struct {
	MediationResult string `json:"mediation_result" binding:"required"`
}

type ResolveBreachRequest struct {
	ResolutionResult string `json:"resolution_result" binding:"required"`
}

type FreezeOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type UnfreezeOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CreateLitigationRequest struct {
	CaseNo       string   `json:"case_no"`
	Court        string   `json:"court"`
	ClaimAmount  float64  `json:"claim_amount"`
	Description  string   `json:"description" binding:"required"`
	EvidenceURLs []string `json:"evidence_urls"`
}

type UpdateLitigationStatusRequest struct {
	Status         string `json:"status" binding:"required,oneof=accepted in_progress judged closed"`
	CaseNo         string `json:"case_no"`
	Court          string `json:"court"`
	JudgmentResult string `json:"judgment_result"`
	JudgmentURL    string `json:"judgment_url"`
}

type LitigationRecordInfo struct {
	ID                    uint64     `json:"id"`
	BreachRecordID        uint64     `json:"breach_record_id"`
	OrderID               uint64     `json:"order_id"`
	OrderNo               string     `json:"order_no"`
	PlaintiffEnterpriseID uint64     `json:"plaintiff_enterprise_id"`
	DefendantEnterpriseID uint64     `json:"defendant_enterprise_id"`
	CaseNo                string     `json:"case_no"`
	Court                 string     `json:"court"`
	ClaimAmount           float64    `json:"claim_amount"`
	Description           string     `json:"description"`
	EvidenceURLs          []string   `json:"evidence_urls"`
	Status                string     `json:"status"`
	JudgmentResult        string     `json:"judgment_result,omitempty"`
	JudgmentURL           string     `json:"judgment_url,omitempty"`
	JudgedAt              *time.Time `json:"judged_at,omitempty"`
	ClosedAt              *time.Time `json:"closed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
