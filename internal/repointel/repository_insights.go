package repointel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cred.com/hack25/backend/internal/insights"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// InsightsRepository handles persistence operations for the new specialized insights tables
type InsightsRepository struct {
	DB *sqlx.DB
}

// NewInsightsRepository creates a new insights repository
func NewInsightsRepository(dbConn *sql.DB) *InsightsRepository {
	return &InsightsRepository{
		DB: sqlx.NewDb(dbConn, "postgres"),
	}
}

// Log returns a logrus entry with the repository context
func (r *InsightsRepository) log() *logrus.Entry {
	return logrus.WithField("component", "insights-repository")
}

// SaveFunctionInsight saves a function insight to the dedicated table
func (r *InsightsRepository) SaveFunctionInsight(repoID int64, functionID int64, insight *insights.FunctionInsight, model string) (int64, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"function_id":   functionID,
		"model":         model,
	}).Info("Saving function insight to database")

	// Convert insight to JSON
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal function insight: %w", err)
	}

	// Insert into function_insights table
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
		repoID,
		functionID,
		insightJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save function insight")
		return 0, fmt.Errorf("failed to save function insight: %w", err)
	}

	return id, nil
}

// GetFunctionInsight retrieves a function insight from the database
func (r *InsightsRepository) GetFunctionInsight(repoID int64, functionID int64) (*insights.FunctionInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"function_id":   functionID,
	}).Debug("Getting function insight")

	query := `
		SELECT data
		FROM code_analyzer.function_insights
		WHERE repository_id = $1 AND function_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var data []byte
	err := r.DB.QueryRow(query, repoID, functionID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No insight found
		}
		return nil, fmt.Errorf("failed to get function insight: %w", err)
	}

	// Unmarshal the data
	var insight insights.FunctionInsight
	if err := json.Unmarshal(data, &insight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal function insight: %w", err)
	}

	return &insight, nil
}

// SaveSymbolInsight saves a symbol insight to the dedicated table
func (r *InsightsRepository) SaveSymbolInsight(repoID int64, symbolID int64, insight *insights.SymbolInsight, model string) (int64, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"symbol_id":     symbolID,
		"model":         model,
	}).Info("Saving symbol insight to database")

	// Convert insight to JSON
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal symbol insight: %w", err)
	}

	// Insert into symbol_insights table
	query := `
		INSERT INTO code_analyzer.symbol_insights (
			repository_id, 
			symbol_id, 
			data,
			model
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err = r.DB.QueryRow(
		query,
		repoID,
		symbolID,
		insightJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save symbol insight")
		return 0, fmt.Errorf("failed to save symbol insight: %w", err)
	}

	return id, nil
}

// GetSymbolInsight retrieves a symbol insight from the database
func (r *InsightsRepository) GetSymbolInsight(repoID int64, symbolID int64) (*insights.SymbolInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"symbol_id":     symbolID,
	}).Debug("Getting symbol insight")

	query := `
		SELECT data
		FROM code_analyzer.symbol_insights
		WHERE repository_id = $1 AND symbol_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var data []byte
	err := r.DB.QueryRow(query, repoID, symbolID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No insight found
		}
		return nil, fmt.Errorf("failed to get symbol insight: %w", err)
	}

	// Unmarshal the data
	var insight insights.SymbolInsight
	if err := json.Unmarshal(data, &insight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal symbol insight: %w", err)
	}

	return &insight, nil
}

// SaveStructInsight saves a struct insight to the dedicated table
func (r *InsightsRepository) SaveStructInsight(repoID int64, symbolID int64, insight *insights.StructInsight, model string) (int64, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"symbol_id":     symbolID,
		"model":         model,
	}).Info("Saving struct insight to database")

	// Convert insight to JSON
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal struct insight: %w", err)
	}

	// Insert into struct_insights table
	query := `
		INSERT INTO code_analyzer.struct_insights (
			repository_id, 
			symbol_id, 
			data,
			model
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err = r.DB.QueryRow(
		query,
		repoID,
		symbolID,
		insightJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save struct insight")
		return 0, fmt.Errorf("failed to save struct insight: %w", err)
	}

	return id, nil
}

// GetStructInsight retrieves a struct insight from the database
func (r *InsightsRepository) GetStructInsight(repoID int64, symbolID int64) (*insights.StructInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"symbol_id":     symbolID,
	}).Debug("Getting struct insight")

	query := `
		SELECT data
		FROM code_analyzer.struct_insights
		WHERE repository_id = $1 AND symbol_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var data []byte
	err := r.DB.QueryRow(query, repoID, symbolID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No insight found
		}
		return nil, fmt.Errorf("failed to get struct insight: %w", err)
	}

	// Unmarshal the data
	var insight insights.StructInsight
	if err := json.Unmarshal(data, &insight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal struct insight: %w", err)
	}

	return &insight, nil
}

// SaveFileInsight saves a file insight to the dedicated table
func (r *InsightsRepository) SaveFileInsight(repoID int64, fileID int64, insight *insights.FileInsight, model string) (int64, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"file_id":       fileID,
		"model":         model,
	}).Info("Saving file insight to database")

	// Convert insight to JSON
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal file insight: %w", err)
	}

	// Insert into file_insights table
	query := `
		INSERT INTO code_analyzer.file_insights (
			repository_id, 
			file_id, 
			data,
			model
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err = r.DB.QueryRow(
		query,
		repoID,
		fileID,
		insightJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save file insight")
		return 0, fmt.Errorf("failed to save file insight: %w", err)
	}

	return id, nil
}

// GetFileInsight retrieves a file insight from the database
func (r *InsightsRepository) GetFileInsight(repoID int64, fileID int64) (*insights.FileInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"file_id":       fileID,
	}).Debug("Getting file insight")

	query := `
		SELECT data
		FROM code_analyzer.file_insights
		WHERE repository_id = $1 AND file_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var data []byte
	err := r.DB.QueryRow(query, repoID, fileID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No insight found
		}
		return nil, fmt.Errorf("failed to get file insight: %w", err)
	}

	// Unmarshal the data
	var insight insights.FileInsight
	if err := json.Unmarshal(data, &insight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file insight: %w", err)
	}

	return &insight, nil
}

// SaveRepositoryInsight saves a repository insight to the dedicated table
func (r *InsightsRepository) SaveRepositoryInsight(repoID int64, insight *insights.RepositoryInsight, model string) (int64, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
		"model":         model,
	}).Info("Saving repository insight to database")

	// Convert insight to JSON
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal repository insight: %w", err)
	}

	// Insert into repository_insights table
	query := `
		INSERT INTO code_analyzer.repository_insights (
			repository_id, 
			data,
			model
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err = r.DB.QueryRow(
		query,
		repoID,
		insightJSON,
		model,
	).Scan(&id)

	if err != nil {
		r.log().WithError(err).Error("Failed to save repository insight")
		return 0, fmt.Errorf("failed to save repository insight: %w", err)
	}

	return id, nil
}

// GetRepositoryInsight retrieves a repository insight from the database
func (r *InsightsRepository) GetRepositoryInsight(repoID int64) (*insights.RepositoryInsight, error) {
	r.log().WithFields(logrus.Fields{
		"repository_id": repoID,
	}).Debug("Getting repository insight")

	query := `
		SELECT data
		FROM code_analyzer.repository_insights
		WHERE repository_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var data []byte
	err := r.DB.QueryRow(query, repoID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No insight found
		}
		return nil, fmt.Errorf("failed to get repository insight: %w", err)
	}

	// Unmarshal the data
	var insight insights.RepositoryInsight
	if err := json.Unmarshal(data, &insight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repository insight: %w", err)
	}

	return &insight, nil
}
