package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func init() {
	_ = godotenv.Load(".env")
}
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
func main() {
	router := gin.Default()

	router.Use(cors)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.OPTIONS("/v1/chat/completions", optionsHandler)
	router.POST("/v1/chat/completions", Authorization, CreateChatCompletions)
	router.Any("/backend-api/*path", Authorization, Proxy)
	err := router.Run(getEnv("SERVER_HOST", ":8080"))
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		return
	}
}
