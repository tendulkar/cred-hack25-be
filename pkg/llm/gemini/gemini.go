package gemini

import (
	"context"
	"errors"
	"strings"

	"cred.com/hack25/backend/pkg/llm/client"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client is an implementation of the LLMClient interface for Google's Gemini
type Client struct {
	genaiClient *genai.Client
	apiKey      string
}

// NewClient creates a new Gemini client
func NewClient(apiKey string) (*Client, error) {
	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.Errorf("Failed to create Gemini client: %v", err)
		return nil, err
	}

	return &Client{
		genaiClient: genaiClient,
		apiKey:      apiKey,
	}, nil
}

// Completion implements the Completion method of the LLMClient interface
func (c *Client) Completion(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error) {
	// Create a new Gemini model instance
	model := c.genaiClient.GenerativeModel(req.Model.Name)

	// Set generation configuration
	temperature := req.Model.Temperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	topP := req.Model.TopP
	if req.TopP != nil {
		topP = *req.TopP
	}

	maxTokens := req.Model.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	model.SetTemperature(float32(temperature))
	model.SetTopP(float32(topP))
	model.SetMaxOutputTokens(int32(maxTokens))

	// Convert client messages to Gemini content
	var contents []*genai.Content
	for _, msg := range req.Messages {
		role := msg.Role
		// Map OpenAI roles to Gemini roles
		switch role {
		case "system":
			// Gemini doesn't have a system role, so treat it as user
			role = "user"
		case "assistant":
			role = "model"
		}

		content := &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		contents = append(contents, content)
	}

	// Generate completion
	resp, err := model.GenerateContent(ctx, contents...)
	if err != nil {
		logger.Errorf("Gemini completion error: %v", err)
		return nil, err
	}

	// Check if we have a response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		logger.Error("Gemini returned no content")
		return nil, errors.New("no content returned")
	}

	// Extract text from response
	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			text += string(textPart)
		}
	}

	// Gemini doesn't provide token usage, so we'll estimate
	estimatedPromptTokens := 0
	for _, msg := range req.Messages {
		estimatedPromptTokens += len(strings.Split(msg.Content, " ")) // Very rough estimate
	}
	estimatedCompletionTokens := len(strings.Split(text, " ")) // Very rough estimate

	return &client.CompletionResponse{
		Text:         text,
		FinishReason: string(resp.Candidates[0].FinishReason),
		TokenUsage: &client.TokenUsage{
			PromptTokens:     estimatedPromptTokens,
			CompletionTokens: estimatedCompletionTokens,
			TotalTokens:      estimatedPromptTokens + estimatedCompletionTokens,
		},
		ModelName: req.Model.Name,
	}, nil
}

// StreamCompletion implements the StreamCompletion method of the LLMClient interface
func (c *Client) StreamCompletion(ctx context.Context, req client.CompletionRequest, callback func(chunk string) error) error {
	// Create a new Gemini model instance
	model := c.genaiClient.GenerativeModel(req.Model.Name)

	// Set generation configuration
	temperature := req.Model.Temperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	topP := req.Model.TopP
	if req.TopP != nil {
		topP = *req.TopP
	}

	maxTokens := req.Model.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	model.SetTemperature(float32(temperature))
	model.SetTopP(float32(topP))
	model.SetMaxOutputTokens(int32(maxTokens))

	// Convert client messages to Gemini content
	var contents []*genai.Content
	for _, msg := range req.Messages {
		role := msg.Role
		// Map OpenAI roles to Gemini roles
		switch role {
		case "system":
			// Gemini doesn't have a system role, so treat it as user
			role = "user"
		case "assistant":
			role = "model"
		}

		content := &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		contents = append(contents, content)
	}

	// Generate streaming completion
	iter := model.GenerateContentStream(ctx, contents...)
	
	for {
		resp, err := iter.Next()
		if err != nil {
			if err.Error() == "stop iteration" {
				// Stream completed successfully
				break
			}
			logger.Errorf("Error streaming from Gemini: %v", err)
			return err
		}

		// Check if we have content
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			continue
		}

		// Extract text from response
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				if err := callback(string(textPart)); err != nil {
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
	// Gemini doesn't natively support embeddings in the same way as OpenAI
	// This is a placeholder implementation
	return nil, errors.New("embeddings not supported by Gemini client")
}

// IsValidModel checks if the given model name is a valid Gemini model
func IsValidModel(modelName string) bool {
	// Add more models as needed
	validModels := []string{
		"gemini-pro",
		"gemini-pro-vision",
	}

	modelName = strings.TrimPrefix(modelName, "google:")
	for _, m := range validModels {
		if m == modelName {
			return true
		}
	}
	return false
}

// Close closes the Gemini client
func (c *Client) Close() {
	if c.genaiClient != nil {
		c.genaiClient.Close()
	}
}
