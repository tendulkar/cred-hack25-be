package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// CodeAnalyzerRepository handles interactions with the code analyzer database tables
type CodeAnalyzerRepository struct {
	DB *sqlx.DB
}

// NewCodeAnalyzerRepository creates a new CodeAnalyzerRepository
func NewCodeAnalyzerRepository(dbConn *sql.DB) *CodeAnalyzerRepository {
	return &CodeAnalyzerRepository{
		DB: sqlx.NewDb(dbConn, "postgres"),
	}
}

// log returns a logrus entry with the repository context
func (r *CodeAnalyzerRepository) log() *logrus.Entry {
	return logger.Log.WithField("component", "code-analyzer-repository")
}

// fieldsToLogrus converts logger.Fields to logrus.Fields
func fieldsToLogrus(fields logger.Fields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for k, v := range fields {
		logrusFields[k] = v
	}
	return logrusFields
}

// CreateRepository creates a new repository in the database
func (r *CodeAnalyzerRepository) CreateRepository(repo *models.Repository) error {
	r.log().WithField("url", repo.URL).Info("Creating repository in database")

	query := `
		INSERT INTO code_analyzer.repositories (kind, url, name, owner, local_path, index_status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		repo.Kind,
		repo.URL,
		repo.Name,
		repo.Owner,
		repo.LocalPath,
		repo.IndexStatus,
	).Scan(&repo.ID, &repo.CreatedAt, &repo.UpdatedAt)

	if err != nil {
		r.log().WithField("error", err).Error("Failed to create repository")
	}
	return err
}

// UpdateRepositoryStatus updates the status of a repository
func (r *CodeAnalyzerRepository) UpdateRepositoryStatus(id int64, status string, errorMsg string) error {
	r.log().WithFields(logrus.Fields{
		"id":     id,
		"status": status,
	}).Info("Updating repository status")

	var lastIndexed *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now()
		lastIndexed = &now
	}

	query := `
		UPDATE code_analyzer.repositories
		SET index_status = $1, index_error = $2, last_indexed = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.DB.Exec(query, status, errorMsg, lastIndexed, id)
	if err != nil {
		r.log().WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Failed to update repository status")
	}
	return err
}

// GetRepositoryByURL gets a repository by its URL
func (r *CodeAnalyzerRepository) GetRepositoryByURL(url string) (*models.Repository, error) {
	r.log().WithField("url", url).Debug("Getting repository by URL")

	var repo models.Repository
	query := `
		SELECT id, kind, url, name, owner, local_path, last_indexed, index_status, index_error, created_at, updated_at
		FROM code_analyzer.repositories
		WHERE url = $1
	`

	err := r.DB.Get(&repo, query, url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().WithField("url", url).Debug("Repository not found")
			return nil, nil // Repository not found
		}
		r.log().WithFields(logrus.Fields{
			"url":   url,
			"error": err,
		}).Error("Error getting repository by URL")
		return nil, err
	}

	return &repo, nil
}

// GetRepositoryByID gets a repository by its ID
func (r *CodeAnalyzerRepository) GetRepositoryByID(id int64) (*models.Repository, error) {
	r.log().WithField("id", id).Debug("Getting repository by ID")

	var repo models.Repository
	query := `
		SELECT id, kind, url, name, owner, local_path, last_indexed, index_status, index_error, created_at, updated_at
		FROM code_analyzer.repositories
		WHERE id = $1
	`

	err := r.DB.Get(&repo, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().WithField("id", id).Debug("Repository not found")
			return nil, nil // Repository not found
		}
		r.log().WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Error getting repository by ID")
		return nil, err
	}

	return &repo, nil
}

// CreateRepositoryFile creates a new file entry in the database
func (r *CodeAnalyzerRepository) CreateRepositoryFile(file *models.RepositoryFile) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id":   file.RepositoryID,
		"file_path": file.FilePath,
	})).Debug("Creating repository file")

	query := `
		INSERT INTO code_analyzer.repository_files (repository_id, file_path, package, last_analyzed)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (repository_id, file_path) 
		DO UPDATE SET package = $3, last_analyzed = $4, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		file.RepositoryID,
		file.FilePath,
		file.Package,
		file.LastAnalyzed,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id":   file.RepositoryID,
			"file_path": file.FilePath,
			"error":     err,
		})).Error("Failed to create repository file")
	}
	return err
}

// GetRepositoryFiles gets all files for a repository
func (r *CodeAnalyzerRepository) GetRepositoryFiles(repoID int64) ([]models.RepositoryFile, error) {
	r.log().WithField("repo_id", repoID).Debug("Getting all files for repository")

	var files []models.RepositoryFile
	query := `
		SELECT id, repository_id, file_path, package, last_analyzed, created_at, updated_at
		FROM code_analyzer.repository_files
		WHERE repository_id = $1
		ORDER BY file_path
	`

	err := r.DB.Select(&files, query, repoID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"error":   err,
		})).Error("Failed to get repository files")
		return nil, err
	}

	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"count":   len(files),
	})).Debug("Retrieved repository files")
	return files, nil
}

// GetRepositoryFileByPath gets a specific file by path
func (r *CodeAnalyzerRepository) GetRepositoryFileByPath(repoID int64, filePath string) (*models.RepositoryFile, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id":   repoID,
		"file_path": filePath,
	})).Debug("Getting repository file by path")

	var file models.RepositoryFile
	query := `
		SELECT id, repository_id, file_path, package, last_analyzed, created_at, updated_at
		FROM code_analyzer.repository_files
		WHERE repository_id = $1 AND file_path = $2
	`

	err := r.DB.Get(&file, query, repoID, filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"repo_id":   repoID,
				"file_path": filePath,
			})).Debug("File not found")
			return nil, nil // File not found
		}
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id":   repoID,
			"file_path": filePath,
			"error":     err,
		})).Error("Error getting repository file by path")
		return nil, err
	}

	return &file, nil
}

// BatchCreateFunctions inserts functions and their related data
func (r *CodeAnalyzerRepository) BatchCreateFunctions(functions []models.RepositoryFunction) error {
	r.log().WithField("count", len(functions)).Info("Creating repository functions in batch")

	tx, err := r.DB.Begin()
	if err != nil {
		r.log().WithField("error", err).Error("Failed to begin transaction for batch function creation")
		return err
	}
	defer tx.Rollback()

	// Prepare the function insert statement
	fnStmt, err := tx.Prepare(`
		INSERT INTO code_analyzer.repository_functions (
			repository_id, file_id, name, kind, receiver, exported, parameters, results, code_block, line
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (repository_id, file_id, name, line) 
		DO UPDATE SET 
			kind = $4, receiver = $5, exported = $6, parameters = $7, results = $8, code_block = $9,
			updated_at = NOW()
		RETURNING id
	`)
	if err != nil {
		r.log().WithField("error", err).Error("Failed to prepare function insert statement")
		return err
	}
	defer fnStmt.Close()

	for _, fn := range functions {
		// Convert params, results to JSON
		paramsJSON, err := json.Marshal(fn.Parameters)
		if err != nil {
			r.log().WithField("error", err).Error("Failed to marshal function parameters")
			return err
		}

		resultsJSON, err := json.Marshal(fn.Results)
		if err != nil {
			r.log().WithField("error", err).Error("Failed to marshal function results")
			return err
		}

		// Insert function and get its ID
		var functionID int64
		err = fnStmt.QueryRow(
			fn.RepositoryID,
			fn.FileID,
			fn.Name,
			fn.Kind,
			fn.Receiver,
			fn.Exported,
			paramsJSON,
			resultsJSON,
			fn.CodeBlock,
			fn.Line,
		).Scan(&functionID)
		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function": fn.Name,
				"error":    err,
			})).Error("Failed to insert function")
			return err
		}

		// Process function calls if available
		if fn.Calls != "" {
			var calls []map[string]interface{}
			if err := json.Unmarshal([]byte(fn.Calls), &calls); err != nil {
				r.log().WithField("error", err).Error("Failed to unmarshal function calls")
				continue // Skip calls for this function but continue with others
			}

			// Insert each call
			for _, call := range calls {
				calleeName, _ := call["callee"].(string)
				calleePackage, _ := call["package"].(string)
				line, _ := call["line"].(float64)

				var paramsJSON []byte
				if params, ok := call["parameters"]; ok {
					paramsJSON, _ = json.Marshal(params)
				}

				_, err = tx.Exec(`
					INSERT INTO code_analyzer.function_calls (
						caller_id, callee_name, callee_package, line, parameters
					)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (caller_id, callee_name, line) DO NOTHING
				`, functionID, calleeName, calleePackage, int(line), paramsJSON)

				if err != nil {
					r.log().WithFields(fieldsToLogrus(logger.Fields{
						"callee": calleeName,
						"error":  err,
					})).Error("Failed to insert function call")
					// Continue with other calls
				}
			}
		}

		// Process function references if available
		if fn.References != "" {
			var refs []map[string]interface{}
			if err := json.Unmarshal([]byte(fn.References), &refs); err != nil {
				r.log().WithField("error", err).Error("Failed to unmarshal function references")
				continue // Skip references for this function but continue with others
			}

			// Insert each reference
			for _, ref := range refs {
				refType, _ := ref["type"].(string)
				fileID, _ := ref["file_id"].(float64)
				line, _ := ref["line"].(float64)
				column, _ := ref["column"].(float64)
				context, _ := ref["context"].(string)

				_, err = tx.Exec(`
					INSERT INTO code_analyzer.function_references (
						function_id, reference_type, file_id, line, column_position, context
					)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (function_id, file_id, line, column_position) DO NOTHING
				`, functionID, refType, int64(fileID), int(line), int(column), context)

				if err != nil {
					r.log().WithFields(fieldsToLogrus(logger.Fields{
						"ref_type": refType,
						"line":     line,
						"error":    err,
					})).Error("Failed to insert function reference")
					// Continue with other references
				}
			}
		}

		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id":    functionID,
			"statement_info": fn.StatementInfo,
			"statements":     fn.Statements,
		})).Info("Processing function statements")

		// Process function statements if available
		// recursively process statements and their children
		var processStatements func(statements []models.FunctionStatement, parentID *int64) error
		processStatements = func(statements []models.FunctionStatement, parentID *int64) error {
			for _, stmt := range statements {
				// Set the parent ID if provided
				stmt.ParentStatementID = parentID

				r.log().WithFields(fieldsToLogrus(logger.Fields{
					"function_id": functionID,
					"stmt_type":   stmt.StatementType,
					"line":        stmt.Line,
					"parent_id":   parentID,
					"text":        stmt.Text,
				})).Info("Inserting function statement")

				// Insert the statement
				var newID int64
				err = tx.QueryRow(`
            INSERT INTO code_analyzer.function_statements (
                function_id, statement_type, text, line, conditions, variables, calls, parent_statement_id
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id
        `, functionID, stmt.StatementType, stmt.Text, stmt.Line, stmt.Conditions, stmt.Variables, stmt.Calls, stmt.ParentStatementID).Scan(&newID)

				if err != nil {
					r.log().WithFields(fieldsToLogrus(logger.Fields{
						"stmt_type": stmt.StatementType,
						"line":      stmt.Line,
						"error":     err,
					})).Error("Failed to insert function statement")
					continue // Continue with other statements
				}

				// Process children if any
				if len(stmt.Children) > 0 {
					if err := processStatements(stmt.Children, &newID); err != nil {
						return err
					}
				}

			}
			return nil
		}

		// Start processing from root statements
		if err := processStatements(fn.Statements, nil); err != nil {
			r.log().WithField("error", err).Error("Failed to process function statements")
		}
	}

	if err := tx.Commit(); err != nil {
		r.log().WithField("error", err).Error("Failed to commit transaction for batch function creation")
		return err
	}

	r.log().WithField("count", len(functions)).Info("Successfully created repository functions in batch")
	return nil
}

// BatchCreateSymbols inserts symbols and their related data
func (r *CodeAnalyzerRepository) BatchCreateSymbols(symbols []models.RepositorySymbol) error {
	r.log().WithField("count", len(symbols)).Info("Creating repository symbols in batch")

	tx, err := r.DB.Begin()
	if err != nil {
		r.log().WithField("error", err).Error("Failed to begin transaction for batch symbol creation")
		return err
	}
	defer tx.Rollback()

	// Prepare the symbol insert statement
	symStmt, err := tx.Prepare(`
		INSERT INTO code_analyzer.repository_symbols (
			repository_id, file_id, name, kind, type, value, exported, fields, methods, line
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (repository_id, file_id, name, line) 
		DO UPDATE SET 
			kind = $4, type = $5, value = $6, exported = $7, fields = $8, methods = $9,
			updated_at = NOW()
		RETURNING id
	`)
	if err != nil {
		r.log().WithField("error", err).Error("Failed to prepare symbol insert statement")
		return err
	}
	defer symStmt.Close()

	for _, sym := range symbols {
		// Convert fields, methods to JSON
		fieldsJSON, err := json.Marshal(sym.Fields)
		if err != nil {
			r.log().WithField("error", err).Error("Failed to marshal symbol fields")
			return err
		}

		methodsJSON, err := json.Marshal(sym.Methods)
		if err != nil {
			r.log().WithField("error", err).Error("Failed to marshal symbol methods")
			return err
		}

		// Insert symbol and get its ID
		var symbolID int64
		err = symStmt.QueryRow(
			sym.RepositoryID,
			sym.FileID,
			sym.Name,
			sym.Kind,
			sym.Type,
			sym.Value,
			sym.Exported,
			fieldsJSON,
			methodsJSON,
			sym.Line,
		).Scan(&symbolID)
		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"symbol": sym.Name,
				"error":  err,
			})).Error("Failed to insert symbol")
			return err
		}

		// Process symbol references if available
		if sym.References != "" {
			var refs []map[string]interface{}
			if err := json.Unmarshal([]byte(sym.References), &refs); err != nil {
				r.log().WithField("error", err).Error("Failed to unmarshal symbol references")
				continue // Skip references for this symbol but continue with others
			}

			// Insert each reference
			for _, ref := range refs {
				refType, _ := ref["type"].(string)
				fileID, _ := ref["file_id"].(float64)
				line, _ := ref["line"].(float64)
				column, _ := ref["column"].(float64)
				context, _ := ref["context"].(string)

				_, err = tx.Exec(`
					INSERT INTO code_analyzer.symbol_references (
						symbol_id, reference_type, file_id, line, column_position, context
					)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (symbol_id, file_id, line, column_position) DO NOTHING
				`, symbolID, refType, int64(fileID), int(line), int(column), context)

				if err != nil {
					r.log().WithFields(fieldsToLogrus(logger.Fields{
						"ref_type": refType,
						"line":     line,
						"error":    err,
					})).Error("Failed to insert symbol reference")
					// Continue with other references
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		r.log().WithField("error", err).Error("Failed to commit transaction for batch symbol creation")
		return err
	}

	r.log().WithField("count", len(symbols)).Info("Successfully created repository symbols in batch")
	return nil
}

// GetRepositoryFunctions gets functions for a repository or specific file
func (r *CodeAnalyzerRepository) GetRepositoryFunctions(repoID int64, fileID int64) ([]models.RepositoryFunction, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"file_id": fileID,
	})).Debug("Getting repository functions")

	var functions []models.RepositoryFunction
	var query string
	var args []interface{}

	if fileID > 0 {
		query = `
			SELECT id, repository_id, file_id, name, kind, receiver, exported, 
				parameters, results, code_block, line, created_at, updated_at
			FROM code_analyzer.repository_functions
			WHERE repository_id = $1 AND file_id = $2
			ORDER BY line
		`
		args = []interface{}{repoID, fileID}
	} else {
		query = `
			SELECT id, repository_id, file_id, name, kind, receiver, exported, 
				parameters, results, code_block, line, created_at, updated_at
			FROM code_analyzer.repository_functions
			WHERE repository_id = $1
			ORDER BY file_id, line
		`
		args = []interface{}{repoID}
	}

	err := r.DB.Select(&functions, query, args...)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"file_id": fileID,
			"error":   err,
		})).Error("Failed to get repository functions")
		return nil, err
	}

	// For each function, load calls, references, and statements
	for i := range functions {
		// Load calls
		var calls []models.FunctionCall
		callsQuery := `
			SELECT id, caller_id, callee_name, callee_package, callee_id, line, parameters, created_at, updated_at
			FROM code_analyzer.function_calls
			WHERE caller_id = $1
			ORDER BY line
		`
		if err := r.DB.Select(&calls, callsQuery, functions[i].ID); err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function_id": functions[i].ID,
				"error":       err,
			})).Error("Failed to load function calls")
		} else {
			// Convert to JSON string for backward compatibility
			callsJSON, _ := json.Marshal(calls)
			functions[i].Calls = string(callsJSON)
		}

		// Load references
		var refs []models.FunctionReference
		refsQuery := `
			SELECT id, function_id, reference_type, file_id, line, column_position, context, created_at, updated_at
			FROM code_analyzer.function_references
			WHERE function_id = $1
			ORDER BY line, column_position
		`
		if err := r.DB.Select(&refs, refsQuery, functions[i].ID); err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function_id": functions[i].ID,
				"error":       err,
			})).Error("Failed to load function references")
		} else {
			// Convert to JSON string for backward compatibility
			refsJSON, _ := json.Marshal(refs)
			functions[i].References = string(refsJSON)
		}

		// Load statements
		var stmts []models.FunctionStatement
		stmtsQuery := `
			SELECT id, function_id, statement_type, text, line, conditions, variables, calls, 
			       parent_statement_id, created_at, updated_at
			FROM code_analyzer.function_statements
			WHERE function_id = $1
			ORDER BY line
		`
		if err := r.DB.Select(&stmts, stmtsQuery, functions[i].ID); err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function_id": functions[i].ID,
				"error":       err,
			})).Error("Failed to load function statements")
		} else {
			// Convert to JSON string for backward compatibility
			stmtsJSON, _ := json.Marshal(stmts)
			functions[i].StatementInfo = string(stmtsJSON)
		}
	}

	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"file_id": fileID,
		"count":   len(functions),
	})).Debug("Retrieved repository functions")
	return functions, nil
}

// GetRepositorySymbols gets symbols for a repository or specific file
func (r *CodeAnalyzerRepository) GetRepositorySymbols(repoID int64, fileID int64) ([]models.RepositorySymbol, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"file_id": fileID,
	})).Debug("Getting repository symbols")

	var symbols []models.RepositorySymbol
	var query string
	var args []interface{}

	if fileID > 0 {
		query = `
			SELECT id, repository_id, file_id, name, kind, type, value, exported, 
				fields, methods, line, created_at, updated_at
			FROM code_analyzer.repository_symbols
			WHERE repository_id = $1 AND file_id = $2
			ORDER BY line
		`
		args = []interface{}{repoID, fileID}
	} else {
		query = `
			SELECT id, repository_id, file_id, name, kind, type, value, exported, 
				fields, methods, line, created_at, updated_at
			FROM code_analyzer.repository_symbols
			WHERE repository_id = $1
			ORDER BY file_id, line
		`
		args = []interface{}{repoID}
	}

	err := r.DB.Select(&symbols, query, args...)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"file_id": fileID,
			"error":   err,
		})).Error("Failed to get repository symbols")
		return nil, err
	}

	// For each symbol, load references
	for i := range symbols {
		var refs []models.SymbolReference
		refsQuery := `
			SELECT id, symbol_id, reference_type, file_id, line, column_position, context, created_at, updated_at
			FROM code_analyzer.symbol_references
			WHERE symbol_id = $1
			ORDER BY line, column_position
		`
		if err := r.DB.Select(&refs, refsQuery, symbols[i].ID); err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"symbol_id": symbols[i].ID,
				"error":     err,
			})).Error("Failed to load symbol references")
		} else {
			// Convert to JSON string for backward compatibility
			refsJSON, _ := json.Marshal(refs)
			symbols[i].References = string(refsJSON)
		}
	}

	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"file_id": fileID,
		"count":   len(symbols),
	})).Debug("Retrieved repository symbols")
	return symbols, nil
}

// GetFunctionCalls gets all function calls for a function ID
func (r *CodeAnalyzerRepository) GetFunctionCalls(functionID int64) ([]models.FunctionCall, error) {
	r.log().WithField("function_id", functionID).Debug("Getting function calls")

	var calls []models.FunctionCall
	query := `
		SELECT id, caller_id, callee_name, callee_package, callee_id, line, parameters, created_at, updated_at
		FROM code_analyzer.function_calls
		WHERE caller_id = $1
		ORDER BY line
	`

	err := r.DB.Select(&calls, query, functionID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id": functionID,
			"error":       err,
		})).Error("Failed to get function calls")
		return nil, err
	}

	return calls, nil
}

// GetFunctionReferences gets all references to a function
func (r *CodeAnalyzerRepository) GetFunctionReferences(functionID int64) ([]models.FunctionReference, error) {
	r.log().WithField("function_id", functionID).Debug("Getting function references")

	var refs []models.FunctionReference
	query := `
		SELECT id, function_id, reference_type, file_id, line, column_position, context, created_at, updated_at
		FROM code_analyzer.function_references
		WHERE function_id = $1
		ORDER BY line, column_position
	`

	err := r.DB.Select(&refs, query, functionID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id": functionID,
			"error":       err,
		})).Error("Failed to get function references")
		return nil, err
	}

	return refs, nil
}

// GetFunctionStatements gets all statements for a function
func (r *CodeAnalyzerRepository) GetFunctionStatements(functionID int64) ([]models.FunctionStatement, error) {
	r.log().WithField("function_id", functionID).Debug("Getting function statements")

	var stmts []models.FunctionStatement
	query := `
		SELECT id, function_id, statement_type, text, line, conditions, variables, calls, 
			   parent_statement_id, created_at, updated_at
		FROM code_analyzer.function_statements
		WHERE function_id = $1
		ORDER BY line
	`

	err := r.DB.Select(&stmts, query, functionID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id": functionID,
			"error":       err,
		})).Error("Failed to get function statements")
		return nil, err
	}

	return stmts, nil
}

// GetSymbolReferences gets all references to a symbol
func (r *CodeAnalyzerRepository) GetSymbolReferences(symbolID int64) ([]models.SymbolReference, error) {
	r.log().WithField("symbol_id", symbolID).Debug("Getting symbol references")

	var refs []models.SymbolReference
	query := `
		SELECT id, symbol_id, reference_type, file_id, line, column_position, context, created_at, updated_at
		FROM code_analyzer.symbol_references
		WHERE symbol_id = $1
		ORDER BY line, column_position
	`

	err := r.DB.Select(&refs, query, symbolID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"symbol_id": symbolID,
			"error":     err,
		})).Error("Failed to get symbol references")
		return nil, err
	}

	return refs, nil
}

// AddFunctionStatement adds a new function statement to the database
func (r *CodeAnalyzerRepository) AddFunctionStatement(stmt *models.FunctionStatement) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"function_id":    stmt.FunctionID,
		"statement_type": stmt.StatementType,
		"line":           stmt.Line,
	})).Debug("Adding function statement")

	// Convert JSON fields
	conditionsJSON, err := json.Marshal(stmt.Conditions)
	if err != nil {
		return err
	}

	variablesJSON, err := json.Marshal(stmt.Variables)
	if err != nil {
		return err
	}

	callsJSON, err := json.Marshal(stmt.Calls)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO code_analyzer.function_statements (
			function_id, statement_type, text, line, conditions, variables, calls, parent_statement_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = r.DB.QueryRow(
		query,
		stmt.FunctionID,
		stmt.StatementType,
		stmt.Text,
		stmt.Line,
		conditionsJSON,
		variablesJSON,
		callsJSON,
		stmt.ParentStatementID,
	).Scan(&stmt.ID, &stmt.CreatedAt, &stmt.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id": stmt.FunctionID,
			"line":        stmt.Line,
			"error":       err,
		})).Error("Failed to add function statement")
	}
	return err
}

// BatchCreateFunctionStatements adds multiple function statements in a transaction
func (r *CodeAnalyzerRepository) BatchCreateFunctionStatements(statements []models.FunctionStatement) error {
	if len(statements) == 0 {
		return nil
	}

	r.log().WithField("count", len(statements)).Debug("Batch adding function statements")

	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range statements {
		conditionsJSON, err := json.Marshal(statements[i].Conditions)
		if err != nil {
			return err
		}

		variablesJSON, err := json.Marshal(statements[i].Variables)
		if err != nil {
			return err
		}

		callsJSON, err := json.Marshal(statements[i].Calls)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO code_analyzer.function_statements (
				function_id, statement_type, text, line, conditions, variables, calls, parent_statement_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, created_at, updated_at
		`

		err = tx.QueryRow(
			query,
			statements[i].FunctionID,
			statements[i].StatementType,
			statements[i].Text,
			statements[i].Line,
			conditionsJSON,
			variablesJSON,
			callsJSON,
			statements[i].ParentStatementID,
		).Scan(&statements[i].ID, &statements[i].CreatedAt, &statements[i].UpdatedAt)

		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function_id": statements[i].FunctionID,
				"line":        statements[i].Line,
				"error":       err,
			})).Error("Failed to add function statement in batch")
			return err
		}
	}

	return tx.Commit()
}

// AddFunctionCall adds a new function call to the database
func (r *CodeAnalyzerRepository) AddFunctionCall(call *models.FunctionCall) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"caller_id":      call.CallerID,
		"callee_name":    call.CalleeName,
		"callee_package": call.CalleePackage,
	})).Debug("Adding function call")

	// Convert JSON field
	paramsJSON, err := json.Marshal(call.Parameters)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO code_analyzer.function_calls (
			caller_id, callee_name, callee_package, callee_id, line, parameters
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (caller_id, callee_name, line) DO UPDATE
		SET callee_package = $3, callee_id = $4, parameters = $6, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err = r.DB.QueryRow(
		query,
		call.CallerID,
		call.CalleeName,
		call.CalleePackage,
		call.CalleeID,
		call.Line,
		paramsJSON,
	).Scan(&call.ID, &call.CreatedAt, &call.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"caller_id":   call.CallerID,
			"callee_name": call.CalleeName,
			"error":       err,
		})).Error("Failed to add function call")
	}
	return err
}

// UpdateFunctionCall updates an existing function call
func (r *CodeAnalyzerRepository) UpdateFunctionCall(call *models.FunctionCall) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"id":             call.ID,
		"caller_id":      call.CallerID,
		"callee_name":    call.CalleeName,
		"callee_package": call.CalleePackage,
	})).Debug("Updating function call")

	// Convert JSON field
	paramsJSON, err := json.Marshal(call.Parameters)
	if err != nil {
		return err
	}

	query := `
		UPDATE code_analyzer.function_calls
		SET callee_package = $1, callee_id = $2, parameters = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err = r.DB.Exec(
		query,
		call.CalleePackage,
		call.CalleeID,
		paramsJSON,
		call.ID,
	)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"id":    call.ID,
			"error": err,
		})).Error("Failed to update function call")
	}
	return err
}

// AddFunctionReference adds a new function reference to the database
func (r *CodeAnalyzerRepository) AddFunctionReference(ref *models.FunctionReference) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"function_id":     ref.FunctionID,
		"reference_type":  ref.ReferenceType,
		"file_id":         ref.FileID,
		"line":            ref.Line,
		"column_position": ref.ColumnPosition,
	})).Debug("Adding function reference")

	query := `
		INSERT INTO code_analyzer.function_references (
			function_id, reference_type, file_id, line, column_position, context
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (function_id, file_id, line, column_position) DO UPDATE
		SET reference_type = $2, context = $6, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		ref.FunctionID,
		ref.ReferenceType,
		ref.FileID,
		ref.Line,
		ref.ColumnPosition,
		ref.Context,
	).Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"function_id": ref.FunctionID,
			"file_id":     ref.FileID,
			"line":        ref.Line,
			"error":       err,
		})).Error("Failed to add function reference")
	}
	return err
}

// BatchAddFunctionReferences adds multiple function references in a transaction
func (r *CodeAnalyzerRepository) BatchAddFunctionReferences(refs []models.FunctionReference) error {
	if len(refs) == 0 {
		return nil
	}

	r.log().WithField("count", len(refs)).Debug("Batch adding function references")

	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range refs {
		query := `
			INSERT INTO code_analyzer.function_references (
				function_id, reference_type, file_id, line, column_position, context
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (function_id, file_id, line, column_position) DO UPDATE
			SET reference_type = $2, context = $6, updated_at = NOW()
			RETURNING id, created_at, updated_at
		`

		err = tx.QueryRow(
			query,
			refs[i].FunctionID,
			refs[i].ReferenceType,
			refs[i].FileID,
			refs[i].Line,
			refs[i].ColumnPosition,
			refs[i].Context,
		).Scan(&refs[i].ID, &refs[i].CreatedAt, &refs[i].UpdatedAt)

		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"function_id": refs[i].FunctionID,
				"file_id":     refs[i].FileID,
				"line":        refs[i].Line,
				"error":       err,
			})).Error("Failed to add function reference in batch")
			return err
		}
	}

	return tx.Commit()
}

// AddSymbolReference adds a new symbol reference to the database
func (r *CodeAnalyzerRepository) AddSymbolReference(ref *models.SymbolReference) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"symbol_id":       ref.SymbolID,
		"reference_type":  ref.ReferenceType,
		"file_id":         ref.FileID,
		"line":            ref.Line,
		"column_position": ref.ColumnPosition,
	})).Debug("Adding symbol reference")

	query := `
		INSERT INTO code_analyzer.symbol_references (
			symbol_id, reference_type, file_id, line, column_position, context
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (symbol_id, file_id, line, column_position) DO UPDATE
		SET reference_type = $2, context = $6, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		ref.SymbolID,
		ref.ReferenceType,
		ref.FileID,
		ref.Line,
		ref.ColumnPosition,
		ref.Context,
	).Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"symbol_id": ref.SymbolID,
			"file_id":   ref.FileID,
			"line":      ref.Line,
			"error":     err,
		})).Error("Failed to add symbol reference")
	}
	return err
}

// BatchAddSymbolReferences adds multiple symbol references in a transaction
func (r *CodeAnalyzerRepository) BatchAddSymbolReferences(refs []models.SymbolReference) error {
	if len(refs) == 0 {
		return nil
	}

	r.log().WithField("count", len(refs)).Debug("Batch adding symbol references")

	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range refs {
		query := `
			INSERT INTO code_analyzer.symbol_references (
				symbol_id, reference_type, file_id, line, column_position, context
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (symbol_id, file_id, line, column_position) DO UPDATE
			SET reference_type = $2, context = $6, updated_at = NOW()
			RETURNING id, created_at, updated_at
		`

		err = tx.QueryRow(
			query,
			refs[i].SymbolID,
			refs[i].ReferenceType,
			refs[i].FileID,
			refs[i].Line,
			refs[i].ColumnPosition,
			refs[i].Context,
		).Scan(&refs[i].ID, &refs[i].CreatedAt, &refs[i].UpdatedAt)

		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"symbol_id": refs[i].SymbolID,
				"file_id":   refs[i].FileID,
				"line":      refs[i].Line,
				"error":     err,
			})).Error("Failed to add symbol reference in batch")
			return err
		}
	}

	return tx.Commit()
}

// GetNestedFunctionStatements gets all statements for a function with their nested relationships
func (r *CodeAnalyzerRepository) GetNestedFunctionStatements(functionID int64) ([]models.FunctionStatement, error) {
	r.log().WithField("function_id", functionID).Debug("Getting nested function statements")

	// First, get all statements for this function
	stmts, err := r.GetFunctionStatements(functionID)
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup by ID
	stmtMap := make(map[int64]*models.FunctionStatement, len(stmts))
	for i := range stmts {
		stmtMap[stmts[i].ID] = &stmts[i]
		// Initialize children slice
		stmts[i].Children = []models.FunctionStatement{}
	}

	// Create the hierarchy
	rootStmts := []models.FunctionStatement{}
	for i := range stmts {
		if stmts[i].ParentStatementID == nil {
			// This is a root statement
			rootStmts = append(rootStmts, stmts[i])
		} else {
			// This is a child statement
			parent, exists := stmtMap[*stmts[i].ParentStatementID]
			if exists {
				parent.Children = append(parent.Children, stmts[i])
			}
		}
	}

	return rootStmts, nil
}

// GetFunctionByRepoAndName gets a function by repository ID and function name
func (r *CodeAnalyzerRepository) GetFunctionByRepoAndName(repoID int64, name string) (*models.RepositoryFunction, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"name":    name,
	})).Debug("Getting function by repo ID and name")

	var fn models.RepositoryFunction
	query := `
		SELECT id, repository_id, file_id, name, kind, receiver, exported, parameters, results, 
		       code_block, line, created_at, updated_at
		FROM code_analyzer.repository_functions
		WHERE repository_id = $1 AND name = $2
		LIMIT 1
	`

	err := r.DB.Get(&fn, query, repoID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"repo_id": repoID,
				"name":    name,
			})).Debug("Function not found")
			return nil, nil // Function not found
		}
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"name":    name,
			"error":   err,
		})).Error("Error getting function by repo ID and name")
		return nil, err
	}

	return &fn, nil
}

// GetSymbolByRepoAndName gets a symbol by repository ID and symbol name
func (r *CodeAnalyzerRepository) GetSymbolByRepoAndName(repoID int64, name string) (*models.RepositorySymbol, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"name":    name,
	})).Debug("Getting symbol by repo ID and name")

	var symbol models.RepositorySymbol
	query := `
		SELECT id, repository_id, file_id, name, kind, type, value, exported, fields, methods,
		       line, created_at, updated_at
		FROM code_analyzer.repository_symbols
		WHERE repository_id = $1 AND name = $2
		LIMIT 1
	`

	err := r.DB.Get(&symbol, query, repoID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"repo_id": repoID,
				"name":    name,
			})).Debug("Symbol not found")
			return nil, nil // Symbol not found
		}
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"name":    name,
			"error":   err,
		})).Error("Error getting symbol by repo ID and name")
		return nil, err
	}

	return &symbol, nil
}

// RemoveRepositoryData removes all data for a repository
func (r *CodeAnalyzerRepository) RemoveRepositoryData(repoID int64) error {
	r.log().WithField("repo_id", repoID).Info("Removing all data for repository")

	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Let the cascading delete constraints handle the deletion of related records
	query := `DELETE FROM code_analyzer.repositories WHERE id = $1`
	_, err = tx.Exec(query, repoID)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"error":   err,
		})).Error("Failed to remove repository data")
		return err
	}

	return tx.Commit()
}
