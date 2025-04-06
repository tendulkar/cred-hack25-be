package repointel

import (
	"cred.com/hack25/backend/internal/insights"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/pkg/llm/structured"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Service represents the repository intelligence service
type Service struct {
	codeAnalyzerRepo  *repository.CodeAnalyzerRepository
	liteLLMBaseURL    string
	apiKey            string
	defaultModel      string
	structuredService *structured.Service
}

// NewService creates a new repository intelligence service
func NewService(
	codeAnalyzerRepo *repository.CodeAnalyzerRepository,
	liteLLMBaseURL string,
	apiKey string,
	defaultModel string,
) *Service {
	// Create logger for service
	serviceLogger := logger.Log.WithField("component", "repointel-service")

	// Create structured service config
	structuredConfig := structured.ServiceConfig{
		LiteLLMBaseURL: liteLLMBaseURL,
		APIKey:         apiKey,
		DefaultModel:   defaultModel,
		UseJSONFormat:  true, // Use structured JSON output format
	}

	// Create structured service
	structuredService := structured.NewService(
		codeAnalyzerRepo,
		structuredConfig,
		serviceLogger,
	)

	return &Service{
		codeAnalyzerRepo:  codeAnalyzerRepo,
		liteLLMBaseURL:    liteLLMBaseURL,
		apiKey:            apiKey,
		defaultModel:      defaultModel,
		structuredService: structuredService,
	}
}

// log returns a logrus entry with the service context
func (s *Service) log() *logrus.Entry {
	return logger.Log.WithField("component", "repointel-service")
}

// GenerateFunctionInsight generates insights for a function
func (s *Service) GenerateFunctionInsight(repoID int64, functionID int64, modelName string) (*insights.FunctionInsight, error) {
	// Delegate to structured service
	return s.structuredService.GenerateFunctionInsight(repoID, functionID, modelName)
}

// GenerateSymbolInsight generates insights for a symbol
func (s *Service) GenerateSymbolInsight(repoID int64, symbolID int64, modelName string) (*insights.SymbolInsight, error) {
	// Delegate to structured service
	return s.structuredService.GenerateSymbolInsight(repoID, symbolID, modelName)
}

// GenerateStructInsight generates insights for a struct
func (s *Service) GenerateStructInsight(repoID int64, symbolID int64, modelName string) (*insights.StructInsight, error) {
	// Delegate to structured service
	return s.structuredService.GenerateStructInsight(repoID, symbolID, modelName)
}

// // prepareFunctionPrompt prepares the prompt for function analysis
// func (s *Service) prepareFunctionPrompt(function *models.RepositoryFunction, calls []models.FunctionCall) string {
// 	var callNames []string
// 	for _, call := range calls {
// 		callNames = append(callNames, call.CalleeName)
// 	}

// 	prompt := fmt.Sprintf(`Analyze this Go function and return your analysis as a JSON object:

// Function Details:
// - Name: %s
// - Receiver: %s
// - Params: %s
// - Results: %s
// - Code:
// %s
// - Function calls: %s

// Provide your analysis in the following JSON format:

// {
//   "intent": {
//     "problem": "The problem this function solves",
//     "goal": "The goal of this function",
//     "result": "What the function accomplishes"
//   },
//   "params": [
//     {
//       "name": "param1",
//       "type": "string",
//       "purpose": "what this parameter is used for"
//     }
//   ],
//   "returns": [
//     {
//       "type": "string",
//       "purpose": "what this return value represents"
//     }
//   ],
//   "network": [
//     {
//       "protocol": "http/grpc/ws",
//       "endpoint": "endpoint path or name",
//       "purpose": "why this call is made"
//     }
//   ],
//   "database": [
//     {
//       "engine": "postgres/mysql/mongo",
//       "action": "query/insert/update/delete",
//       "purpose": "why this database operation is performed"
//     }
//   ],
//   "object_store": [
//     {
//       "provider": "s3/gcs/azure",
//       "bucket": "bucket name",
//       "action": "get/put/delete",
//       "key_pattern": "pattern of keys accessed",
//       "purpose": "why this storage operation is performed"
//     }
//   ],
//   "compute": [
//     {
//       "category": "calculation/transformation/processing",
//       "description": "what computation is performed"
//     }
//   ],
//   "observability": [
//     {
//       "type": "logging/metrics/tracing",
//       "purpose": "what is being observed"
//     }
//   ],
//   "quality": [
//     {
//       "category": "error handling/validation/performance",
//       "description": "quality consideration description"
//     }
//   ],
//   "frameworks": [
//     {
//       "name": "framework name",
//       "purpose": "how the framework is used"
//     }
//   ],
//   "patterns": [
//     {
//       "name": "pattern name",
//       "description": "how the pattern is applied"
//     }
//   ],
//   "related": ["list", "of", "related", "function", "names"],
//   "notes": "Any additional notes or observations"
// }

// Focus on accuracy and ensure your JSON is properly formatted. Leave arrays empty if there's no relevant information.`,
// 		function.Name,
// 		function.Receiver,
// 		function.Parameters,
// 		function.Results,
// 		function.CodeBlock,
// 		strings.Join(callNames, ", "))

// 	return prompt
// }

// // prepareSymbolPrompt prepares the prompt for symbol analysis
// func (s *Service) prepareSymbolPrompt(symbol *models.RepositorySymbol, refs []models.SymbolReference) string {
// 	var refContexts []string
// 	for _, ref := range refs {
// 		refContexts = append(refContexts, fmt.Sprintf("%s", ref.Context))
// 	}

// 	prompt := fmt.Sprintf(`Analyze this Go symbol and return your analysis as a JSON object:

// Symbol Details:
// - Name: %s
// - Type: %s
// - References in code: %s

// Provide your analysis in the following JSON format:

// {
//   "concept": {
//     "domain": "The knowledge domain this symbol belongs to",
//     "name": "Formal name or concept this symbol represents",
//     "description": "Description of the concept"
//   },
//   "decision": {
//     "problem": "The problem this symbol helps solve",
//     "rationale": "Why this symbol exists in the codebase",
//     "alternatives": "Potential alternative approaches"
//   },
//   "used_by": ["list", "of", "functions", "or", "structs", "that", "use", "this", "symbol"],
//   "patterns": [
//     {
//       "name": "pattern name",
//       "description": "how the pattern is applied"
//     }
//   ],
//   "quality": [
//     {
//       "category": "naming/encapsulation/consistency",
//       "description": "quality consideration description"
//     }
//   ]
// }

// Focus on accuracy and ensure your JSON is properly formatted. Leave arrays empty if there's no relevant information.`,
// 		symbol.Name,
// 		symbol.Type,
// 		strings.Join(refContexts, ", "))

// 	return prompt
// }

// // prepareStructPrompt prepares the prompt for struct analysis
// func (s *Service) prepareStructPrompt(symbol *models.RepositorySymbol, refs []models.SymbolReference) string {
// 	var refContexts []string
// 	for _, ref := range refs {
// 		refContexts = append(refContexts, fmt.Sprintf("%s", ref.Context))
// 	}

// 	prompt := fmt.Sprintf(`Analyze this Go struct and return your analysis as a JSON object:

// Struct Details:
// - Name: %s
// - Fields: %s
// - References in code: %s

// Provide your analysis in the following JSON format:

// {
//   "concept": {
//     "domain": "The knowledge domain this struct belongs to",
//     "name": "Formal name or concept this struct represents",
//     "description": "Description of the concept"
//   },
//   "fields": [
//     {
//       "name": "fieldName",
//       "type": "fieldType",
//       "purpose": "what this field represents"
//     }
//   ],
//   "relations": [
//     {
//       "pattern": "relation pattern (e.g., Aggregate Root, Value Object)",
//       "description": "how this struct relates to others"
//     }
//   ],
//   "persistence": {
//     "engine": "storage engine if applicable",
//     "table": "database table if applicable",
//     "strategy": "persistence strategy"
//   },
//   "observability": [
//     {
//       "type": "logging/metrics/tracing",
//       "purpose": "what is being observed"
//     }
//   ],
//   "quality": [
//     {
//       "category": "encapsulation/validation/integrity",
//       "description": "quality consideration description"
//     }
//   ],
//   "patterns": [
//     {
//       "name": "pattern name",
//       "description": "how the pattern is applied"
//     }
//   ]
// }

// Focus on accuracy and ensure your JSON is properly formatted. Leave objects empty or with null values if there's no relevant information.`,
// 		symbol.Name,
// 		symbol.Fields,
// 		strings.Join(refContexts, ", "))

// 	return prompt
// }

// // callLiteLLM makes a call to LiteLLM
// func (s *Service) callLiteLLM(model string, prompt string) (string, error) {
// 	s.log().WithFields(logrus.Fields{
// 		"model":         model,
// 		"prompt_length": len(prompt),
// 	}).Info("Calling LiteLLM")

// 	// Prepare the request
// 	reqBody := insights.LLMRequest{
// 		Model: model,
// 		Messages: []insights.LLMMessage{
// 			{
// 				Role: "user",
// 				Text: prompt,
// 			},
// 		},
// 		MaxTokens: 8192,
// 	}

// 	jsonReq, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request: %w", err)
// 	}

// 	// Create the HTTP request
// 	req, err := http.NewRequest("POST", s.liteLLMBaseURL+"/chat/completions", bytes.NewBuffer(jsonReq))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create request: %w", err)
// 	}

// 	// Set headers
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+s.apiKey)

// 	// Make the request
// 	resp, err := s.httpClient.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to make request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response: %w", err)
// 	}

// 	// Check for errors
// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("LiteLLM returned error: %s", string(body))
// 	}

// 	// Parse the response
// 	var llmResp struct {
// 		Choices []struct {
// 			Message struct {
// 				Content string `json:"content"`
// 			} `json:"message"`
// 		} `json:"choices"`
// 	}

// 	if err := json.Unmarshal(body, &llmResp); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	if len(llmResp.Choices) == 0 {
// 		return "", fmt.Errorf("no choices in response")
// 	}

// 	return llmResp.Choices[0].Message.Content, nil
// }

// // parseFunctionInsight parses the LLM response into a FunctionInsight using the new model structure
// func (s *Service) parseFunctionInsight(content string) (*insights.FunctionInsight, error) {
// 	// Extract the JSON part from the response
// 	jsonStr, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		// If we can't extract JSON, log it and fall back to the old parsing method
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"content_length": len(content),
// 		}).Warn("Failed to extract JSON from LLM response, falling back to text parsing")
// 		return s.parseFunctionInsightLegacy(content)
// 	}

// 	// Try to parse the JSON directly into the FunctionInsight structure
// 	insight := &insights.FunctionInsight{}
// 	if err := json.Unmarshal([]byte(jsonStr), insight); err != nil {
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"json_length": len(jsonStr),
// 		}).Warn("Failed to unmarshal JSON into FunctionInsight, falling back to text parsing")
// 		return s.parseFunctionInsightLegacy(content)
// 	}

// 	// Ensure we have at least the minimum required fields
// 	if insight.Intent.Problem == "" && insight.Intent.Goal == "" {
// 		s.log().Warn("Parsed FunctionInsight is missing required intent fields, falling back to text parsing")
// 		return s.parseFunctionInsightLegacy(content)
// 	}

// 	return insight, nil
// }

// // extractJSONFromResponse attempts to extract a valid JSON object from an LLM response
// func extractJSONFromResponse(response string) (string, error) {
// 	// First, look for JSON block markers in markdown
// 	jsonPattern := "```(?:json)?\\s*(.+?)\\s*```"
// 	re := regexp.MustCompile(jsonPattern)
// 	matches := re.FindStringSubmatch(response)
// 	if len(matches) > 1 {
// 		// Try to validate the extracted JSON
// 		jsonStr := matches[1]
// 		if json.Valid([]byte(jsonStr)) {
// 			return jsonStr, nil
// 		}
// 	}

// 	// If the above failed, try to find a JSON object directly
// 	// by finding the first { and the last }
// 	firstBrace := strings.Index(response, "{")
// 	lastBrace := strings.LastIndex(response, "}")

// 	if firstBrace != -1 && lastBrace != -1 && firstBrace < lastBrace {
// 		jsonStr := response[firstBrace : lastBrace+1]
// 		if json.Valid([]byte(jsonStr)) {
// 			return jsonStr, nil
// 		}
// 	}

// 	return "", fmt.Errorf("no valid JSON found in response")
// }

// // parseFunctionInsightLegacy is the old text-based parsing method kept for fallback
// func (s *Service) parseFunctionInsightLegacy(content string) (*insights.FunctionInsight, error) {
// 	lines := strings.Split(content, "\n")

// 	// Initialize the new function insight with the modern structure
// 	insight := &insights.FunctionInsight{
// 		Intent:  insights.Narrative{},
// 		Params:  []insights.IOParam{},
// 		Returns: []insights.IOParam{},
// 	}

// 	// Parse the function's purpose/goal
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Trying to extract the core purpose/intent
// 		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "Why the function is implemented:") {
// 			if i+1 < len(lines) {
// 				purpose := strings.TrimSpace(lines[i+1])

// 				// Fill the narrative structure with extracted info
// 				insight.Intent.Problem = "Need to " + purpose // Basic transformation
// 				insight.Intent.Goal = "Successfully " + purpose
// 				insight.Intent.Result = "Function completes " + purpose
// 			}
// 		}
// 	}

// 	// Extract information about systems it interacts with
// 	var networkCalls []insights.NetworkCall
// 	var databaseOps []insights.DatabaseOp
// 	var objectStoreOps []insights.ObjectStoreOp

// 	// Look for database operations
// 	for _, line := range lines {
// 		line = strings.ToLower(strings.TrimSpace(line))

// 		// Check for database operations
// 		if strings.Contains(line, "database") || strings.Contains(line, "sql") ||
// 			strings.Contains(line, "query") || strings.Contains(line, "postgres") ||
// 			strings.Contains(line, "mysql") || strings.Contains(line, "mongo") {

// 			// Determine database type
// 			engine := "postgres" // Default assumption
// 			if strings.Contains(line, "mysql") {
// 				engine = "mysql"
// 			} else if strings.Contains(line, "mongo") {
// 				engine = "mongo"
// 			}

// 			// Determine action type
// 			action := "query" // Default
// 			if strings.Contains(line, "insert") || strings.Contains(line, "add") || strings.Contains(line, "create") {
// 				action = "insert"
// 			} else if strings.Contains(line, "update") || strings.Contains(line, "modify") {
// 				action = "update"
// 			} else if strings.Contains(line, "delete") || strings.Contains(line, "remove") {
// 				action = "delete"
// 			} else if strings.Contains(line, "select") || strings.Contains(line, "get") ||
// 				strings.Contains(line, "fetch") || strings.Contains(line, "retrieve") {
// 				action = "select"
// 			}

// 			// Add to database operations
// 			databaseOps = append(databaseOps, insights.DatabaseOp{
// 				Engine:  engine,
// 				Action:  action,
// 				Purpose: line, // Using the whole line as purpose for now
// 			})
// 		}

// 		// Check for API/network calls
// 		if strings.Contains(line, "api") || strings.Contains(line, "http") ||
// 			strings.Contains(line, "request") || strings.Contains(line, "endpoint") ||
// 			strings.Contains(line, "service") {

// 			// Determine protocol
// 			protocol := "http" // Default assumption
// 			if strings.Contains(line, "grpc") {
// 				protocol = "grpc"
// 			} else if strings.Contains(line, "websocket") || strings.Contains(line, "ws") {
// 				protocol = "ws"
// 			}

// 			// Add to network calls
// 			networkCalls = append(networkCalls, insights.NetworkCall{
// 				Protocol: protocol,
// 				Endpoint: "external-service", // Placeholder
// 				Purpose:  line,               // Using the line as purpose for now
// 			})
// 		}

// 		// Check for object storage operations
// 		if strings.Contains(line, "storage") || strings.Contains(line, "s3") ||
// 			strings.Contains(line, "object") || strings.Contains(line, "file") ||
// 			strings.Contains(line, "blob") {

// 			// Determine provider
// 			provider := "s3" // Default assumption
// 			if strings.Contains(line, "gcs") || strings.Contains(line, "google") {
// 				provider = "gcs"
// 			}

// 			// Determine action
// 			action := "get" // Default
// 			if strings.Contains(line, "upload") || strings.Contains(line, "put") ||
// 				strings.Contains(line, "save") || strings.Contains(line, "write") {
// 				action = "put"
// 			} else if strings.Contains(line, "delete") || strings.Contains(line, "remove") {
// 				action = "delete"
// 			}

// 			// Add to storage operations
// 			objectStoreOps = append(objectStoreOps, insights.ObjectStoreOp{
// 				Provider:   provider,
// 				Bucket:     "data-bucket", // Placeholder
// 				Action:     action,
// 				KeyPattern: "*",  // Placeholder
// 				Purpose:    line, // Using the line as purpose for now
// 			})
// 		}
// 	}

// 	// Add any found operations to the insight
// 	if len(databaseOps) > 0 {
// 		insight.Database = databaseOps
// 	}

// 	if len(networkCalls) > 0 {
// 		insight.Network = networkCalls
// 	}

// 	if len(objectStoreOps) > 0 {
// 		insight.ObjectStore = objectStoreOps
// 	}

// 	// Add any additional information as notes
// 	for _, line := range lines {
// 		if strings.Contains(strings.ToLower(line), "additional") ||
// 			strings.Contains(strings.ToLower(line), "note") {
// 			insight.Notes = line
// 			break
// 		}
// 	}

// 	return insight, nil
// }

// // parseSymbolInsight parses the LLM response into a SymbolInsight
// func (s *Service) parseSymbolInsight(content string) (*insights.SymbolInsight, error) {
// 	// Extract the JSON part from the response
// 	jsonStr, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		// If we can't extract JSON, log it and fall back to the old parsing method
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"content_length": len(content),
// 		}).Warn("Failed to extract JSON from LLM response, falling back to text parsing")
// 		return s.parseSymbolInsightLegacy(content)
// 	}

// 	// Try to parse the JSON directly into the SymbolInsight structure
// 	insight := &insights.SymbolInsight{}
// 	if err := json.Unmarshal([]byte(jsonStr), insight); err != nil {
// 		// If direct unmarshaling fails, try more flexible approach with a map
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"json_length": len(jsonStr),
// 		}).Warn("Failed to unmarshal JSON directly, trying map approach")

// 		// Create a map to hold the raw JSON data
// 		var rawData map[string]interface{}
// 		if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
// 			s.log().WithFields(logrus.Fields{
// 				"error": err.Error(),
// 				"json_length": len(jsonStr),
// 			}).Warn("Failed to unmarshal JSON into map, falling back to text parsing")
// 			return s.parseSymbolInsightLegacy(content)
// 		}

// 		// Extract concept info
// 		if concept, ok := rawData["concept"].(map[string]interface{}); ok {
// 			knowledgeRef := insights.KnowledgeRef{}
// 			if name, ok := concept["name"].(string); ok {
// 				knowledgeRef.Concept = name
// 			}
// 			if description, ok := concept["description"].(string); ok {
// 				knowledgeRef.Description = description
// 			}
// 			if uri, ok := concept["ontology_uri"].(string); ok {
// 				knowledgeRef.OntologyURI = uri
// 			}
// 			insight.Concept = knowledgeRef
// 		}

// 		// Extract decision/rationale
// 		if decision, ok := rawData["decision"].(map[string]interface{}); ok {
// 			narrative := insights.Narrative{}
// 			if problem, ok := decision["problem"].(string); ok {
// 				narrative.Problem = problem
// 			}
// 			if rationale, ok := decision["rationale"].(string); ok {
// 				narrative.Goal = rationale
// 			}
// 			if alternatives, ok := decision["alternatives"].(string); ok {
// 				narrative.Result = alternatives
// 			}
// 			insight.Decision = narrative
// 		}

// 		// Extract used by
// 		if usedBy, ok := rawData["used_by"].([]interface{}); ok {
// 			for _, item := range usedBy {
// 				if str, ok := item.(string); ok {
// 					insight.UsedBy = append(insight.UsedBy, str)
// 				}
// 			}
// 		}

// 		// Extract patterns
// 		if patterns, ok := rawData["patterns"].([]interface{}); ok {
// 			for _, p := range patterns {
// 				if pMap, ok := p.(map[string]interface{}); ok {
// 					pattern := insights.CodingPattern{}
// 					if name, ok := pMap["name"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if rationale, ok := pMap["rationale"].(string); ok {
// 						pattern.Rationale = rationale
// 					} else if desc, ok := pMap["description"].(string); ok {
// 						// Fallback for description field
// 						pattern.Rationale = desc
// 					}
// 					insight.Patterns = append(insight.Patterns, pattern)
// 				}
// 			}
// 		}

// 		// Extract quality metrics
// 		if quality, ok := rawData["quality"].([]interface{}); ok {
// 			for _, q := range quality {
// 				if qMap, ok := q.(map[string]interface{}); ok {
// 					metric := insights.QualityMetric{}
// 					if category, ok := qMap["category"].(string); ok {
// 						metric.Metric = category
// 					} else if metricName, ok := qMap["metric"].(string); ok {
// 						metric.Metric = metricName
// 					}

// 					// Default values
// 					metric.Value = 1.0
// 					metric.Status = "pass"

// 					// Try to extract value if available
// 					if val, ok := qMap["value"].(float64); ok {
// 						metric.Value = val
// 					}

// 					insight.Quality = append(insight.Quality, metric)
// 				}
// 			}
// 		}
// 	}

// 	// Ensure we have at least basic content
// 	if (insight.Concept.Concept == "" && insight.Concept.Description == "") &&
// 	   (insight.Decision.Problem == "" && insight.Decision.Goal == "") &&
// 	   len(insight.UsedBy) == 0 {
// 		s.log().Warn("Parsed SymbolInsight is missing required fields, falling back to text parsing")
// 		return s.parseSymbolInsightLegacy(content)
// 	}

// 	return insight, nil
// }

// // parseSymbolInsightLegacy is the old text-based parsing method kept for fallback
// func (s *Service) parseSymbolInsightLegacy(content string) (*insights.SymbolInsight, error) {
// 	// Simple parsing logic
// 	lines := strings.Split(content, "\n")
// 	insight := &insights.SymbolInsight{}
// 	knowledgeRef := insights.KnowledgeRef{}
// 	narrative := insights.Narrative{}

// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "1.") || strings.Contains(line, "DBpedia") {
// 			if i+1 < len(lines) {
// 				// Extract concept information
// 				text := strings.TrimSpace(lines[i+1])
// 				knowledgeRef.Concept = text
// 				knowledgeRef.Description = "Based on text analysis"
// 				insight.Concept = knowledgeRef
// 			}
// 		} else if strings.HasPrefix(line, "2.") || strings.Contains(line, "used for") {
// 			if i+1 < len(lines) {
// 				// Extract usage as part of the decision narrative
// 				narrative.Goal = strings.TrimSpace(lines[i+1])

// 				// Also add to UsedBy if it mentions specific components
// 				usage := strings.TrimSpace(lines[i+1])
// 				parts := strings.Split(usage, " by ")
// 				if len(parts) > 1 {
// 					users := strings.Split(parts[1], ", ")
// 					insight.UsedBy = append(insight.UsedBy, users...)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "3.") || strings.Contains(line, "why") {
// 			if i+1 < len(lines) {
// 				// Extract rationale as part of the decision narrative
// 				narrative.Problem = strings.TrimSpace(lines[i+1])
// 			}
// 		} else if strings.HasPrefix(line, "4.") || strings.Contains(line, "additional") {
// 			if i+1 < len(lines) {
// 				// Extract additional info as part of the decision narrative
// 				narrative.Result = strings.TrimSpace(lines[i+1])

// 				// Also look for patterns in the additional info
// 				text := strings.ToLower(strings.TrimSpace(lines[i+1]))
// 				if strings.Contains(text, "pattern") || strings.Contains(text, "practice") {
// 					pattern := insights.CodingPattern{
// 						Name: "identified-pattern",
// 						Rationale: text,
// 					}
// 					insight.Patterns = append(insight.Patterns, pattern)
// 				}
// 			}
// 		}
// 	}

// 	// Assign the narrative we built
// 	insight.Decision = narrative

// 	return insight, nil
// }

// // parseStructInsight parses the LLM response into a StructInsight
// func (s *Service) parseStructInsight(content string) (*insights.StructInsight, error) {
// 	// Extract the JSON part from the response
// 	jsonStr, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		// If we can't extract JSON, log it and fall back to the old parsing method
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"content_length": len(content),
// 		}).Warn("Failed to extract JSON from LLM response, falling back to text parsing")
// 		return s.parseStructInsightLegacy(content)
// 	}

// 	// Try to parse the JSON directly into the StructInsight structure
// 	insight := &insights.StructInsight{}
// 	if err := json.Unmarshal([]byte(jsonStr), insight); err != nil {
// 		// If direct unmarshaling fails, try more flexible approach with a map
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"json_length": len(jsonStr),
// 		}).Warn("Failed to unmarshal JSON directly, trying map approach")

// 		// Create a map to hold the raw JSON data
// 		var rawData map[string]interface{}
// 		if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
// 			s.log().WithFields(logrus.Fields{
// 				"error": err.Error(),
// 				"json_length": len(jsonStr),
// 			}).Warn("Failed to unmarshal JSON into map, falling back to text parsing")
// 			return s.parseStructInsightLegacy(content)
// 		}

// 		// Extract concept info
// 		if concept, ok := rawData["concept"].(map[string]interface{}); ok {
// 			knowledgeRef := insights.KnowledgeRef{}
// 			if name, ok := concept["name"].(string); ok {
// 				knowledgeRef.Concept = name
// 			}
// 			if description, ok := concept["description"].(string); ok {
// 				knowledgeRef.Description = description
// 			}
// 			if uri, ok := concept["ontology_uri"].(string); ok {
// 				knowledgeRef.OntologyURI = uri
// 			}
// 			insight.Concept = knowledgeRef
// 		}

// 		// Extract fields
// 		if fields, ok := rawData["fields"].([]interface{}); ok {
// 			for _, f := range fields {
// 				if fMap, ok := f.(map[string]interface{}); ok {
// 					param := insights.IOParam{}
// 					if name, ok := fMap["name"].(string); ok {
// 						param.Name = name
// 					}
// 					if typ, ok := fMap["type"].(string); ok {
// 						param.Type = typ
// 					}
// 					if purpose, ok := fMap["purpose"].(string); ok {
// 						param.Meaning = purpose
// 					}
// 					insight.Fields = append(insight.Fields, param)
// 				}
// 			}
// 		}

// 		// Extract relations
// 		if relations, ok := rawData["relations"].([]interface{}); ok {
// 			for _, r := range relations {
// 				if rMap, ok := r.(map[string]interface{}); ok {
// 					pattern := insights.DesignPattern{}
// 					if name, ok := rMap["pattern"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if desc, ok := rMap["description"].(string); ok {
// 						pattern.Reason = desc
// 					}
// 					insight.Relations = append(insight.Relations, pattern)
// 				}
// 			}
// 		}

// 		// Extract persistence
// 		if persistence, ok := rawData["persistence"].(map[string]interface{}); ok {
// 			dbOp := insights.DatabaseOp{}
// 			if engine, ok := persistence["engine"].(string); ok {
// 				dbOp.Engine = engine
// 			}
// 			if table, ok := persistence["table"].(string); ok {
// 				dbOp.Table = table
// 			}
// 			if strategy, ok := persistence["strategy"].(string); ok {
// 				dbOp.Purpose = strategy
// 			}
// 			insight.Persistence = dbOp
// 		}

// 		// Extract observability
// 		if observability, ok := rawData["observability"].([]interface{}); ok {
// 			for _, o := range observability {
// 				if oMap, ok := o.(map[string]interface{}); ok {
// 					hook := insights.ObservabilityHook{}
// 					if typ, ok := oMap["type"].(string); ok {
// 						hook.Type = typ
// 					}
// 					if purpose, ok := oMap["purpose"].(string); ok {
// 						hook.Detail = purpose
// 					}
// 					insight.Observability = append(insight.Observability, hook)
// 				}
// 			}
// 		}

// 		// Extract quality metrics
// 		if quality, ok := rawData["quality"].([]interface{}); ok {
// 			for _, q := range quality {
// 				if qMap, ok := q.(map[string]interface{}); ok {
// 					metric := insights.QualityMetric{}
// 					if category, ok := qMap["category"].(string); ok {
// 						metric.Metric = category
// 					} else if metricName, ok := qMap["metric"].(string); ok {
// 						metric.Metric = metricName
// 					}

// 					// Default values
// 					metric.Value = 1.0
// 					metric.Status = "pass"

// 					insight.Quality = append(insight.Quality, metric)
// 				}
// 			}
// 		}

// 		// Extract patterns
// 		if patterns, ok := rawData["patterns"].([]interface{}); ok {
// 			for _, p := range patterns {
// 				if pMap, ok := p.(map[string]interface{}); ok {
// 					pattern := insights.CodingPattern{}
// 					if name, ok := pMap["name"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if rationale, ok := pMap["rationale"].(string); ok {
// 						pattern.Rationale = rationale
// 					} else if desc, ok := pMap["description"].(string); ok {
// 						pattern.Rationale = desc
// 					}
// 					insight.Patterns = append(insight.Patterns, pattern)
// 				}
// 			}
// 		}
// 	}

// 	// Ensure we have at least basic content
// 	if insight.Concept.Concept == "" && len(insight.Fields) == 0 {
// 		s.log().Warn("Parsed StructInsight is missing required fields, falling back to text parsing")
// 		return s.parseStructInsightLegacy(content)
// 	}

// 	return insight, nil
// }

// // parseStructInsightLegacy is the old text-based parsing method kept for fallback
// func (s *Service) parseStructInsightLegacy(content string) (*insights.StructInsight, error) {
// 	// Simple parsing logic
// 	lines := strings.Split(content, "\n")
// 	insight := &insights.StructInsight{}
// 	knowledgeRef := insights.KnowledgeRef{}

// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "1.") || strings.Contains(line, "DBpedia") {
// 			if i+1 < len(lines) {
// 				knowledgeRef.Concept = strings.TrimSpace(lines[i+1])
// 				knowledgeRef.Description = "Based on text analysis"
// 				insight.Concept = knowledgeRef
// 			}
// 		} else if strings.HasPrefix(line, "2.") || strings.Contains(line, "used for") {
// 			if i+1 < len(lines) {
// 				// Add as a field with meaning
// 				param := insights.IOParam{
// 					Name:    "usage",
// 					Type:    "string",
// 					Meaning: strings.TrimSpace(lines[i+1]),
// 				}
// 				insight.Fields = append(insight.Fields, param)
// 			}
// 		} else if strings.HasPrefix(line, "3.") || strings.Contains(line, "why") {
// 			if i+1 < len(lines) {
// 				// Add as a relationship pattern
// 				pattern := insights.DesignPattern{
// 					Name:   "purpose",
// 					Reason: strings.TrimSpace(lines[i+1]),
// 				}
// 				insight.Relations = append(insight.Relations, pattern)
// 			}
// 		} else if strings.HasPrefix(line, "4.") || strings.Contains(line, "related") {
// 			if i+1 < len(lines) {
// 				relatedText := strings.TrimSpace(lines[i+1])
// 				relatedStructs := strings.Split(relatedText, ", ")

// 				// Create a relationship pattern for each related struct
// 				for _, rel := range relatedStructs {
// 					pattern := insights.DesignPattern{
// 						Name:   "related",
// 						Reason: "Related to " + rel,
// 						AppliesTo: []string{rel},
// 					}
// 					insight.Relations = append(insight.Relations, pattern)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "5.") || strings.Contains(line, "data model") {
// 			if i+1 < len(lines) {
// 				// Add as persistence
// 				insight.Persistence = insights.DatabaseOp{
// 					Engine:  "database",
// 					Purpose: strings.TrimSpace(lines[i+1]),
// 				}
// 			}
// 		}
// 	}

// 	return insight, nil
// }

// // GenerateFileInsight generates insights for a file
// func (s *Service) GenerateFileInsight(repoID int64, fileID int64, modelName string) (*insights.FileInsight, error) {
// 	// Delegate to structured service
// 	return s.structuredService.GenerateFileInsight(repoID, fileID, modelName)
// }

// // GenerateRepositoryInsight generates insights for an entire repository
// func (s *Service) GenerateRepositoryInsight(repoID int64, modelName string) (*insights.RepositoryInsight, error) {
// 	// Delegate to structured service
// 	return s.structuredService.GenerateRepositoryInsight(repoID, modelName)
// }

// // prepareFilePrompt prepares the prompt for file analysis
// func (s *Service) prepareFilePrompt(repo *models.Repository, file *models.RepositoryFile, functions []models.RepositoryFunction, symbols []models.RepositorySymbol) string {
// 	// Create a list of function names
// 	var functionNames []string
// 	for _, fn := range functions {
// 		functionNames = append(functionNames, fn.Name)
// 	}

// 	// Create a list of symbol names
// 	var symbolNames []string
// 	for _, sym := range symbols {
// 		symbolNames = append(symbolNames, fmt.Sprintf("%s (%s)", sym.Name, sym.Kind))
// 	}

// 	prompt := fmt.Sprintf(`Analyze this Go file and return your analysis as a JSON object:

// File Details:
// - Path: %s
// - Package: %s
// - Repository: %s (%s)
// - Functions (%d): %s
// - Symbols (%d): %s

// Provide your analysis in the following JSON format:

// {
//   "responsibilities": {
//     "main_purpose": "The primary purpose of this file",
//     "details": "More detailed explanation of responsibilities"
//   },
//   "contains": [
//     "list", "of", "important", "components", "in", "this", "file"
//   ],
//   "dependencies": [
//     {
//       "name": "dependency name",
//       "type": "import/package/internal",
//       "purpose": "why this dependency is used"
//     }
//   ],
//   "observability": [
//     {
//       "type": "logging/metrics/tracing",
//       "purpose": "what is being observed"
//     }
//   ],
//   "quality": [
//     {
//       "category": "organization/cohesion/maintainability",
//       "description": "quality consideration description"
//     }
//   ],
//   "patterns": [
//     {
//       "name": "pattern name",
//       "description": "how the pattern is applied"
//     }
//   ]
// }

// Focus on accuracy and ensure your JSON is properly formatted. Leave arrays empty if there's no relevant information.`,
// 		file.FilePath,
// 		file.Package,
// 		repo.Name,
// 		repo.URL,
// 		len(functionNames),
// 		strings.Join(functionNames, ", "),
// 		len(symbolNames),
// 		strings.Join(symbolNames, ", "))

// 	return prompt
// }

// // prepareRepositoryPrompt prepares the prompt for repository analysis
// func (s *Service) prepareRepositoryPrompt(repo *models.Repository, files []models.RepositoryFile, functions []models.RepositoryFunction, symbols []models.RepositorySymbol) string {
// 	// Group files by package
// 	packages := make(map[string][]string)
// 	for _, file := range files {
// 		packages[file.Package] = append(packages[file.Package], file.FilePath)
// 	}

// 	// Create summary of packages
// 	var packageSummary []string
// 	for pkg, files := range packages {
// 		packageSummary = append(packageSummary, fmt.Sprintf("%s (%d files)", pkg, len(files)))
// 	}

// 	// Get a sample of important functions (limit to 20)
// 	var functionSample []string
// 	maxFunctions := 20
// 	if len(functions) > maxFunctions {
// 		functions = functions[:maxFunctions]
// 	}
// 	for _, fn := range functions {
// 		functionSample = append(functionSample, fn.Name)
// 	}

// 	// Get a sample of important symbols (limit to 20, prioritize structs)
// 	var structSample []string
// 	var otherSymbolSample []string
// 	for _, sym := range symbols {
// 		if strings.Contains(strings.ToLower(sym.Kind), "struct") {
// 			structSample = append(structSample, fmt.Sprintf("%s", sym.Name))
// 		} else {
// 			otherSymbolSample = append(otherSymbolSample, fmt.Sprintf("%s (%s)", sym.Name, sym.Kind))
// 		}
// 		// Limit samples
// 		if len(structSample) >= 10 {
// 			break
// 		}
// 		if len(otherSymbolSample) >= 10 {
// 			break
// 		}
// 	}

// 	prompt := fmt.Sprintf(`Analyze this Go repository and return your analysis as a JSON object:

// Repository Details:
// - Name: %s (%s)
// - Owner: %s
// - Total files: %d
// - Packages (%d): %s
// - Important structs: %s
// - Other symbols: %s
// - Sample functions: %s

// Provide your analysis in the following JSON format:

// {
//   "domain": {
//     "name": "Domain name or area this repository belongs to",
//     "description": "Description of the domain"
//   },
//   "architecture": {
//     "pattern": "Main architectural pattern (e.g., MVC, CQRS, Clean)",
//     "description": "Description of the architecture",
//     "strengths": "Strengths of the chosen architecture",
//     "weaknesses": "Weaknesses or potential issues with the architecture"
//   },
//   "frameworks": [
//     {
//       "name": "Framework name",
//       "purpose": "How the framework is used"
//     }
//   ],
//   "design_patterns": [
//     {
//       "name": "Pattern name",
//       "description": "How the pattern is applied",
//       "location": "Where this pattern is used"
//     }
//   ],
//   "coding_patterns": [
//     {
//       "name": "Pattern name",
//       "description": "How the pattern is applied"
//     }
//   ],
//   "critical_paths": [
//     "list", "of", "critical", "execution", "paths"
//   ],
//   "tech_debt": [
//     {
//       "category": "architecture/code/testing/documentation",
//       "description": "Description of the tech debt",
//       "severity": "low/medium/high"
//     }
//   ]
// }

// Focus on accuracy and ensure your JSON is properly formatted. Leave arrays empty if there's no relevant information.`,
// 		repo.Name,
// 		repo.URL,
// 		repo.Owner,
// 		len(files),
// 		len(packages),
// 		strings.Join(packageSummary, ", "),
// 		strings.Join(structSample, ", "),
// 		strings.Join(otherSymbolSample, ", "),
// 		strings.Join(functionSample, ", "))

// 	return prompt
// }

// // parseFileInsight parses the LLM response into a FileInsight
// func (s *Service) parseFileInsight(content string) (*insights.FileInsight, error) {
// 	// Extract the JSON part from the response
// 	jsonStr, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		// If we can't extract JSON, log it and fall back to the old parsing method
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"content_length": len(content),
// 		}).Warn("Failed to extract JSON from LLM response, falling back to text parsing")
// 		return s.parseFileInsightLegacy(content)
// 	}

// 	// Try to parse the JSON directly into the FileInsight structure
// 	insight := &insights.FileInsight{}
// 	if err := json.Unmarshal([]byte(jsonStr), insight); err != nil {
// 		// If direct unmarshaling fails, try more flexible approach with a map
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"json_length": len(jsonStr),
// 		}).Warn("Failed to unmarshal JSON directly, trying map approach")

// 		// Create a map to hold the raw JSON data
// 		var rawData map[string]interface{}
// 		if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
// 			s.log().WithFields(logrus.Fields{
// 				"error": err.Error(),
// 				"json_length": len(jsonStr),
// 			}).Warn("Failed to unmarshal JSON into map, falling back to text parsing")
// 			return s.parseFileInsightLegacy(content)
// 		}

// 		// Extract responsibilities
// 		if resp, ok := rawData["responsibilities"].(map[string]interface{}); ok {
// 			narrative := insights.Narrative{}
// 			if mainPurpose, ok := resp["main_purpose"].(string); ok {
// 				narrative.Problem = mainPurpose
// 			}
// 			if details, ok := resp["details"].(string); ok {
// 				narrative.Goal = details
// 				narrative.Result = "File implements this functionality"
// 			}
// 			insight.Responsibilities = narrative
// 		}

// 		// Extract contains - list of components
// 		if contains, ok := rawData["contains"].([]interface{}); ok {
// 			for _, item := range contains {
// 				if str, ok := item.(string); ok {
// 					insight.Contains = append(insight.Contains, str)
// 				}
// 			}
// 		}

// 		// Extract dependencies
// 		if dependencies, ok := rawData["dependencies"].([]interface{}); ok {
// 			for _, d := range dependencies {
// 				if dMap, ok := d.(map[string]interface{}); ok {
// 					framework := insights.FrameworkUsage{}
// 					if name, ok := dMap["name"].(string); ok {
// 						framework.Name = name
// 					}
// 					if purpose, ok := dMap["purpose"].(string); ok {
// 						framework.Purpose = purpose
// 					}
// 					insight.Dependencies = append(insight.Dependencies, framework)
// 				}
// 			}
// 		}

// 		// Extract observability
// 		if observability, ok := rawData["observability"].([]interface{}); ok {
// 			for _, o := range observability {
// 				if oMap, ok := o.(map[string]interface{}); ok {
// 					hook := insights.ObservabilityHook{}
// 					if typ, ok := oMap["type"].(string); ok {
// 						hook.Type = typ
// 					}
// 					if purpose, ok := oMap["purpose"].(string); ok {
// 						hook.Detail = purpose
// 					}
// 					insight.Observability = append(insight.Observability, hook)
// 				}
// 			}
// 		}

// 		// Extract quality metrics
// 		if quality, ok := rawData["quality"].([]interface{}); ok {
// 			for _, q := range quality {
// 				if qMap, ok := q.(map[string]interface{}); ok {
// 					metric := insights.QualityMetric{}
// 					if category, ok := qMap["category"].(string); ok {
// 						metric.Metric = category
// 					} else if metricName, ok := qMap["metric"].(string); ok {
// 						metric.Metric = metricName
// 					}

// 					// Default values
// 					metric.Value = 1.0
// 					metric.Status = "pass"

// 					insight.Quality = append(insight.Quality, metric)
// 				}
// 			}
// 		}

// 		// Extract patterns
// 		if patterns, ok := rawData["patterns"].([]interface{}); ok {
// 			for _, p := range patterns {
// 				if pMap, ok := p.(map[string]interface{}); ok {
// 					pattern := insights.CodingPattern{}
// 					if name, ok := pMap["name"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if rationale, ok := pMap["rationale"].(string); ok {
// 						pattern.Rationale = rationale
// 					} else if desc, ok := pMap["description"].(string); ok {
// 						pattern.Rationale = desc
// 					}
// 					insight.Patterns = append(insight.Patterns, pattern)
// 				}
// 			}
// 		}
// 	}

// 	// Ensure we have at least basic content
// 	if (insight.Responsibilities.Problem == "" && insight.Responsibilities.Goal == "") &&
// 	   len(insight.Contains) == 0 {
// 		s.log().Warn("Parsed FileInsight is missing required fields, falling back to text parsing")
// 		return s.parseFileInsightLegacy(content)
// 	}

// 	return insight, nil
// }

// // parseFileInsightLegacy is the old text-based parsing method kept for fallback
// func (s *Service) parseFileInsightLegacy(content string) (*insights.FileInsight, error) {
// 	// Simple parsing logic
// 	lines := strings.Split(content, "\n")
// 	insight := &insights.FileInsight{}
// 	narrative := insights.Narrative{}

// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "1.") || strings.Contains(line, "purpose") {
// 			if i+1 < len(lines) {
// 				// Primary purpose goes into the narrative
// 				narrative.Problem = strings.TrimSpace(lines[i+1])
// 				narrative.Goal = "Implement the described functionality"
// 				narrative.Result = "File fulfills its purpose"
// 			}
// 		} else if strings.HasPrefix(line, "2.") || strings.Contains(line, "components") || strings.Contains(line, "responsibilities") {
// 			if i+1 < len(lines) {
// 				// Try to extract component list
// 				componentLines := extractListItems(lines, i+1)
// 				for _, item := range componentLines {
// 					insight.Contains = append(insight.Contains, item)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "3.") || strings.Contains(line, "dependencies") {
// 			if i+1 < len(lines) {
// 				// Try to extract dependencies as a list
// 				dependencyLines := extractListItems(lines, i+1)
// 				for _, item := range dependencyLines {
// 					// Convert to framework usage
// 					dep := insights.FrameworkUsage{
// 						Name:    item,
// 						Purpose: "Used by this file",
// 					}
// 					insight.Dependencies = append(insight.Dependencies, dep)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "4.") || strings.Contains(line, "data flow") {
// 			if i+1 < len(lines) {
// 				// Data flow information gets added to the narrative result
// 				narrative.Result = strings.TrimSpace(lines[i+1])
// 			}
// 		}
// 	}

// 	// Assign the narrative we built
// 	insight.Responsibilities = narrative

// 	return insight, nil
// }

// // parseRepositoryInsight parses the LLM response into a RepositoryInsight
// func (s *Service) parseRepositoryInsight(content string) (*insights.RepositoryInsight, error) {
// 	// Extract the JSON part from the response
// 	jsonStr, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		// If we can't extract JSON, log it and fall back to the old parsing method
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"content_length": len(content),
// 		}).Warn("Failed to extract JSON from LLM response, falling back to text parsing")
// 		return s.parseRepositoryInsightLegacy(content)
// 	}

// 	// Try to parse the JSON directly into the RepositoryInsight structure
// 	insight := &insights.RepositoryInsight{}
// 	if err := json.Unmarshal([]byte(jsonStr), insight); err != nil {
// 		// If direct unmarshaling fails, try more flexible approach with a map
// 		s.log().WithFields(logrus.Fields{
// 			"error": err.Error(),
// 			"json_length": len(jsonStr),
// 		}).Warn("Failed to unmarshal JSON directly, trying map approach")

// 		// Create a map to hold the raw JSON data
// 		var rawData map[string]interface{}
// 		if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
// 			s.log().WithFields(logrus.Fields{
// 				"error": err.Error(),
// 				"json_length": len(jsonStr),
// 			}).Warn("Failed to unmarshal JSON into map, falling back to text parsing")
// 			return s.parseRepositoryInsightLegacy(content)
// 		}

// 		// Extract domain info
// 		if domain, ok := rawData["domain"].(map[string]interface{}); ok {
// 			knowledgeRef := insights.KnowledgeRef{}
// 			if name, ok := domain["name"].(string); ok {
// 				knowledgeRef.Concept = name
// 			}
// 			if description, ok := domain["description"].(string); ok {
// 				knowledgeRef.Description = description
// 			}
// 			if uri, ok := domain["ontology_uri"].(string); ok {
// 				knowledgeRef.OntologyURI = uri
// 			}
// 			insight.Domain = knowledgeRef
// 		}

// 		// Extract architecture
// 		if arch, ok := rawData["architecture"].(map[string]interface{}); ok {
// 			archPattern := insights.ArchitecturePattern{}
// 			if name, ok := arch["name"].(string); ok {
// 				archPattern.Name = name
// 			}
// 			if reason, ok := arch["reason"].(string); ok {
// 				archPattern.Reason = reason
// 			}
// 			insight.Architecture = archPattern
// 		}

// 		// Extract frameworks
// 		if frameworks, ok := rawData["frameworks"].([]interface{}); ok {
// 			for _, f := range frameworks {
// 				if fMap, ok := f.(map[string]interface{}); ok {
// 					framework := insights.FrameworkUsage{}
// 					if name, ok := fMap["name"].(string); ok {
// 						framework.Name = name
// 					}
// 					if purpose, ok := fMap["purpose"].(string); ok {
// 						framework.Purpose = purpose
// 					}
// 					if version, ok := fMap["version"].(string); ok {
// 						framework.Version = version
// 					}
// 					insight.Frameworks = append(insight.Frameworks, framework)
// 				}
// 			}
// 		}

// 		// Extract design patterns
// 		if patterns, ok := rawData["patterns"].([]interface{}); ok {
// 			for _, p := range patterns {
// 				if pMap, ok := p.(map[string]interface{}); ok {
// 					pattern := insights.DesignPattern{}
// 					if name, ok := pMap["name"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if reason, ok := pMap["reason"].(string); ok {
// 						pattern.Reason = reason
// 					}
// 					if applies, ok := pMap["applies_to"].([]interface{}); ok {
// 						for _, a := range applies {
// 							if str, ok := a.(string); ok {
// 								pattern.AppliesTo = append(pattern.AppliesTo, str)
// 							}
// 						}
// 					}
// 					insight.DesignPatterns = append(insight.DesignPatterns, pattern)
// 				}
// 			}
// 		}

// 		// Extract critical paths if present (interfaces would be converted to critical paths)
// 		if interfaces, ok := rawData["interfaces"].([]interface{}); ok {
// 			for _, i := range interfaces {
// 				if iMap, ok := i.(map[string]interface{}); ok {
// 					if name, ok := iMap["name"].(string); ok {
// 						// Add interface name as a critical path
// 						insight.CriticalPaths = append(insight.CriticalPaths, name)
// 					}
// 				} else if str, ok := i.(string); ok {
// 					insight.CriticalPaths = append(insight.CriticalPaths, str)
// 				}
// 			}
// 		}

// 		// Extract quality metrics as tech debt
// 		if quality, ok := rawData["quality"].([]interface{}); ok {
// 			for _, q := range quality {
// 				if qMap, ok := q.(map[string]interface{}); ok {
// 					metric := insights.QualityMetric{}
// 					if metricName, ok := qMap["metric"].(string); ok {
// 						metric.Metric = metricName
// 					}

// 					// Default values
// 					metric.Value = 1.0
// 					metric.Status = "pass"

// 					// Try to extract value if available
// 					if val, ok := qMap["value"].(float64); ok {
// 						metric.Value = val
// 					}

// 					insight.TechDebt = append(insight.TechDebt, metric)
// 				}
// 			}
// 		}

// 		// Extract key components and add them to appropriate places
// 		if components, ok := rawData["key_components"].([]interface{}); ok {
// 			for _, c := range components {
// 				if str, ok := c.(string); ok {
// 					// Add as a design pattern with a special name to indicate it's a key component
// 					pattern := insights.DesignPattern{
// 						Name:   "key-component",
// 						Reason: str,
// 						AppliesTo: []string{str},
// 					}
// 					insight.DesignPatterns = append(insight.DesignPatterns, pattern)
// 				}
// 			}
// 		}

// 		// Extract coding patterns
// 		if codingPatterns, ok := rawData["coding_patterns"].([]interface{}); ok {
// 			for _, p := range codingPatterns {
// 				if pMap, ok := p.(map[string]interface{}); ok {
// 					pattern := insights.CodingPattern{}
// 					if name, ok := pMap["name"].(string); ok {
// 						pattern.Name = name
// 					}
// 					if rationale, ok := pMap["rationale"].(string); ok {
// 						pattern.Rationale = rationale
// 					} else if desc, ok := pMap["description"].(string); ok {
// 						pattern.Rationale = desc
// 					}
// 					// Also try to extract example if available
// 					if example, ok := pMap["example"].(string); ok {
// 						pattern.Example = example
// 					}
// 					insight.CodingPatterns = append(insight.CodingPatterns, pattern)
// 				}
// 			}
// 		}

// 		// Extract recommendations as part of technical debt
// 		if recommendations, ok := rawData["recommendations"].(string); ok && recommendations != "" {
// 			metric := insights.QualityMetric{
// 				Metric:  "recommendations",
// 				Value:   1.0,
// 				Status:  "info",  // Use info status to indicate this is just informational
// 			}
// 			insight.TechDebt = append(insight.TechDebt, metric)
// 		}
// 	}

// 	// Ensure we have at least basic content
// 	if insight.Domain.Concept == "" && insight.Architecture.Name == "" {
// 		s.log().Warn("Parsed RepositoryInsight is missing required fields, falling back to text parsing")
// 		return s.parseRepositoryInsightLegacy(content)
// 	}

// 	return insight, nil
// }

// // parseRepositoryInsightLegacy is the old text-based parsing method kept for fallback
// func (s *Service) parseRepositoryInsightLegacy(content string) (*insights.RepositoryInsight, error) {
// 	// Simple parsing logic
// 	lines := strings.Split(content, "\n")
// 	insight := &insights.RepositoryInsight{}

// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "1.") || strings.Contains(line, "purpose") || strings.Contains(line, "domain") {
// 			if i+1 < len(lines) {
// 				// Extract domain/purpose as a KnowledgeRef
// 				knowledgeRef := insights.KnowledgeRef{
// 					Concept:     strings.TrimSpace(lines[i+1]),
// 					Description: "Based on text analysis",
// 				}
// 				insight.Domain = knowledgeRef
// 			}
// 		} else if strings.HasPrefix(line, "2.") || strings.Contains(line, "architecture") {
// 			if i+1 < len(lines) {
// 				// Extract architecture pattern
// 				archPattern := insights.ArchitecturePattern{
// 					Name:   strings.TrimSpace(lines[i+1]),
// 					Reason: "Based on repository analysis",
// 				}
// 				insight.Architecture = archPattern
// 			}
// 		} else if strings.HasPrefix(line, "3.") || strings.Contains(line, "key components") {
// 			if i+1 < len(lines) {
// 				// Extract key components and convert them to design patterns
// 				componentLines := extractListItems(lines, i+1)
// 				for _, item := range componentLines {
// 					pattern := insights.DesignPattern{
// 						Name:   "key-component",
// 						Reason: item,
// 						AppliesTo: []string{item},
// 					}
// 					insight.DesignPatterns = append(insight.DesignPatterns, pattern)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "4.") || strings.Contains(line, "data flow") {
// 			if i+1 < len(lines) {
// 				// Data flow gets added as a special framework usage
// 				dataFlow := insights.FrameworkUsage{
// 					Name:    "data-flow",
// 					Purpose: strings.TrimSpace(lines[i+1]),
// 				}
// 				insight.Frameworks = append(insight.Frameworks, dataFlow)
// 			}
// 		} else if strings.HasPrefix(line, "5.") || strings.Contains(line, "dependencies") {
// 			if i+1 < len(lines) {
// 				// Extract dependencies as framework usages
// 				dependencyLines := extractListItems(lines, i+1)
// 				for _, item := range dependencyLines {
// 					// Parse out version if present
// 					parts := strings.Split(item, " (")
// 					name := parts[0]
// 					version := ""
// 					if len(parts) > 1 {
// 						version = strings.TrimSuffix(parts[1], ")")
// 					}

// 					framework := insights.FrameworkUsage{
// 						Name:    name,
// 						Version: version,
// 						Purpose: "Used in repository",
// 					}
// 					insight.Frameworks = append(insight.Frameworks, framework)
// 				}
// 			}
// 		} else if strings.HasPrefix(line, "6.") || strings.Contains(line, "recommendations") {
// 			if i+1 < len(lines) {
// 				// Add recommendations as tech debt metric
// 				metric := insights.QualityMetric{
// 					Metric:  "recommendations",
// 					Value:   1.0,
// 					Status:  "info",
// 				}
// 				insight.TechDebt = append(insight.TechDebt, metric)
// 			}
// 		}
// 	}

// 	return insight, nil
// }

// // extractListItems extracts a list of bullet points from text
// func extractListItems(lines []string, startIndex int) []string {
// 	var items []string

// 	// Max number of lines to check for list items
// 	maxLines := 10
// 	for i := startIndex; i < len(lines) && i < startIndex+maxLines; i++ {
// 		line := strings.TrimSpace(lines[i])

// 		// Check if this is a bullet point item
// 		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") || strings.HasPrefix(line, "*") {
// 			item := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "-"), "•"), "*"))
// 			items = append(items, item)
// 		} else if len(line) > 0 && len(items) == 0 {
// 			// If no bullet points found but there's text, use the first line
// 			items = append(items, line)
// 			break
// 		} else if len(line) == 0 && len(items) > 0 {
// 			// Empty line after we found some items - probably end of list
// 			break
// 		}
// 	}

// 	return items
// }
