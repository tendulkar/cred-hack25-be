package handlers

import (
	"io"
	"net/http"

	"cred.com/hack25/backend/internal/service"
	"cred.com/hack25/backend/pkg/llm/client"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// LLMHandler handles HTTP requests for LLM operations
type LLMHandler struct {
	llmService *service.LLMService
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(llmService *service.LLMService) *LLMHandler {
	return &LLMHandler{
		llmService: llmService,
	}
}

// Chat handles a request to chat with an LLM
func (h *LLMHandler) Chat(c *gin.Context) {
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid chat request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		logger.Warn("Empty messages in chat request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages cannot be empty"})
		return
	}

	// Process chat request
	resp, err := h.llmService.Chat(c.Request.Context(), req)
	if err != nil {
		logger.Errorf("Chat error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process chat request"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// StreamChat handles a streaming request to chat with an LLM
func (h *LLMHandler) StreamChat(c *gin.Context) {
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid stream chat request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		logger.Warn("Empty messages in stream chat request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages cannot be empty"})
		return
	}

	// Force streaming
	req.Stream = true

	// Set up SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Create a channel to signal client disconnection
	clientGone := c.Request.Context().Done()
	
	// Create a function to send SSE events
	sendEvent := func(chunk string) error {
		select {
		case <-clientGone:
			return io.EOF
		default:
			c.Writer.Write([]byte("data: " + chunk + "\n\n"))
			c.Writer.Flush()
			return nil
		}
	}

	// Stream the response
	err := h.llmService.StreamChat(c.Request.Context(), req, sendEvent)
	if err != nil && err != io.EOF {
		logger.Errorf("Stream chat error: %v", err)
		// We've already started sending events, so we can't change the status code now
		sendEvent("Error: " + err.Error())
	}

	// Send end of stream marker
	sendEvent("[DONE]")
}

// Embedding generates an embedding for the given text
func (h *LLMHandler) Embedding(c *gin.Context) {
	var req struct {
		Text      string `json:"text" binding:"required"`
		ModelName string `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid embedding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate embedding
	embedding, err := h.llmService.GenerateEmbedding(c.Request.Context(), req.Text, req.ModelName)
	if err != nil {
		logger.Errorf("Embedding error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate embedding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"embedding": embedding})
}

// Models returns a list of available models
func (h *LLMHandler) Models(c *gin.Context) {
	// Get the list of default models
	models := client.DefaultModels()
	
	// Format for response
	var modelList []map[string]interface{}
	for fullName, model := range models {
		modelList = append(modelList, map[string]interface{}{
			"id":          fullName,
			"name":        model.Name,
			"provider":    model.Provider,
			"max_tokens":  model.MaxTokens,
			"temperature": model.Temperature,
			"top_p":       model.TopP,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{"models": modelList})
}
