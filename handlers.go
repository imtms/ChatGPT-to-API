package main

import (
	official_types "chatgpt-to-api/typings/official"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		AccessToken = strings.Replace(authHeader, "Bearer ", "", 1)
	}
	// Convert the chat request to a ChatGPT request
	translatedRequest := ConvertAPIRequest(inRequest)

	if inRequest.Stream {
		response, err := sendConversationRequest(translatedRequest, AccessToken)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "error sending request",
			})
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(response.Body)
		HandlerStream(c, response, translatedRequest, AccessToken)
		c.String(200, "data: [DONE]\n\n")
	} else {
		c.String(500, "only support stream")
	}
}
