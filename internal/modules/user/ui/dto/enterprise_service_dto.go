package dto

import "time"

type ApplicationDocumentRequest struct {
	DocumentType string `json:"document_type" binding:"required"`
	FileURL      string `json:"file_url" binding:"required,url"`
}

type ApplyEnterpriseServiceRequest struct {
	Service   string                       `json:"service" binding:"required,oneof=sales procurement logistics warehouse finance"`
	Documents []ApplicationDocumentRequest `json:"documents" binding:"required,min=1,dive"`
}

type AuditEnterpriseServiceRequest struct {
	Status int    `json:"status" binding:"required,oneof=2 3"`
	Reason string `json:"reason"`
}

type UpdateEnterpriseServicePermissionRequest struct {
	Status int    `json:"status" binding:"required,oneof=2 3 4"`
	Reason string `json:"reason"`
}

type EnterpriseServicePermissionInfo struct {
	ID             uint64     `json:"id"`
	EnterpriseID   uint64     `json:"enterprise_id"`
	EnterpriseCode string     `json:"enterprise_code"`
	Service        string     `json:"service"`
	SupplyRole     string     `json:"supply_role"`
	Status         int        `json:"status"`
	AuditReason    string     `json:"audit_reason"`
	ApprovedAt     *time.Time `json:"approved_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type EnterpriseApplicationDocumentInfo struct {
	ID            uint64    `json:"id"`
	ApplicationID uint64    `json:"application_id"`
	DocumentType  string    `json:"document_type"`
	FileURL       string    `json:"file_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type EnterpriseServiceApplicationInfo struct {
	ID             uint64                               `json:"id"`
	EnterpriseID   uint64                               `json:"enterprise_id"`
	EnterpriseCode string                               `json:"enterprise_code"`
	Service        string                               `json:"service"`
	SupplyRole     string                               `json:"supply_role"`
	Status         int                                  `json:"status"`
	AuditReason    string                               `json:"audit_reason"`
	SubmittedBy    uint64                               `json:"submitted_by"`
	AuditedBy      uint64                               `json:"audited_by"`
	AuditedAt      *time.Time                           `json:"audited_at"`
	Documents      []*EnterpriseApplicationDocumentInfo `json:"documents"`
	CreatedAt      time.Time                            `json:"created_at"`
	UpdatedAt      time.Time                            `json:"updated_at"`
}

type EnterpriseServiceOverview struct {
	Permissions       []*EnterpriseServicePermissionInfo `json:"permissions"`
	RequiredDocuments map[string][]string                `json:"required_documents"`
}

type EnterpriseServiceApplicationListRequest struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	PageSize       int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	EnterpriseCode string `form:"enterprise_code"`
	Service        string `form:"service" binding:"omitempty,oneof=sales procurement logistics warehouse finance"`
	Status         *int   `form:"status"`
}

type EnterpriseServiceApplicationListResponse struct {
	List     []*EnterpriseServiceApplicationInfo `json:"list"`
	Total    int64                               `json:"total"`
	Page     int                                 `json:"page"`
	PageSize int                                 `json:"page_size"`
}
