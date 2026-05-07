package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
)

const (
	ClaimsKey         = "claims"
	TokenRefreshedKey = "X-Token-Refreshed"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, ecode.ErrTokenMissing)
			c.Abort()
			return
		}

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

		refreshedClaims, newToken := checkAndRefreshClaims(c, claims)
		if refreshedClaims != nil {
			claims = refreshedClaims
			c.Header(TokenRefreshedKey, newToken)
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

func checkAndRefreshClaims(c *gin.Context, claims *auth.Claims) (*auth.Claims, string) {
	db := database.GetDB()
	if db == nil {
		return nil, ""
	}

	var user entity.User
	if err := db.WithContext(c.Request.Context()).Select("id, role, enterprise_code").First(&user, claims.UserID).Error; err != nil {
		return nil, ""
	}

	needsRefresh := false
	currentUserType := claims.UserType
	currentEnterpriseCode := claims.EnterpriseCode
	currentEnterpriseID := claims.EnterpriseID

	if string(user.Role) != claims.UserType {
		currentUserType = string(user.Role)
		needsRefresh = true
	}

	if user.EnterpriseCode != claims.EnterpriseCode {
		currentEnterpriseCode = user.EnterpriseCode
		needsRefresh = true
	}

	if needsRefresh {
		var enterpriseID uint64
		var enterpriseRole string
		if user.EnterpriseCode != "" {
			var enterprise entity.Enterprise
			if err := db.WithContext(c.Request.Context()).Select("id, enterprise_code, enterprise_role").Where("enterprise_code = ?", user.EnterpriseCode).First(&enterprise).Error; err == nil {
				enterpriseID = enterprise.ID
				enterpriseRole = enterprise.EnterpriseRole
			}
		}

		newToken, err := auth.GenerateToken(
			user.ID, claims.Phone, claims.Email,
			currentUserType, enterpriseID, currentEnterpriseCode, enterpriseRole,
		)
		if err != nil {
			return nil, ""
		}

		newClaims, err := auth.ParseToken(newToken.AccessToken)
		if err != nil {
			return nil, ""
		}

		return newClaims, newToken.AccessToken
	}

	if currentEnterpriseID == 0 && user.EnterpriseCode != "" {
		var enterprise entity.Enterprise
		if err := db.WithContext(c.Request.Context()).Select("id, enterprise_code, enterprise_role").Where("enterprise_code = ?", user.EnterpriseCode).First(&enterprise).Error; err == nil {
			if enterprise.ID != claims.EnterpriseID || enterprise.EnterpriseRole != claims.EnterpriseRole {
				newToken, err := auth.GenerateToken(
					user.ID, claims.Phone, claims.Email,
					currentUserType, enterprise.ID, currentEnterpriseCode, enterprise.EnterpriseRole,
				)
				if err != nil {
					return nil, ""
				}
				newClaims, err := auth.ParseToken(newToken.AccessToken)
				if err != nil {
					return nil, ""
				}
				return newClaims, newToken.AccessToken
			}
		}
	}

	return nil, ""
}

func GetClaims(c *gin.Context) *auth.Claims {
	if claims, exists := c.Get(ClaimsKey); exists {
		return claims.(*auth.Claims)
	}
	return nil
}

func GetUserID(c *gin.Context) uint64 {
	if claims := GetClaims(c); claims != nil {
		return claims.UserID
	}
	return 0
}

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
			refreshedClaims, newToken := checkAndRefreshClaims(c, claims)
			if refreshedClaims != nil {
				claims = refreshedClaims
				c.Header(TokenRefreshedKey, newToken)
			}
			c.Set(ClaimsKey, claims)
		}
		c.Next()
	}
}
