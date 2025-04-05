package client

import (
	"cred.com/hack25/backend/pkg/llm/interfaces"
)

// Re-export types from interfaces package
type (
	Model              = interfaces.Model
	Message            = interfaces.Message
	CompletionRequest  = interfaces.CompletionRequest
	CompletionResponse = interfaces.CompletionResponse
	TokenUsage         = interfaces.TokenUsage
	EmbeddingResponse  = interfaces.EmbeddingResponse
	LLMClient          = interfaces.LLMClient
	ModelInfo          = interfaces.ModelInfo
)

// Re-export constants from interfaces package
const (
	RoleSystem    = interfaces.RoleSystem
	RoleUser      = interfaces.RoleUser
	RoleAssistant = interfaces.RoleAssistant
)

// DefaultModels returns default models for different providers
func DefaultModels() map[string]Model {
	return interfaces.DefaultModels()
}

// IsValidModelWithProvider checks if a model name includes a provider
func IsValidModelWithProvider(modelName string) bool {
	return interfaces.IsValidModelWithProvider(modelName)
}

// SplitModelName splits a model name into provider and name
func SplitModelName(modelName string) ModelInfo {
	return interfaces.SplitModelName(modelName)
}
