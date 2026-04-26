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

type ProcurementListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
}

type SubmitReviewRequest struct {
	Remark string `json:"remark"`
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
