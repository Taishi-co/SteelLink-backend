package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type RateLimitConfig struct {
	Name   string
	Limit  int64
	Window time.Duration
	KeyFn  func(*gin.Context) string
}

func RedisRateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Limit <= 0 || cfg.Window <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = "default"
	}
	return func(c *gin.Context) {
		identity := ""
		if cfg.KeyFn != nil {
			identity = strings.TrimSpace(cfg.KeyFn(c))
		}
		if identity == "" {
			identity = clientIdentity(c)
		}
		key := "rate:" + name + ":" + cache.HashString(identity)
		count, ok := cache.Increment(c.Request.Context(), key, cfg.Window)
		if !ok {
			c.Next()
			return
		}
		if count > cfg.Limit {
			retryAfter := int(cache.TTL(c.Request.Context(), key).Seconds())
			if retryAfter <= 0 {
				retryAfter = int(cfg.Window.Seconds())
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, response.Response{
				Code:    1000001,
				Message: fmt.Sprintf("请求过于频繁，请 %d 秒后重试", retryAfter),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RateLimitByUserOrIP(c *gin.Context) string {
	if claims := GetClaims(c); claims != nil && claims.UserID > 0 {
		return fmt.Sprintf("user:%d", claims.UserID)
	}
	return clientIdentity(c)
}

func clientIdentity(c *gin.Context) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		if value != "" {
			return "ip:" + value
		}
	}
	return "ip:" + c.ClientIP()
}
