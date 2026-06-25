package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	RedisURL          string
	StandardOrigin    string
	ReviewOrigin      string
	BlockCountries    []string
	RateLimitPerMin   int
	KillSwitchEnabled bool
	GlobalMode        string
}

func Default() *Config {
	return &Config{
		Port:              "8080",
		RedisURL:          "redis://localhost:6379",
		StandardOrigin:    "https://standard.yourdomain.com",
		ReviewOrigin:      "https://review.yourdomain.com",
		BlockCountries:    []string{"RU", "CN"},
		RateLimitPerMin:   40,
		KillSwitchEnabled: false,
		GlobalMode:        "review",
	}
}

func Load() *Config {
	cfg := Default()
	cfg.Port = getEnv("PORT", cfg.Port)
	cfg.RedisURL = getEnv("REDIS_URL", cfg.RedisURL)
	cfg.StandardOrigin = getEnv("STANDARD_ORIGIN", cfg.StandardOrigin)
	cfg.ReviewOrigin = getEnv("REVIEW_ORIGIN", cfg.ReviewOrigin)
	cfg.GlobalMode = getEnv("GLOBAL_MODE", cfg.GlobalMode)
	cfg.BlockCountries = splitEnv("BLOCK_COUNTRIES", cfg.BlockCountries)
	cfg.RateLimitPerMin = intEnv("RATE_LIMIT_PER_MIN", cfg.RateLimitPerMin)
	cfg.KillSwitchEnabled = boolEnv("KILL_SWITCH_ENABLED", cfg.KillSwitchEnabled)
	return cfg
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}
