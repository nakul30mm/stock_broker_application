package utils

import (
	"context"
	"fmt"
	"stock_broker_application/src/constants"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     constants.RedisAddr,
		Password: constants.RedisPassword,
		DB:       constants.RedisDB,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	redisClient = rdb
	return rdb, nil
}

func GetRedisClient() *redis.Client {
	return redisClient
}
