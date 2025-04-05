package litellm

import (
	"context"
	"os"
	"testing"
	"time"

	"cred.com/hack25/backend/pkg/llm/interfaces"
)

// TestLiteLLMRealAPICall tests real API calls to the Anthropic models via LiteLLM
// This test will only run if LITELLM_API_KEY environment variable is set
func TestLiteLLMRealAPICall(t *testing.T) {
	// Skip test if no API key is provided
	apiKey := os.Getenv("LITELLM_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: LITELLM_API_KEY environment variable not set")
	}

	// Create a new LiteLLM client with the Rabbithole API endpoint
	baseURL := "https://api.rabbithole.cred.club"
	client := NewClient(apiKey, baseURL)

	// Set a timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a completion request using the default Claude model
	model := interfaces.Model{
		Name:        "claude-3-sonnet-20240229",
		Provider:    "litellm",
		MaxTokens:   8000,
		Temperature: 0.7,
		TopP:        1.0,
	}

	messages := []interfaces.Message{
		{
			Role:    interfaces.RoleSystem,
			Content: "You are a helpful assistant. Keep your responses concise.",
		},
		{
			Role:    interfaces.RoleUser,
			Content: "What time is it?",
		},
	}

	req := interfaces.CompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   100, // Keep response short for testing
		Temperature: &model.Temperature,
		TopP:        &model.TopP,
	}

	// Make the API call
	t.Log("Making real API call to Anthropic model via LiteLLM...")
	resp, err := client.Completion(ctx, req)
	if err != nil {
		t.Fatalf("Error making completion request: %v", err)
	}

	// Check the response
	if resp.Text == "" {
		t.Error("Empty response received")
	} else {
		t.Logf("Response: %s", resp.Text)
	}

	if resp.TokenUsage == nil {
		t.Error("No token usage information received")
	} else {
		t.Logf("Token usage - Prompt: %d, Completion: %d, Total: %d",
			resp.TokenUsage.PromptTokens,
			resp.TokenUsage.CompletionTokens,
			resp.TokenUsage.TotalTokens)
	}

	// Test streaming completion
	t.Log("Testing streaming completion...")
	chunks := []string{}
	streamErr := client.StreamCompletion(ctx, req, func(chunk string) error {
		chunks = append(chunks, chunk)
		t.Logf("Received chunk: %s", chunk)
		return nil
	})

	if streamErr != nil {
		t.Fatalf("Error in streaming completion: %v", streamErr)
	}

	if len(chunks) == 0 {
		t.Error("No streaming chunks received")
	} else {
		t.Logf("Received %d streaming chunks", len(chunks))
	}

	// Test embeddings
	t.Log("Testing embeddings...")
	embeddingText := "This is a test for embedding generation"
	embeddingResp, err := client.Embedding(ctx, embeddingText, "text-embedding-ada-002")
	if err != nil {
		t.Fatalf("Error generating embedding: %v", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		t.Error("Empty embedding received")
	} else {
		t.Logf("Received embedding with %d dimensions", len(embeddingResp.Embedding))
	}
}
