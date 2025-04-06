package repointel

import (
	"os"
	"testing"

	"cred.com/hack25/backend/internal/config"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/pkg/database"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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
	logger.Init(logrus.InfoLevel, "repointel-test")
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

	// Setup repository
	repo := repository.NewCodeAnalyzerRepository(db.Conn)

	// Get LiteLLM configuration from environment
	liteLLMURL := "https://api.rabbithole.cred.club"
	liteLLMAPIKey := "sk-_ANTPTNsfl9XBVA5Q4jvyg"
	// liteLLMModel := "claude-3-7-sonnet"
	liteLLMModel := "claude-3-7-sonnet-20250219"

	if liteLLMURL == "" || liteLLMAPIKey == "" {
		t.Skip("Skipping test: LITELLM_URL, LITELLM_API_KEY, or LITELLM_DEFAULT_MODEL not set")
	}

	// Create service
	service := NewService(repo, liteLLMURL, liteLLMAPIKey, liteLLMModel)
	return service
}

func TestGenerateFunctionInsight_RealAPI(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Setup service with real dependencies
	service := setupService(t)

	// Test parameters
	// Note: You need to have a real repository and function in your test database
	// Replace these values with actual IDs from your test database
	repoID := int64(1)     // Replace with a real repository ID
	functionID := int64(1) // Replace with a real function ID
	modelName := ""        // Empty to use default model

	// Get a function from the database first to make sure we have valid IDs
	functions, err := service.codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
	if err != nil || len(functions) == 0 {
		t.Skipf("Skipping test: No functions found for repository ID %d: %v", repoID, err)
	}

	// Use the first function we find
	functionID = functions[0].ID

	// Call the method being tested
	insight, err := service.GenerateFunctionInsight(repoID, functionID, modelName)
	t.Logf("Generated insight: %+v", insight)
	// Assert
	require.NoError(t, err, "Should not return an error")
	require.NotNil(t, insight, "Should return an insight")

}

// Optional: Add a more comprehensive test that verifies database operations
// func TestGenerateFunctionInsight_FullIntegration(t *testing.T) {
// 	// Skip if running in CI environment
// 	if os.Getenv("CI") == "true" {
// 		t.Skip("Skipping test in CI environment")
// 	}

// 	// Setup service with real dependencies
// 	service := setupService(t)

// 	// Create a test repository if needed
// 	// (You may want to use an existing repository instead)
// 	repo := &models.Repository{
// 		Name:        "test-repo-for-insight",
// 		URL:         "https://github.com/test/repo",
// 	}

// 	repoID, err := service.codeAnalyzerRepo.CreateRepository(repo)
// 	if err != nil {
// 		t.Skipf("Skipping test: Failed to create test repository: %v", err)
// 	}

// 	t.Cleanup(func() {
// 		// Clean up test data (if desired)
// 		// Note: You might want to keep the test data for debugging
// 		// service.codeAnalyzerRepo.DeleteRepository(repoID)
// 	})

// 	// Get functions for the repository
// 	functions, err := service.codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
// 	if err != nil || len(functions) == 0 {
// 		t.Skipf("Skipping test: No functions found for repository ID %d: %v", repoID, err)
// 	}

// 	// Use the first function
// 	functionID := functions[0].ID
// 	modelName := "" // Use default model

// 	// Call the method being tested
// 	insight, err := service.GenerateFunctionInsight(repoID, functionID, modelName)

// 	// Assertions
// 	require.NoError(t, err, "Should not return an error")
// 	require.NotNil(t, insight, "Should return an insight")

// 	// Check insight content
// 	assert.NotEmpty(t, insight.Purpose, "Purpose should not be empty")
// 	assert.NotEmpty(t, insight.SystemInteractions, "SystemInteractions should not be empty")

// 	// Verify insight was saved to database
// 	savedInsights, err := service.repository.GetFunctionInsights(repoID, functionID)
// 	require.NoError(t, err, "Should retrieve saved insights")
// 	assert.GreaterOrEqual(t, len(savedInsights), 1, "Should have at least one saved insight")
// }
