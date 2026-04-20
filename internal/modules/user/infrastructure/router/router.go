package router

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
)

// RegisterRoutes 注册用户模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 初始化处理器
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()

	// 认证路由（无需登录）
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/sms/send", authHandler.SendSms)
		auth.POST("/sms/login", authHandler.SmsLogin)
		auth.POST("/password/reset", authHandler.ResetPassword)
	}

	// 需要登录的路由
	users := r.Group("/users")
	users.Use(middleware.AuthRequired())
	{
		// 当前用户操作
		users.GET("/me", userHandler.GetCurrentUser)
		users.PUT("/me", userHandler.UpdateCurrentUser)
		users.POST("/me/enterprise/join", userHandler.JoinEnterprise)
		users.POST("/me/enterprise/leave", userHandler.LeaveEnterprise)

		// 密码操作
		users.PUT("/me/password", authHandler.ChangePassword)

		// 用户管理（管理员）
		users.GET("", userHandler.ListUsers)
		users.GET("/:id", userHandler.GetUserByID)
		users.PUT("/:id/status", userHandler.UpdateUserStatus)
	}
}
