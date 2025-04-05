package litellm

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
	"time"

	"cred.com/hack25/backend/pkg/llm/interfaces"
	"cred.com/hack25/backend/pkg/logger"
)

// Client is an implementation of the LLMClient interface for LiteLLM
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new LiteLLM client
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.rabbithole.cred.club"
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Message represents a chat message in LiteLLM format
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a chat completion request to LiteLLM
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	TopP        float32   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage in a completion response
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionResponse represents a chat completion response from LiteLLM
type CompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// StreamResponse represents a streaming chat completion response
type StreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Completion implements the Completion method of the LLMClient interface
func (c *Client) Completion(ctx context.Context, req interfaces.CompletionRequest) (*interfaces.CompletionResponse, error) {
	// Convert messages to LiteLLM format
	messages := make([]Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = Message{
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

	// Create request payload
	completionReq := CompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(completionReq)
	if err != nil {
		logger.Errorf("Failed to marshal LiteLLM request: %v", err)
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"pass-through/anthropic_proxy_route_anthropic",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		logger.Errorf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("LiteLLM HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("LiteLLM API error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("LiteLLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var completionResp CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		logger.Errorf("Failed to decode LiteLLM response: %v", err)
		return nil, err
	}

	// Check if we have any choices
	if len(completionResp.Choices) == 0 {
		logger.Error("LiteLLM returned no choices")
		return nil, errors.New("no completion choices returned")
	}

	// Return the response
	return &interfaces.CompletionResponse{
		Text:         completionResp.Choices[0].Message.Content,
		FinishReason: completionResp.Choices[0].FinishReason,
		TokenUsage: &interfaces.TokenUsage{
			PromptTokens:     completionResp.Usage.PromptTokens,
			CompletionTokens: completionResp.Usage.CompletionTokens,
			TotalTokens:      completionResp.Usage.TotalTokens,
		},
		ModelName: model,
	}, nil
}

// StreamCompletion implements the StreamCompletion method of the LLMClient interface
func (c *Client) StreamCompletion(ctx context.Context, req interfaces.CompletionRequest, callback func(chunk string) error) error {
	// Convert messages to LiteLLM format
	messages := make([]Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = Message{
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

	// Create request payload
	completionReq := CompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
		Stream:      true,
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(completionReq)
	if err != nil {
		logger.Errorf("Failed to marshal LiteLLM stream request: %v", err)
		return err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"pass-through/anthropic_proxy_route_anthropic",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		logger.Errorf("Failed to create HTTP stream request: %v", err)
		return err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("LiteLLM HTTP stream request error: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("LiteLLM API stream error (status %d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("LiteLLM API stream error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read the response line by line
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

		// SSE format starts with "data: "
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		// Extract the JSON payload
		data := bytes.TrimPrefix(line, []byte("data: "))

		// Check for the stream end marker
		if string(data) == "[DONE]" {
			break
		}

		// Parse the chunk
		var streamResp StreamResponse
		if err := json.Unmarshal(data, &streamResp); err != nil {
			logger.Errorf("Failed to decode stream chunk: %v", err)
			continue
		}

		// Process the chunk
		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
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
func (c *Client) Embedding(ctx context.Context, text string, modelName string) (*interfaces.EmbeddingResponse, error) {
	// Create the embedding request
	type EmbeddingRequest struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}

	// Default to text-embedding-ada-002 if not specified
	if modelName == "" {
		modelName = "text-embedding-ada-002"
	}

	req := EmbeddingRequest{
		Model: modelName,
		Input: []string{text},
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(req)
	if err != nil {
		logger.Errorf("Failed to marshal LiteLLM embedding request: %v", err)
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"pass-through/anthropic_proxy_route_anthropic/embeddings",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		logger.Errorf("Failed to create HTTP embedding request: %v", err)
		return nil, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("LiteLLM HTTP embedding request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("LiteLLM API embedding error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("LiteLLM API embedding error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	type EmbeddingData struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	}

	type EmbeddingResponse struct {
		Object string          `json:"object"`
		Data   []EmbeddingData `json:"data"`
		Model  string          `json:"model"`
		Usage  struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		logger.Errorf("Failed to decode LiteLLM embedding response: %v", err)
		return nil, err
	}

	// Check if we have any data
	if len(embeddingResp.Data) == 0 {
		logger.Error("LiteLLM returned no embedding data")
		return nil, errors.New("no embedding data returned")
	}

	// Return the response
	return &interfaces.EmbeddingResponse{
		Embedding: embeddingResp.Data[0].Embedding,
		TokenUsage: &interfaces.TokenUsage{
			PromptTokens:     embeddingResp.Usage.PromptTokens,
			CompletionTokens: 0,
			TotalTokens:      embeddingResp.Usage.TotalTokens,
		},
		ModelName: modelName,
	}, nil
}

// IsValidModel checks if the given model name is registered with LiteLLM
// Note: LiteLLM proxy can route to various models, so we're not validating specific models
func IsValidModel(modelName string) bool {
	// LiteLLM proxy can route to many models, so we assume it's valid
	// In a production environment, you might want to verify against a list of supported models
	return true
}
