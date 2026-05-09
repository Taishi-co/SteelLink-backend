package dto

type CreateNotificationRequest struct {
	RecipientUserIDs       []uint64               `json:"recipient_user_ids"`
	RecipientEnterpriseIDs []uint64               `json:"recipient_enterprise_ids"`
	RecipientRoles         []string               `json:"recipient_roles"`
	SendToAll              bool                   `json:"send_to_all"`
	Type                   string                 `json:"type" binding:"omitempty,oneof=admin order audit system"`
	Title                  string                 `json:"title" binding:"required,max=120"`
	Content                string                 `json:"content" binding:"required,max=5000"`
	LinkURL                string                 `json:"link_url" binding:"omitempty,max=500"`
	Metadata               map[string]interface{} `json:"metadata"`
}
