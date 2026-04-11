package utils

import (
	"context"
	"fmt"
	"stock_broker_application/src/constants"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     constants.RedisAddr,
		Password: constants.RedisPassword,
		DB:       constants.RedisDB,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	redisClient = redisClient
	return redisClient, nil
}

func GetRedisClient() *redis.Client {
	return redisClient
}
