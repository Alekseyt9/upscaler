package idcheck

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type IdCheckService struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewIdCheckService создает новый сервис для проверки ключей идемпотентности через Redis с заданным TTL.
func NewIdCheckService(redisAddr string, ttl time.Duration) *IdCheckService {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr, // например, "localhost:6379"
		DB:   0,         // используется 0-й Redis DB
	})

	return &IdCheckService{
		redisClient: rdb,
		ttl:         ttl,
	}
}

// CheckAndSave проверяет наличие ключа и сохраняет его с заранее заданным TTL, если ключ отсутствует.
// Возвращает true, если ключ был добавлен, и false, если ключ уже существует.
func (s *IdCheckService) CheckAndSave(ctx context.Context, key string) bool {
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Ошибка при проверке ключа в Redis: %v", err)
		return false
	}

	if exists > 0 {
		// Ключ уже существует
		return false
	}

	// Ключ не существует, сохраняем его с TTL
	err = s.redisClient.Set(ctx, key, 1, s.ttl).Err()
	if err != nil {
		log.Printf("Ошибка при сохранении ключа в Redis: %v", err)
		return false
	}

	return true
}
