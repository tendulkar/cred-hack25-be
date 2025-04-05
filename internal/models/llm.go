package models

// LLMMessage represents a message in a conversation with an LLM
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMChatRequest represents a request to chat with an LLM
type LLMChatRequest struct {
	Messages    []LLMMessage `json:"messages"`
	ModelName   string       `json:"model,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature *float32     `json:"temperature,omitempty"`
	TopP        *float32     `json:"top_p,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

// LLMChatResponse represents a response from chatting with an LLM
type LLMChatResponse struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	ModelName    string `json:"model,omitempty"`
}

// LLMTokenUsage represents token usage information
type LLMTokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Role constants for message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)
