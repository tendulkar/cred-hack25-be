package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/goanalyzer"
	analyzerModels "cred.com/hack25/backend/pkg/goanalyzer/models"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ServiceLogger provides a service-specific logging wrapper
type ServiceLogger struct {
	service string
}

// logWithFields logs with service context and additional fields
func (sl *ServiceLogger) logWithFields(level logrus.Level, msg string, fields ...interface{}) {
	// Create a logrus Fields map from the variadic fields (expecting key-value pairs)
	logFields := logrus.Fields{"service": sl.service}

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				logFields[key] = fields[i+1]
			}
		}
	}

	entry := logger.WithFields(logger.Fields(logFields))

	switch level {
	case logger.DebugLevel:
		entry.Debug(msg)
	case logger.InfoLevel:
		entry.Info(msg)
	case logger.WarnLevel:
		entry.Warn(msg)
	case logger.ErrorLevel:
		entry.Error(msg)
	}
}

// Debug logs a debug message with context fields
func (sl *ServiceLogger) Debug(msg string, fields ...interface{}) {
	sl.logWithFields(logger.DebugLevel, msg, fields...)
}

// Info logs an info message with context fields
func (sl *ServiceLogger) Info(msg string, fields ...interface{}) {
	sl.logWithFields(logger.InfoLevel, msg, fields...)
}

// Warn logs a warning message with context fields
func (sl *ServiceLogger) Warn(msg string, fields ...interface{}) {
	sl.logWithFields(logger.WarnLevel, msg, fields...)
}

// Error logs an error message with context fields
func (sl *ServiceLogger) Error(msg string, fields ...interface{}) {
	sl.logWithFields(logger.ErrorLevel, msg, fields...)
}

// NewServiceLogger creates a new service-specific logger
func NewServiceLogger(serviceName string) *ServiceLogger {
	return &ServiceLogger{
		service: serviceName,
	}
}

// CodeAnalyzerRepository defines the methods required for the repository
type CodeAnalyzerRepository interface {
	CreateRepository(repo *models.Repository) error
	UpdateRepositoryStatus(id int64, status string, errorMsg string) error
	GetRepositoryByURL(url string) (*models.Repository, error)
	GetRepositoryByID(id int64) (*models.Repository, error)
	CreateRepositoryFile(file *models.RepositoryFile) error
	GetRepositoryFiles(repoID int64) ([]models.RepositoryFile, error)
	GetRepositoryFileByPath(repoID int64, filePath string) (*models.RepositoryFile, error)
	BatchCreateFunctions(functions []models.RepositoryFunction) error
	BatchCreateSymbols(symbols []models.RepositorySymbol) error
	GetRepositoryFunctions(repoID int64, fileID int64) ([]models.RepositoryFunction, error)
	GetRepositorySymbols(repoID int64, fileID int64) ([]models.RepositorySymbol, error)
	BatchCreateFunctionStatements(statements []models.FunctionStatement) error
}

// CodeAnalyzerService handles code analysis operations
type CodeAnalyzerService struct {
	repo         CodeAnalyzerRepository
	analyzer     *goanalyzer.Analyzer
	workspaceDir string
	logger       *ServiceLogger
}

// NewCodeAnalyzerService creates a new code analyzer service
func NewCodeAnalyzerService(repo CodeAnalyzerRepository, workspaceDir string) *CodeAnalyzerService {
	if workspaceDir == "" {
		// Default to a temp directory
		workspaceDir = os.TempDir()
	}

	log := NewServiceLogger("code-analyzer-service")

	return &CodeAnalyzerService{
		repo:         repo,
		analyzer:     goanalyzer.New(),
		workspaceDir: workspaceDir,
		logger:       log,
	}
}

// IndexRepository starts the process of analyzing a repository
func (s *CodeAnalyzerService) IndexRepository(url string) (*models.IndexRepositoryResponse, error) {
	s.logger.Info("Starting repository indexing", "url", url)

	// Parse the URL to extract owner/repo
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		s.logger.Error("Invalid GitHub URL format", "url", url)
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]
	s.logger.Debug("Parsed repository info", "owner", owner, "name", name)

	// Check if repository exists in database
	existingRepo, err := s.repo.GetRepositoryByURL(url)
	if err != nil {
		s.logger.Error("Error checking repository existence", "error", err)
		return nil, fmt.Errorf("error checking repository: %w", err)
	}

	if existingRepo != nil {
		s.logger.Info("Repository found in database", "id", existingRepo.ID, "status", existingRepo.IndexStatus)

		// Repository already exists
		// If already indexing, just return status
		if existingRepo.IndexStatus == "in_progress" {
			s.logger.Info("Repository indexing already in progress", "id", existingRepo.ID)
			return &models.IndexRepositoryResponse{
				ID:          existingRepo.ID,
				URL:         existingRepo.URL,
				IndexStatus: existingRepo.IndexStatus,
				Message:     "Repository indexing in progress",
			}, nil
		}

		// Update status to in_progress
		s.logger.Info("Updating repository status to in_progress", "id", existingRepo.ID)
		err = s.repo.UpdateRepositoryStatus(existingRepo.ID, "in_progress", "")
		if err != nil {
			s.logger.Error("Error updating repository status", "id", existingRepo.ID, "error", err)
			return nil, fmt.Errorf("error updating repository status: %w", err)
		}

		// Start analyzing in a goroutine
		s.logger.Info("Starting repository processing in background", "id", existingRepo.ID)
		s.processRepository(existingRepo.ID, existingRepo.Kind, url, owner, name, existingRepo.LocalPath)

		return &models.IndexRepositoryResponse{
			ID:          existingRepo.ID,
			URL:         existingRepo.URL,
			IndexStatus: "in_progress",
			Message:     "Repository indexing completed",
		}, nil
	}

	// Create a new repository entry
	localPath := filepath.Join(s.workspaceDir, owner, name)
	s.logger.Info("Creating new repository entry", "path", localPath)

	newRepo := &models.Repository{
		Kind:        "github",
		URL:         url,
		Name:        name,
		Owner:       owner,
		LocalPath:   localPath,
		IndexStatus: "in_progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.repo.CreateRepository(newRepo)
	if err != nil {
		s.logger.Error("Error creating repository", "error", err)
		return nil, fmt.Errorf("error creating repository: %w", err)
	}
	s.logger.Info("Repository created successfully", "id", newRepo.ID)

	// Start analyzing in a goroutine
	s.logger.Info("Starting repository processing in background", "id", newRepo.ID)
	err = s.processRepository(newRepo.ID, newRepo.Kind, url, owner, name, localPath)
	if err != nil {
		s.logger.Error("Error processing repository", "error", err)
		return nil, fmt.Errorf("error processing repository: %w", err)
	}

	return &models.IndexRepositoryResponse{
		ID:          newRepo.ID,
		URL:         newRepo.URL,
		IndexStatus: "in_progress",
		Message:     "Repository indexing completed",
	}, nil
}

// processRepository clones the repository and analyzes its code
func (s *CodeAnalyzerService) processRepository(repoID int64, kind, url, owner, name, localPath string) error {
	s.logger.Info("Processing repository", "id", repoID, "kind", kind, "url", url)
	var err error

	// Make sure the local directory exists
	s.logger.Debug("Creating directory", "path", filepath.Dir(localPath))
	if err = os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		errMsg := fmt.Sprintf("Error creating directories: %v", err)
		s.logger.Error(errMsg, "path", filepath.Dir(localPath))
		s.repo.UpdateRepositoryStatus(repoID, "failed", errMsg)
		return err
	}

	// Clone or update repository
	if _, err = os.Stat(localPath); os.IsNotExist(err) {
		// Repository doesn't exist locally, clone it
		s.logger.Info("Cloning repository", "url", url, "path", localPath)
		err = s.cloneRepository(kind, url, localPath)
		if err != nil {
			errMsg := fmt.Sprintf("Error cloning repository: %v", err)
			s.logger.Error(errMsg, "url", url)
			s.repo.UpdateRepositoryStatus(repoID, "failed", errMsg)
			return err
		}
		s.logger.Info("Repository cloned successfully", "path", localPath)
	} else {
		// Repository exists, update it
		s.logger.Info("Updating existing repository", "path", localPath)
		err = s.updateRepository(localPath)
		if err != nil {
			errMsg := fmt.Sprintf("Error updating repository: %v", err)
			s.logger.Error(errMsg, "path", localPath)
			s.repo.UpdateRepositoryStatus(repoID, "failed", errMsg)
			return err
		}
		s.logger.Info("Repository updated successfully", "path", localPath)
	}

	// Analyze the repository
	s.logger.Info("Starting code analysis", "repoID", repoID, "path", localPath)
	err = s.analyzeRepository(repoID, localPath)
	if err != nil {
		errMsg := fmt.Sprintf("Error analyzing repository: %v", err)
		s.logger.Error(errMsg, "path", localPath)
		s.repo.UpdateRepositoryStatus(repoID, "failed", errMsg)
		return err
	}
	s.logger.Info("Repository analysis completed successfully", "repoID", repoID)

	// Update status to completed
	s.logger.Info("Updating repository status to completed", "repoID", repoID)
	s.repo.UpdateRepositoryStatus(repoID, "completed", "")
	return nil
}

// cloneRepository clones a repository from a remote URL
func (s *CodeAnalyzerService) cloneRepository(kind, url, localPath string) error {
	var gitURL string
	switch kind {
	case "github":
		gitURL = url + ".git"
	default:
		gitURL = url
	}
	s.logger.Debug("Determined git URL", "kind", kind, "url", gitURL)

	s.logger.Debug("Executing git clone command", "url", gitURL, "path", localPath)
	cmd := exec.Command("git", "clone", gitURL, localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Git clone failed", "error", err, "output", string(output))
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}
	return nil
}

// updateRepository pulls the latest changes from the remote
func (s *CodeAnalyzerService) updateRepository(localPath string) error {
	s.logger.Debug("Executing git pull command", "path", localPath)
	cmd := exec.Command("git", "-C", localPath, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Git pull failed", "error", err, "output", string(output))
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}
	return nil
}

// analyzeRepository analyzes all Go files in the repository
func (s *CodeAnalyzerService) analyzeRepository(repoID int64, localPath string) error {
	s.logger.Info("Finding Go files in repository", "path", localPath)

	// Find all Go files
	var goFiles []string
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("Error accessing path during walk", "path", path, "error", err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Exclude vendored files
			if !strings.Contains(path, "/vendor/") && !strings.Contains(path, "/.git/") {
				goFiles = append(goFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Error("Error walking directory", "error", err)
		return fmt.Errorf("error walking directory: %w", err)
	}

	s.logger.Info("Found Go files to analyze", "count", len(goFiles))

	// Process each file
	for i, filePath := range goFiles {
		if i > 0 && i%100 == 0 {
			s.logger.Info("Analysis progress", "processed", i, "total", len(goFiles))
		}

		// Get relative path from repo root
		relPath, err := filepath.Rel(localPath, filePath)
		if err != nil {
			s.logger.Warn("Unable to get relative path for file", "file", filePath, "error", err)
			continue
		}

		s.logger.Debug("Analyzing file", "file", relPath)

		// Analyze the file
		analysis, err := s.analyzer.AnalyzeFile(filePath)
		if err != nil {
			s.logger.Warn("Error analyzing file", "file", relPath, "error", err)
			continue
		}
		s.logger.Debug("File analyzed successfully", "file", relPath, "package", analysis.Package)

		// Create repository file entry
		file := &models.RepositoryFile{
			RepositoryID: repoID,
			FilePath:     relPath,
			Package:      analysis.Package,
			LastAnalyzed: time.Now(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = s.repo.CreateRepositoryFile(file)
		if err != nil {
			s.logger.Error("Error creating file entry", "file", relPath, "error", err)
			return fmt.Errorf("error creating file entry: %w", err)
		}
		s.logger.Debug("File entry created", "file", relPath, "fileID", file.ID)

		// Convert functions and symbols to repository models
		functions, symbols, _ := models.FileAnalysisToRepositoryModels(analysis, repoID, file.ID)
		s.logger.Debug("Extracted entities from file", "file", relPath, "functions", len(functions), "symbols", len(symbols))

		// Store functions and symbols
		if len(functions) > 0 {
			err = s.repo.BatchCreateFunctions(functions)
			if err != nil {
				s.logger.Error("Error creating function entries", "file", relPath, "error", err)
				return fmt.Errorf("error creating function entries: %w", err)
			}
			s.logger.Debug("Function entries created", "file", relPath, "count", len(functions))
		}

		if len(symbols) > 0 {
			err = s.repo.BatchCreateSymbols(symbols)
			if err != nil {
				s.logger.Error("Error creating symbol entries", "file", relPath, "error", err)
				return fmt.Errorf("error creating symbol entries: %w", err)
			}
			s.logger.Debug("Symbol entries created", "file", relPath, "count", len(symbols))
		}
	}

	s.logger.Info("Repository analysis completed", "files_processed", len(goFiles))
	return nil
}

// GetRepositoryIndex retrieves the analysis for a repository or specific file
func (s *CodeAnalyzerService) GetRepositoryIndex(url, filePath string) (*models.GetIndexResponse, error) {
	s.logger.Info("Getting repository index", "url", url, "filePath", filePath)

	// Get repository by URL
	repo, err := s.repo.GetRepositoryByURL(url)
	if err != nil {
		s.logger.Error("Error retrieving repository", "url", url, "error", err)
		return nil, fmt.Errorf("error retrieving repository: %w", err)
	}

	if repo == nil {
		s.logger.Warn("Repository not found", "url", url)
		return nil, fmt.Errorf("repository not found")
	}
	s.logger.Debug("Repository found", "id", repo.ID, "status", repo.IndexStatus)

	response := &models.GetIndexResponse{
		Repository: repo,
	}

	// If a specific file was requested
	if filePath != "" {
		s.logger.Debug("Getting specific file", "filePath", filePath)
		file, err := s.repo.GetRepositoryFileByPath(repo.ID, filePath)
		if err != nil {
			s.logger.Error("Error retrieving file", "filePath", filePath, "error", err)
			return nil, fmt.Errorf("error retrieving file: %w", err)
		}

		if file == nil {
			s.logger.Warn("File not found", "filePath", filePath)
			return nil, fmt.Errorf("file not found")
		}
		s.logger.Debug("File found", "fileID", file.ID)

		// Get functions and symbols for this file
		s.logger.Debug("Getting functions for file", "fileID", file.ID)
		functions, err := s.repo.GetRepositoryFunctions(repo.ID, file.ID)
		if err != nil {
			s.logger.Error("Error retrieving functions", "fileID", file.ID, "error", err)
			return nil, fmt.Errorf("error retrieving functions: %w", err)
		}
		s.logger.Debug("Functions retrieved", "count", len(functions))

		s.logger.Debug("Getting symbols for file", "fileID", file.ID)
		symbols, err := s.repo.GetRepositorySymbols(repo.ID, file.ID)
		if err != nil {
			s.logger.Error("Error retrieving symbols", "fileID", file.ID, "error", err)
			return nil, fmt.Errorf("error retrieving symbols: %w", err)
		}
		s.logger.Debug("Symbols retrieved", "count", len(symbols))

		response.Files = []models.RepositoryFile{*file}
		response.Functions = functions
		response.Symbols = symbols

		s.logger.Info("File index data retrieved successfully", "filePath", filePath,
			"functions", len(functions), "symbols", len(symbols))
	} else {
		// Get all files
		s.logger.Debug("Getting all files for repository", "repoID", repo.ID)
		files, err := s.repo.GetRepositoryFiles(repo.ID)
		if err != nil {
			s.logger.Error("Error retrieving files", "error", err)
			return nil, fmt.Errorf("error retrieving files: %w", err)
		}
		response.Files = files
		s.logger.Info("Repository index data retrieved successfully", "fileCount", len(files))
	}

	return response, nil
}

// AnalyzeGoFile analyzes a single Go file and returns the analysis
// This is a direct analysis without storing in the database
func (s *CodeAnalyzerService) AnalyzeGoFile(filePath string) (*analyzerModels.FileAnalysis, error) {
	s.logger.Info("Analyzing single Go file", "filePath", filePath)

	analysis, err := s.analyzer.AnalyzeFile(filePath)
	if err != nil {
		s.logger.Error("Error analyzing file", "filePath", filePath, "error", err)
		return nil, err
	}

	s.logger.Info("File analyzed successfully", "filePath", filePath,
		"package", analysis.Package,
		"functions", len(analysis.Functions),
		"symbols", len(analysis.Constants)+len(analysis.Variables)+len(analysis.Types))

	return analysis, nil
}
