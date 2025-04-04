package config

// LLMConfig contains configuration for the LLM clients
type LLMConfig struct {
	DefaultModelName string `env:"LLM_DEFAULT_MODEL" envDefault:"openai:gpt-3.5-turbo"`
	OpenAI           OpenAIConfig
	Gemini           GeminiConfig
	Sonnet           SonnetConfig
}

// OpenAIConfig contains configuration for the OpenAI client
type OpenAIConfig struct {
	APIKey      string `env:"OPENAI_API_KEY"`
	DefaultModel string `env:"OPENAI_DEFAULT_MODEL" envDefault:"gpt-3.5-turbo"`
}

// GeminiConfig contains configuration for the Gemini client
type GeminiConfig struct {
	APIKey      string `env:"GEMINI_API_KEY"`
	DefaultModel string `env:"GEMINI_DEFAULT_MODEL" envDefault:"gemini-pro"`
}

// SonnetConfig contains configuration for the Sonnet client
type SonnetConfig struct {
	APIKey      string `env:"SONNET_API_KEY"`
	BaseURL     string `env:"SONNET_BASE_URL" envDefault:"https://api.sonnet.ai/v1"`
	DefaultModel string `env:"SONNET_DEFAULT_MODEL" envDefault:"sonnet-3.5-pro"`
}
