package middleware

import (
	"net/http"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/internal/pkg/cache"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const IdempotencyKeyHeader = "Idempotency-Key"

func RedisIdempotency(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		idempotencyKey := strings.TrimSpace(c.GetHeader(IdempotencyKeyHeader))
		if idempotencyKey == "" {
			c.Next()
			return
		}

		claims := GetClaims(c)
		owner := clientIdentity(c)
		if claims != nil && claims.UserID > 0 {
			owner = "user:" + cache.Hash(claims.UserID)
		}
		key := cache.Key("idempotency:"+name, struct {
			Owner string `json:"owner"`
			Path  string `json:"path"`
			Key   string `json:"key"`
		}{Owner: owner, Path: c.FullPath(), Key: idempotencyKey})

		if !cache.TryLock(c.Request.Context(), key, "locked", cache.IdempotencyKeyTTL) {
			c.JSON(http.StatusConflict, response.Response{
				Code:    1000001,
				Message: "重复提交，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
