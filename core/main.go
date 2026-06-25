package main

import (
	"log"

	"nexusflow/core/config"
	"nexusflow/core/decision"
	"nexusflow/core/fingerprint"
	"nexusflow/core/redisclient"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	redisclient.Init(cfg)

	r := gin.Default()

	r.POST("/decide", func(c *gin.Context) {
		var fp fingerprint.Fingerprint
		if err := c.ShouldBindJSON(&fp); err != nil {
			c.JSON(400, gin.H{"error": "invalid json"})
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
			c.JSON(400, gin.H{"error": "invalid json"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/killswitch", func(c *gin.Context) {
		// Kill switch logic
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "running"})
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
