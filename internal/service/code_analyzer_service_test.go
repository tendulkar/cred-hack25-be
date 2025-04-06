package service

import (
	"encoding/json"
	"os"
	"testing"

	"cred.com/hack25/backend/internal/config"
	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/pkg/database"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
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

func setupCodeAnalyzerService(t *testing.T) *CodeAnalyzerService {
	// Initialize logger
	logger.Init(logrus.InfoLevel, "code-analyzer-test")

	// Load configuration or use defaults
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	// Set database configuration if not loaded from config
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

	// Get LiteLLM configuration
	liteLLMURL := os.Getenv("LITELLM_URL")
	if liteLLMURL == "" {
		liteLLMURL = "https://api.rabbithole.cred.club"
	}

	liteLLMAPIKey := os.Getenv("LITELLM_API_KEY")
	if liteLLMAPIKey == "" {
		liteLLMAPIKey = "dummy-key-for-testing"
	}

	liteLLMModel := os.Getenv("LITELLM_DEFAULT_MODEL")
	if liteLLMModel == "" {
		liteLLMModel = "gpt-4o"
	}

	// Create service
	tempDir := "/tmp/code-analyzer-test"
	service := NewCodeAnalyzerService(repo, tempDir, liteLLMURL, liteLLMAPIKey, liteLLMModel, nil)
	return service
}

func TestGetRepositoryIndex_RealDB(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Setup service with real dependencies
	service := setupCodeAnalyzerService(t)

	// Get a repository from the database
	// Note: Make sure you have at least one repository in your test database
	repo, err := service.repo.GetRepositoryByURL("https://github.com/zorig/gopeerflix")
	if err != nil || repo == nil {
		t.Skipf("Skipping test: No repositories found in database: %v", err)
		return
	}

	// Use the first repository
	testRepo := repo
	t.Logf("Testing with repository: %s (ID: %d)", testRepo.URL, testRepo.ID)

	// Test 1: Get entire repository index
	t.Run("GetEntireRepositoryIndex", func(t *testing.T) {
		response, err := service.GetRepositoryIndex(testRepo.URL, "")
		assert.NoError(t, err, "Should not return an error")
		responseJson, err := json.MarshalIndent(response, "", "  ")
		t.Logf("Repository index response: %s", string(responseJson))

		// Check for errors
		require.NoError(t, err, "Should not return an error")
		require.NotNil(t, response, "Should return a response")

		// Basic validation
		assert.Equal(t, testRepo.ID, response.Repository.ID, "Repository ID should match")
		assert.NotEmpty(t, response.Files, "Should return files (legacy field)")
		assert.NotNil(t, response.IndexedFilesMap, "Should have IndexedFilesMap")

		// Check the hierarchical structure
		if len(response.Files) > 0 {
			// At least one file should be in the map
			assert.GreaterOrEqual(t, len(response.IndexedFilesMap), 1, "Should have at least one file in IndexedFilesMap")

			// Verify that file paths match
			for _, file := range response.Files {
				// The file path should be a key in the map
				indexedFile, exists := response.IndexedFilesMap[file.FilePath]
				assert.True(t, exists, "File %s should exist in IndexedFilesMap", file.FilePath)

				if exists {
					// The file in the map should match the file in the legacy field
					assert.Equal(t, file.ID, indexedFile.File.ID, "File IDs should match")
					assert.Equal(t, file.FilePath, indexedFile.File.FilePath, "File paths should match")

					// Check functions if they're available
					if len(indexedFile.Functions) > 0 {
						t.Logf("File %s has %d functions", file.FilePath, len(indexedFile.Functions))

						// Pick the first function ID and check its structure
						var firstFuncID int64
						for id := range indexedFile.Functions {
							firstFuncID = id
							break
						}

						// Check function structure
						indexedFunc := indexedFile.Functions[firstFuncID]
						assert.NotNil(t, indexedFunc.RepoFunction, "Function should not be nil")
						assert.Equal(t, firstFuncID, indexedFunc.RepoFunction.ID, "Function ID should match")

						// Check if insights are populated (if available)
						if indexedFunc.Insights != nil {
							t.Logf("Function has insights: %+v", indexedFunc.Insights)
						}
					}
				}
			}
		}
	})

	// Test 2: Get index for a specific file if files exist
	repoFiles, err := service.repo.GetRepositoryFiles(testRepo.ID)
	if err != nil {
		t.Fatalf("Failed to get repository files: %v", err)
	}
	if len(repoFiles) > 0 {
		t.Run("GetSpecificFileIndex", func(t *testing.T) {
			// Get the first file for this repository
			testFile := repoFiles[0]
			t.Logf("Testing with file: %s (ID: %d)", testFile.FilePath, testFile.ID)

			// Get index for this specific file
			response, err := service.GetRepositoryIndex(testRepo.URL, testFile.FilePath)

			// Check for errors
			require.NoError(t, err, "Should not return an error")
			require.NotNil(t, response, "Should return a response")

			// Basic validation
			assert.Equal(t, testRepo.ID, response.Repository.ID, "Repository ID should match")
			assert.Len(t, response.Files, 1, "Should return exactly one file (legacy field)")
			assert.Equal(t, testFile.ID, response.Files[0].ID, "File ID should match")

			// Check the file in the map
			indexedFile, exists := response.IndexedFilesMap[testFile.FilePath]
			assert.True(t, exists, "File should exist in IndexedFilesMap")

			if exists {
				assert.Equal(t, testFile.ID, indexedFile.File.ID, "File ID should match")

				// Check functions and symbols
				assert.NotNil(t, indexedFile.Functions, "Functions map should not be nil")
				assert.NotNil(t, indexedFile.Symbols, "Symbols map should not be nil")

				// Check if functions in legacy field match those in the map
				assert.Equal(t, len(response.Functions), len(indexedFile.Functions),
					"Number of functions should match between legacy and new structure")

				// Check if symbols in legacy field match those in the map
				assert.Equal(t, len(response.Symbols), len(indexedFile.Symbols),
					"Number of symbols should match between legacy and new structure")

				// Print insights metrics if available
				t.Logf("File has %d functions with insights", countFunctionsWithInsights(indexedFile))
			}
		})
	}
}

// Helper function to count functions with insights
func countFunctionsWithInsights(indexedFile *models.IndexedFile) int {
	count := 0
	for _, fn := range indexedFile.Functions {
		if fn.Insights != nil {
			count++
		}
	}
	return count
}
