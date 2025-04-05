package client

import (
	"cred.com/hack25/backend/pkg/llm/interfaces"
)

// ModelDetails is an alias for interfaces.ModelInfo
type ModelDetails = interfaces.ModelInfo

// For backward compatibility, keep these functions but make them call the interfaces versions
// These functions are deprecated and will be removed in a future version

// GetModelDetails is a backward compatibility function that calls SplitModelName
func GetModelDetails(modelName string) ModelDetails {
	return interfaces.SplitModelName(modelName)
}

// InferProviderFromModel tries to infer the provider from a model name
func InferProviderFromModel(modelName string) string {
	if interfaces.IsValidModelWithProvider(modelName) {
		return interfaces.SplitModelName(modelName).Provider
	}

	// Try to infer provider from model name
	if modelName == "" {
		return ""
	}

	switch {
	case len(modelName) >= 4 && modelName[:4] == "gpt-":
		return "openai"
	case len(modelName) >= 7 && modelName[:7] == "gemini-":
		return "google"
	case len(modelName) >= 7 && modelName[:7] == "sonnet-":
		return "sonnet"
	default:
		return ""
	}
}
