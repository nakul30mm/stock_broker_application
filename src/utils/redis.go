package utils

import (
	"context"
	"stock_broker_application/src/constants"
	"sync"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient         *redis.Client
	once                sync.Once
	redisErr            error
	mockWalaRedisClient *redis.Client
	mockRedisController redismock.ClientMock
)

func InitRedis(ctx context.Context) {
	//once.Do creates a connectio pool only once
	client := redis.NewClient(&redis.Options{
		Addr:     constants.RedisAddr,
		Password: constants.RedisPassword,
		DB:       constants.RedisDB,
	})

	_, pingErr := client.Ping(ctx).Result()
	if pingErr != nil {
		return
	}

	redisClient = client
	redisErr = pingErr
}

func GetRedisClient(ctx context.Context, isReal bool) (*redis.Client, redismock.ClientMock, error) {
	if isReal {
		if redisClient != nil {
			return redisClient, nil, nil
		}
		once.Do(func() {
			InitRedis(ctx)
		})
		return redisClient, nil, redisErr
	}

	if mockWalaRedisClient != nil {
		return mockWalaRedisClient, mockRedisController, nil
	}
	once.Do(func() {
		getMockRedisClient(&testing.T{})
	})
	return mockWalaRedisClient, mockRedisController, nil
	// return redisClient
}

func getMockRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { client.Close() })
	mockWalaRedisClient = client
	mockRedisController = mock
}
