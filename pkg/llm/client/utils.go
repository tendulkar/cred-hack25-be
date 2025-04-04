package client

import (
	"strings"
)

// ModelDetails holds the provider and name of a model
type ModelDetails struct {
	Provider string
	Name     string
}

// IsValidModelWithProvider checks if a model string has a valid provider prefix
func IsValidModelWithProvider(modelName string) bool {
	return strings.Contains(modelName, ":")
}

// SplitModelName splits a model name with provider (e.g., "openai:gpt-4") into provider and name
func SplitModelName(modelName string) ModelDetails {
	if !IsValidModelWithProvider(modelName) {
		// If no provider is specified, try to infer it
		if strings.HasPrefix(modelName, "gpt-") {
			return ModelDetails{Provider: "openai", Name: modelName}
		} else if strings.HasPrefix(modelName, "gemini-") {
			return ModelDetails{Provider: "google", Name: modelName}
		} else if strings.HasPrefix(modelName, "sonnet-") {
			return ModelDetails{Provider: "sonnet", Name: modelName}
		}
		// Default to returning as-is
		return ModelDetails{Provider: "", Name: modelName}
	}

	parts := strings.SplitN(modelName, ":", 2)
	return ModelDetails{
		Provider: parts[0],
		Name:     parts[1],
	}
}
