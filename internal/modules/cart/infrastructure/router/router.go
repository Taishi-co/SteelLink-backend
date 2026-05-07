package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	cartHandler := handler.NewCartHandler()

	cart := r.Group("/cart")
	cart.Use(middleware.AuthRequired(), middleware.RequireTradeCapability("procurement"))
	{
		cart.GET("/items", cartHandler.List)
		cart.POST("/items", cartHandler.Add)
		cart.PUT("/items/:id", cartHandler.Update)
		cart.DELETE("/items/:id", cartHandler.Remove)
		cart.GET("/count", cartHandler.Count)
		cart.DELETE("", cartHandler.Clear)
	}
}
