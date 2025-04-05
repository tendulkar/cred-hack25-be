package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/github"
	"cred.com/hack25/backend/pkg/logger"
)

// CodeAnalysisService handles code analysis operations
type CodeAnalysisService struct {
	githubClient github.Client
	llmService   *LLMService
}

// NewCodeAnalysisService creates a new code analysis service
func NewCodeAnalysisService(llmService *LLMService) *CodeAnalysisService {
	return &CodeAnalysisService{
		llmService: llmService,
	}
}

// AnalyzeRepository analyzes a GitHub repository
func (s *CodeAnalysisService) AnalyzeRepository(ctx context.Context, repoURL, authToken string) (*models.RepositoryAnalysisResult, error) {
	// Create GitHub client if not already created
	if s.githubClient == nil {
		s.githubClient = github.NewClient(authToken)
	}

	// Fetch repository
	logger.Infof("Fetching repository: %s", repoURL)
	repo, err := s.githubClient.FetchRepository(repoURL, authToken)
	if err != nil {
		logger.Errorf("Failed to fetch repository: %v", err)
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	logger.Infof("Successfully fetched repository: %s/%s", repo.Owner, repo.Name)

	// Create result object
	result := &models.RepositoryAnalysisResult{
		RepoURL: repoURL,
		Owner:   repo.Owner,
		Name:    repo.Name,
		Files:   []models.FileAnalysisResult{},
	}

	// Get all file paths
	filePaths := repo.GetAllFilePaths()

	// Build directory tree for post-order traversal
	dirTree := buildDirectoryTree(filePaths)

	// Perform post-order traversal
	var processedFiles []string
	postOrderTraversal(dirTree, "", func(filePath string) {
		// Skip if not a file
		if !repo.IsFile(filePath) {
			return
		}

		// Skip binary files and non-code files
		if !isCodeFile(filePath) {
			return
		}

		// Analyze file
		fileContent, err := repo.GetFileContent(filePath)
		if err != nil {
			logger.Warnf("Failed to get content for file %s: %v", filePath, err)
			return
		}

		// Analyze code
		fileAnalysis, err := s.analyzeCode(ctx, filePath, fileContent)
		if err != nil {
			logger.Warnf("Failed to analyze file %s: %v", filePath, err)
			return
		}

		// Add to result
		result.Files = append(result.Files, *fileAnalysis)
		processedFiles = append(processedFiles, filePath)
	})

	logger.Infof("Analyzed %d files in repository %s/%s", len(processedFiles), repo.Owner, repo.Name)
	return result, nil
}

// analyzeCode analyzes code content using LLM
func (s *CodeAnalysisService) analyzeCode(ctx context.Context, filePath, fileContent string) (*models.FileAnalysisResult, error) {
	// Prepare prompt for LLM
	prompt := fmt.Sprintf(`
Analyze the following code file and extract the following information:
1. Dependencies (imports, includes, etc.)
2. Global variables
3. Constants
4. Init function (if any) with its functionality
5. Structs with their fields and types
6. Methods with their input parameters, output parameters, and functionality
7. For each method or function, break down the workflow into logical steps
8. For each workflow step, identify:
   - Step name
   - Step type (external system, database, logic, function call)
   - Type details (external system name, database schema, operation name, function call details)
   - Step description
   - Dependencies
   - Input variables/objects
   - Output variables/objects
   - Workflow name

File path: %s

Code:
%s

Provide the analysis in JSON format.
`, filePath, fileContent)

	// Create a chat request for the LLM
	req := models.LLMChatRequest{
		Messages: []models.LLMMessage{
			{Role: models.RoleSystem, Content: "You are a code analysis assistant that extracts structured information from code files."},
			{Role: models.RoleUser, Content: prompt},
		},
		ModelName: "openai:gpt-4", // Use a powerful model for code analysis
	}

	// Send request to LLM
	response, err := s.llmService.ChatWithModels(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Parse LLM response
	result := &models.FileAnalysisResult{
		Path: filePath,
	}

	// Try to parse the JSON response
	// In a production environment, we would handle this more robustly
	err = json.Unmarshal([]byte(response.Text), result)
	if err != nil {
		logger.Warnf("Failed to parse LLM response as JSON: %v", err)
		// Return basic result with just the path
	}

	return result, nil
}

// buildDirectoryTree builds a tree representation of directories
func buildDirectoryTree(filePaths []string) map[string][]string {
	tree := make(map[string][]string)

	// Group files by directory
	for _, filePath := range filePaths {
		dir := path.Dir(filePath)
		if dir == "." {
			dir = ""
		}

		if _, exists := tree[dir]; !exists {
			tree[dir] = []string{}
		}
		tree[dir] = append(tree[dir], filePath)
	}

	// Sort files in each directory
	for dir, files := range tree {
		sort.Strings(files)
		tree[dir] = files
	}

	return tree
}

// postOrderTraversal performs a post-order traversal of the directory tree
func postOrderTraversal(tree map[string][]string, dir string, processFile func(string)) {
	// Get files in current directory
	files, exists := tree[dir]
	if !exists {
		return
	}

	// Process each file
	for _, file := range files {
		// If it's a directory, process it first
		fileDir := path.Join(dir, path.Base(file))
		if _, exists := tree[fileDir]; exists {
			postOrderTraversal(tree, fileDir, processFile)
		}

		// Process the file
		processFile(file)
	}
}

// isCodeFile checks if a file is a code file based on its extension
func isCodeFile(filePath string) bool {
	ext := strings.ToLower(path.Ext(filePath))

	// List of common code file extensions
	codeExtensions := map[string]bool{
		".go":    true,
		".java":  true,
		".js":    true,
		".ts":    true,
		".py":    true,
		".rb":    true,
		".php":   true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".swift": true,
		".kt":    true,
		".rs":    true,
		".scala": true,
		".sh":    true,
		".pl":    true,
		".r":     true,
		".jsx":   true,
		".tsx":   true,
	}

	return codeExtensions[ext]
}
