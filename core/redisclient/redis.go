package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"nexusflow/core/config"
)

var Client *redis.Client
var Ctx = context.Background()

func Init(cfg *config.Config) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		Client = nil
		return
	}
	Client = redis.NewClient(opt)
}

func IsRateLimited(ip string, cfg *config.Config) bool {
	if Client == nil || ip == "" {
		return false
	}

	key := "rate:" + ip
	count, err := Client.Incr(Ctx, key).Result()
	if err != nil {
		return false
	}
	if count == 1 {
		Client.Expire(Ctx, key, time.Minute)
	}
	return int(count) > cfg.RateLimitPerMin
}

func IsBlocked(ip string) bool {
	if Client == nil || ip == "" {
		return false
	}
	blocked, err := Client.SIsMember(Ctx, "blocked_ips", ip).Result()
	return err == nil && blocked
}
