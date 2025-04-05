package handlers

import (
	"net/http"

	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/internal/service"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CodeAnalysisHandler handles HTTP requests for code analysis operations
type CodeAnalysisHandler struct {
	codeAnalysisService *service.CodeAnalysisService
}

// NewCodeAnalysisHandler creates a new code analysis handler
func NewCodeAnalysisHandler(codeAnalysisService *service.CodeAnalysisService) *CodeAnalysisHandler {
	return &CodeAnalysisHandler{
		codeAnalysisService: codeAnalysisService,
	}
}

// AnalyzeRepository handles a request to analyze a GitHub repository
func (h *CodeAnalysisHandler) AnalyzeRepository(c *gin.Context) {
	var req models.CodeAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid repository analysis request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.RepoURL == "" {
		logger.Warn("Empty repository URL in analysis request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository URL cannot be empty"})
		return
	}

	// Process analysis request
	result, err := h.codeAnalysisService.AnalyzeRepository(c.Request.Context(), req.RepoURL, req.AuthToken)
	if err != nil {
		logger.Errorf("Repository analysis error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze repository"})
		return
	}

	c.JSON(http.StatusOK, result)
}
