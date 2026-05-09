package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/redis/go-redis/v9"
)

const (
	AuthSnapshotTTL       = 5 * time.Minute
	EnterpriseServiceTTL  = 5 * time.Minute
	ProductListTTL        = 60 * time.Second
	ProductDetailTTL      = 3 * time.Minute
	ServiceOfferListTTL   = 2 * time.Minute
	ServiceOfferDetailTTL = 5 * time.Minute
	MarketStatsTTL        = 45 * time.Second
	AIProcurementTTL      = 20 * time.Minute
	AIProcurementMatchTTL = 10 * time.Minute
	PlatformStatsTTL      = 60 * time.Second
	DashboardStatsTTL     = 60 * time.Second
	IdempotencyKeyTTL     = 5 * time.Minute
	VectorSyncLockTTL     = 30 * time.Minute
	VectorSyncStatusTTL   = 24 * time.Hour
)

func Hash(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		raw = []byte("")
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func HashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func Key(prefix string, value any) string {
	return prefix + ":" + Hash(value)
}

func GetJSON(ctx context.Context, key string, target any) bool {
	client := database.GetRedis()
	if client == nil {
		return false
	}
	raw, err := client.Get(ctx, key).Result()
	if err != nil || raw == "" {
		return false
	}
	return json.Unmarshal([]byte(raw), target) == nil
}

func SetJSON(ctx context.Context, key string, value any, ttl time.Duration) {
	client := database.GetRedis()
	if client == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = client.Set(ctx, key, raw, ttl).Err()
}

func Exists(ctx context.Context, key string) bool {
	client := database.GetRedis()
	if client == nil {
		return false
	}
	n, err := client.Exists(ctx, key).Result()
	return err == nil && n > 0
}

func SetString(ctx context.Context, key, value string, ttl time.Duration) {
	client := database.GetRedis()
	if client == nil {
		return
	}
	_ = client.Set(ctx, key, value, ttl).Err()
}

func GetString(ctx context.Context, key string) (string, bool) {
	client := database.GetRedis()
	if client == nil {
		return "", false
	}
	value, err := client.Get(ctx, key).Result()
	return value, err == nil
}

func Increment(ctx context.Context, key string, ttl time.Duration) (int64, bool) {
	client := database.GetRedis()
	if client == nil {
		return 0, false
	}
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, false
	}
	if count == 1 && ttl > 0 {
		_ = client.Expire(ctx, key, ttl).Err()
	}
	return count, true
}

func TTL(ctx context.Context, key string) time.Duration {
	client := database.GetRedis()
	if client == nil {
		return 0
	}
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil || ttl < 0 {
		return 0
	}
	return ttl
}

func TryLock(ctx context.Context, key, token string, ttl time.Duration) bool {
	client := database.GetRedis()
	if client == nil {
		return true
	}
	ok, err := client.SetNX(ctx, key, token, ttl).Result()
	return err == nil && ok
}

func Unlock(ctx context.Context, key, token string) {
	client := database.GetRedis()
	if client == nil {
		return
	}
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`
	_ = client.Eval(ctx, script, []string{key}, token).Err()
}

func IsRedisNil(err error) bool {
	return err == redis.Nil
}

func Delete(ctx context.Context, keys ...string) {
	client := database.GetRedis()
	if client == nil || len(keys) == 0 {
		return
	}
	_ = client.Del(ctx, keys...).Err()
}

func DeleteByPattern(ctx context.Context, pattern string) {
	client := database.GetRedis()
	if client == nil || pattern == "" {
		return
	}
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			Delete(ctx, keys...)
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		Delete(ctx, keys...)
	}
}
