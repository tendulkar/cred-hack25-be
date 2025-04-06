package config

import (
	"os"
	"strconv"
	"time"

	"cred.com/hack25/backend/pkg/database"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	Environment string
	Server      ServerConfig
	Database    database.Config
	JWT         JWTConfig
	LLM         LLMConfig
	LogLevel    logrus.Level
	LogFile     string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	SigningAlgorithm string
}

// Load loads the application configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set default values
	config := &Config{
		Environment: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "6060"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 10)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 120)) * time.Second,
		},
		Database: database.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "code_analyser"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-secret-key"),
			AccessTokenTTL:   time.Duration(getEnvAsInt("JWT_ACCESS_TOKEN_TTL", 15)) * time.Minute,
			RefreshTokenTTL:  time.Duration(getEnvAsInt("JWT_REFRESH_TOKEN_TTL", 24*7)) * time.Hour,
			SigningAlgorithm: getEnv("JWT_SIGNING_ALGORITHM", "HS256"),
		},
		LLM: LLMConfig{
			DefaultModelName: getEnv("LLM_DEFAULT_MODEL", "openai:gpt-3.5-turbo"),
			OpenAI: OpenAIConfig{
				APIKey:       getEnv("OPENAI_API_KEY", ""),
				DefaultModel: getEnv("OPENAI_DEFAULT_MODEL", "gpt-3.5-turbo"),
			},
			Gemini: GeminiConfig{
				APIKey:       getEnv("GEMINI_API_KEY", ""),
				DefaultModel: getEnv("GEMINI_DEFAULT_MODEL", "gemini-pro"),
			},
			Sonnet: SonnetConfig{
				APIKey:       getEnv("SONNET_API_KEY", ""),
				BaseURL:      getEnv("SONNET_BASE_URL", "https://api.sonnet.ai/v1"),
				DefaultModel: getEnv("SONNET_DEFAULT_MODEL", "sonnet-3.5-pro"),
			},
			LiteLLM: LiteLLMConfig{
				APIKey:       getEnv("LITELLM_API_KEY", "sk-_ANTPTNsfl9XBVA5Q4jvyg"),
				BaseURL:      getEnv("LITELLM_BASE_URL", "https://api.rabbithole.cred.club"),
				DefaultModel: getEnv("LITELLM_DEFAULT_MODEL", "gpt-4o"),
			},
		},
		LogLevel: getLogLevel(getEnv("LOG_LEVEL", "info")),
		LogFile:  getEnv("LOG_FILE", ""),
	}

	// Initialize logger
	logger.Init(config.LogLevel, config.LogFile)

	logger.WithFields(logger.Fields{
		"environment": config.Environment,
		"server_port": config.Server.Port,
		"db_host":     config.Database.Host,
		"db_name":     config.Database.DBName,
		"log_level":   config.LogLevel.String(),
	}).Info("Configuration loaded")

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getLogLevel converts a string log level to a logrus.Level
func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
