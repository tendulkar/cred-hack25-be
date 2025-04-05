# LiteLLM Client for Rabbithole API

This package provides an implementation of the LLMClient interface for interacting with Anthropic models through the Rabbithole API endpoint at CRED.

## Features

- Chat completions (both regular and streaming)
- Embeddings
- Integration with Anthropic Claude models

## Supported Models

- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-7-sonnet`

## Usage

### Configuration

Set the following environment variables:

```
LITELLM_API_KEY=your_api_key
LITELLM_BASE_URL=https://api.rabbithole.cred.club
LITELLM_DEFAULT_MODEL=claude-3-sonnet-20240229
```

### Using in Code

```go
import (
    "cred.com/hack25/backend/pkg/llm/interfaces"
    "cred.com/hack25/backend/pkg/llm/litellm"
)

// Create a client
client := litellm.NewClient(apiKey, baseURL)

// Create a completion request
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
        Content: "You are a helpful assistant.",
    },
    {
        Role:    interfaces.RoleUser,
        Content: "Hello, how are you?",
    },
}

req := interfaces.CompletionRequest{
    Model:       model,
    Messages:    messages,
    MaxTokens:   100,
    Temperature: &model.Temperature,
    TopP:        &model.TopP,
}

// Make the API call
resp, err := client.Completion(ctx, req)
if err != nil {
    // Handle error
}

// Process response
fmt.Println(resp.Text)
```

## Testing

The package includes unit tests that make real API calls to verify functionality. To run these tests, set the `LITELLM_API_KEY` environment variable:

```bash
export LITELLM_API_KEY=your_api_key
cd pkg/llm/litellm
go test -v
```

These tests will:
1. Make a completion request to the Anthropic Claude model
2. Test streaming capabilities
3. Generate embeddings

If the API key is not provided, the tests will be skipped.
