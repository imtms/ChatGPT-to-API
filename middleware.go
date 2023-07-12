package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Next()
}

func Authorization(c *gin.Context) {
	if c.Request.Header.Get("Authorization") == "" {
		c.JSON(401, gin.H{"error": "No API key provided. Get one at https://discord.gg/9K2BvbXEHT"})
	} else if strings.HasPrefix(c.Request.Header.Get("Authorization"), "Bearer sk-") {
		c.JSON(401, gin.H{"error": "You tried to use the official API key which is not supported."})
	} else if strings.HasPrefix(c.Request.Header.Get("Authorization"), "Bearer eyJhbGciOiJSUzI1NiI") {
		return
	} else {
		c.JSON(401, gin.H{"error": "Invalid API key."})
	}
	c.Next()
}
