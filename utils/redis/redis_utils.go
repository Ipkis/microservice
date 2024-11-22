package redis_utils

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func InitRedisClient(addr, password string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) RevokeToken(token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	_, err := r.client.Set(context.Background(), "revoked:"+token, "true", ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}
	return nil
}

func (r *RedisClient) IsTokenRevoked(token string) (bool, error) {
	val, err := r.client.Get(context.Background(), "revoked:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get revoked token status: %v", err)
	}
	return val == "true", nil
}
