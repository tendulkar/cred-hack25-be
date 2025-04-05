package models

import (
	"time"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// Repository represents a code repository that has been indexed
type Repository struct {
	ID               int64      `json:"id" db:"id"`
	Kind             string     `json:"kind" db:"kind"`                 // "github", "gitlab", etc.
	URL              string     `json:"url" db:"url"`                   // Original URL
	Name             string     `json:"name" db:"name"`                 // Repository name
	Owner            string     `json:"owner" db:"owner"`               // Repository owner/organization
	LocalPath        string     `json:"local_path" db:"local_path"`     // Where it's stored locally
	LastIndexed      *time.Time `json:"last_indexed" db:"last_indexed"` // When it was last analyzed
	IndexStatus      string     `json:"index_status" db:"index_status"` // "in_progress", "completed", "failed"
	IndexStatusError string     `json:"index_error" db:"index_error"`   // Error message if indexing failed
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// RepositoryFile represents an analyzed file in a repository
type RepositoryFile struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	FilePath     string    `json:"file_path" db:"file_path"` // Relative path within repo
	Package      string    `json:"package" db:"package"`     // Go package name
	LastAnalyzed time.Time `json:"last_analyzed" db:"last_analyzed"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RepositoryFunction represents an analyzed function in a file
type RepositoryFunction struct {
	ID            int64     `json:"id" db:"id"`
	RepositoryID  int64     `json:"repository_id" db:"repository_id"`
	FileID        int64     `json:"file_id" db:"file_id"`
	Name          string    `json:"name" db:"name"`                     // Function name
	Kind          string    `json:"kind" db:"kind"`                     // "function" or "method"
	Receiver      string    `json:"receiver" db:"receiver"`             // For methods
	Exported      bool      `json:"exported" db:"exported"`             // If it's exported
	Parameters    string    `json:"parameters" db:"parameters"`         // JSON array of parameters
	Results       string    `json:"results" db:"results"`               // JSON array of results
	CodeBlock     string    `json:"code_block" db:"code_block"`         // Full code
	Line          int       `json:"line" db:"line"`                     // Starting line
	Calls         string    `json:"calls" db:"calls"`                   // JSON array of function calls
	CalledBy      string    `json:"called_by" db:"called_by"`           // JSON array of functions calling this
	References    string    `json:"references" db:"references"`         // JSON array of references
	StatementInfo string    `json:"statement_info" db:"statement_info"` // JSON of parsed statement info
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// RepositorySymbol represents other symbols in the repository (vars, consts, types)
type RepositorySymbol struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	FileID       int64     `json:"file_id" db:"file_id"`
	Name         string    `json:"name" db:"name"`
	Kind         string    `json:"kind" db:"kind"`             // "variable", "constant", "type", "struct", "interface"
	Type         string    `json:"type" db:"type"`             // Type information
	Value        string    `json:"value" db:"value"`           // For constants and variables
	Exported     bool      `json:"exported" db:"exported"`     // If it's exported
	Fields       string    `json:"fields" db:"fields"`         // JSON array of fields (for structs)
	Methods      string    `json:"methods" db:"methods"`       // JSON array of methods
	Line         int       `json:"line" db:"line"`             // Starting line
	References   string    `json:"references" db:"references"` // JSON array of references
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// IndexRepositoryRequest is used to request repository indexing
type IndexRepositoryRequest struct {
	URL string `json:"url" validate:"required"`
}

// IndexRepositoryResponse is the response for a repository indexing request
type IndexRepositoryResponse struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	IndexStatus string `json:"index_status"`
	Message     string `json:"message"`
}

// GetIndexRequest is used to request repository index info
type GetIndexRequest struct {
	URL      string `json:"url" validate:"required"`
	FilePath string `json:"file_path,omitempty"`
}

// GetIndexResponse is the response for a get index request
type GetIndexResponse struct {
	Repository *Repository          `json:"repository"`
	Files      []RepositoryFile     `json:"files,omitempty"`
	Functions  []RepositoryFunction `json:"functions,omitempty"`
	Symbols    []RepositorySymbol   `json:"symbols,omitempty"`
}

// FileAnalysisToRepositoryModels converts a FileAnalysis to repository models
func FileAnalysisToRepositoryModels(analysis *models.FileAnalysis, repoID int64, fileID int64) ([]RepositoryFunction, []RepositorySymbol) {
	var functions []RepositoryFunction
	var symbols []RepositorySymbol

	// Convert functions
	for _, fn := range analysis.Functions {
		repoFn := RepositoryFunction{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         fn.Name,
			Kind:         fn.Kind,
			Receiver:     fn.Receiver,
			Exported:     fn.Exported,
			CodeBlock:    fn.CodeBlock,
			Line:         fn.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		functions = append(functions, repoFn)
	}

	// Convert constants
	for _, c := range analysis.Constants {
		symbol := RepositorySymbol{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         c.Name,
			Kind:         "constant",
			Type:         c.Type,
			Value:        c.Value,
			Exported:     c.Exported,
			Line:         c.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		symbols = append(symbols, symbol)
	}

	// Convert variables
	for _, v := range analysis.Variables {
		symbol := RepositorySymbol{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         v.Name,
			Kind:         "variable",
			Type:         v.Type,
			Value:        v.Value,
			Exported:     v.Exported,
			Line:         v.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		symbols = append(symbols, symbol)
	}

	// Convert types
	for _, t := range analysis.Types {
		symbol := RepositorySymbol{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         t.Name,
			Kind:         "type",
			Type:         t.Type,
			Exported:     t.Exported,
			Line:         t.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		symbols = append(symbols, symbol)
	}

	// Convert structs
	for _, s := range analysis.Structs {
		symbol := RepositorySymbol{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         s.Name,
			Kind:         "struct",
			Exported:     s.Exported,
			Line:         s.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		symbols = append(symbols, symbol)
	}

	// Convert interfaces
	for _, i := range analysis.Interfaces {
		symbol := RepositorySymbol{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         i.Name,
			Kind:         "interface",
			Exported:     i.Exported,
			Line:         i.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		symbols = append(symbols, symbol)
	}

	return functions, symbols
}
