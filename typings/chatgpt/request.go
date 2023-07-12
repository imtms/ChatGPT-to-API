package chatgpt

import (
	"os"

	"github.com/google/uuid"
)

type chatgptMessage struct {
	ID      uuid.UUID      `json:"id"`
	Author  chatgptAuthor  `json:"author"`
	Content chatgptContent `json:"content"`
}

type chatgptContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type chatgptAuthor struct {
	Role string `json:"role"`
}

type GptRequest struct {
	Action                     string           `json:"action"`
	Messages                   []chatgptMessage `json:"messages"`
	ParentMessageID            string           `json:"parent_message_id,omitempty"`
	ConversationID             string           `json:"conversation_id,omitempty"`
	Model                      string           `json:"model"`
	HistoryAndTrainingDisabled bool             `json:"history_and_training_disabled"`
	ArkoseToken                string           `json:"arkose_token,omitempty"`
	PluginIDs                  []string         `json:"plugin_ids,omitempty"`
}

func NewChatGPTRequest() GptRequest {
	enableHistory := os.Getenv("ENABLE_HISTORY") == ""
	return GptRequest{
		Action:                     "next",
		ParentMessageID:            uuid.NewString(),
		Model:                      "text-davinci-002-render-sha",
		HistoryAndTrainingDisabled: !enableHistory,
	}
}

func (c *GptRequest) AddMessage(role string, content string) {
	c.Messages = append(c.Messages, chatgptMessage{
		ID:      uuid.New(),
		Author:  chatgptAuthor{Role: role},
		Content: chatgptContent{ContentType: "text", Parts: []string{content}},
	})
}
