package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	enterpriseHandler := handler.NewEnterpriseHandler()

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/sms/send", authHandler.SendSms)
		auth.POST("/sms/login", authHandler.SmsLogin)
		auth.POST("/password/reset", authHandler.ResetPassword)
	}

	users := r.Group("/users")
	users.Use(middleware.AuthRequired())
	{
		users.GET("/me", userHandler.GetCurrentUser)
		users.PUT("/me", userHandler.UpdateCurrentUser)
		users.POST("/me/enterprise/join", userHandler.JoinEnterprise)
		users.POST("/me/enterprise/leave", userHandler.LeaveEnterprise)

		users.POST("/me/identity", userHandler.SubmitIdentity)
		users.GET("/me/identity", userHandler.GetIdentityStatus)

		users.GET("/me/dashboard", userHandler.GetDashboard)

		subAccounts := users.Group("")
		subAccounts.Use(middleware.RequireUserType("enterprise"))
		{
			subAccounts.POST("/me/sub-accounts", userHandler.CreateSubAccount)
			subAccounts.GET("/me/sub-accounts", userHandler.ListSubAccounts)
			subAccounts.PUT("/me/sub-accounts/:sub_id", userHandler.UpdateSubAccount)
		}

		users.PUT("/me/password", authHandler.ChangePassword)

		userAdmin := users.Group("")
		userAdmin.Use(middleware.RequireUserType("operator", "super_admin"))
		{
			userAdmin.GET("", userHandler.ListUsers)
			userAdmin.GET("/:id", userHandler.GetUserByID)
			userAdmin.PUT("/:id/status", userHandler.UpdateUserStatus)
		}
	}

	enterprises := r.Group("/enterprises")
	enterprises.Use(middleware.AuthRequired())
	{
		enterprises.POST("", enterpriseHandler.Create)
		enterprises.GET("", enterpriseHandler.List)
		enterprises.GET("/:id", enterpriseHandler.GetByID)
		enterprises.POST("/:id/verify", enterpriseHandler.Verify)

		entAdmin := enterprises.Group("")
		entAdmin.Use(middleware.RequireUserType("operator", "super_admin"))
		{
			entAdmin.PUT("/:id/audit", enterpriseHandler.Audit)
		}
	}
}
