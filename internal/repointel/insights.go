package repointel

import (
	"fmt"

	"cred.com/hack25/backend/pkg/logger"
	"github.com/sirupsen/logrus"
)

// InsightsManager provides a reusable interface for generating and storing insights
type InsightsManager struct {
	service *Service
	repo    *Repository
	logger  *logrus.Entry
}

// NewInsightsManager creates a new insights manager
func NewInsightsManager(service *Service, repo *Repository) *InsightsManager {
	return &InsightsManager{
		service: service,
		repo:    repo,
		logger:  logger.Log.WithField("component", "repointel-insights-manager"),
	}
}

// GenerateAndSaveFunctionInsight generates and saves an insight for a function
func (im *InsightsManager) GenerateAndSaveFunctionInsight(repoID int64, functionID int64, modelName string) (*FunctionInsight, error) {
	im.logger.WithFields(logrus.Fields{
		"repo_id":     repoID,
		"function_id": functionID,
		"model_name":  modelName,
	}).Debug("Generating function insight")

	// Generate the insight
	insight, err := im.service.GenerateFunctionInsight(repoID, functionID, modelName)
	if err != nil {
		im.logger.WithError(err).Error("Failed to generate function insight")
		return nil, fmt.Errorf("failed to generate function insight: %w", err)
	}

	// Save the insight
	err = im.repo.SaveFunctionInsight(repoID, functionID, insight, modelName)
	if err != nil {
		im.logger.WithError(err).Error("Failed to save function insight")
		return nil, fmt.Errorf("failed to save function insight: %w", err)
	}

	im.logger.WithFields(logrus.Fields{
		"repo_id":     repoID,
		"function_id": functionID,
	}).Info("Function insight generated and saved successfully")

	return insight, nil
}

// GetFunctionInsights retrieves insights for a specific function
func (im *InsightsManager) GetFunctionInsights(repoID int64, functionID int64) ([]*FunctionInsight, error) {
	im.logger.WithFields(logrus.Fields{
		"repo_id":     repoID,
		"function_id": functionID,
	}).Debug("Retrieving function insights")

	record, err := im.repo.GetFunctionInsight(repoID, functionID)
	if err != nil {
		im.logger.WithError(err).Error("Failed to retrieve function insights")
		return nil, fmt.Errorf("failed to retrieve function insights: %w", err)
	}

	return []*FunctionInsight{record}, nil
}
