package goanalyzer

import (
	"fmt"
	"os"
	"path/filepath"

	"cred.com/hack25/backend/pkg/goanalyzer/analyzer"
	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// Analyzer is a facade for the goanalyzer functionality
type Analyzer struct {
	analyzer *analyzer.Analyzer
}

// New creates a new code analyzer
func New() *Analyzer {
	return &Analyzer{
		analyzer: analyzer.New(),
	}
}

// AnalyzeFile analyzes a single Go file
func (a *Analyzer) AnalyzeFile(filePath string) (*models.FileAnalysis, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}

	// Check if the file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", absPath)
	}

	return a.analyzer.AnalyzeFile(absPath)
}

// AnalyzeDirectory analyzes all Go files in a directory
func (a *Analyzer) AnalyzeDirectory(dirPath string) ([]models.FileAnalysis, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}

	// Check if the directory exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	return a.analyzer.AnalyzeDirectory(absPath)
}

// GetCallHierarchy returns the call hierarchy for a specific function
func (a *Analyzer) GetCallHierarchy(filePath, funcName string) []models.CallInfo {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil
	}
	return a.analyzer.GetCallHierarchy(absPath, funcName)
}

// GetReferences returns all references to a symbol
func (a *Analyzer) GetReferences(symbolName string) []models.ReferenceInfo {
	return a.analyzer.GetReferences(symbolName)
}

// GetSymbol returns a symbol by name
func (a *Analyzer) GetSymbol(symbolName string) (models.Symbol, bool) {
	return a.analyzer.GetSymbol(symbolName)
}
