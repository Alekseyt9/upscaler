package idcheck

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisForTest() (*redis.Client, func()) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	cleanup := func() {
		rdb.FlushDB(context.TODO())
		rdb.Close()
	}

	return rdb, cleanup
}

func TestIdCheckService_CheckAndSave(t *testing.T) {
	rdb, cleanup := setupRedisForTest()
	defer cleanup()

	ttl := time.Minute * 5
	service := &IdCheckService{
		redisClient: rdb,
		ttl:         ttl,
	}

	ctx := context.TODO()

	// Test Case 1: Check and save a new key
	key := "test-key-1"
	result := service.CheckAndSave(ctx, key)
	require.True(t, result, "Expected CheckAndSave to return true for a new key")

	// Verify that the key exists in Redis
	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(t, err, "Expected no error when checking key existence in Redis")
	assert.Equal(t, int64(1), exists, "Expected key to exist in Redis after being saved")

	// Test Case 2: Check the same key again (should return false)
	result = service.CheckAndSave(ctx, key)
	require.False(t, result, "Expected CheckAndSave to return false for an existing key")

	// Test Case 3: TTL check
	ttlRemaining, err := rdb.TTL(ctx, key).Result()
	require.NoError(t, err, "Expected no error when checking TTL in Redis")
	assert.Greater(t, ttlRemaining, time.Duration(0), "Expected TTL to be greater than 0")

	// Test Case 4: Check and save a different key
	key2 := "test-key-2"
	result = service.CheckAndSave(ctx, key2)
	require.True(t, result, "Expected CheckAndSave to return true for a different new key")

	// Verify that the second key exists in Redis
	exists, err = rdb.Exists(ctx, key2).Result()
	require.NoError(t, err, "Expected no error when checking existence of the second key in Redis")
	assert.Equal(t, int64(1), exists, "Expected second key to exist in Redis after being saved")
}
