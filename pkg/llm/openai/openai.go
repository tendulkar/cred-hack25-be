package openai

import (
	"context"
	"errors"
	"strings"

	"cred.com/hack25/backend/pkg/llm/client"
	"cred.com/hack25/backend/pkg/logger"
	openai "github.com/sashabaranov/go-openai"
)

// Client is an implementation of the LLMClient interface for OpenAI
type Client struct {
	openaiClient *openai.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) *Client {
	return &Client{
		openaiClient: openai.NewClient(apiKey),
	}
}

// Completion implements the Completion method of the LLMClient interface
func (c *Client) Completion(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error) {
	// Convert client messages to OpenAI messages
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepare the request
	model := req.Model.Name
	maxTokens := req.Model.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	temperature := req.Model.Temperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	topP := req.Model.TopP
	if req.TopP != nil {
		topP = *req.TopP
	}

	completionReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temperature),
		TopP:        float32(topP),
	}

	// Send the request to OpenAI
	resp, err := c.openaiClient.CreateChatCompletion(ctx, completionReq)
	if err != nil {
		logger.Errorf("OpenAI completion error: %v", err)
		return nil, err
	}

	// Check if we have any choices
	if len(resp.Choices) == 0 {
		logger.Error("OpenAI returned no choices")
		return nil, errors.New("no completion choices returned")
	}

	// Return the response
	return &client.CompletionResponse{
		Text:         resp.Choices[0].Message.Content,
		FinishReason: resp.Choices[0].FinishReason,
		TokenUsage: &client.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		ModelName: model,
	}, nil
}

// StreamCompletion implements the StreamCompletion method of the LLMClient interface
func (c *Client) StreamCompletion(ctx context.Context, req client.CompletionRequest, callback func(chunk string) error) error {
	// Convert client messages to OpenAI messages
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepare the request
	model := req.Model.Name
	maxTokens := req.Model.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	temperature := req.Model.Temperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	topP := req.Model.TopP
	if req.TopP != nil {
		topP = *req.TopP
	}

	completionReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(temperature),
		TopP:        float32(topP),
		Stream:      true,
	}

	stream, err := c.openaiClient.CreateChatCompletionStream(ctx, completionReq)
	if err != nil {
		logger.Errorf("OpenAI stream completion error: %v", err)
		return err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, openai.ErrStreamClosed) {
			logger.Info("OpenAI stream closed")
			break
		}
		if err != nil {
			logger.Errorf("Error receiving from OpenAI stream: %v", err)
			return err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				if err := callback(content); err != nil {
					logger.Errorf("Error in stream callback: %v", err)
					return err
				}
			}
		}
	}

	return nil
}

// Embedding implements the Embedding method of the LLMClient interface
func (c *Client) Embedding(ctx context.Context, text string, modelName string) (*client.EmbeddingResponse, error) {
	if modelName == "" {
		modelName = openai.AdaEmbeddingV2
	}

	// Create the embedding request
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: modelName,
	}

	// Get the embedding from OpenAI
	resp, err := c.openaiClient.CreateEmbeddings(ctx, req)
	if err != nil {
		logger.Errorf("OpenAI embedding error: %v", err)
		return nil, err
	}

	// Check if we have any data
	if len(resp.Data) == 0 {
		logger.Error("OpenAI returned no embedding data")
		return nil, errors.New("no embedding data returned")
	}

	// Convert the embedding to []float32
	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	// Return the response
	return &client.EmbeddingResponse{
		Embedding: embedding,
		TokenUsage: &client.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: 0,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		ModelName: modelName,
	}, nil
}

// IsValidModel checks if the given model name is a valid OpenAI model
func IsValidModel(modelName string) bool {
	// Add more models as needed
	validModels := []string{
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo-preview",
	}

	modelName = strings.TrimPrefix(modelName, "openai:")
	for _, m := range validModels {
		if m == modelName {
			return true
		}
	}
	return false
}
