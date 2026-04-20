package dto

import "time"

// CreateEnterpriseRequest 创建企业请求
type CreateEnterpriseRequest struct {
	Name             string `json:"name" binding:"required,max=200"`
	CreditCode       string `json:"credit_code" binding:"required,max=50"`
	EnterpriseRole   string `json:"enterprise_role" binding:"required,oneof=buyer seller"`
	ContactName      string `json:"contact_name" binding:"required"`
	ContactPhone     string `json:"contact_phone" binding:"required,len=11"`
	Address          string `json:"address"`
}

// EnterpriseInfo 企业信息响应
type EnterpriseInfo struct {
	ID               uint64            `json:"id"`
	EnterpriseCode   string            `json:"enterprise_code"`
	Name             string            `json:"name"`
	CreditCode       string            `json:"credit_code"`
	EnterpriseRole   string            `json:"enterprise_role"`
	Status           int               `json:"status"`
	VerificationInfo *VerificationInfo `json:"verification_info,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// VerificationInfo 认证信息
type VerificationInfo struct {
	LegalPerson      string `json:"legal_person"`
	LegalIdCard      string `json:"legal_id_card"`
	BusinessLicense  string `json:"business_license"`
	BankAccount      string `json:"bank_account"`
	BankName         string `json:"bank_name"`
	ContactName      string `json:"contact_name"`
	ContactPhone     string `json:"contact_phone"`
	Address          string `json:"address"`
}

// UpdateEnterpriseRequest 更新企业信息请求
type UpdateEnterpriseRequest struct {
	Name           string `json:"name" binding:"omitempty,max=200"`
	ContactName    string `json:"contact_name"`
	ContactPhone   string `json:"contact_phone" binding:"omitempty,len=11"`
	Address        string `json:"address"`
}

// VerifyEnterpriseRequest 企业认证请求
type VerifyEnterpriseRequest struct {
	LegalPerson     string `json:"legal_person" binding:"required"`
	LegalIdCard     string `json:"legal_id_card" binding:"required"`
	BusinessLicense string `json:"business_license" binding:"required"`
	BankAccount     string `json:"bank_account" binding:"required"`
	BankName        string `json:"bank_name" binding:"required"`
	ContactName     string `json:"contact_name" binding:"required"`
	ContactPhone    string `json:"contact_phone" binding:"required,mobile"`
	Address         string `json:"address" binding:"required"`
}

// EnterpriseListRequest 企业列表请求
type EnterpriseListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name"`
	Status   *int   `form:"status"`
	Keyword  string `form:"keyword"`
}

// EnterpriseListResponse 企业列表响应
type EnterpriseListResponse struct {
	List     []*EnterpriseInfo `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// AuditEnterpriseRequest 审核企业请求
type AuditEnterpriseRequest struct {
	Status  int    `json:"status" binding:"required,oneof=1 2"` // 1=通过, 2=拒绝
	Reason  string `json:"reason"` // 拒绝原因
}
