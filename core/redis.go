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
	opt, _ := redis.ParseURL(cfg.RedisURL)
	Client = redis.NewClient(opt)
}

func IsRateLimited(ip string, cfg *config.Config) bool {
	// ... โค้ดเดิม
	return false
}

func IsBlocked(ip string) bool {
	return false
}