package structured

import (
	"encoding/json"
	"fmt"

	"cred.com/hack25/backend/internal/insights"
	"cred.com/hack25/backend/internal/models"
	"cred.com/hack25/backend/internal/repository"
	"github.com/sirupsen/logrus"
)

// Service provides methods for generating code insights using structured output from LLM
type Service struct {
	codeAnalyzerRepo *repository.CodeAnalyzerRepository
	client           *Client
	promptBuilder    *PromptBuilder
	parser           *Parser
	schemaBuilder    *SchemaBuilder
	logger           *logrus.Entry
}

// ServiceConfig holds configuration for the Service
type ServiceConfig struct {
	LiteLLMBaseURL string
	APIKey         string
	DefaultModel   string
	UseJSONFormat  bool
}

// NewService creates a new Service instance
func NewService(
	codeAnalyzerRepo *repository.CodeAnalyzerRepository,
	config ServiceConfig,
	logger *logrus.Entry,
) *Service {
	// Create the logger context
	serviceLogger := logger.WithField("component", "llm-structured-service")

	// Initialize client
	client := NewClient(
		config.LiteLLMBaseURL,
		config.APIKey,
		config.DefaultModel,
		serviceLogger,
	)

	// Initialize prompt builder
	promptBuilder := NewPromptBuilder(config.UseJSONFormat)

	// Initialize parser
	parser := NewParser(serviceLogger)

	// Initialize schema builder
	schemaBuilder := NewSchemaBuilder()

	return &Service{
		codeAnalyzerRepo: codeAnalyzerRepo,
		client:           client,
		promptBuilder:    promptBuilder,
		parser:           parser,
		schemaBuilder:    schemaBuilder,
		logger:           serviceLogger,
	}
}

// GenerateFunctionInsight generates insights for a function
func (s *Service) GenerateFunctionInsight(repoID int64, functionID int64, modelName string) (*insights.FunctionInsight, error) {
	// Get the function details
	functions, err := s.codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
	if err != nil {
		return nil, err
	}

	// Find the function with the matching ID
	var targetFunction *models.RepositoryFunction
	for i, fn := range functions {
		if fn.ID == functionID {
			targetFunction = &functions[i]
			break
		}
	}

	if targetFunction == nil {
		return nil, ErrNotFound("function", functionID)
	}

	// Get function calls
	functionCalls, err := s.codeAnalyzerRepo.GetFunctionCalls(functionID)
	if err != nil {
		return nil, err
	}

	// Prepare the prompt
	prompt := s.promptBuilder.BuildFunctionPrompt(targetFunction, functionCalls)

	// Get the JSON schema for symbol insights
	schema := s.schemaBuilder.SymbolInsightJSONSchema()

	// Call the LLM with the schema
	response, err := s.client.Call(modelName, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Parse the response into a function insight
	var insight insights.FunctionInsight
	err = json.Unmarshal([]byte(response), &insight)
	if err != nil {
		return nil, err
	}

	return &insight, nil
}

// GenerateSymbolInsight generates insights for a symbol
func (s *Service) GenerateSymbolInsight(repoID int64, symbolID int64, modelName string) (*insights.SymbolInsight, error) {
	// Get the symbol details
	symbols, err := s.codeAnalyzerRepo.GetRepositorySymbols(repoID, 0)
	if err != nil {
		return nil, err
	}

	// Find the symbol with the matching ID
	var targetSymbol *models.RepositorySymbol
	for i, sym := range symbols {
		if sym.ID == symbolID {
			targetSymbol = &symbols[i]
			break
		}
	}

	if targetSymbol == nil {
		return nil, ErrNotFound("symbol", symbolID)
	}

	// Get symbol references
	symbolRefs, err := s.codeAnalyzerRepo.GetSymbolReferences(symbolID)
	if err != nil {
		return nil, err
	}

	// Prepare the prompt
	prompt := s.promptBuilder.BuildSymbolPrompt(targetSymbol, symbolRefs)

	// Get the JSON schema for symbol insights
	schema := s.schemaBuilder.SymbolInsightJSONSchema()

	// Call the LLM with the schema
	response, err := s.client.Call(modelName, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Parse the response into a symbol insight
	var insight insights.SymbolInsight
	err = json.Unmarshal([]byte(response), &insight)
	if err != nil {
		return nil, err
	}

	return &insight, nil
}

// GenerateStructInsight generates insights for a struct
func (s *Service) GenerateStructInsight(repoID int64, symbolID int64, modelName string) (*insights.StructInsight, error) {
	// Get the symbol details (structs are stored as symbols)
	symbols, err := s.codeAnalyzerRepo.GetRepositorySymbols(repoID, 0)
	if err != nil {
		return nil, err
	}

	// Find the symbol with the matching ID
	var targetSymbol *models.RepositorySymbol
	for i, sym := range symbols {
		if sym.ID == symbolID && sym.Kind == "struct" {
			targetSymbol = &symbols[i]
			break
		}
	}

	if targetSymbol == nil {
		return nil, ErrNotFound("struct", symbolID)
	}

	// Get symbol references
	symbolRefs, err := s.codeAnalyzerRepo.GetSymbolReferences(symbolID)
	if err != nil {
		return nil, err
	}

	// Prepare the prompt
	prompt := s.promptBuilder.BuildStructPrompt(targetSymbol, symbolRefs)

	// Get the JSON schema for symbol insights
	schema := s.schemaBuilder.SymbolInsightJSONSchema()

	// Call the LLM with the schema
	response, err := s.client.Call(modelName, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Parse the response into a structured insight
	insight := &insights.StructInsight{}
	err = json.Unmarshal([]byte(response), insight)
	if err != nil {
		return nil, err
	}

	return insight, nil
}

// GenerateFileInsight generates insights for a file
func (s *Service) GenerateFileInsight(repoID int64, fileID int64, modelName string) (*insights.FileInsight, error) {
	// Get repository information
	repo, err := s.codeAnalyzerRepo.GetRepositoryByID(repoID)
	if err != nil {
		return nil, err
	}

	// Get the file details
	file, err := s.codeAnalyzerRepo.GetRepositoryFileByID(repoID, fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, ErrNotFound("file", fileID)
	}

	// Get functions in the file
	functions, err := s.codeAnalyzerRepo.GetRepositoryFunctions(repoID, fileID)
	if err != nil {
		return nil, err
	}

	// Get symbols in the file
	symbols, err := s.codeAnalyzerRepo.GetRepositorySymbols(repoID, fileID)
	if err != nil {
		return nil, err
	}

	// Prepare the prompt
	prompt := s.promptBuilder.BuildFilePrompt(repo, file, functions, symbols)

	// Get the JSON schema for symbol insights
	schema := s.schemaBuilder.SymbolInsightJSONSchema()

	// Call the LLM with the schema
	response, err := s.client.Call(modelName, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Parse the response into a file insight
	var insight insights.FileInsight
	err = json.Unmarshal([]byte(response), &insight)
	if err != nil {
		return nil, err
	}

	return &insight, nil
}

// GenerateRepositoryInsight generates insights for an entire repository
func (s *Service) GenerateRepositoryInsight(repoID int64, modelName string) (*insights.RepositoryInsight, error) {
	// Get repository information
	repo, err := s.codeAnalyzerRepo.GetRepositoryByID(repoID)
	if err != nil {
		return nil, err
	}

	// Get repository files
	files, err := s.codeAnalyzerRepo.GetRepositoryFiles(repoID)
	if err != nil {
		return nil, err
	}

	// Get repository functions
	functions, err := s.codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
	if err != nil {
		return nil, err
	}

	// Get repository symbols
	symbols, err := s.codeAnalyzerRepo.GetRepositorySymbols(repoID, 0)
	if err != nil {
		return nil, err
	}

	// Prepare the prompt
	prompt := s.promptBuilder.BuildRepositoryPrompt(repo, files, functions, symbols)

	// Get the JSON schema for symbol insights
	schema := s.schemaBuilder.SymbolInsightJSONSchema()

	// Call the LLM with the schema
	response, err := s.client.Call(modelName, prompt, schema)
	if err != nil {
		return nil, err
	}

	// Parse the response into a repository insight
	var insight insights.RepositoryInsight
	err = json.Unmarshal([]byte(response), &insight)
	if err != nil {
		return nil, err
	}

	return &insight, nil
}

// ErrNotFound returns a formatted error for when a resource is not found
func ErrNotFound(resourceType string, id int64) error {
	return fmt.Errorf("%s with ID %d not found", resourceType, id)
}
