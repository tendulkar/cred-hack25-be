package models

import (
	"encoding/json"
	"strings"
	"time"

	"cred.com/hack25/backend/internal/insights"
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
	ID            int64               `json:"id" db:"id"`
	RepositoryID  int64               `json:"repository_id" db:"repository_id"`
	FileID        int64               `json:"file_id" db:"file_id"`
	Name          string              `json:"name" db:"name"`                     // Function name
	Kind          string              `json:"kind" db:"kind"`                     // "function" or "method"
	Receiver      string              `json:"receiver" db:"receiver"`             // For methods
	Exported      bool                `json:"exported" db:"exported"`             // If it's exported
	Parameters    string              `json:"parameters" db:"parameters"`         // JSON array of parameters
	Results       string              `json:"results" db:"results"`               // JSON array of results
	CodeBlock     string              `json:"code_block" db:"code_block"`         // Full code
	Line          int                 `json:"line" db:"line"`                     // Starting line
	Calls         string              `json:"calls" db:"calls"`                   // JSON array of function calls
	CalledBy      string              `json:"called_by" db:"called_by"`           // JSON array of functions calling this
	References    string              `json:"references" db:"references"`         // JSON array of references
	StatementInfo string              `json:"statement_info" db:"statement_info"` // JSON of parsed statement info
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" db:"updated_at"`
	Statements    []FunctionStatement `json:"-" db:"-"`
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

// IndexedFunction represents a function with additional metadata like insights
type IndexedFunction struct {
	Function        *RepositoryFunction       `json:"function"`
	FunctionInsight *insights.FunctionInsight `json:"function_insight,omitempty"`
	Insights        interface{}               `json:"insights,omitempty"`
}

// IndexedSymbol represents a symbol with additional metadata
type IndexedSymbol struct {
	Symbol   *RepositorySymbol `json:"symbol"`
	Insights interface{}       `json:"insights,omitempty"`
}

// IndexedFile represents a file with its functions and symbols
type IndexedFile struct {
	File      *RepositoryFile            `json:"file"`
	Functions map[int64]*IndexedFunction `json:"functions,omitempty"` // Map of function ID to IndexedFunction
	Symbols   map[int64]*IndexedSymbol   `json:"symbols,omitempty"`   // Map of symbol ID to IndexedSymbol
	Insights  interface{}                `json:"insights,omitempty"`
}

// GetIndexResponse is the response for a get index request
type GetIndexResponse struct {
	Repository      *Repository             `json:"repository"`
	IndexedFilesMap map[string]*IndexedFile `json:"indexed_files_map,omitempty"` // Map of file path to IndexedFile

	// Legacy fields - kept for backward compatibility
	Files     []RepositoryFile       `json:"files,omitempty"`
	Functions []RepositoryFunction   `json:"functions,omitempty"`
	Symbols   []RepositorySymbol     `json:"symbols,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // For additional data like insights
}

// FileAnalysisToRepositoryModels converts a FileAnalysis to repository models
func FileAnalysisToRepositoryModels(analysis *models.FileAnalysis, repoID int64, fileID int64) ([]RepositoryFunction, []RepositorySymbol, []FunctionStatement, []FunctionCall, []FunctionReference, []FileDependency) {
	var functions []RepositoryFunction
	var symbols []RepositorySymbol
	var statements []FunctionStatement
	var calls []FunctionCall
	var references []FunctionReference
	var dependencies []FileDependency

	// Convert functions
	for _, fn := range analysis.Functions {
		// Convert parameters to JSON
		paramsJSON, _ := json.Marshal(fn.Parameters)

		// Convert results to JSON
		resultsJSON, _ := json.Marshal(fn.Results)

		repoFn := RepositoryFunction{
			RepositoryID: repoID,
			FileID:       fileID,
			Name:         fn.Name,
			Kind:         fn.Kind,
			Receiver:     fn.Receiver,
			Exported:     fn.Exported,
			Parameters:   string(paramsJSON),
			Results:      string(resultsJSON),
			CodeBlock:    fn.CodeBlock,
			Line:         fn.Position.Line,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		repoFn.Statements = convertStatements(fn.StatementAnalysis, nil)
		functions = append(functions, repoFn)

		// We'll need to associate statements with this function later
		// Store the index of this function for reference when we have its ID
		fnIndex := len(functions) - 1

		// Convert statement analysis to function statements
		convertedStatements := convertStatements(fn.StatementAnalysis, nil)
		if len(convertedStatements) > 0 {
			// Store the function index so we can update with the real function ID later
			for i := range convertedStatements {
				convertedStatements[i].FunctionIndex = fnIndex
			}
			statements = append(statements, convertedStatements...)
		}
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

	// Convert function calls
	for _, call := range analysis.Calls {
		// Find caller function index
		callerIndex := -1
		for i, fn := range functions {
			// Match function by name and line number
			if call.Caller == fn.Name {
				callerIndex = i
				break
			}
		}

		// Skip if we couldn't find the caller function
		if callerIndex == -1 {
			continue
		}

		// Convert parameters to JSON
		paramsJSON, _ := json.Marshal(call.Parameters)

		fnCall := FunctionCall{
			CallerID:      0, // Will be set after function IDs are assigned
			CalleeName:    call.Callee,
			CalleePackage: call.CalleePath,
			Line:          call.Position.Line,
			Parameters:    string(paramsJSON),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Store the function index so we can update with the real caller ID later
		fnCall.CallerID = int64(callerIndex) // Temporarily store index, will be replaced
		calls = append(calls, fnCall)
	}

	// Log the references we're processing
	// s.logger.Info("Processing function references", "references", analysis.References)

	// Convert function references
	for _, ref := range analysis.References {
		// Find referenced function index
		functionIndex := -1
		for i, fn := range functions {
			// Match function by name
			if ref.Symbol == fn.Name {
				functionIndex = i
				break
			}
		}

		// Skip if we couldn't find the referenced function
		if functionIndex == -1 {
			continue
		}

		fnRef := FunctionReference{
			FunctionID:     0, // Will be set after function IDs are assigned
			ReferenceType:  ref.RefType,
			FileID:         fileID,
			Line:           ref.Position.Line,
			ColumnPosition: ref.Position.Column,
			Context:        "", // Could extract a snippet from the source file if needed
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Store the function index so we can update with the real function ID later
		fnRef.FunctionID = int64(functionIndex) // Temporarily store index, will be replaced
		references = append(references, fnRef)
	}

	// Convert imports to dependencies
	for _, imp := range analysis.Imports {
		// Determine if it's a standard library import
		isStdlib := false
		if !strings.Contains(imp.Value, ".") {
			// Standard library packages typically don't have dots in their names
			isStdlib = true
		}

		dep := FileDependency{
			RepositoryID: repoID,
			FileID:       fileID,
			ImportPath:   imp.Value,
			Alias:        imp.Name,
			IsStdlib:     isStdlib,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		dependencies = append(dependencies, dep)
	}

	return functions, symbols, statements, calls, references, dependencies
}

// convertStatements recursively converts StatementInfo to FunctionStatement models
func convertStatements(stmtInfos []models.StatementInfo, parentID *int64) []FunctionStatement {
	var statements []FunctionStatement

	for _, stmt := range stmtInfos {
		// Encode JSON fields
		conditionsJSON, _ := json.Marshal(stmt.Conditions)
		variablesJSON, _ := json.Marshal(stmt.Variables)
		callsJSON, _ := json.Marshal(stmt.Calls)

		// Create statement
		repoStmt := FunctionStatement{
			// FunctionID will be set later when we have actual function IDs
			StatementType:     stmt.Type,
			Text:              stmt.Text,
			Line:              stmt.Position.Line,
			Conditions:        string(conditionsJSON),
			Variables:         string(variablesJSON),
			Calls:             string(callsJSON),
			ParentStatementID: parentID,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			// Temporary field to track parentage before DB insertion
			FunctionIndex: -1, // Will be set by the caller
		}

		statements = append(statements, repoStmt)

		// Set this statement as the parent for its sub-statements
		stmtIndex := len(statements) - 1

		// Process nested statements
		if len(stmt.SubStatements) > 0 {
			// Use a placeholder ID that will be replaced after DB insertion
			placeholderID := int64(-(stmtIndex + 1)) // Negative to ensure it doesn't conflict with real IDs
			childStmts := convertStatements(stmt.SubStatements, &placeholderID)

			// Associate with the same function
			for i := range childStmts {
				childStmts[i].ParentIndex = stmtIndex
			}

			statements = append(statements, childStmts...)
		}
	}

	return statements
}
