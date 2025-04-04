package sonnet

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cred.com/hack25/backend/pkg/llm/client"
	"cred.com/hack25/backend/pkg/logger"
)

// Client is an implementation of the LLMClient interface for Sonnet
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// SonnetCompletionRequest represents a request to the Sonnet API
type SonnetCompletionRequest struct {
	Model       string             `json:"model"`
	Messages    []SonnetMessage    `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float32            `json:"temperature,omitempty"`
	TopP        float32            `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// SonnetMessage represents a message in the Sonnet API
type SonnetMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SonnetCompletionResponse represents a response from the Sonnet API
type SonnetCompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []SonnetChoice `json:"choices"`
	Usage   SonnetUsage    `json:"usage"`
}

// SonnetChoice represents a choice in the Sonnet API response
type SonnetChoice struct {
	Index        int            `json:"index"`
	Message      SonnetMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

// SonnetUsage represents token usage in the Sonnet API response
type SonnetUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// SonnetStreamResponse represents a streaming response chunk from Sonnet
type SonnetStreamResponse struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []SonnetStreamChoice `json:"choices"`
}

// SonnetStreamChoice represents a choice in a streaming response
type SonnetStreamChoice struct {
	Index        int                 `json:"index"`
	Delta        SonnetStreamDelta   `json:"delta"`
	FinishReason *string             `json:"finish_reason"`
}

// SonnetStreamDelta represents delta content in a streaming response
type SonnetStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// NewClient creates a new Sonnet client
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.sonnet.ai/v1"
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 0}, // No timeout for streaming
	}
}

// Completion implements the Completion method of the LLMClient interface
func (c *Client) Completion(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error) {
	// Convert client messages to Sonnet messages
	messages := make([]SonnetMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = SonnetMessage{
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

	sonnetReq := SonnetCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(sonnetReq)
	if err != nil {
		logger.Errorf("Failed to marshal Sonnet request: %v", err)
		return nil, err
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("Failed to send request to Sonnet: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Sonnet API error: Status %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("sonnet API error: %s", resp.Status)
	}

	// Decode response
	var sonnetResp SonnetCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sonnetResp); err != nil {
		logger.Errorf("Failed to decode Sonnet response: %v", err)
		return nil, err
	}

	// Check if we have any choices
	if len(sonnetResp.Choices) == 0 {
		logger.Error("Sonnet returned no choices")
		return nil, errors.New("no completion choices returned")
	}

	// Return the response
	return &client.CompletionResponse{
		Text:         sonnetResp.Choices[0].Message.Content,
		FinishReason: sonnetResp.Choices[0].FinishReason,
		TokenUsage: &client.TokenUsage{
			PromptTokens:     sonnetResp.Usage.PromptTokens,
			CompletionTokens: sonnetResp.Usage.CompletionTokens,
			TotalTokens:      sonnetResp.Usage.TotalTokens,
		},
		ModelName: model,
	}, nil
}

// StreamCompletion implements the StreamCompletion method of the LLMClient interface
func (c *Client) StreamCompletion(ctx context.Context, req client.CompletionRequest, callback func(chunk string) error) error {
	// Convert client messages to Sonnet messages
	messages := make([]SonnetMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = SonnetMessage{
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

	sonnetReq := SonnetCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
		Stream:      true,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(sonnetReq)
	if err != nil {
		logger.Errorf("Failed to marshal Sonnet request: %v", err)
		return err
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to create HTTP request: %v", err)
		return err
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("Failed to send request to Sonnet: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Sonnet API error: Status %d, Body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("sonnet API error: %s", resp.Status)
	}

	// Process the stream
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Errorf("Error reading stream: %v", err)
			return err
		}

		// Skip empty lines
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE format: lines starting with "data: "
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		// Extract JSON data
		data := line[6:] // Skip "data: "
		
		// Check for stream end
		if string(data) == "[DONE]" {
			break
		}

		// Parse the JSON
		var streamResp SonnetStreamResponse
		if err := json.Unmarshal(data, &streamResp); err != nil {
			logger.Errorf("Failed to parse stream data: %v", err)
			continue
		}

		// Process each choice in the response
		for _, choice := range streamResp.Choices {
			if choice.Delta.Content != "" {
				if err := callback(choice.Delta.Content); err != nil {
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
		modelName = "sonnet-embedding-001"
	}

	// Prepare the request
	type EmbeddingRequest struct {
		Input string `json:"input"`
		Model string `json:"model"`
	}

	embReq := EmbeddingRequest{
		Input: text,
		Model: modelName,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(embReq)
	if err != nil {
		logger.Errorf("Failed to marshal Sonnet embedding request: %v", err)
		return nil, err
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/embeddings", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("Failed to send request to Sonnet: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Sonnet API error: Status %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("sonnet API error: %s", resp.Status)
	}

	// Decode response
	type EmbeddingResponse struct {
		Embedding []float32 `json:"embedding"`
		Usage     struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		logger.Errorf("Failed to decode Sonnet embedding response: %v", err)
		return nil, err
	}

	return &client.EmbeddingResponse{
		Embedding: embResp.Embedding,
		TokenUsage: &client.TokenUsage{
			PromptTokens:     embResp.Usage.PromptTokens,
			CompletionTokens: 0,
			TotalTokens:      embResp.Usage.TotalTokens,
		},
		ModelName: modelName,
	}, nil
}

// IsValidModel checks if the given model name is a valid Sonnet model
func IsValidModel(modelName string) bool {
	// Add more models as needed
	validModels := []string{
		"sonnet-3.5-turbo",
		"sonnet-4",
		"sonnet-embedding-001",
	}

	modelName = strings.TrimPrefix(modelName, "sonnet:")
	for _, m := range validModels {
		if m == modelName {
			return true
		}
	}
	return false
}
