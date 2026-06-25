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
		c.ShouldBindJSON(&fp)
		result := decision.MakeDecision(fp, cfg)
		c.JSON(200, result)
	})

	r.POST("/behavior", func(c *gin.Context) {
		var data map[string]interface{}
		c.ShouldBindJSON(&data)
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/killswitch", func(c *gin.Context) {
		// Kill switch logic
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "running"})
	})

	r.Run(":" + cfg.Port)
}