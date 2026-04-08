package business

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type LogoutService struct {
	rdb *redis.Client
}

func NewLogoutService(rdb *redis.Client) *LogoutService {
	return &LogoutService{
		rdb: rdb,
	}
}

func (service *LogoutService) Logout(ctx context.Context, token string, ttl time.Duration) error {
	if service.rdb == nil {
		return errors.New("redis client not initialized")
	}
	key := fmt.Sprintf("BLACKLISTED_TOKEN_%s", token)
	exists, err := service.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return errors.New("token is already blacklisted")
	}
	if ttl <= 0 {
		return nil
	}
	return service.rdb.Set(ctx, key, "1", ttl).Err() //replace TTL with TokenExpiry-time.now()
}
