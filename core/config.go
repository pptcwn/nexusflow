package config

type Config struct {
	Port              string
	RedisURL          string
	SURL           string
	MOrigin       string
	BlockCountries    []string
	RateLimitPerMin   int
}

func Load() *Config {
	return &Config{
		Port:            "8080",
		RedisURL:        "redis://localhost:6379",
		SafeURL:         "https://s.yourdomain.com",
		MoneyOrigin:     "https://m.yourdomain.com",
		BlockCountries:  []string{"RU", "CN"},
		RateLimitPerMin: 40,
	}
}