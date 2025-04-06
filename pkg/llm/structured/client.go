package structured

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ResponseFormat represents the format specification for the LLM response
type ResponseFormat struct {
	// Type should be "json_object" for structured JSON output
	Type string `json:"type"`
}

// LiteLLMRequest represents a request to the LiteLLM API
type LiteLLMRequest struct {
	Model          string           `json:"model"`
	Messages       []LiteLLMMessage `json:"messages"`
	MaxTokens      int              `json:"max_tokens,omitempty"`
	Stream         bool             `json:"stream"`
	ResponseFormat *ResponseFormat  `json:"response_format,omitempty"`
}

// LiteLLMMessage represents a message in the LiteLLM API request
type LiteLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LiteLLMResponse represents a response from the LiteLLM API
type LiteLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Client is a simplified LLM client for structured output generation
type Client struct {
	baseURL      string
	apiKey       string
	defaultModel string
	httpClient   *http.Client
	logger       *logrus.Entry
}

// NewClient creates a new LLM client
func NewClient(baseURL, apiKey, defaultModel string, logger *logrus.Entry) *Client {
	return &Client{
		baseURL:      baseURL,
		apiKey:       apiKey,
		defaultModel: defaultModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // 2 minute timeout
		},
		logger: logger,
	}
}

// Call makes a request to the LLM API with the given prompt
func (c *Client) Call(model, prompt string, schema json.RawMessage) (string, error) {
	if model == "" {
		model = c.defaultModel
	}

	// Prepare the request body
	requestBody := LiteLLMRequest{
		Model: model,
		Messages: []LiteLLMMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 8192, // Adjust as needed for your use case
		Stream:    false,
	}

	// Add response format for JSON output if schema is provided
	if len(schema) > 0 {
		// According to OpenAI documentation, just set the type to json_object
		// The schema should be included in the prompt itself
		requestBody.ResponseFormat = &ResponseFormat{
			Type: "json_object",
		}
		
		// Extract a readable version of the schema to include in the prompt
		var schemaObj map[string]interface{}
		if err := json.Unmarshal(schema, &schemaObj); err == nil {
			prettySchema, err := json.MarshalIndent(schemaObj, "", "  ")
			if err == nil {
				// Add the schema to the prompt to guide the model
				schemaPrompt := "\n\nYou must respond with a JSON object that conforms to this schema:\n```json\n" + 
					string(prettySchema) + "\n```\n\nThe response should be a valid JSON object with no other text."
				prompt += schemaPrompt
				requestBody.Messages[0].Content = prompt
			}
		}
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Log request
	c.logger.WithFields(logrus.Fields{
		"model":  model,
		"url":    c.baseURL,
		"tokens": len(prompt) / 4, // Rough estimate
	}).Info("Sending request to LLM API")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if the response is OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var llmResponse LiteLLMResponse
	if err := json.Unmarshal(bodyBytes, &llmResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors in the response
	if llmResponse.Error != nil {
		return "", fmt.Errorf("LLM API returned error: %s", llmResponse.Error.Message)
	}

	// Ensure we have a valid response
	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("LLM API returned empty choices")
	}

	// Extract the content
	content := llmResponse.Choices[0].Message.Content

	// Log response size
	c.logger.WithFields(logrus.Fields{
		"model":        model,
		"response_len": len(content),
	}).Info("Received response from LLM API")

	return content, nil
}
