package repository

import (
	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/pkg/logger"
)

// AddFileDependency adds a new file dependency to the database
func (r *CodeAnalyzerRepository) AddFileDependency(dep *models.FileDependency) error {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"file_id":      dep.FileID,
		"import_path":  dep.ImportPath,
		"alias":        dep.Alias,
		"is_stdlib":    dep.IsStdlib,
	})).Debug("Adding file dependency")

	query := `
		INSERT INTO code_analyzer.file_dependencies (
			repository_id, file_id, import_path, alias, is_stdlib
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (file_id, import_path) DO UPDATE
		SET alias = $4, is_stdlib = $5, updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		dep.RepositoryID,
		dep.FileID,
		dep.ImportPath,
		dep.Alias,
		dep.IsStdlib,
	).Scan(&dep.ID, &dep.CreatedAt, &dep.UpdatedAt)

	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"file_id":     dep.FileID,
			"import_path": dep.ImportPath,
			"error":       err,
		})).Error("Failed to add file dependency")
	}
	return err
}

// BatchAddFileDependencies adds multiple file dependencies in a transaction
func (r *CodeAnalyzerRepository) BatchAddFileDependencies(deps []models.FileDependency) error {
	if len(deps) == 0 {
		return nil
	}

	r.log().WithField("count", len(deps)).Debug("Batch adding file dependencies")

	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range deps {
		query := `
			INSERT INTO code_analyzer.file_dependencies (
				repository_id, file_id, import_path, alias, is_stdlib
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (file_id, import_path) DO UPDATE
			SET alias = $4, is_stdlib = $5, updated_at = NOW()
			RETURNING id, created_at, updated_at
		`

		err = tx.QueryRow(
			query,
			deps[i].RepositoryID,
			deps[i].FileID,
			deps[i].ImportPath,
			deps[i].Alias,
			deps[i].IsStdlib,
		).Scan(&deps[i].ID, &deps[i].CreatedAt, &deps[i].UpdatedAt)

		if err != nil {
			r.log().WithFields(fieldsToLogrus(logger.Fields{
				"file_id":     deps[i].FileID,
				"import_path": deps[i].ImportPath,
				"error":       err,
			})).Error("Failed to add file dependency in batch")
			return err
		}
	}

	r.log().WithField("count", len(deps)).Info("Successfully added file dependencies in batch")
	return tx.Commit()
}

// GetFileDependencies gets all dependencies for a repository or specific file
func (r *CodeAnalyzerRepository) GetFileDependencies(repoID int64, fileID int64) ([]models.FileDependency, error) {
	r.log().WithFields(fieldsToLogrus(logger.Fields{
		"repo_id": repoID,
		"file_id": fileID,
	})).Debug("Getting file dependencies")

	var deps []models.FileDependency
	var query string
	var args []interface{}

	if fileID > 0 {
		query = `
			SELECT id, repository_id, file_id, import_path, alias, is_stdlib, created_at, updated_at
			FROM code_analyzer.file_dependencies
			WHERE repository_id = $1 AND file_id = $2
		`
		args = []interface{}{repoID, fileID}
	} else {
		query = `
			SELECT id, repository_id, file_id, import_path, alias, is_stdlib, created_at, updated_at
			FROM code_analyzer.file_dependencies
			WHERE repository_id = $1
		`
		args = []interface{}{repoID}
	}

	err := r.DB.Select(&deps, query, args...)
	if err != nil {
		r.log().WithFields(fieldsToLogrus(logger.Fields{
			"repo_id": repoID,
			"file_id": fileID,
			"error":   err,
		})).Error("Failed to get file dependencies")
		return nil, err
	}

	return deps, nil
}
