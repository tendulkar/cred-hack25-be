package interfaces

import (
	"context"
	"strings"
)

// Model represents an LLM model with its configuration
type Model struct {
	// Name is the identifier of the model (e.g., "gpt-4", "gemini-pro")
	Name string
	// Provider is the provider of the model (e.g., "openai", "google", "sonnet")
	Provider string
	// MaxTokens is the maximum number of tokens the model can process
	MaxTokens int
	// Temperature controls randomness (0-1, higher is more random)
	Temperature float32
	// TopP controls diversity of generated text (0-1)
	TopP float32
}

// Message represents a message in a conversation
type Message struct {
	// Role is the sender of the message (system, user, assistant)
	Role string
	// Content is the text content of the message
	Content string
}

// CompletionRequest represents a request for a text completion
type CompletionRequest struct {
	// Model is the model to use for generation
	Model Model
	// Messages is the conversation history
	Messages []Message
	// MaxTokens is the maximum number of tokens to generate (overrides model.MaxTokens if set)
	MaxTokens int
	// Temperature controls randomness (overrides model.Temperature if set)
	Temperature *float32
	// TopP controls diversity (overrides model.TopP if set)
	TopP *float32
	// Stream indicates whether to stream the response
	Stream bool
}

// CompletionResponse represents a response from a completion request
type CompletionResponse struct {
	// Text is the generated text
	Text string
	// FinishReason describes why the generation stopped
	FinishReason string
	// TokenUsage contains token usage information
	TokenUsage *TokenUsage
	// ModelName is the name of the model that generated the response
	ModelName string
}

// TokenUsage tracks token usage for billing and rate limiting
type TokenUsage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int
	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int
	// TotalTokens is the total number of tokens used
	TotalTokens int
}

// EmbeddingResponse represents a response from an embedding request
type EmbeddingResponse struct {
	// Embedding is the vector representation of the input
	Embedding []float32
	// TokenUsage contains token usage information
	TokenUsage *TokenUsage
	// ModelName is the name of the model that generated the embedding
	ModelName string
}

// LLMClient is the interface that all LLM clients must implement
type LLMClient interface {
	// Completion generates text based on the provided messages
	Completion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// StreamCompletion streams the completion response
	StreamCompletion(ctx context.Context, req CompletionRequest, callback func(chunk string) error) error

	// Embedding generates an embedding vector for the provided text
	Embedding(ctx context.Context, text string, modelName string) (*EmbeddingResponse, error)
}

// DefaultModels returns default models for different providers
func DefaultModels() map[string]Model {
	return map[string]Model{
		"openai:gpt-3.5-turbo": {
			Name:        "gpt-3.5-turbo",
			Provider:    "openai",
			MaxTokens:   4096,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"openai:gpt-4": {
			Name:        "gpt-4",
			Provider:    "openai",
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"google:gemini-pro": {
			Name:        "gemini-pro",
			Provider:    "google",
			MaxTokens:   4096,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"sonnet:sonnet-3.5-turbo": {
			Name:        "sonnet-3.5-turbo",
			Provider:    "sonnet",
			MaxTokens:   4096,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"litellm:claude-3-opus-20240229": {
			Name:        "claude-3-opus-20240229",
			Provider:    "litellm",
			MaxTokens:   10000,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"litellm:claude-3-sonnet-20240229": {
			Name:        "claude-3-sonnet-20240229",
			Provider:    "litellm",
			MaxTokens:   8000,
			Temperature: 0.7,
			TopP:        1.0,
		},
		"litellm:claude-3-7-sonnet": {
			Name:        "claude-3-7-sonnet",
			Provider:    "litellm",
			MaxTokens:   12000,
			Temperature: 0.7,
			TopP:        1.0,
		},
	}
}

// Role constants for message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ModelInfo contains provider and name information
type ModelInfo struct {
	Provider string
	Name     string
}

// IsValidModelWithProvider checks if a model name includes a provider
func IsValidModelWithProvider(modelName string) bool {
	return strings.Contains(modelName, ":")
}

// SplitModelName splits a model name into provider and name
func SplitModelName(modelName string) ModelInfo {
	parts := strings.SplitN(modelName, ":", 2)
	if len(parts) != 2 {
		return ModelInfo{
			Provider: "",
			Name:     modelName,
		}
	}
	return ModelInfo{
		Provider: parts[0],
		Name:     parts[1],
	}
}
