package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

const (
	// Context key for user claims
	ClaimsKey = "claims"
)

// AuthRequired validates JWT token and injects claims into context
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, ecode.ErrTokenMissing)
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, ecode.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			if err == ecode.ErrTokenExpired {
				response.Error(c, ecode.ErrTokenExpired)
			} else {
				response.Error(c, ecode.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		// Store claims in context
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves claims from gin context
func GetClaims(c *gin.Context) *auth.Claims {
	if claims, exists := c.Get(ClaimsKey); exists {
		return claims.(*auth.Claims)
	}
	return nil
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) uint64 {
	if claims := GetClaims(c); claims != nil {
		return claims.UserID
	}
	return 0
}

// OptionalAuth extracts claims if token exists but doesn't require it
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ParseToken(tokenString)
		if err == nil {
			c.Set(ClaimsKey, claims)
		}
		c.Next()
	}
}
