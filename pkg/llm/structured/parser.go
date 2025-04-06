package structured

import (
	"github.com/sirupsen/logrus"
)

// Parser handles converting structured LLM responses into insight objects
type Parser struct {
	logger *logrus.Entry
}

// NewParser creates a new Parser instance
func NewParser(logger *logrus.Entry) *Parser {
	return &Parser{
		logger: logger,
	}
}

// // extractJSONFromResponse attempts to extract valid JSON from an LLM response
// func extractJSONFromResponse(response string) (string, error) {
// 	// Pattern for JSON blocks in markdown code blocks or plain
// 	// Pattern for JSON blocks in markdown code blocks or plain
// 	jsonPattern := regexp.MustCompile(`(?s)(?:```(?:json)?\s*([\s\S]*?)```|(\{[\s\S]*?\}))`)
// 	matches := jsonPattern.FindStringSubmatch(response)

// 	if len(matches) > 1 {
// 		for i := 1; i < len(matches); i++ {
// 			if matches[i] != "" {
// 				// Found a JSON object
// 				return matches[i], nil
// 			}
// 		}
// 	}

// 	// Try a different approach for unquoted JSON
// 	unquotedPattern := regexp.MustCompile(`(?s)\{.*?\}`)
// 	unquotedMatches := unquotedPattern.FindString(response)
// 	if unquotedMatches != "" {
// 		return unquotedMatches, nil
// 	}

// 	return "", fmt.Errorf("no valid JSON found in response")
// }

// // ParseFunctionInsight parses LLM response into a FunctionInsight
// func (p *Parser) ParseFunctionInsight(content string) (*insights.FunctionInsight, error) {
// 	// Try to extract JSON from the response
// 	jsonData, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		p.logger.WithError(err).Warn("Failed to extract JSON from response, falling back to legacy parser")
// 		// Return legacy parsing result if needed
// 		return ParseFunctionInsightLegacy(content)
// 	}

// 	// Parse the structured schema
// 	var schema FunctionInsightSchema
// 	if err := json.Unmarshal([]byte(jsonData), &schema); err != nil {
// 		p.logger.WithError(err).Warn("Failed to parse JSON schema, falling back to legacy parser")
// 		// Return legacy parsing result if needed
// 		return ParseFunctionInsightLegacy(content)
// 	}

// 	// Map the schema to the insights type
// 	insight := &insights.FunctionInsight{
// 		Intent: insights.Narrative{
// 			Problem: schema.Intent.Problem,
// 			Goal:    schema.Intent.Goal,
// 			Result:  schema.Intent.Result,
// 		},
// 	}

// 	// Map parameters
// 	for _, param := range schema.Params {
// 		insight.Params = append(insight.Params, insights.IOParam{
// 			Name:    param.Name,
// 			Type:    param.Type,
// 			Meaning: param.Purpose,
// 		})
// 	}

// 	// Map return values
// 	for _, ret := range schema.Returns {
// 		insight.Returns = append(insight.Returns, insights.IOParam{
// 			Type:    ret.Type,
// 			Meaning: ret.Purpose,
// 		})
// 	}

// 	// Map network calls
// 	for _, network := range schema.Network {
// 		insight.Network = append(insight.Network, insights.NetworkCall{
// 			Protocol: network.Protocol,
// 			Endpoint: network.Endpoint,
// 			Purpose:  network.Purpose,
// 		})
// 	}

// 	// Map database operations
// 	for _, db := range schema.Database {
// 		insight.Database = append(insight.Database, insights.DatabaseOp{
// 			Engine:  db.Engine,
// 			Action:  db.Action,
// 			Purpose: db.Purpose,
// 		})
// 	}

// 	// Map object store operations
// 	for _, obj := range schema.ObjectStore {
// 		insight.ObjectStore = append(insight.ObjectStore, insights.ObjectStoreOp{
// 			Provider:   obj.Provider,
// 			Bucket:     obj.Bucket,
// 			Action:     obj.Action,
// 			KeyPattern: obj.KeyPattern,
// 			Purpose:    obj.Purpose,
// 		})
// 	}

// 	// Map compute operations
// 	for _, compute := range schema.Compute {
// 		insight.Compute = append(insight.Compute, insights.ComputeTask{
// 			Service: compute.Category,
// 			Purpose: compute.Description,
// 		})
// 	}

// 	// Map observability hooks
// 	for _, obs := range schema.Observability {
// 		insight.Observability = append(insight.Observability, insights.ObservabilityHook{
// 			Type:   obs.Type,
// 			Detail: obs.Purpose,
// 		})
// 	}

// 	// Map quality metrics
// 	for _, quality := range schema.Quality {
// 		insight.Quality = append(insight.Quality, insights.QualityMetric{
// 			Metric: quality.Category,
// 			Status: "info", // Default status
// 			Value:  0,      // Default value
// 		})
// 	}

// 	// Map frameworks
// 	for _, framework := range schema.Frameworks {
// 		insight.Frameworks = append(insight.Frameworks, insights.FrameworkUsage{
// 			Name:    framework.Name,
// 			Purpose: framework.Purpose,
// 		})
// 	}

// 	// Map patterns
// 	for _, pattern := range schema.Patterns {
// 		insight.Patterns = append(insight.Patterns, insights.CodingPattern{
// 			Name:      pattern.Name,
// 			Rationale: pattern.Description,
// 		})
// 	}

// 	// Map related functions and notes
// 	insight.Related = schema.Related
// 	insight.Notes = schema.Notes

// 	return insight, nil
// }

// // ParseSymbolInsight parses LLM response into a SymbolInsight
// func (p *Parser) ParseSymbolInsight(content string) (*insights.SymbolInsight, error) {
// 	// Try to extract JSON from the response
// 	jsonData, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		p.logger.WithError(err).Warn("Failed to extract JSON from response, falling back to legacy parser")
// 		return ParseSymbolInsightLegacy(content)
// 	}

// 	// Parse the structured schema
// 	var schema SymbolInsightSchema
// 	if err := json.Unmarshal([]byte(jsonData), &schema); err != nil {
// 		p.logger.WithError(err).Warn("Failed to parse JSON schema, falling back to legacy parser")
// 		return ParseSymbolInsightLegacy(content)
// 	}

// 	// Map the schema to the insights type
// 	insight := &insights.SymbolInsight{
// 		Concept: insights.KnowledgeRef{
// 			Concept:     schema.Concept.Name,
// 			Description: schema.Concept.Description,
// 			OntologyURI: schema.Concept.OntologyURI,
// 		},
// 		Decision: insights.Narrative{
// 			Problem: schema.Decision.Problem,
// 			Goal:    schema.Decision.Rationale,
// 			Result:  schema.Decision.Alternatives,
// 		},
// 		UsedBy: schema.UsedBy,
// 	}

// 	// Map patterns
// 	for _, pattern := range schema.Patterns {
// 		insight.Patterns = append(insight.Patterns, insights.CodingPattern{
// 			Name:      pattern.Name,
// 			Rationale: pattern.Description,
// 			Example:   pattern.Rationale, // Use Rationale as Example
// 		})
// 	}

// 	// Map quality metrics
// 	for _, quality := range schema.Quality {
// 		insight.Quality = append(insight.Quality, insights.QualityMetric{
// 			Metric:    quality.Category,
// 			Value:     quality.Value,
// 			Status:    quality.Status,
// 			Threshold: 0, // Default threshold
// 		})
// 	}

// 	return insight, nil
// }

// // ParseStructInsight parses LLM response into a StructInsight
// func (p *Parser) ParseStructInsight(content string) (*insights.StructInsight, error) {
// 	// Try to extract JSON from the response
// 	jsonData, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		p.logger.WithError(err).Warn("Failed to extract JSON from response, falling back to legacy parser")
// 		return ParseStructInsightLegacy(content)
// 	}

// 	// Parse the structured schema
// 	var schema StructInsightSchema
// 	if err := json.Unmarshal([]byte(jsonData), &schema); err != nil {
// 		p.logger.WithError(err).Warn("Failed to parse JSON schema, falling back to legacy parser")
// 		return ParseStructInsightLegacy(content)
// 	}

// 	// Map the schema to the insights type
// 	insight := &insights.StructInsight{
// 		Concept: insights.KnowledgeRef{
// 			Concept:     schema.Concept.Name,
// 			Description: schema.Concept.Description,
// 			OntologyURI: schema.Concept.OntologyURI,
// 		},
// 		Persistence: insights.DatabaseOp{
// 			Engine:  schema.Persistence.Engine,
// 			Table:   schema.Persistence.Table,
// 			Purpose: schema.Persistence.Strategy,
// 		},
// 	}

// 	// Map fields
// 	for _, field := range schema.Fields {
// 		insight.Fields = append(insight.Fields, insights.IOParam{
// 			Name:    field.Name,
// 			Type:    field.Type,
// 			Meaning: field.Purpose,
// 		})
// 	}

// 	// Map relations
// 	for _, relation := range schema.Relations {
// 		insight.Relations = append(insight.Relations, insights.DesignPattern{
// 			Name:   relation.Pattern,
// 			Reason: relation.Description,
// 		})
// 	}

// 	// Map observability hooks
// 	for _, obs := range schema.Observability {
// 		insight.Observability = append(insight.Observability, insights.ObservabilityHook{
// 			Type:   obs.Type,
// 			Detail: obs.Purpose,
// 		})
// 	}

// 	// Map quality metrics
// 	for _, quality := range schema.Quality {
// 		insight.Quality = append(insight.Quality, insights.QualityMetric{
// 			Metric: quality.Category,
// 			Value:  quality.Value,
// 			Status: quality.Status,
// 		})
// 	}

// 	// Map patterns
// 	for _, pattern := range schema.Patterns {
// 		insight.Patterns = append(insight.Patterns, insights.CodingPattern{
// 			Name:      pattern.Name,
// 			Rationale: pattern.Description,
// 			Example:   pattern.Rationale, // Use Rationale as Example
// 		})
// 	}

// 	return insight, nil
// }

// // ParseFileInsight parses LLM response into a FileInsight
// func (p *Parser) ParseFileInsight(content string) (*insights.FileInsight, error) {
// 	// Try to extract JSON from the response
// 	jsonData, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		p.logger.WithError(err).Warn("Failed to extract JSON from response, falling back to legacy parser")
// 		return ParseFileInsightLegacy(content)
// 	}

// 	// Parse the structured schema
// 	var schema FileInsightSchema
// 	if err := json.Unmarshal([]byte(jsonData), &schema); err != nil {
// 		p.logger.WithError(err).Warn("Failed to parse JSON schema, falling back to legacy parser")
// 		return ParseFileInsightLegacy(content)
// 	}

// 	// Map the schema to the insights type
// 	insight := &insights.FileInsight{
// 		Responsibilities: insights.Narrative{
// 			Problem: "", // Not directly mapped
// 			Goal:    schema.Responsibilities.MainPurpose,
// 			Result:  schema.Responsibilities.Details,
// 		},
// 		Contains: schema.Contains,
// 	}

// 	// Map dependencies
// 	for _, dep := range schema.Dependencies {
// 		insight.Dependencies = append(insight.Dependencies, insights.FrameworkUsage{
// 			Name:    dep.Name,
// 			Purpose: dep.Purpose,
// 		})
// 	}

// 	// Map observability hooks
// 	for _, obs := range schema.Observability {
// 		insight.Observability = append(insight.Observability, insights.ObservabilityHook{
// 			Type:   obs.Type,
// 			Detail: obs.Purpose,
// 		})
// 	}

// 	// Map quality metrics
// 	for _, quality := range schema.Quality {
// 		insight.Quality = append(insight.Quality, insights.QualityMetric{
// 			Metric: quality.Category,
// 			Value:  quality.Value,
// 			Status: quality.Status,
// 		})
// 	}

// 	// Map patterns
// 	for _, pattern := range schema.Patterns {
// 		insight.Patterns = append(insight.Patterns, insights.CodingPattern{
// 			Name:      pattern.Name,
// 			Rationale: pattern.Description,
// 			Example:   pattern.Rationale, // Use Rationale as Example
// 		})
// 	}

// 	return insight, nil
// }

// // ParseRepositoryInsight parses LLM response into a RepositoryInsight
// func (p *Parser) ParseRepositoryInsight(content string) (*insights.RepositoryInsight, error) {
// 	// Try to extract JSON from the response
// 	jsonData, err := extractJSONFromResponse(content)
// 	if err != nil {
// 		p.logger.WithError(err).Warn("Failed to extract JSON from response, falling back to legacy parser")
// 		return ParseRepositoryInsightLegacy(content)
// 	}

// 	// Parse the structured schema
// 	var schema RepositoryInsightSchema
// 	if err := json.Unmarshal([]byte(jsonData), &schema); err != nil {
// 		p.logger.WithError(err).Warn("Failed to parse JSON schema, falling back to legacy parser")
// 		return ParseRepositoryInsightLegacy(content)
// 	}

// 	// Map the schema to the insights type
// 	insight := &insights.RepositoryInsight{
// 		Domain: insights.KnowledgeRef{
// 			Concept:     schema.Domain.Name,
// 			Description: schema.Domain.Description,
// 			OntologyURI: schema.Domain.OntologyURI,
// 		},
// 		Architecture: insights.ArchitecturePattern{
// 			Name:   schema.Architecture.Pattern,
// 			Reason: schema.Architecture.Reason,
// 		},
// 		CriticalPaths: schema.CriticalPaths,
// 	}

// 	// Map frameworks
// 	for _, framework := range schema.Frameworks {
// 		insight.Frameworks = append(insight.Frameworks, insights.FrameworkUsage{
// 			Name:    framework.Name,
// 			Version: framework.Version,
// 			Purpose: framework.Purpose,
// 		})
// 	}

// 	// Map design patterns
// 	for _, pattern := range schema.DesignPatterns {
// 		insight.DesignPatterns = append(insight.DesignPatterns, insights.DesignPattern{
// 			Name:      pattern.Name,
// 			Reason:    pattern.Reason,
// 			AppliesTo: pattern.AppliesTo,
// 		})
// 	}

// 	// Map coding patterns
// 	for _, pattern := range schema.CodingPatterns {
// 		insight.CodingPatterns = append(insight.CodingPatterns, insights.CodingPattern{
// 			Name:      pattern.Name,
// 			Rationale: pattern.Rationale,
// 			Example:   pattern.Example,
// 		})
// 	}

// 	// Map tech debt (quality metrics)
// 	for _, debt := range schema.TechDebt {
// 		insight.TechDebt = append(insight.TechDebt, insights.QualityMetric{
// 			Metric: debt.Category,
// 			Value:  debt.Value,
// 			Status: debt.Status,
// 		})
// 	}

// 	return insight, nil
// }

// // Legacy parsers for fallback

// // ParseFunctionInsightLegacy handles text-based parsing of function insights
// func ParseFunctionInsightLegacy(content string) (*insights.FunctionInsight, error) {
// 	lines := strings.Split(content, "\n")
// 	insight := &insights.FunctionInsight{}

// 	// Basic parsing logic - this is simplified
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Look for sections with common headers
// 		if strings.HasPrefix(line, "Intent:") || strings.HasPrefix(line, "# Intent") {
// 			// Parse intent section
// 			for j := i + 1; j < len(lines) && j < i + 10; j++ {
// 				subline := strings.TrimSpace(lines[j])
// 				if strings.HasPrefix(subline, "Problem:") {
// 					insight.Intent.Problem = strings.TrimPrefix(subline, "Problem:")
// 				}
// 				if strings.HasPrefix(subline, "Goal:") {
// 					insight.Intent.Goal = strings.TrimPrefix(subline, "Goal:")
// 				}
// 				if strings.HasPrefix(subline, "Result:") {
// 					insight.Intent.Result = strings.TrimPrefix(subline, "Result:")
// 				}
// 			}
// 		}

// 		// Additional section parsing would go here
// 	}

// 	return insight, nil
// }

// // ParseSymbolInsightLegacy handles text-based parsing of symbol insights
// func ParseSymbolInsightLegacy(content string) (*insights.SymbolInsight, error) {
// 	// Simplified legacy parser
// 	insight := &insights.SymbolInsight{
// 		Concept: insights.KnowledgeRef{},
// 		Decision: insights.Narrative{},
// 	}

// 	lines := strings.Split(content, "\n")
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Basic parsing logic
// 		if strings.HasPrefix(line, "Concept:") {
// 			for j := i + 1; j < len(lines) && j < i + 5; j++ {
// 				if strings.Contains(lines[j], "Domain:") {
// 					insight.Concept.Concept = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 				if strings.Contains(lines[j], "Description:") {
// 					insight.Concept.Description = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 			}
// 		}
// 	}

// 	return insight, nil
// }

// // ParseStructInsightLegacy handles text-based parsing of struct insights
// func ParseStructInsightLegacy(content string) (*insights.StructInsight, error) {
// 	// Simplified legacy parser
// 	insight := &insights.StructInsight{
// 		Concept: insights.KnowledgeRef{},
// 		Persistence: insights.DatabaseOp{},
// 	}

// 	lines := strings.Split(content, "\n")
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Basic parsing logic
// 		if strings.HasPrefix(line, "Concept:") {
// 			for j := i + 1; j < len(lines) && j < i + 5; j++ {
// 				if strings.Contains(lines[j], "Domain:") {
// 					insight.Concept.Concept = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 				if strings.Contains(lines[j], "Description:") {
// 					insight.Concept.Description = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 			}
// 		}
// 	}

// 	return insight, nil
// }

// // ParseFileInsightLegacy handles text-based parsing of file insights
// func ParseFileInsightLegacy(content string) (*insights.FileInsight, error) {
// 	// Simplified legacy parser
// 	insight := &insights.FileInsight{
// 		Responsibilities: insights.Narrative{},
// 	}

// 	lines := strings.Split(content, "\n")
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Basic parsing logic
// 		if strings.HasPrefix(line, "Responsibilities:") {
// 			for j := i + 1; j < len(lines) && j < i + 5; j++ {
// 				if strings.Contains(lines[j], "Main Purpose:") {
// 					insight.Responsibilities.Goal = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 				if strings.Contains(lines[j], "Details:") {
// 					insight.Responsibilities.Result = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 			}
// 		}
// 	}

// 	return insight, nil
// }

// // ParseRepositoryInsightLegacy handles text-based parsing of repository insights
// func ParseRepositoryInsightLegacy(content string) (*insights.RepositoryInsight, error) {
// 	// Simplified legacy parser
// 	insight := &insights.RepositoryInsight{
// 		Domain: insights.KnowledgeRef{},
// 		Architecture: insights.ArchitecturePattern{},
// 	}

// 	lines := strings.Split(content, "\n")
// 	for i, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// Basic parsing logic
// 		if strings.HasPrefix(line, "Domain:") {
// 			for j := i + 1; j < len(lines) && j < i + 5; j++ {
// 				if strings.Contains(lines[j], "Name:") {
// 					insight.Domain.Concept = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 				if strings.Contains(lines[j], "Description:") {
// 					insight.Domain.Description = strings.TrimSpace(strings.Split(lines[j], ":")[1])
// 				}
// 			}
// 		}
// 	}

// 	return insight, nil
// }
