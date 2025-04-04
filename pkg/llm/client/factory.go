package client

import (
	"errors"
	"fmt"
	"strings"
	
	"cred.com/hack25/backend/pkg/llm/gemini"
	"cred.com/hack25/backend/pkg/llm/openai"
	"cred.com/hack25/backend/pkg/llm/sonnet"
	"cred.com/hack25/backend/pkg/logger"
)

// Factory creates and manages LLM clients
type Factory struct {
	openaiClient *openai.Client
	geminiClient *gemini.Client
	sonnetClient *sonnet.Client
	// Other clients can be added here
}

// NewFactory creates a new LLM client factory
func NewFactory(config Config) (*Factory, error) {
	factory := &Factory{}

	// Initialize OpenAI client if configured
	if config.OpenAIAPIKey != "" {
		factory.openaiClient = openai.NewClient(config.OpenAIAPIKey)
		logger.Info("OpenAI client initialized")
	}

	// Initialize Gemini client if configured
	if config.GeminiAPIKey != "" {
		client, err := gemini.NewClient(config.GeminiAPIKey)
		if err != nil {
			logger.Errorf("Failed to initialize Gemini client: %v", err)
			return nil, err
		}
		factory.geminiClient = client
		logger.Info("Gemini client initialized")
	}

	// Initialize Sonnet client if configured
	if config.SonnetAPIKey != "" {
		factory.sonnetClient = sonnet.NewClient(config.SonnetAPIKey, config.SonnetBaseURL)
		logger.Info("Sonnet client initialized")
	}

	return factory, nil
}

// GetClient returns the appropriate LLM client for the given model
func (f *Factory) GetClient(modelName string) (LLMClient, error) {
	// Determine provider from model name
	provider := ""
	if strings.Contains(modelName, ":") {
		parts := strings.SplitN(modelName, ":", 2)
		provider = parts[0]
		modelName = parts[1]
	} else {
		// Try to infer provider from model name
		if strings.HasPrefix(modelName, "gpt-") {
			provider = "openai"
		} else if strings.HasPrefix(modelName, "gemini-") {
			provider = "google"
		} else if strings.HasPrefix(modelName, "sonnet-") {
			provider = "sonnet"
		}
	}

	// Return the appropriate client
	switch provider {
	case "openai":
		if f.openaiClient == nil {
			return nil, errors.New("openai client not initialized")
		}
		if !openai.IsValidModel(modelName) {
			return nil, fmt.Errorf("invalid OpenAI model: %s", modelName)
		}
		return f.openaiClient, nil
	case "google":
		if f.geminiClient == nil {
			return nil, errors.New("gemini client not initialized")
		}
		if !gemini.IsValidModel(modelName) {
			return nil, fmt.Errorf("invalid Gemini model: %s", modelName)
		}
		return f.geminiClient, nil
	case "sonnet":
		if f.sonnetClient == nil {
			return nil, errors.New("sonnet client not initialized")
		}
		if !sonnet.IsValidModel(modelName) {
			return nil, fmt.Errorf("invalid Sonnet model: %s", modelName)
		}
		return f.sonnetClient, nil
	default:
		return nil, fmt.Errorf("unsupported provider for model: %s", modelName)
	}
}

// Close closes all clients
func (f *Factory) Close() {
	if f.geminiClient != nil {
		f.geminiClient.Close()
	}
	// Add other cleanup as needed
}

// Config holds configuration for all LLM clients
type Config struct {
	OpenAIAPIKey  string
	GeminiAPIKey  string
	SonnetAPIKey  string
	SonnetBaseURL string
}
