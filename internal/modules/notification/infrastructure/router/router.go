package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/notification/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := handler.NewHandler()

	notifications := r.Group("/notifications")
	notifications.Use(middleware.AuthRequired())
	{
		notifications.GET("", h.List)
		notifications.GET("/unread-count", h.UnreadCount)
		notifications.PUT("/read-all", h.MarkAllRead)
		notifications.PUT("/:id/read", h.MarkRead)
	}

	admin := r.Group("/admin/notifications")
	admin.Use(middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"))
	{
		admin.POST("", h.CreateByAdmin)
	}
}
