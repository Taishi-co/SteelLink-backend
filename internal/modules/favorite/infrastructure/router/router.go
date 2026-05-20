package router

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/ui/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	favoriteHandler := handler.NewFavoriteHandler()

	favorites := r.Group("/favorites")
	favorites.Use(middleware.AuthRequired())
	{
		favorites.POST("", favoriteHandler.Add)
		favorites.GET("", favoriteHandler.List)
		favorites.GET("/count", favoriteHandler.Count)
		favorites.GET("/check/:product_id", favoriteHandler.Check)
		favorites.DELETE("/product/:product_id", favoriteHandler.RemoveByProduct)
		favorites.DELETE("/:id", favoriteHandler.Remove)
	}
}
