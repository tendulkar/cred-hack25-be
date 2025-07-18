package service

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/internal/repointel"
	"cred.com/hack25/backend/internal/repository"
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
	AddFunctionCall(call *models.FunctionCall) error
	AddFunctionReference(ref *models.FunctionReference) error
	AddFileDependency(dep *models.FileDependency) error
	BatchAddFileDependencies(deps []models.FileDependency) error
	GetFileDependencies(repoID int64, fileID int64) ([]models.FileDependency, error)
}

// CodeAnalyzerService handles code analysis operations
type CodeAnalyzerService struct {
	repo                CodeAnalyzerRepository
	analyzer            *goanalyzer.Analyzer
	workspaceDir        string
	logger              *ServiceLogger
	insightsManager     *repointel.InsightsManager
	liteLLMBaseURL      string
	liteLLMAPIKey       string
	liteLLMDefaultModel string
}

// NewCodeAnalyzerService creates a new code analyzer service
func NewCodeAnalyzerService(repo CodeAnalyzerRepository, workspaceDir string, liteLLMURL, liteLLMAPIKey, liteLLMDefaultModel string, insightsManager *repointel.InsightsManager) *CodeAnalyzerService {
	if workspaceDir == "" {
		// Default to a temp directory
		workspaceDir = os.TempDir()
	}

	log := NewServiceLogger("code-analyzer-service")

	// Use environment variables as fallback if not provided
	if liteLLMURL == "" {
		liteLLMURL = os.Getenv("LITELLM_BASE_URL")
	}
	if liteLLMAPIKey == "" {
		liteLLMAPIKey = os.Getenv("LITELLM_API_KEY")
	}
	if liteLLMDefaultModel == "" {
		liteLLMDefaultModel = os.Getenv("LITELLM_DEFAULT_MODEL")
	}
	s := &CodeAnalyzerService{
		repo:                repo,
		analyzer:            goanalyzer.New(),
		workspaceDir:        workspaceDir,
		logger:              log,
		liteLLMBaseURL:      liteLLMURL,
		liteLLMAPIKey:       liteLLMAPIKey,
		liteLLMDefaultModel: liteLLMDefaultModel,
		insightsManager:     insightsManager,
	}

	// Initialize insights components if LiteLLM credentials are provided
	if liteLLMURL != "" && liteLLMAPIKey != "" && liteLLMDefaultModel != "" {
		// We'll initialize the insights manager after creating the service
		// since we need to pass the raw DB connection to the repointel repository
		log.Info("LiteLLM credentials found, repository intelligence will be available")
	} else {
		log.Warn("LiteLLM credentials not provided, repository intelligence features will be unavailable")
	}

	return s
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
	// case "github":
	// 	gitURL = url + ".git"
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
		functions, symbols, _, funcCalls, funcRefs, fileDeps := models.FileAnalysisToRepositoryModels(analysis, repoID, file.ID)
		s.logger.Info("Extracted entities from file", "file", relPath, "functions", len(functions), "symbols", len(symbols),
			"calls", len(funcCalls), "references", len(funcRefs), "dependencies", len(fileDeps))

		// Store functions and symbols
		if len(functions) > 0 {
			err = s.repo.BatchCreateFunctions(functions)
			if err != nil {
				s.logger.Error("Error creating function entries", "file", relPath, "error", err)
				return fmt.Errorf("error creating function entries: %w", err)
			}
			s.logger.Debug("Function entries created", "file", relPath, "count", len(functions))

			// Now process function calls and references using the real function IDs
			if len(funcCalls) > 0 {
				// Update caller IDs with real database IDs
				for i := range funcCalls {
					// The CallerID currently contains the index into the functions slice
					fnIndex := funcCalls[i].CallerID
					if fnIndex >= 0 && int(fnIndex) < len(functions) {
						funcCalls[i].CallerID = functions[fnIndex].ID
					}
				}

				// Store function calls in the database
				for _, call := range funcCalls {
					err = s.repo.AddFunctionCall(&call)
					if err != nil {
						s.logger.Warn("Error creating function call", "caller_id", call.CallerID, "callee", call.CalleeName, "error", err)
						// Continue with other calls, don't fail the entire analysis
					}
				}
				s.logger.Debug("Function calls created", "file", relPath, "count", len(funcCalls))
			}

			// Process function references
			if len(funcRefs) > 0 {
				// Update function IDs with real database IDs
				for i := range funcRefs {
					// The FunctionID currently contains the index into the functions slice
					fnIndex := funcRefs[i].FunctionID
					if fnIndex >= 0 && int(fnIndex) < len(functions) {
						funcRefs[i].FunctionID = functions[fnIndex].ID
					}
				}

				// Store function references in the database
				for _, ref := range funcRefs {
					err = s.repo.AddFunctionReference(&ref)
					if err != nil {
						s.logger.Warn("Error creating function reference", "function_id", ref.FunctionID, "type", ref.ReferenceType, "error", err)
						// Continue with other references, don't fail the entire analysis
					}
				}
				s.logger.Debug("Function references created", "file", relPath, "count", len(funcRefs))
			}

			// Store insights
			for _, function := range functions {
				s.logger.Info("Storing insights for repository", "file", relPath)
				_, err = s.insightsManager.GenerateAndSaveFunctionInsight(repoID, function.ID, "gpt-4o")
				if err != nil {
					s.logger.Error("Error storing insights", "file", relPath, "error", err)
					return fmt.Errorf("error storing insights: %w", err)
				}
				s.logger.Debug("Insights stored", "file", relPath)
			}

		}

		if len(symbols) > 0 {
			err = s.repo.BatchCreateSymbols(symbols)
			if err != nil {
				s.logger.Error("Error creating symbol entries", "file", relPath, "error", err)
				return fmt.Errorf("error creating symbol entries: %w", err)
			}
			s.logger.Debug("Symbol entries created", "file", relPath, "count", len(symbols))
		}

		// Store file dependencies
		if len(fileDeps) > 0 {
			s.logger.Info("Adding file dependencies", "file", relPath, "count", len(fileDeps),
				"fileDependencies", fileDeps)
			err = s.repo.BatchAddFileDependencies(fileDeps)
			if err != nil {
				s.logger.Error("Error creating file dependency entries", "file", relPath, "error", err)
				// Don't fail the entire analysis for dependency errors, just log it
				s.logger.Warn("Continuing analysis despite dependency errors")
			} else {
				s.logger.Debug("File dependency entries created", "file", relPath, "count", len(fileDeps))
			}
		}

	}

	s.logger.Info("Repository analysis completed", "files_processed", len(goFiles))
	return nil
}

// InitializeInsightsManager initializes the insights manager if LiteLLM credentials are available
func (s *CodeAnalyzerService) InitializeInsightsManager(dbConn *sql.DB) {
	if s.liteLLMBaseURL != "" && s.liteLLMAPIKey != "" && s.liteLLMDefaultModel != "" {
		s.logger.Info("Initializing repository intelligence components")

		// Create repointel repository
		insightRepo := repointel.NewRepository(dbConn)

		// Create repointel service
		insightService := repointel.NewService(
			s.repo.(*repository.CodeAnalyzerRepository), // Cast to concrete type
			s.liteLLMBaseURL,
			s.liteLLMAPIKey,
			s.liteLLMDefaultModel,
		)

		// Create insights manager
		s.insightsManager = repointel.NewInsightsManager(insightService, insightRepo)
		s.logger.Info("Repository intelligence components initialized successfully")
	}
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
		Repository:      repo,
		IndexedFilesMap: make(map[string]*models.IndexedFile),
		Metadata:        make(map[string]interface{}),
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

		// Create the indexed file structure
		indexedFile := &models.IndexedFile{
			File:      file,
			Functions: make(map[int64]*models.IndexedFunction),
			Symbols:   make(map[int64]*models.IndexedSymbol),
		}

		// Get functions and symbols for this file
		s.logger.Debug("Getting functions for file", "fileID", file.ID)
		functions, err := s.repo.GetRepositoryFunctions(repo.ID, file.ID)
		if err != nil {
			s.logger.Error("Error retrieving functions", "fileID", file.ID, "error", err)
			return nil, fmt.Errorf("error retrieving functions: %w", err)
		}
		s.logger.Debug("Functions retrieved", "count", len(functions))

		// Populate the indexed functions with insights
		for i := range functions {
			functionPtr := &functions[i]
			indexedFunc := &models.IndexedFunction{
				Function: functionPtr,
			}

			// Add function insights if insights manager is available
			if s.insightsManager != nil {
				insightList, err := s.insightsManager.GetFunctionInsights(repo.ID, functionPtr.ID)
				if err != nil {
					s.logger.Warn("Error retrieving function insight", "function_id", functionPtr.ID, "error", err)
				} else if len(insightList) > 0 {
					// Use the latest insight
					indexedFunc.Insights = insightList[0]
				}
			}

			// Add to the indexed file
			indexedFile.Functions[functionPtr.ID] = indexedFunc
		}

		s.logger.Debug("Getting symbols for file", "fileID", file.ID)
		symbols, err := s.repo.GetRepositorySymbols(repo.ID, file.ID)
		if err != nil {
			s.logger.Error("Error retrieving symbols", "fileID", file.ID, "error", err)
			return nil, fmt.Errorf("error retrieving symbols: %w", err)
		}
		s.logger.Debug("Symbols retrieved", "count", len(symbols))

		// Populate the indexed symbols
		for i := range symbols {
			sym := &symbols[i]
			indexedSymbol := &models.IndexedSymbol{
				Symbol: sym,
			}

			// Add to the indexed file
			indexedFile.Symbols[sym.ID] = indexedSymbol
		}

		// Add the indexed file to the map
		response.IndexedFilesMap[filePath] = indexedFile

		// For backward compatibility
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

		// Process each file
		for i := range files {
			filePtr := &files[i]

			// Create indexed file structure
			indexedFile := &models.IndexedFile{
				File:      filePtr,
				Functions: make(map[int64]*models.IndexedFunction),
				Symbols:   make(map[int64]*models.IndexedSymbol),
			}

			// Get functions for this file
			fileFunctions, err := s.repo.GetRepositoryFunctions(repo.ID, filePtr.ID)
			if err != nil {
				s.logger.Warn("Error retrieving functions for file", "fileID", filePtr.ID, "error", err)
				// Continue with other files even if this one fails
			} else {
				// Process functions
				for j := range fileFunctions {
					functionPtr := &fileFunctions[j]
					indexedFunc := &models.IndexedFunction{
						Function: functionPtr,
					}

					// Add function insights if available
					if s.insightsManager != nil {
						insightList, err := s.insightsManager.GetFunctionInsights(repo.ID, functionPtr.ID)
						if err == nil && len(insightList) > 0 {
							indexedFunc.Insights = insightList[0]
						}
					}

					// Add to indexed file
					indexedFile.Functions[functionPtr.ID] = indexedFunc
				}
			}

			// Get symbols for this file
			fileSymbols, err := s.repo.GetRepositorySymbols(repo.ID, filePtr.ID)
			if err != nil {
				s.logger.Warn("Error retrieving symbols for file", "fileID", filePtr.ID, "error", err)
				// Continue with other files even if this one fails
			} else {
				// Process symbols
				for j := range fileSymbols {
					symbolPtr := &fileSymbols[j]
					indexedFile.Symbols[symbolPtr.ID] = &models.IndexedSymbol{Symbol: symbolPtr}
				}
			}

			// File insights are not implemented yet
			// This is where we would add file-level insights in the future

			// Add to response map
			response.IndexedFilesMap[filePtr.FilePath] = indexedFile
		}

		// For backward compatibility
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
