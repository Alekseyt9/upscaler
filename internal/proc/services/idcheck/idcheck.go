// Package idcheck provides a service for checking and storing idempotency keys using Redis.
// It helps to ensure that certain operations are performed only once by storing keys with a TTL (Time to Live).
package idcheck

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// IdCheckService is a service for checking idempotency keys in Redis.
// It stores keys with a specified TTL to prevent duplicate processing of tasks.
type IdCheckService struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewIdCheckService creates a new service for checking idempotency keys using Redis with a specified TTL.
//
// Parameters:
//   - redisAddr: The address of the Redis server.
//   - ttl: The time-to-live (TTL) for keys stored in Redis.
//
// Returns:
//   - A pointer to an IdCheckService instance.
func NewIdCheckService(redisAddr string, ttl time.Duration) *IdCheckService {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	return &IdCheckService{
		redisClient: rdb,
		ttl:         ttl,
	}
}

// CheckAndSave checks the existence of a key and saves it with a predefined TTL if it does not exist.
// Returns true if the key was added and false if the key already exists.
//
// Parameters:
//   - ctx: The context for handling request-scoped values and cancellations.
//   - key: The key to check and potentially store in Redis.
//
// Returns:
//   - A boolean indicating whether the key was newly added (true) or already existed (false).
func (s *IdCheckService) CheckAndSave(ctx context.Context, key string) bool {
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Error checking key in Redis: %v", err)
		return false
	}

	if exists > 0 {
		return false
	}

	err = s.redisClient.Set(ctx, key, 1, s.ttl).Err()
	if err != nil {
		log.Printf("Error saving key in Redis: %v", err)
		return false
	}

	return true
}

func (s *IdCheckService) Close() {
	s.redisClient.FlushDB(context.TODO())
	s.redisClient.Close()
}
