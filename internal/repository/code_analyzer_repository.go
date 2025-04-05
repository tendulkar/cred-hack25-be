package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"cred.com/hack25/backend/internal/models"
	"github.com/jmoiron/sqlx"
)

// CodeAnalyzerRepository handles interactions with the code analyzer database tables
type CodeAnalyzerRepository struct {
	DB *sqlx.DB
}

// NewCodeAnalyzerRepository creates a new CodeAnalyzerRepository
func NewCodeAnalyzerRepository(db *sqlx.DB) *CodeAnalyzerRepository {
	return &CodeAnalyzerRepository{DB: db}
}

// CreateRepository creates a new repository in the database
func (r *CodeAnalyzerRepository) CreateRepository(repo *models.Repository) error {
	query := `
		INSERT INTO repositories (kind, url, name, owner, local_path, index_status)
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

	return err
}

// UpdateRepositoryStatus updates the status of a repository
func (r *CodeAnalyzerRepository) UpdateRepositoryStatus(id int64, status string, errorMsg string) error {
	var lastIndexed *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now()
		lastIndexed = &now
	}

	query := `
		UPDATE repositories
		SET index_status = $1, index_error = $2, last_indexed = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.DB.Exec(query, status, errorMsg, lastIndexed, id)
	return err
}

// GetRepositoryByURL gets a repository by its URL
func (r *CodeAnalyzerRepository) GetRepositoryByURL(url string) (*models.Repository, error) {
	var repo models.Repository
	query := `
		SELECT id, kind, url, name, owner, local_path, last_indexed, index_status, index_error, created_at, updated_at
		FROM repositories
		WHERE url = $1
	`

	err := r.DB.Get(&repo, query, url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Repository not found
		}
		return nil, err
	}

	return &repo, nil
}

// GetRepositoryByID gets a repository by its ID
func (r *CodeAnalyzerRepository) GetRepositoryByID(id int64) (*models.Repository, error) {
	var repo models.Repository
	query := `
		SELECT id, kind, url, name, owner, local_path, last_indexed, index_status, index_error, created_at, updated_at
		FROM repositories
		WHERE id = $1
	`

	err := r.DB.Get(&repo, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Repository not found
		}
		return nil, err
	}

	return &repo, nil
}

// CreateRepositoryFile creates a new file entry in the database
func (r *CodeAnalyzerRepository) CreateRepositoryFile(file *models.RepositoryFile) error {
	query := `
		INSERT INTO repository_files (repository_id, file_path, package, last_analyzed)
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

	return err
}

// GetRepositoryFiles gets all files for a repository
func (r *CodeAnalyzerRepository) GetRepositoryFiles(repoID int64) ([]models.RepositoryFile, error) {
	var files []models.RepositoryFile
	query := `
		SELECT id, repository_id, file_path, package, last_analyzed, created_at, updated_at
		FROM repository_files
		WHERE repository_id = $1
		ORDER BY file_path
	`

	err := r.DB.Select(&files, query, repoID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// GetRepositoryFileByPath gets a specific file by path
func (r *CodeAnalyzerRepository) GetRepositoryFileByPath(repoID int64, filePath string) (*models.RepositoryFile, error) {
	var file models.RepositoryFile
	query := `
		SELECT id, repository_id, file_path, package, last_analyzed, created_at, updated_at
		FROM repository_files
		WHERE repository_id = $1 AND file_path = $2
	`

	err := r.DB.Get(&file, query, repoID, filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // File not found
		}
		return nil, err
	}

	return &file, nil
}

// BatchCreateFunctions inserts or updates multiple functions at once
func (r *CodeAnalyzerRepository) BatchCreateFunctions(functions []models.RepositoryFunction) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare the statement
	stmt, err := tx.Prepare(`
		INSERT INTO repository_functions (
			repository_id, file_id, name, kind, receiver, exported, 
			parameters, results, code_block, line, calls, called_by, references, statement_info
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (repository_id, file_id, name, line) 
		DO UPDATE SET 
			kind = $4, receiver = $5, exported = $6, parameters = $7, results = $8, 
			code_block = $9, calls = $11, called_by = $12, references = $13, 
			statement_info = $14, updated_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, fn := range functions {
		// Convert params, results to JSON
		paramsJSON, err := json.Marshal(fn.Parameters)
		if err != nil {
			return err
		}

		resultsJSON, err := json.Marshal(fn.Results)
		if err != nil {
			return err
		}

		callsJSON, err := json.Marshal(fn.Calls)
		if err != nil {
			return err
		}

		calledByJSON, err := json.Marshal(fn.CalledBy)
		if err != nil {
			return err
		}

		referencesJSON, err := json.Marshal(fn.References)
		if err != nil {
			return err
		}

		statementInfoJSON, err := json.Marshal(fn.StatementInfo)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
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
			callsJSON,
			calledByJSON,
			referencesJSON,
			statementInfoJSON,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchCreateSymbols inserts or updates multiple symbols at once
func (r *CodeAnalyzerRepository) BatchCreateSymbols(symbols []models.RepositorySymbol) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare the statement
	stmt, err := tx.Prepare(`
		INSERT INTO repository_symbols (
			repository_id, file_id, name, kind, type, value, exported, 
			fields, methods, line, references
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (repository_id, file_id, name, line) 
		DO UPDATE SET 
			kind = $4, type = $5, value = $6, exported = $7, fields = $8, 
			methods = $9, references = $11, updated_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sym := range symbols {
		// Convert fields, methods to JSON
		fieldsJSON, err := json.Marshal(sym.Fields)
		if err != nil {
			return err
		}

		methodsJSON, err := json.Marshal(sym.Methods)
		if err != nil {
			return err
		}

		referencesJSON, err := json.Marshal(sym.References)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
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
			referencesJSON,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetRepositoryFunctions gets functions for a repository or specific file
func (r *CodeAnalyzerRepository) GetRepositoryFunctions(repoID int64, fileID int64) ([]models.RepositoryFunction, error) {
	var functions []models.RepositoryFunction
	var query string
	var args []interface{}

	if fileID > 0 {
		query = `
			SELECT id, repository_id, file_id, name, kind, receiver, exported, 
				parameters, results, code_block, line, calls, called_by, references, statement_info,
				created_at, updated_at
			FROM repository_functions
			WHERE repository_id = $1 AND file_id = $2
			ORDER BY line
		`
		args = []interface{}{repoID, fileID}
	} else {
		query = `
			SELECT id, repository_id, file_id, name, kind, receiver, exported, 
				parameters, results, code_block, line, calls, called_by, references, statement_info,
				created_at, updated_at
			FROM repository_functions
			WHERE repository_id = $1
			ORDER BY file_id, line
		`
		args = []interface{}{repoID}
	}

	err := r.DB.Select(&functions, query, args...)
	if err != nil {
		return nil, err
	}

	return functions, nil
}

// GetRepositorySymbols gets symbols for a repository or specific file
func (r *CodeAnalyzerRepository) GetRepositorySymbols(repoID int64, fileID int64) ([]models.RepositorySymbol, error) {
	var symbols []models.RepositorySymbol
	var query string
	var args []interface{}

	if fileID > 0 {
		query = `
			SELECT id, repository_id, file_id, name, kind, type, value, exported, 
				fields, methods, line, references, created_at, updated_at
			FROM repository_symbols
			WHERE repository_id = $1 AND file_id = $2
			ORDER BY line
		`
		args = []interface{}{repoID, fileID}
	} else {
		query = `
			SELECT id, repository_id, file_id, name, kind, type, value, exported, 
				fields, methods, line, references, created_at, updated_at
			FROM repository_symbols
			WHERE repository_id = $1
			ORDER BY file_id, line
		`
		args = []interface{}{repoID}
	}

	err := r.DB.Select(&symbols, query, args...)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}
