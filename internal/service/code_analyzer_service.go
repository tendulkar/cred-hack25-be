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
)

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
}

// CodeAnalyzerService handles code analysis operations
type CodeAnalyzerService struct {
	repo         CodeAnalyzerRepository
	analyzer     *goanalyzer.Analyzer
	workspaceDir string
}

// NewCodeAnalyzerService creates a new code analyzer service
func NewCodeAnalyzerService(repo CodeAnalyzerRepository, workspaceDir string) *CodeAnalyzerService {
	if workspaceDir == "" {
		// Default to a temp directory
		workspaceDir = os.TempDir()
	}

	return &CodeAnalyzerService{
		repo:         repo,
		analyzer:     goanalyzer.New(),
		workspaceDir: workspaceDir,
	}
}

// IndexRepository starts the process of analyzing a repository
func (s *CodeAnalyzerService) IndexRepository(url string) (*models.IndexRepositoryResponse, error) {
	// Parse the URL to extract owner/repo
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]

	// Check if repository exists in database
	existingRepo, err := s.repo.GetRepositoryByURL(url)
	if err != nil {
		return nil, fmt.Errorf("error checking repository: %w", err)
	}

	if existingRepo != nil {
		// Repository already exists
		// If already indexing, just return status
		if existingRepo.IndexStatus == "in_progress" {
			return &models.IndexRepositoryResponse{
				ID:          existingRepo.ID,
				URL:         existingRepo.URL,
				IndexStatus: existingRepo.IndexStatus,
				Message:     "Repository indexing in progress",
			}, nil
		}

		// Update status to in_progress
		err = s.repo.UpdateRepositoryStatus(existingRepo.ID, "in_progress", "")
		if err != nil {
			return nil, fmt.Errorf("error updating repository status: %w", err)
		}

		// Start analyzing in a goroutine
		go s.processRepository(existingRepo.ID, existingRepo.Kind, url, owner, name, existingRepo.LocalPath)

		return &models.IndexRepositoryResponse{
			ID:          existingRepo.ID,
			URL:         existingRepo.URL,
			IndexStatus: "in_progress",
			Message:     "Repository indexing started",
		}, nil
	}

	// Create a new repository entry
	localPath := filepath.Join(s.workspaceDir, owner, name)
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
		return nil, fmt.Errorf("error creating repository: %w", err)
	}

	// Start analyzing in a goroutine
	go s.processRepository(newRepo.ID, newRepo.Kind, url, owner, name, localPath)

	return &models.IndexRepositoryResponse{
		ID:          newRepo.ID,
		URL:         newRepo.URL,
		IndexStatus: "in_progress",
		Message:     "Repository indexing started",
	}, nil
}

// processRepository clones the repository and analyzes its code
func (s *CodeAnalyzerService) processRepository(repoID int64, kind, url, owner, name, localPath string) {
	var err error

	// Make sure the local directory exists
	if err = os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		s.repo.UpdateRepositoryStatus(repoID, "failed", fmt.Sprintf("Error creating directories: %v", err))
		return
	}

	// Clone or update repository
	if _, err = os.Stat(localPath); os.IsNotExist(err) {
		// Repository doesn't exist locally, clone it
		err = s.cloneRepository(kind, url, localPath)
		if err != nil {
			s.repo.UpdateRepositoryStatus(repoID, "failed", fmt.Sprintf("Error cloning repository: %v", err))
			return
		}
	} else {
		// Repository exists, update it
		err = s.updateRepository(localPath)
		if err != nil {
			s.repo.UpdateRepositoryStatus(repoID, "failed", fmt.Sprintf("Error updating repository: %v", err))
			return
		}
	}

	// Analyze the repository
	err = s.analyzeRepository(repoID, localPath)
	if err != nil {
		s.repo.UpdateRepositoryStatus(repoID, "failed", fmt.Sprintf("Error analyzing repository: %v", err))
		return
	}

	// Update status to completed
	s.repo.UpdateRepositoryStatus(repoID, "completed", "")
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

	cmd := exec.Command("git", "clone", gitURL, localPath)
	return cmd.Run()
}

// updateRepository pulls the latest changes from the remote
func (s *CodeAnalyzerService) updateRepository(localPath string) error {
	cmd := exec.Command("git", "-C", localPath, "pull")
	return cmd.Run()
}

// analyzeRepository analyzes all Go files in the repository
func (s *CodeAnalyzerService) analyzeRepository(repoID int64, localPath string) error {
	// Find all Go files
	var goFiles []string
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Process each file
	for _, filePath := range goFiles {
		// Get relative path from repo root
		relPath, err := filepath.Rel(localPath, filePath)
		if err != nil {
			continue
		}

		// Analyze the file
		analysis, err := s.analyzer.AnalyzeFile(filePath)
		if err != nil {
			continue
		}

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
			return fmt.Errorf("error creating file entry: %w", err)
		}

		// Convert functions and symbols to repository models
		functions, symbols := models.FileAnalysisToRepositoryModels(analysis, repoID, file.ID)

		// Store functions and symbols
		if len(functions) > 0 {
			err = s.repo.BatchCreateFunctions(functions)
			if err != nil {
				return fmt.Errorf("error creating function entries: %w", err)
			}
		}

		if len(symbols) > 0 {
			err = s.repo.BatchCreateSymbols(symbols)
			if err != nil {
				return fmt.Errorf("error creating symbol entries: %w", err)
			}
		}
	}

	return nil
}

// GetRepositoryIndex retrieves the analysis for a repository or specific file
func (s *CodeAnalyzerService) GetRepositoryIndex(url, filePath string) (*models.GetIndexResponse, error) {
	// Get repository by URL
	repo, err := s.repo.GetRepositoryByURL(url)
	if err != nil {
		return nil, fmt.Errorf("error retrieving repository: %w", err)
	}

	if repo == nil {
		return nil, fmt.Errorf("repository not found")
	}

	response := &models.GetIndexResponse{
		Repository: repo,
	}

	// If a specific file was requested
	if filePath != "" {
		file, err := s.repo.GetRepositoryFileByPath(repo.ID, filePath)
		if err != nil {
			return nil, fmt.Errorf("error retrieving file: %w", err)
		}

		if file == nil {
			return nil, fmt.Errorf("file not found")
		}

		// Get functions and symbols for this file
		functions, err := s.repo.GetRepositoryFunctions(repo.ID, file.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving functions: %w", err)
		}

		symbols, err := s.repo.GetRepositorySymbols(repo.ID, file.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving symbols: %w", err)
		}

		response.Files = []models.RepositoryFile{*file}
		response.Functions = functions
		response.Symbols = symbols
	} else {
		// Get all files
		files, err := s.repo.GetRepositoryFiles(repo.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving files: %w", err)
		}

		response.Files = files
	}

	return response, nil
}

// AnalyzeGoFile analyzes a single Go file and returns the analysis
// This is a direct analysis without storing in the database
func (s *CodeAnalyzerService) AnalyzeGoFile(filePath string) (*analyzerModels.FileAnalysis, error) {
	return s.analyzer.AnalyzeFile(filePath)
}
