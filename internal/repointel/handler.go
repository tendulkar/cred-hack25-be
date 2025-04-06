package repointel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"cred.com/hack25/backend/pkg/logger"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles HTTP requests for repository intelligence
type Handler struct {
	service    *Service
	repository *Repository
}

// NewHandler creates a new repository intelligence handler
func NewHandler(service *Service, repository *Repository) *Handler {
	return &Handler{
		service:    service,
		repository: repository,
	}
}

// log returns a logrus entry with the handler context
func (h *Handler) log() *logrus.Entry {
	return logger.Log.WithField("component", "repointel-handler")
}

// RegisterRoutes registers the routes for the handler
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/repointel/function/{repoId}/{functionId}", h.GetFunctionInsight).Methods("GET")
	router.HandleFunc("/api/repointel/function/{repoId}/{functionId}", h.GenerateFunctionInsight).Methods("POST")

	router.HandleFunc("/api/repointel/symbol/{repoId}/{symbolId}", h.GetSymbolInsight).Methods("GET")
	router.HandleFunc("/api/repointel/symbol/{repoId}/{symbolId}", h.GenerateSymbolInsight).Methods("POST")

	router.HandleFunc("/api/repointel/struct/{repoId}/{symbolId}", h.GetStructInsight).Methods("GET")
	router.HandleFunc("/api/repointel/struct/{repoId}/{symbolId}", h.GenerateStructInsight).Methods("POST")

	router.HandleFunc("/api/repointel/repo/{repoId}/insights", h.ListInsights).Methods("GET")
}

// GetFunctionInsight handles GET requests for function insights
func (h *Handler) GetFunctionInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	functionIDStr := vars["functionId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	functionID, err := strconv.ParseInt(functionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	insight, err := h.repository.GetFunctionInsight(repoID, functionID)
	if err != nil {
		h.log().WithError(err).Error("Failed to get function insight")
		http.Error(w, "Failed to get function insight", http.StatusInternalServerError)
		return
	}

	if insight == nil {
		http.Error(w, "Function insight not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// GenerateFunctionInsight handles POST requests to generate function insights
func (h *Handler) GenerateFunctionInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	functionIDStr := vars["functionId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	functionID, err := strconv.ParseInt(functionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Parse request body for optional model name
	var req struct {
		ModelName string `json:"model_name"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Create insights manager and generate the insight
	insightsManager := NewInsightsManager(h.service, h.repository)
	insight, err := insightsManager.GenerateAndSaveFunctionInsight(repoID, functionID, req.ModelName)
	if err != nil {
		h.log().WithError(err).Error("Failed to generate and save function insight")
		http.Error(w, fmt.Sprintf("Failed to generate function insight: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// GetSymbolInsight handles GET requests for symbol insights
func (h *Handler) GetSymbolInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	symbolIDStr := vars["symbolId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	symbolID, err := strconv.ParseInt(symbolIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid symbol ID", http.StatusBadRequest)
		return
	}

	insight, err := h.repository.GetSymbolInsight(repoID, symbolID)
	if err != nil {
		h.log().WithError(err).Error("Failed to get symbol insight")
		http.Error(w, "Failed to get symbol insight", http.StatusInternalServerError)
		return
	}

	if insight == nil {
		http.Error(w, "Symbol insight not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// GenerateSymbolInsight handles POST requests to generate symbol insights
func (h *Handler) GenerateSymbolInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	symbolIDStr := vars["symbolId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	symbolID, err := strconv.ParseInt(symbolIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid symbol ID", http.StatusBadRequest)
		return
	}

	// Parse request body for optional model name
	var req struct {
		ModelName string `json:"model_name"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Generate the insight
	insight, err := h.service.GenerateSymbolInsight(repoID, symbolID, req.ModelName)
	if err != nil {
		h.log().WithError(err).Error("Failed to generate symbol insight")
		http.Error(w, fmt.Sprintf("Failed to generate symbol insight: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the insight
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		h.log().WithError(err).Error("Failed to marshal symbol insight")
		http.Error(w, "Failed to marshal symbol insight", http.StatusInternalServerError)
		return
	}

	insightRecord := &InsightRecord{
		RepositoryID: repoID,
		SymbolID:     &symbolID,
		Type:         InsightTypeSymbol,
		Data:         string(insightJSON),
		Model:        req.ModelName,
	}

	err = h.repository.SaveInsight(insightRecord)
	if err != nil {
		h.log().WithError(err).Error("Failed to save symbol insight")
		http.Error(w, "Failed to save symbol insight", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// GetStructInsight handles GET requests for struct insights
func (h *Handler) GetStructInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	symbolIDStr := vars["symbolId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	symbolID, err := strconv.ParseInt(symbolIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid symbol ID", http.StatusBadRequest)
		return
	}

	insight, err := h.repository.GetStructInsight(repoID, symbolID)
	if err != nil {
		h.log().WithError(err).Error("Failed to get struct insight")
		http.Error(w, "Failed to get struct insight", http.StatusInternalServerError)
		return
	}

	if insight == nil {
		http.Error(w, "Struct insight not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// GenerateStructInsight handles POST requests to generate struct insights
func (h *Handler) GenerateStructInsight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]
	symbolIDStr := vars["symbolId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	symbolID, err := strconv.ParseInt(symbolIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid symbol ID", http.StatusBadRequest)
		return
	}

	// Parse request body for optional model name
	var req struct {
		ModelName string `json:"model_name"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Generate the insight
	insight, err := h.service.GenerateStructInsight(repoID, symbolID, req.ModelName)
	if err != nil {
		h.log().WithError(err).Error("Failed to generate struct insight")
		http.Error(w, fmt.Sprintf("Failed to generate struct insight: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the insight
	insightJSON, err := json.Marshal(insight)
	if err != nil {
		h.log().WithError(err).Error("Failed to marshal struct insight")
		http.Error(w, "Failed to marshal struct insight", http.StatusInternalServerError)
		return
	}

	insightRecord := &InsightRecord{
		RepositoryID: repoID,
		SymbolID:     &symbolID,
		Type:         InsightTypeStruct,
		Data:         string(insightJSON),
		Model:        req.ModelName,
	}

	err = h.repository.SaveInsight(insightRecord)
	if err != nil {
		h.log().WithError(err).Error("Failed to save struct insight")
		http.Error(w, "Failed to save struct insight", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insight)
}

// ListInsights handles GET requests to list all insights for a repository
func (h *Handler) ListInsights(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoIDStr := vars["repoId"]

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	insights, err := h.repository.ListInsightsByRepository(repoID)
	if err != nil {
		h.log().WithError(err).Error("Failed to list insights")
		http.Error(w, "Failed to list insights", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insights)
}
