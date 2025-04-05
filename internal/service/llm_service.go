package service

import (
	"context"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/llm/client"
	"cred.com/hack25/backend/pkg/llm/interfaces"
	"cred.com/hack25/backend/pkg/logger"
)

// LLMService handles interactions with LLM clients
type LLMService struct {
	clientFactory *client.Factory
	defaultModel  interfaces.Model
}

// ChatRequest represents a request to chat with an LLM
type ChatRequest struct {
	Messages    []interfaces.Message `json:"messages"`
	ModelName   string               `json:"model,omitempty"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Temperature *float32             `json:"temperature,omitempty"`
	TopP        *float32             `json:"top_p,omitempty"`
	Stream      bool                 `json:"stream,omitempty"`
}

// ChatResponse represents a response from chatting with an LLM
type ChatResponse struct {
	Text         string                 `json:"text"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	TokenUsage   *interfaces.TokenUsage `json:"token_usage,omitempty"`
	ModelName    string                 `json:"model,omitempty"`
}

// NewLLMService creates a new LLM service
func NewLLMService(clientFactory *client.Factory, defaultModelName string) *LLMService {
	defaultModel := interfaces.DefaultModels()["openai:gpt-3.5-turbo"] // Default fallback

	if model, ok := interfaces.DefaultModels()[defaultModelName]; ok {
		defaultModel = model
	}

	return &LLMService{
		clientFactory: clientFactory,
		defaultModel:  defaultModel,
	}
}

// Chat sends a chat request to the appropriate LLM
func (s *LLMService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Determine model to use
	modelName := s.defaultModel.Name
	provider := s.defaultModel.Provider

	if req.ModelName != "" {
		modelName = req.ModelName
		// Extract provider if provided in format "provider:model"
		if interfaces.IsValidModelWithProvider(req.ModelName) {
			parts := interfaces.SplitModelName(req.ModelName)
			provider = parts.Provider
			modelName = parts.Name
		}
	}

	// Get the model configuration
	model := interfaces.Model{
		Name:        modelName,
		Provider:    provider,
		MaxTokens:   s.defaultModel.MaxTokens,
		Temperature: s.defaultModel.Temperature,
		TopP:        s.defaultModel.TopP,
	}

	// If the model exists in our defaults, use those settings
	fullModelName := provider + ":" + modelName
	if defaultModel, ok := interfaces.DefaultModels()[fullModelName]; ok {
		model = defaultModel
	}

	// Get the appropriate client
	llmClient, err := s.clientFactory.GetClient(fullModelName)
	if err != nil {
		logger.Errorf("Failed to get LLM client: %v", err)
		return nil, err
	}

	// Convert the request to a client request
	clientReq := interfaces.CompletionRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	// Send the request
	resp, err := llmClient.Completion(ctx, clientReq)
	if err != nil {
		logger.Errorf("LLM completion error: %v", err)
		return nil, err
	}

	return &ChatResponse{
		Text:         resp.Text,
		FinishReason: resp.FinishReason,
		TokenUsage:   resp.TokenUsage,
		ModelName:    resp.ModelName,
	}, nil
}

// StreamChat streams a chat response from the appropriate LLM
func (s *LLMService) StreamChat(ctx context.Context, req ChatRequest, callback func(chunk string) error) error {
	// Determine model to use
	modelName := s.defaultModel.Name
	provider := s.defaultModel.Provider

	if req.ModelName != "" {
		modelName = req.ModelName
		// Extract provider if provided in format "provider:model"
		if interfaces.IsValidModelWithProvider(req.ModelName) {
			parts := interfaces.SplitModelName(req.ModelName)
			provider = parts.Provider
			modelName = parts.Name
		}
	}

	// Get the model configuration
	model := interfaces.Model{
		Name:        modelName,
		Provider:    provider,
		MaxTokens:   s.defaultModel.MaxTokens,
		Temperature: s.defaultModel.Temperature,
		TopP:        s.defaultModel.TopP,
	}

	// If the model exists in our defaults, use those settings
	fullModelName := provider + ":" + modelName
	if defaultModel, ok := interfaces.DefaultModels()[fullModelName]; ok {
		model = defaultModel
	}

	// Get the appropriate client
	llmClient, err := s.clientFactory.GetClient(fullModelName)
	if err != nil {
		logger.Errorf("Failed to get LLM client: %v", err)
		return err
	}

	// Convert the request to a client request
	clientReq := interfaces.CompletionRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      true,
	}

	// Send the streaming request
	err = llmClient.StreamCompletion(ctx, clientReq, callback)
	if err != nil {
		logger.Errorf("LLM stream completion error: %v", err)
		return err
	}

	return nil
}

// ChatWithModels sends a chat request using the models package types
func (s *LLMService) ChatWithModels(ctx context.Context, req models.LLMChatRequest) (*models.LLMChatResponse, error) {
	// Convert models.LLMMessage to interfaces.Message
	messages := make([]interfaces.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = interfaces.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create a ChatRequest from the LLMChatRequest
	chatReq := ChatRequest{
		Messages:    messages,
		ModelName:   req.ModelName,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	// Call the regular Chat method
	resp, err := s.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// Convert ChatResponse to models.LLMChatResponse
	return &models.LLMChatResponse{
		Text:         resp.Text,
		FinishReason: resp.FinishReason,
		ModelName:    resp.ModelName,
	}, nil
}

// GenerateEmbedding generates an embedding for the given text
func (s *LLMService) GenerateEmbedding(ctx context.Context, text string, modelName string) ([]float32, error) {
	// Use default embedding model if not specified
	if modelName == "" {
		modelName = "openai:text-embedding-ada-002"
	}

	// Get the appropriate client
	llmClient, err := s.clientFactory.GetClient(modelName)
	if err != nil {
		logger.Errorf("Failed to get LLM client for embedding: %v", err)
		return nil, err
	}

	// Determine the actual model name (without provider prefix)
	actualModelName := modelName
	if interfaces.IsValidModelWithProvider(modelName) {
		actualModelName = interfaces.SplitModelName(modelName).Name
	}

	// Generate the embedding
	resp, err := llmClient.Embedding(ctx, text, actualModelName)
	if err != nil {
		logger.Errorf("Failed to generate embedding: %v", err)
		return nil, err
	}

	return resp.Embedding, nil
}
