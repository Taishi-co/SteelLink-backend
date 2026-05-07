package middleware

import (
	"github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func RequireUserType(userTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		for _, ut := range userTypes {
			if claims.UserType == ut {
				c.Next()
				return
			}
		}
		response.Error(c, ecode.ErrForbidden.WithMessage("无权访问：用户类型不匹配"))
		c.Abort()
	}
}

func RequireEnterpriseRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		if claims.EnterpriseCode == "" {
			response.Error(c, ecode.ErrForbidden.WithMessage("无权访问：未加入企业"))
			c.Abort()
			return
		}
		for _, r := range roles {
			if claims.EnterpriseRole == r {
				c.Next()
				return
			}
		}
		response.Error(c, ecode.ErrForbidden.WithMessage("无权访问：企业角色不匹配"))
		c.Abort()
	}
}

func RequireEnterpriseService(services ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		if claims.UserType != string(entity.TypeEnterprise) || claims.EnterpriseID == 0 {
			response.Error(c, ecode.ErrForbidden.WithMessage("无权访问：未开通企业服务"))
			c.Abort()
			return
		}

		db := database.GetDB()
		if db == nil {
			response.Error(c, ecode.ErrDatabase.WithMessage("数据库未初始化"))
			c.Abort()
			return
		}

		var total int64
		for table, tableServices := range entity.PermissionTablesForServices(services) {
			var count int64
			if err := db.WithContext(c.Request.Context()).
				Table(table).
				Where("enterprise_id = ? AND service IN ? AND status = ?", claims.EnterpriseID, tableServices, entity.ServicePermissionApproved).
				Count(&count).Error; err != nil {
				response.Error(c, ecode.ErrDatabase.WithMessage("查询企业能力权限失败"))
				c.Abort()
				return
			}
			total += count
		}
		if total == 0 {
			response.Error(c, ecode.ErrForbidden.WithMessage("无权访问：企业未开通对应能力"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireTradeCapability(capabilities ...string) gin.HandlerFunc {
	return RequireEnterpriseService(capabilities...)
}

func RequireProviderService(services ...string) gin.HandlerFunc {
	return RequireEnterpriseService(services...)
}

func RequireFulfillmentParticipant() gin.HandlerFunc {
	return RequireEnterpriseService("procurement", "sales", "warehouse", "logistics", "finance")
}

func GetEnterpriseCode(c *gin.Context) string {
	if claims := GetClaims(c); claims != nil {
		return claims.EnterpriseCode
	}
	return ""
}

func GetEnterpriseRole(c *gin.Context) string {
	if claims := GetClaims(c); claims != nil {
		return claims.EnterpriseRole
	}
	return ""
}

func HasEnterpriseRole(c *gin.Context, role string) bool {
	claims := GetClaims(c)
	if claims == nil {
		return false
	}
	return claims.EnterpriseRole == role
}

func IsEnterpriseUser(c *gin.Context) bool {
	claims := GetClaims(c)
	if claims == nil {
		return false
	}
	return claims.EnterpriseID != 0
}

func EnsureClaims(c *gin.Context) *auth.Claims {
	return GetClaims(c)
}
