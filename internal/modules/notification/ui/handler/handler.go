package handler

import (
	"strconv"

	notificationService "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/application/service"
	notificationEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/notification/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *notificationService.NotificationService
}

func NewHandler() *Handler {
	return &Handler{service: notificationService.NewNotificationService()}
}

func (h *Handler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	var status *int
	if raw := c.Query("status"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			response.BadRequest(c, "无效的消息状态")
			return
		}
		status = &parsed
	}
	list, total, unread, err := h.service.List(c.Request.Context(), notificationService.ListQuery{
		UserID:   userID,
		Status:   status,
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", intQuery(c, "pageSize", 20)),
	})
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"list": list, "total": total, "unread_count": unread})
}

func (h *Handler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	count, err := h.service.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的消息ID")
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "已读"})
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"message": "全部已读"})
}

func (h *Handler) CreateByAdmin(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var req dto.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	created, err := h.service.Create(c.Request.Context(), notificationService.CreateNotificationInput{
		RecipientUserIDs:       req.RecipientUserIDs,
		RecipientEnterpriseIDs: req.RecipientEnterpriseIDs,
		RecipientRoles:         req.RecipientRoles,
		SendToAll:              req.SendToAll,
		SenderUserID:           claims.UserID,
		Type:                   notificationEntity.NotificationType(req.Type),
		Title:                  req.Title,
		Content:                req.Content,
		LinkURL:                req.LinkURL,
		Metadata:               req.Metadata,
	})
	if err != nil {
		response.Error(c, err.(*ecode.Error))
		return
	}
	response.Success(c, gin.H{"created": len(created)})
}

func intQuery(c *gin.Context, key string, fallback int) int {
	if raw := c.Query(key); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			return value
		}
	}
	return fallback
}
