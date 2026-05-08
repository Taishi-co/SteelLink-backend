package dto

import (
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
)

type ParseRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type CreateProcurementRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type ConfirmInstructionRequest struct {
	Prompt string                   `json:"prompt"`
	Parsed entity.ParsedInstruction `json:"parsed" binding:"required"`
}

type MatchProductsRequest struct {
	Parsed entity.ParsedInstruction `json:"parsed" binding:"required"`
}

type MatchLogisticsWarehouseRequest struct {
	Parsed           entity.ParsedInstruction `json:"parsed" binding:"required"`
	SelectedProducts entity.MatchedProducts   `json:"selected_products" binding:"required"`
}

type GenerateOrderRequest struct {
	Prompt           string                   `json:"prompt"`
	Parsed           entity.ParsedInstruction `json:"parsed" binding:"required"`
	SelectedProduct  entity.MatchedProduct    `json:"selected_product"`
	SelectedProducts entity.MatchedProducts   `json:"selected_products"`
	OrderForm        OrderForm                `json:"order_form" binding:"required"`
}

type OrderForm struct {
	Quantity        float64 `json:"quantity"`
	ContactName     string  `json:"contact_name" binding:"required"`
	ContactPhone    string  `json:"contact_phone" binding:"required"`
	WarehouseName   string  `json:"warehouse_name" binding:"required"`
	DeliveryAddress string  `json:"delivery_address" binding:"required"`
	DeliveryMethod  string  `json:"delivery_method" binding:"required"`
	PaymentMethod   string  `json:"payment_method" binding:"required"`
	DeliveryNote    string  `json:"delivery_note"`
	Remark          string  `json:"remark"`
}

type MatchFinanceRequest struct {
	FinancingRequired bool `json:"financing_required"`
}

type ProcurementListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
}

type SubmitReviewRequest struct {
	Remark      string `json:"remark"`
	SaveHistory *bool  `json:"save_history"`
}

type VectorSyncRequest struct {
	Target       string `json:"target"`
	BatchSize    int    `json:"batch_size"`
	Limit        int    `json:"limit"`
	CreateSchema *bool  `json:"create_schema"`
}

type VectorSyncResponse struct {
	Target              string    `json:"target"`
	Synced              int       `json:"synced"`
	LastID              uint64    `json:"last_id"`
	LastSourceUpdatedAt time.Time `json:"last_source_updated_at,omitempty"`
	ReservedTargets     []string  `json:"reserved_targets,omitempty"`
	Message             string    `json:"message"`
}

type StageInfo struct {
	Code        entity.WorkflowStage `json:"code"`
	Name        string               `json:"name"`
	Status      string               `json:"status"`
	Description string               `json:"description"`
}

type ParseResponse struct {
	Parsed          entity.ParsedInstruction `json:"parsed"`
	MatchedProducts entity.MatchedProducts   `json:"matched_products"`
	OrderDraft      entity.OrderDraft        `json:"order_draft"`
	FinancePlans    entity.FinancePlans      `json:"finance_plans"`
	AuditItems      entity.AuditItems        `json:"audit_items"`
	Stages          []StageInfo              `json:"stages"`
}

type MatchProductsResponse struct {
	MatchedProducts entity.MatchedProducts `json:"matched_products"`
	Stages          []StageInfo            `json:"stages"`
}

type MatchLogisticsWarehouseResponse struct {
	MatchedProducts entity.MatchedProducts `json:"matched_products"`
	Stages          []StageInfo            `json:"stages"`
}

type GenerateOrderResponse struct {
	Procurement *ProcurementInfo  `json:"procurement"`
	OrderDraft  entity.OrderDraft `json:"order_draft"`
}

type ProcurementInfo struct {
	ID                uint64                   `json:"id"`
	ProcurementNo     string                   `json:"procurement_no"`
	Prompt            string                   `json:"prompt"`
	ProductName       string                   `json:"product_name"`
	Category          string                   `json:"category"`
	Quantity          float64                  `json:"quantity"`
	Unit              string                   `json:"unit"`
	Usage             string                   `json:"usage"`
	FinancingRequired bool                     `json:"financing_required"`
	CurrentStage      entity.WorkflowStage     `json:"current_stage"`
	Status            entity.ProcurementStatus `json:"status"`
	ParsedInstruction entity.ParsedInstruction `json:"parsed_instruction"`
	MatchedProducts   entity.MatchedProducts   `json:"matched_products"`
	OrderDraft        entity.OrderDraft        `json:"order_draft"`
	FinancePlans      entity.FinancePlans      `json:"finance_plans"`
	AuditItems        entity.AuditItems        `json:"audit_items"`
	Stages            []StageInfo              `json:"stages"`
	HistorySaved      bool                     `json:"history_saved"`
	EnterpriseID      uint64                   `json:"enterprise_id"`
	CreatedBy         uint64                   `json:"created_by"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	ReviewSubmittedAt *time.Time               `json:"review_submitted_at"`
}

type ProcurementListResponse struct {
	List     []*ProcurementInfo `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}
