package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/pkg/logger"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"go.uber.org/zap"
)

// Recovery returns a recovery middleware that catches panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)

				// Return internal error response
				response.Error(c, nil)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
