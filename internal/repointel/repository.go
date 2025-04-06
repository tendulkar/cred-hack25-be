package repointel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cred.com/hack25/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Repository handles persistence operations for repository insights
type Repository struct {
	DB *sqlx.DB
}

// NewRepository creates a new insights repository
func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		DB: sqlx.NewDb(dbConn, "postgres"),
	}
}

// log returns a logrus entry with the repository context
func (r *Repository) log() *logrus.Entry {
	return logger.Log.WithField("component", "repointel-repository")
}

// SaveInsight saves an insight to the database
func (r *Repository) SaveInsight(insight *InsightRecord) error {
	r.log().WithFields(logrus.Fields{
		"repository_id": insight.RepositoryID,
		"type":          insight.Type,
	}).Info("Saving insight to database")

	query := `
		INSERT INTO code_analyzer.insights (
			repository_id, 
			file_id, 
			function_id, 
			symbol_id, 
			path,
			type, 
			data,
			model
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(
		query,
		insight.RepositoryID,
		insight.FileID,
		insight.FunctionID,
		insight.SymbolID,
		insight.Path,
		insight.Type,
		insight.Data,
		insight.Model,
	).Scan(&insight.ID, &insight.CreatedAt, &insight.UpdatedAt)

	if err != nil {
		r.log().WithField("error", err).Error("Failed to save insight")
		return fmt.Errorf("failed to save insight: %w", err)
	}

	return nil
}

// SaveFunctionInsight saves a function insight to the function_insights table
func (r *Repository) SaveFunctionInsight(repositoryID, functionID int64, insight *FunctionInsight, model string) error {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"function_id":   functionID,
	}).Info("Saving function insight to database")

	// Convert the insight data to JSON
	dataJSON, err := json.Marshal(insight)
	if err != nil {
		r.log().WithError(err).Error("Failed to marshal function insight data")
		return fmt.Errorf("failed to marshal function insight data: %w", err)
	}

	query := `
		INSERT INTO code_analyzer.function_insights (
			repository_id, 
			function_id, 
			data,
			model
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err = r.DB.QueryRow(
		query,
		repositoryID,
		functionID,
		dataJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save function insight")
		return fmt.Errorf("failed to save function insight: %w", err)
	}

	return nil
}

// GetFunctionInsight retrieves the latest function insight from the database
func (r *Repository) GetFunctionInsight(repositoryID, functionID int64) (*FunctionInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"function_id":   functionID,
	}).Info("Getting function insight from database")

	query := `
		SELECT 
			id, 
			repository_id, 
			function_id, 
			data, 
			model, 
			created_at, 
			updated_at
		FROM code_analyzer.function_insights
		WHERE repository_id = $1 AND function_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	type functionInsightRow struct {
		ID           int64     `db:"id"`
		RepositoryID int64     `db:"repository_id"`
		FunctionID   int64     `db:"function_id"`
		Data         []byte    `db:"data"`
		Model        string    `db:"model"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	var row functionInsightRow
	err := r.DB.Get(&row, query, repositoryID, functionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().Debug("Function insight not found")
			return nil, nil
		}
		r.log().WithError(err).Error("Failed to get function insight")
		return nil, fmt.Errorf("failed to get function insight: %w", err)
	}

	// Unmarshal the JSON data into the FunctionInsight struct
	var insight FunctionInsight
	if err := json.Unmarshal(row.Data, &insight); err != nil {
		r.log().WithError(err).Error("Failed to unmarshal function insight data")
		return nil, fmt.Errorf("failed to unmarshal function insight data: %w", err)
	}

	return &insight, nil
}

// GetSymbolInsight retrieves a symbol insight from the database
func (r *Repository) GetSymbolInsight(repositoryID int64, symbolID int64) (*SymbolInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"symbol_id":     symbolID,
	}).Debug("Getting symbol insight")

	var record InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1 AND symbol_id = $2 AND type = $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.DB.Get(&record, query, repositoryID, symbolID, InsightTypeSymbol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().Debug("Symbol insight not found")
			return nil, nil
		}
		r.log().WithField("error", err).Error("Failed to get symbol insight")
		return nil, fmt.Errorf("failed to get symbol insight: %w", err)
	}

	var insight SymbolInsight
	if err := json.Unmarshal([]byte(record.Data), &insight); err != nil {
		r.log().WithField("error", err).Error("Failed to unmarshal symbol insight")
		return nil, fmt.Errorf("failed to unmarshal symbol insight: %w", err)
	}

	return &insight, nil
}

// GetStructInsight retrieves a struct insight from the database
func (r *Repository) GetStructInsight(repositoryID int64, symbolID int64) (*StructInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"symbol_id":     symbolID,
	}).Debug("Getting struct insight")

	var record InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1 AND symbol_id = $2 AND type = $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.DB.Get(&record, query, repositoryID, symbolID, InsightTypeStruct)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().Debug("Struct insight not found")
			return nil, nil
		}
		r.log().WithField("error", err).Error("Failed to get struct insight")
		return nil, fmt.Errorf("failed to get struct insight: %w", err)
	}

	var insight StructInsight
	if err := json.Unmarshal([]byte(record.Data), &insight); err != nil {
		r.log().WithField("error", err).Error("Failed to unmarshal struct insight")
		return nil, fmt.Errorf("failed to unmarshal struct insight: %w", err)
	}

	return &insight, nil
}

// GetFileInsight retrieves a file insight from the database
func (r *Repository) GetFileInsight(repositoryID int64, fileID int64) (*FileInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"file_id":       fileID,
	}).Debug("Getting file insight")

	var record InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1 AND file_id = $2 AND type = $3 AND function_id IS NULL AND symbol_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.DB.Get(&record, query, repositoryID, fileID, InsightTypeFile)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().Debug("File insight not found")
			return nil, nil
		}
		r.log().WithField("error", err).Error("Failed to get file insight")
		return nil, fmt.Errorf("failed to get file insight: %w", err)
	}

	var insight FileInsight
	if err := json.Unmarshal([]byte(record.Data), &insight); err != nil {
		r.log().WithField("error", err).Error("Failed to unmarshal file insight")
		return nil, fmt.Errorf("failed to unmarshal file insight: %w", err)
	}

	return &insight, nil
}

// GetRepositoryInsight retrieves a repository insight from the database
func (r *Repository) GetRepositoryInsight(repositoryID int64) (*RepositoryInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
	}).Debug("Getting repository insight")

	var record InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1 AND type = $2 AND file_id IS NULL AND function_id IS NULL AND symbol_id IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.DB.Get(&record, query, repositoryID, InsightTypeRepository)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log().Debug("Repository insight not found")
			return nil, nil
		}
		r.log().WithField("error", err).Error("Failed to get repository insight")
		return nil, fmt.Errorf("failed to get repository insight: %w", err)
	}

	var insight RepositoryInsight
	if err := json.Unmarshal([]byte(record.Data), &insight); err != nil {
		r.log().WithField("error", err).Error("Failed to unmarshal repository insight")
		return nil, fmt.Errorf("failed to unmarshal repository insight: %w", err)
	}

	return &insight, nil
}

// ListInsightsByRepository lists all insights for a repository
func (r *Repository) ListInsightsByRepository(repositoryID int64) ([]InsightRecord, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
	}).Debug("Listing insights for repository")

	var records []InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1
		ORDER BY created_at DESC
	`

	err := r.DB.Select(&records, query, repositoryID)
	if err != nil {
		r.log().WithField("error", err).Error("Failed to list insights")
		return nil, fmt.Errorf("failed to list insights: %w", err)
	}

	return records, nil
}

// GetFunctionInsights retrieves all function insights for a specific function
func (r *Repository) GetFunctionInsights(repositoryID int64, functionID int64) ([]InsightRecord, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"function_id":   functionID,
	}).Debug("Getting function insights")

	var records []InsightRecord
	query := `
		SELECT id, repository_id, file_id, function_id, symbol_id, type, data, model_name, created_at, updated_at
		FROM code_analyzer.insights
		WHERE repository_id = $1 AND function_id = $2 AND type = $3
		ORDER BY created_at DESC
	`

	err := r.DB.Select(&records, query, repositoryID, functionID, InsightTypeFunction)
	if err != nil {
		r.log().WithField("error", err).Error("Failed to get function insights")
		return nil, fmt.Errorf("failed to get function insights: %w", err)
	}

	return records, nil
}
