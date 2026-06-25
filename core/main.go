package main

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"nexusflow/core/config"
	"nexusflow/core/decision"
	"nexusflow/core/fingerprint"
	"nexusflow/core/redisclient"

	"github.com/gin-gonic/gin"
)

const maxJSONBodyBytes int64 = 64 * 1024

func main() {
	cfg := config.Load()
	redisclient.Init(cfg)

	r := setupRouter(cfg)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(limitJSONBody(maxJSONBodyBytes))

	r.POST("/decide", func(c *gin.Context) {
		var fp fingerprint.Fingerprint
		if err := c.ShouldBindJSON(&fp); err != nil {
			writeJSONBindError(c, err)
			return
		}
		if isBlockedCountry(fp.Country, cfg.BlockCountries) {
			c.JSON(200, decision.Decision{Mode: "review", AllowProgressive: false})
			return
		}
		if redisclient.IsRateLimited(fp.IP, cfg) {
			c.JSON(200, decision.Decision{Mode: "review", AllowProgressive: false})
			return
		}
		result := decision.MakeDecision(fp, cfg)
		c.JSON(200, result)
	})

	r.POST("/behavior", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			writeJSONBindError(c, err)
			return
		}
		c.Status(204)
	})

	r.POST("/killswitch", func(c *gin.Context) {
		// Kill switch logic
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "running"})
	})

	return r
}

func limitJSONBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && (c.FullPath() == "/decide" || c.FullPath() == "/behavior") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func writeJSONBindError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
}

func isBlockedCountry(country string, blockedCountries []string) bool {
	country = strings.ToUpper(strings.TrimSpace(country))
	if country == "" {
		return false
	}
	for _, blocked := range blockedCountries {
		if country == strings.ToUpper(strings.TrimSpace(blocked)) {
			return true
		}
	}
	return false
}
