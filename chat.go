package main

import (
	"bufio"
	"bytes"
	"chatgpt-to-api/typings"
	chatgpt_types "chatgpt-to-api/typings/chatgpt"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"

	official_types "chatgpt-to-api/typings/official"
)

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(tls_client.Safari_Ipad_15_6),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		// Disable SSL verification
		tls_client.WithInsecureSkipVerify(),
	}
	client, _ = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	httpProxy = os.Getenv("HTTP_PROXY")
	ApiDomain = getEnv("API_Domain", "https://chat.openai.com")
)

func init() {
	if httpProxy != "" {
		err := client.SetProxy(httpProxy)
		if err != nil {
			return
		}
	}
}
func sendConversationRequest(message chatgpt_types.GptRequest, accessToken string) (*http.Response, error) {
	bodyJson, err := json.Marshal(message)
	if err != nil {
		return &http.Response{}, err
	}

	request, err := http.NewRequest(http.MethodPost, ApiDomain+"/backend-api/conversation", bytes.NewBuffer(bodyJson))
	if err != nil {
		return &http.Response{}, err
	}
	if os.Getenv("PUID") != "" {
		request.Header.Set("Cookie", "_puid="+os.Getenv("PUID")+";")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	if err != nil {
		return &http.Response{}, err
	}
	response, err := client.Do(request)
	if response.StatusCode != 200 {
		str, err2 := io.ReadAll(response.Body)
		if err2 != nil {
			return nil, err2
		}
		return nil, errors.New(string(str))
	}
	return response, nil
}

func HandlerStream(c *gin.Context, response *http.Response, translatedRequest chatgpt_types.GptRequest, accessToken string) {
	reader := bufio.NewReader(response.Body)
	var finishReason string
	var previousText typings.StringStruct
	var originalResponse chatgpt_types.GptResponse
	var isRole = true
	c.Header("Content-Type", "text/event-stream")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if len(line) < 6 {
			continue
		}
		// Remove "data: " from the beginning of the line
		line = line[6:]
		// Check if line starts with [DONE]
		if !strings.HasPrefix(line, "[DONE]") {
			// Parse the line as JSON

			err = json.Unmarshal([]byte(line), &originalResponse)
			if err != nil {
				continue
			}
			if originalResponse.Error != nil {
				c.JSON(500, gin.H{"error": originalResponse.Error})
				return
			}
			if originalResponse.Message.Author.Role != "assistant" || originalResponse.Message.Content.Parts == nil {
				continue
			}
			if originalResponse.Message.Metadata.MessageType != "next" && originalResponse.Message.Metadata.MessageType != "continue" || originalResponse.Message.EndTurn != nil {
				continue
			}
			responseString := ConvertToString(&originalResponse, &previousText, isRole)
			isRole = false
			_, err = c.Writer.WriteString(responseString)
			if err != nil {
				return
			}

			// Flush the response writer buffer to ensure that the client receives each line as it's written
			c.Writer.Flush()

			if originalResponse.Message.Metadata.FinishDetails != nil {
				finishReason = originalResponse.Message.Metadata.FinishDetails.Type
			}

		} else {
			if finishReason == "max_tokens" {
				fmt.Println("continuing")
				translatedRequest.Messages = nil
				translatedRequest.Action = "continue"
				translatedRequest.ConversationID = originalResponse.ConversationID
				translatedRequest.ParentMessageID = originalResponse.Message.ID
				response, err := sendConversationRequest(translatedRequest, accessToken)
				if err != nil {
					c.JSON(500, gin.H{
						"error": "error sending request",
					})
					return
				}
				func() {
					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {
							return
						}
					}(response.Body)
				}()
				HandlerStream(c, response, translatedRequest, accessToken)
			} else {
				finalLine := official_types.StopChunk(finishReason)
				_, err := c.Writer.WriteString("data: " + finalLine.String() + "\n\n")
				if err != nil {
					return
				}
			}
		}
	}
}

var arkoseTokenUrl = os.Getenv("ARKOSE_TOKEN_URL")

func ConvertAPIRequest(apiRequest official_types.APIRequest) chatgpt_types.GptRequest {
	chatgptRequest := chatgpt_types.NewChatGPTRequest()
	if strings.HasPrefix(apiRequest.Model, "gpt-3.5") {
		chatgptRequest.Model = "text-davinci-002-render-sha"
	}
	if strings.HasPrefix(apiRequest.Model, "gpt-4") {
		token, err := GetArkoseToken(arkoseTokenUrl)
		if err == nil {
			chatgptRequest.ArkoseToken = token
		} else {
			fmt.Println("Error getting Arkose token: ", err)
		}
		chatgptRequest.Model = apiRequest.Model
	}
	if apiRequest.PluginIDs != nil {
		chatgptRequest.PluginIDs = apiRequest.PluginIDs
		chatgptRequest.Model = "gpt-4-plugins"
	}

	for _, apiMessage := range apiRequest.Messages {
		if apiMessage.Role == "system" {
			apiMessage.Role = "critic"
		}
		chatgptRequest.AddMessage(apiMessage.Role, apiMessage.Content)
	}
	return chatgptRequest
}

func GetArkoseToken(arkoseTokenUrl string) (string, error) {
	resp, err := http.Get(arkoseTokenUrl)
	if err != nil {
		fmt.Println("Error getting Arkose token: ", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error getting Arkose token: ", err)
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	responseMap := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	if err != nil {
		fmt.Println("Error getting Arkose token: ", err)
		return "", err
	}
	token, ok := responseMap["token"]
	if !ok || token == "" {
		fmt.Println("Error getting Arkose token: ", err)
		return "", err
	}
	return token.(string), nil
}

func ConvertToString(chatgptResponse *chatgpt_types.GptResponse, previousText *typings.StringStruct, role bool) string {
	translatedResponse := official_types.NewChatCompletionChunk(strings.ReplaceAll(chatgptResponse.Message.Content.Parts[0], *&previousText.Text, ""))
	if role {
		translatedResponse.Choices[0].Delta.Role = chatgptResponse.Message.Author.Role
	}
	previousText.Text = chatgptResponse.Message.Content.Parts[0]
	return "data: " + translatedResponse.String() + "\n\n"

}
