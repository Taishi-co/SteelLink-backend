package middleware

import (
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	ClaimsKey         = "claims"
	TokenRefreshedKey = "X-Token-Refreshed"
)

type authUserSnapshot struct {
	UserID              uint64   `json:"user_id"`
	UserType            string   `json:"user_type"`
	EnterpriseID        uint64   `json:"enterprise_id"`
	EnterpriseCode      string   `json:"enterprise_code"`
	EnterpriseRole      string   `json:"enterprise_role"`
	ServiceCapabilities []string `json:"service_capabilities"`
}

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
	snapshot, ok := loadAuthUserSnapshot(c, claims.UserID)
	if !ok {
		return nil, ""
	}

	needsRefresh := false
	currentUserType := claims.UserType
	currentEnterpriseCode := claims.EnterpriseCode

	if snapshot.UserType != claims.UserType {
		currentUserType = snapshot.UserType
		needsRefresh = true
	}

	if snapshot.EnterpriseCode != claims.EnterpriseCode {
		currentEnterpriseCode = snapshot.EnterpriseCode
		needsRefresh = true
	}

	if needsRefresh {
		newToken, err := auth.GenerateToken(
			claims.UserID, claims.Phone, claims.Email,
			currentUserType, snapshot.EnterpriseID, currentEnterpriseCode, snapshot.EnterpriseRole,
			snapshot.ServiceCapabilities,
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

	if snapshot.EnterpriseCode != "" {
		if snapshot.EnterpriseID != claims.EnterpriseID || snapshot.EnterpriseRole != claims.EnterpriseRole || !sameStringSet(snapshot.ServiceCapabilities, claims.ServiceCapabilities) {
			newToken, err := auth.GenerateToken(
				claims.UserID, claims.Phone, claims.Email,
				currentUserType, snapshot.EnterpriseID, currentEnterpriseCode, snapshot.EnterpriseRole,
				snapshot.ServiceCapabilities,
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

	return nil, ""
}

func loadAuthUserSnapshot(c *gin.Context, userID uint64) (authUserSnapshot, bool) {
	key := authSnapshotCacheKey(userID)
	var snapshot authUserSnapshot
	if cache.GetJSON(c.Request.Context(), key, &snapshot) {
		return snapshot, true
	}

	db := database.GetDB()
	if db == nil {
		return authUserSnapshot{}, false
	}

	var user entity.User
	if err := db.WithContext(c.Request.Context()).Select("id, role, enterprise_code").First(&user, userID).Error; err != nil {
		return authUserSnapshot{}, false
	}

	snapshot = authUserSnapshot{
		UserID:         user.ID,
		UserType:       string(user.Role),
		EnterpriseCode: user.EnterpriseCode,
	}
	if user.EnterpriseCode != "" {
		var enterprise entity.Enterprise
		if err := db.WithContext(c.Request.Context()).Select("id, enterprise_code, enterprise_role").Where("enterprise_code = ?", user.EnterpriseCode).First(&enterprise).Error; err == nil {
			snapshot.EnterpriseID = enterprise.ID
			snapshot.EnterpriseCode = enterprise.EnterpriseCode
			snapshot.EnterpriseRole = enterprise.EnterpriseRole
			snapshot.ServiceCapabilities = approvedServiceCapabilities(c, enterprise.ID)
		}
	}
	cache.SetJSON(c.Request.Context(), key, snapshot, cache.AuthSnapshotTTL)
	return snapshot, true
}

func authSnapshotCacheKey(userID uint64) string {
	return "auth:user-snapshot:" + cache.Hash(userID)
}

func approvedServiceCapabilities(c *gin.Context, enterpriseID uint64) []string {
	db := database.GetDB()
	if db == nil || enterpriseID == 0 {
		return nil
	}

	capabilities := make([]string, 0, 5)
	for table := range map[string]struct{}{
		entity.TradeCapabilityPermissionTable: {},
		entity.ProviderServicePermissionTable: {},
	} {
		var rows []struct {
			Service string
		}
		if err := db.WithContext(c.Request.Context()).
			Table(table).
			Select("service").
			Where("enterprise_id = ? AND status = ? AND deleted_at IS NULL", enterpriseID, entity.ServicePermissionApproved).
			Find(&rows).Error; err != nil {
			return nil
		}
		for _, row := range rows {
			capabilities = append(capabilities, row.Service)
		}
	}
	return capabilities
}

func sameStringSet(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, item := range a {
		seen[item]++
	}
	for _, item := range b {
		seen[item]--
		if seen[item] < 0 {
			return false
		}
	}
	return true
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
