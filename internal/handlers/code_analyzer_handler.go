package handlers

import (
	"net/http"

	"cred.com/hack25/backend/internal/models"
	analyzerModels "cred.com/hack25/backend/pkg/goanalyzer/models"
	"github.com/gin-gonic/gin"
)

// CodeAnalyzerService defines the service interface for code analyzer operations
type CodeAnalyzerService interface {
	IndexRepository(url string) (*models.IndexRepositoryResponse, error)
	GetRepositoryIndex(url, filePath string) (*models.GetIndexResponse, error)
	AnalyzeGoFile(filePath string) (*analyzerModels.FileAnalysis, error)
}

// CodeAnalyzerHandler handles code analyzer API requests
type CodeAnalyzerHandler struct {
	service CodeAnalyzerService
}

// NewCodeAnalyzerHandler creates a new code analyzer handler
func NewCodeAnalyzerHandler(service CodeAnalyzerService) *CodeAnalyzerHandler {
	return &CodeAnalyzerHandler{
		service: service,
	}
}

// RegisterRoutes registers the code analyzer routes
func (h *CodeAnalyzerHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/code-analyzer")
	{
		group.POST("/repositories", h.IndexRepository)
		group.GET("/repositories", h.GetRepositoryIndex)
		group.POST("/analyze-file", h.AnalyzeFile)
	}
}

// IndexRepository handles the request to index a repository
func (h *CodeAnalyzerHandler) IndexRepository(c *gin.Context) {
	var request models.IndexRepositoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if request.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	response, err := h.service.IndexRepository(request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRepositoryIndex handles the request to get repository index information
func (h *CodeAnalyzerHandler) GetRepositoryIndex(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	filePath := c.Query("file_path")

	response, err := h.service.GetRepositoryIndex(url, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AnalyzeFileRequest represents a request to analyze a single file
type AnalyzeFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

// AnalyzeFile handles the request to analyze a single file
func (h *CodeAnalyzerHandler) AnalyzeFile(c *gin.Context) {
	var request AnalyzeFileRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	analysis, err := h.service.AnalyzeGoFile(request.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
