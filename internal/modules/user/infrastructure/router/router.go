package router

import (
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	enterpriseHandler := handler.NewEnterpriseHandler()
	enterprisePermissionHandler := handler.NewEnterprisePermissionHandler()

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "auth-login",
			Limit:  20,
			Window: 10 * time.Minute,
		}), authHandler.Login)
		auth.POST("/refresh", middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "auth-refresh",
			Limit:  60,
			Window: 10 * time.Minute,
		}), authHandler.RefreshToken)
		auth.POST("/switch-session", middleware.AuthRequired(), authHandler.CreateSwitchSession)
		auth.POST("/switch-session/consume", authHandler.ConsumeSwitchSession)
		auth.POST("/sms/send", middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "sms-send",
			Limit:  10,
			Window: 10 * time.Minute,
		}), authHandler.SendSms)
		auth.POST("/sms/login", middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "sms-login",
			Limit:  20,
			Window: 10 * time.Minute,
		}), authHandler.SmsLogin)
		auth.POST("/password/reset", middleware.RedisRateLimit(middleware.RateLimitConfig{
			Name:   "password-reset",
			Limit:  10,
			Window: 10 * time.Minute,
		}), authHandler.ResetPassword)
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
		enterprises.GET("/me/services", enterprisePermissionHandler.GetMyServices)
		enterprises.POST("/me/services/applications", enterprisePermissionHandler.Apply)

		entAdmin := enterprises.Group("")
		entAdmin.Use(middleware.RequireUserType("operator", "super_admin"))
		{
			entAdmin.PUT("/:id/services/:service/status", enterprisePermissionHandler.UpdatePermissionStatus)
			entAdmin.PUT("/:id/audit", enterpriseHandler.Audit)
		}

		enterprises.GET("/:id", enterpriseHandler.GetByID)
		enterprises.POST("/:id/verify", enterpriseHandler.Verify)
	}

	serviceApplications := r.Group("/enterprise-service-applications")
	serviceApplications.Use(middleware.AuthRequired(), middleware.RequireUserType("operator", "super_admin"))
	{
		serviceApplications.GET("", enterprisePermissionHandler.ListApplications)
		serviceApplications.GET("/:application_id", enterprisePermissionHandler.GetApplication)
		serviceApplications.PUT("/:application_id/audit", enterprisePermissionHandler.Audit)
	}
}
