package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

// InitCache initializes Redis connection
func InitCache(redisURL string) error {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Try default connection if URL parsing fails
		opt = &redis.Options{
			Addr:     redisURL,
			Password: "",
			DB:       0,
		}
	}

	Client = redis.NewClient(opt)

	// Test connection
	_, err = Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// GetShopperBalance gets cached shopper balance
func GetShopperBalance(shopperID string) (int, bool, error) {
	key := fmt.Sprintf("shopper:balance:%s", shopperID)
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	var balance int
	if err := json.Unmarshal([]byte(val), &balance); err != nil {
		return 0, false, err
	}

	return balance, true, nil
}

// SetShopperBalance caches shopper balance with TTL
func SetShopperBalance(shopperID string, balance int, ttl time.Duration) error {
	key := fmt.Sprintf("shopper:balance:%s", shopperID)
	val, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, val, ttl).Err()
}

// InvalidateShopperBalance removes cached balance
func InvalidateShopperBalance(shopperID string) error {
	key := fmt.Sprintf("shopper:balance:%s", shopperID)
	return Client.Del(ctx, key).Err()
}

