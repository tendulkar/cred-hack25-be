package structured

import (
	"os"
	"testing"

	"cred.com/hack25/backend/internal/config"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/pkg/database"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// Get database connection string from environment variable
	dbConnectionString := os.Getenv("TEST_DB_URL")
	if dbConnectionString == "" {
		t.Skip("Skipping test: TEST_DB_URL not set")
	}

	// Connect to the database
	db, err := sqlx.Connect("postgres", dbConnectionString)
	require.NoError(t, err, "Failed to connect to test database")

	return db
}

func setupService(t *testing.T) *Service {
	// Initialize logger
	logger.Init(logrus.InfoLevel, "llm-structured-test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	cfg.Database.Host = "localhost"
	cfg.Database.Port = "5432"
	cfg.Database.User = "code_analyser_user"
	cfg.Database.Password = "code_analyser_password"
	cfg.Database.DBName = "code_analyser"

	// Setup database connection
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		t.Fatal(err)
	}

	// Get LiteLLM configuration from environment
	liteLLMURL := os.Getenv("LITELLM_URL")
	if liteLLMURL == "" {
		liteLLMURL = "https://api.rabbithole.cred.club"
	}

	liteLLMAPIKey := os.Getenv("LITELLM_API_KEY")
	if liteLLMAPIKey == "" {
		liteLLMAPIKey = "sk-_ANTPTNsfl9XBVA5Q4jvyg" // Default key for testing
	}

	liteLLMModel := os.Getenv("LITELLM_DEFAULT_MODEL")
	if liteLLMModel == "" {
		liteLLMModel = "gpt-4o" // Default model for testing
	}

	if liteLLMURL == "" || liteLLMAPIKey == "" {
		t.Skip("Skipping test: LITELLM_URL, LITELLM_API_KEY, or LITELLM_DEFAULT_MODEL not set")
	}

	// Create service config
	config := ServiceConfig{
		LiteLLMBaseURL: liteLLMURL,
		APIKey:         liteLLMAPIKey,
		DefaultModel:   liteLLMModel,
		UseJSONFormat:  true,
	}

	// Create service with the real repository
	// Setup repository
	repo := repository.NewCodeAnalyzerRepository(db.Conn)
	logEntry := logrus.WithField("component", "llm-structured-test")
	service := NewService(repo, config, logEntry)
	return service
}

func TestGenerateFunctionInsight_RealAPI(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Skip if real API tests are disabled
	// if os.Getenv("ENABLE_REAL_API_TESTS") != "true" {
	// 	t.Skip("Skipping test that makes real API calls. Set ENABLE_REAL_API_TESTS=true to run.")
	// }

	// Setup service with real dependencies
	service := setupService(t)

	// Test parameters - using default model name (empty string)
	repoID := int64(1) // Replace with a real repository ID from your database
	modelName := ""    // Empty to use default model

	// Get a function from the database first to make sure we have valid IDs
	functions, err := service.codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
	if err != nil || len(functions) == 0 {
		t.Skipf("Skipping test: No functions found for repository ID %d: %v", repoID, err)
	}

	// Use the first function we find
	functionID := functions[0].ID
	t.Logf("Using function: ID=%d, Name=%s", functionID, functions[0].Name)

	// Call the method being tested
	t.Log("Calling GenerateFunctionInsight with real database and API...")
	insight, err := service.GenerateFunctionInsight(repoID, functionID, modelName)

	// Assert results
	require.NoError(t, err, "GenerateFunctionInsight should not return an error")
	require.NotNil(t, insight, "Insight should not be nil")

	// Validate the structure of the returned insight
	assert.NotEmpty(t, insight.Intent.Problem, "Intent.Problem should not be empty")
	assert.NotEmpty(t, insight.Intent.Goal, "Intent.Goal should not be empty")
	assert.NotEmpty(t, insight.Intent.Result, "Intent.Result should not be empty")

	// Check params
	assert.Greater(t, len(insight.Params), 0, "Function should have parameters")
	for i, param := range insight.Params {
		assert.NotEmpty(t, param.Name, "Parameter %d name should not be empty", i)
		assert.NotEmpty(t, param.Type, "Parameter %d type should not be empty", i)
	}

	// Check returns
	assert.Greater(t, len(insight.Returns), 0, "Function should have return values")

	// Log the complete insight for debugging
	t.Logf("Successfully generated function insight: %+v", insight)
}
