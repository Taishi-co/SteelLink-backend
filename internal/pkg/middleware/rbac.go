package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/pkg/auth"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
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
