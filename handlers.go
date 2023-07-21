package main

import (
	chatgpt_types "chatgpt-to-api/typings/chatgpt"
	official_types "chatgpt-to-api/typings/official"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

var AccessToken = os.Getenv("ACCESS_TOKEN")

func optionsHandler(c *gin.Context) {
	// Set headers for CORS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Proxy(c *gin.Context) {
	if c.Param("path") == "/conversation" {
		var inRequest chatgpt_types.GptRequest
		err := c.BindJSON(&inRequest)
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{
				"message": "Request must be proper JSON",
				"type":    "invalid_request_error",
				"param":   nil,
				"code":    err.Error(),
			}})
			return
		}
		if strings.HasPrefix(inRequest.Model, "gpt-4") {
			arkoseToken, err := GetArkoseToken(arkoseTokenUrl)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			inRequest.ArkoseToken = arkoseToken
		}
		handleConversation(c, inRequest)
	} else {
		Normal(c)
	}
}

func CreateChatCompletions(c *gin.Context) {
	var inRequest official_types.APIRequest
	err := c.BindJSON(&inRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}
	handleConversation(c, ConvertAPIRequest(inRequest))
}
