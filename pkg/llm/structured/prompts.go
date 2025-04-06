package structured

import (
	"fmt"
	"strings"

	"cred.com/hack25/backend/internal/models"
)

// PromptBuilder handles building prompts for different insight types
type PromptBuilder struct {
	useJSONFormat bool
}

// NewPromptBuilder creates a new PromptBuilder
func NewPromptBuilder(useJSONFormat bool) *PromptBuilder {
	return &PromptBuilder{
		useJSONFormat: useJSONFormat,
	}
}

// BuildFunctionPrompt creates a prompt for function analysis
func (p *PromptBuilder) BuildFunctionPrompt(function *models.RepositoryFunction, calls []models.FunctionCall) string {
	var sb strings.Builder

	// System instruction
	sb.WriteString("You are an expert code analyst capable of understanding Go source code in depth.\n\n")

	// Description of what we want
	sb.WriteString("Please analyze the following Go function and extract key insights about its purpose, behavior, and implementation:\n\n")

	// Function details
	sb.WriteString("## Function Details\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", function.Name))

	// Function code
	sb.WriteString("## Source Code\n\n")
	sb.WriteString("```go\n")
	sb.WriteString(function.CodeBlock)
	sb.WriteString("\n```\n\n")

	// Function calls
	if len(calls) > 0 {
		sb.WriteString("## Function Calls\n\n")
		sb.WriteString("This function makes the following calls:\n\n")

		for _, call := range calls {
			sb.WriteString(fmt.Sprintf("- %s\n", call.CalleeName))
		}
		sb.WriteString("\n")
	}

	// Output format instruction
	if p.useJSONFormat {
		sb.WriteString("## Response Format\n\n")
		sb.WriteString("Please provide your analysis in JSON format conforming exactly to the following schema:\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(`{
  "intent": {
    "problem": "The real-world problem this function solves",
    "goal": "What this function aims to accomplish",
    "result": "The concrete outcome produced by this function"
  },
  "params": [
    {
      "name": "paramName",
      "type": "paramType",
      "purpose": "What this parameter is used for"
    }
  ],
  "returns": [
    {
      "type": "returnType",
      "purpose": "What this return value represents"
    }
  ],
  "network": [
    {
      "protocol": "http/grpc/etc",
      "endpoint": "URL or service being called",
      "purpose": "Why this network call is made"
    }
  ],
  "database": [
    {
      "engine": "postgres/mysql/etc",
      "action": "read/write/update",
      "purpose": "Why this database operation is performed"
    }
  ],
  "object_store": [
    {
      "provider": "s3/gcs/etc",
      "bucket": "bucketName",
      "action": "get/put/etc",
      "key_pattern": "pattern of keys used",
      "purpose": "Why this operation is performed"
    }
  ],
  "compute": [
    {
      "category": "type of computation",
      "description": "details about the computation"
    }
  ],
  "observability": [
    {
      "type": "log/metric/trace",
      "purpose": "why this observability data is emitted"
    }
  ],
  "quality": [
    {
      "category": "type of quality concern",
      "description": "details about the quality issue"
    }
  ],
  "frameworks": [
    {
      "name": "framework name",
      "purpose": "how/why the framework is used"
    }
  ],
  "patterns": [
    {
      "name": "pattern name",
      "description": "how the pattern is applied"
    }
  ],
  "related": ["list", "of", "related", "function", "names"],
  "notes": "Any additional insights or observations about the function"
}`)
		sb.WriteString("\n```\n\n")
		sb.WriteString("If you don't have information for a particular field, use an empty array [] or empty string \"\" as appropriate. Do not omit any fields from the schema. Your response should be valid JSON with no other text or explanations outside the JSON object.")
	}

	return sb.String()
}

// BuildSymbolPrompt creates a prompt for symbol analysis
func (p *PromptBuilder) BuildSymbolPrompt(symbol *models.RepositorySymbol, refs []models.SymbolReference) string {
	var sb strings.Builder

	// System instruction
	sb.WriteString("You are an expert code analyst capable of understanding Go source code in depth.\n\n")

	// Description of what we want
	sb.WriteString("Please analyze the following Go symbol and extract key insights about its purpose and usage:\n\n")

	// Symbol details
	sb.WriteString("## Symbol Details\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", symbol.Name))
	sb.WriteString(fmt.Sprintf("Kind: %s\n", symbol.Kind))
	sb.WriteString(fmt.Sprintf("Type: %s\n", symbol.Type))
	sb.WriteString(fmt.Sprintf("Location: Line %d\n\n", symbol.Line))

	// Symbol code/definition
	sb.WriteString("## Definition\n\n")
	sb.WriteString("```go\n")
	sb.WriteString(symbol.Value)
	sb.WriteString("\n```\n\n")

	// Symbol references
	if len(refs) > 0 {
		sb.WriteString("## References\n\n")
		sb.WriteString("This symbol is referenced in the following locations:\n\n")

		for _, ref := range refs {
			sb.WriteString(fmt.Sprintf("- %d (Line %d): %s\n", ref.FileID, ref.Line, ref.Context))
		}
		sb.WriteString("\n")
	}

	// Output format instruction
	if p.useJSONFormat {
		sb.WriteString("## Response Format\n\n")
		sb.WriteString("Please provide your analysis in JSON format conforming exactly to the following schema:\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(`{
  "concept": {
    "domain": "The domain this symbol belongs to (e.g., 'authentication', 'database', etc.)",
    "name": "Conceptual name for this symbol",
    "description": "What this symbol represents in the codebase",
    "ontology_uri": "Optional: A URI to an ontology entry"
  },
  "decision": {
    "problem": "What problem this symbol addresses",
    "rationale": "Why this symbol exists in its current form",
    "alternatives": "What alternatives might have been considered"
  },
  "used_by": ["list", "of", "functions", "or", "components", "using", "this", "symbol"],
  "patterns": [
    {
      "name": "pattern name",
      "description": "how the pattern applies to this symbol",
      "rationale": "why this pattern was chosen"
    }
  ],
  "quality": [
    {
      "category": "type of quality concern",
      "description": "details about the quality issue",
      "metric": "optional metric name",
      "value": 0.0,
      "status": "pass/warn/fail"
    }
  ]
}`)
		sb.WriteString("\n```\n\n")
		sb.WriteString("If you don't have information for a particular field, use an empty array [] or empty string \"\" as appropriate. Do not omit any fields from the schema. Your response should be valid JSON with no other text or explanations outside the JSON object.")
	}

	return sb.String()
}

// BuildStructPrompt creates a prompt for struct analysis
func (p *PromptBuilder) BuildStructPrompt(symbol *models.RepositorySymbol, refs []models.SymbolReference) string {
	var sb strings.Builder

	// System instruction
	sb.WriteString("You are an expert code analyst capable of understanding Go source code in depth.\n\n")

	// Description of what we want
	sb.WriteString("Please analyze the following Go struct and extract key insights about its purpose, design, and usage:\n\n")

	// Struct details
	sb.WriteString("## Struct Details\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", symbol.Name))
	sb.WriteString(fmt.Sprintf("Location: Line %d\n\n", symbol.Line))

	// Struct code/definition
	sb.WriteString("## Definition\n\n")
	sb.WriteString("```go\n")
	sb.WriteString(symbol.Value)
	sb.WriteString("\n```\n\n")

	// Struct references
	if len(refs) > 0 {
		sb.WriteString("## References\n\n")
		sb.WriteString("This struct is referenced in the following locations:\n\n")

		for _, ref := range refs {
			sb.WriteString(fmt.Sprintf("- %d (Line %d): %s\n", ref.FileID, ref.Line, ref.Context))
		}
		sb.WriteString("\n")
	}

	// Output format instruction
	if p.useJSONFormat {
		sb.WriteString("## Response Format\n\n")
		sb.WriteString("Please provide your analysis in JSON format conforming exactly to the following schema:\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(`{
  "concept": {
    "domain": "The domain this struct belongs to (e.g., 'user management', 'payment processing', etc.)",
    "name": "Conceptual name for this struct",
    "description": "What this struct represents in the codebase",
    "ontology_uri": "Optional: A URI to an ontology entry"
  },
  "fields": [
    {
      "name": "fieldName",
      "type": "fieldType",
      "purpose": "What this field is used for"
    }
  ],
  "relations": [
    {
      "pattern": "relationship pattern (e.g., 'has-a', 'is-a', etc.)",
      "description": "description of the relationship"
    }
  ],
  "persistence": {
    "engine": "database engine if applicable",
    "table": "table name if applicable",
    "strategy": "how this struct is persisted"
  },
  "observability": [
    {
      "type": "log/metric/trace",
      "purpose": "why this observability data is emitted"
    }
  ],
  "quality": [
    {
      "category": "type of quality concern",
      "description": "details about the quality issue",
      "metric": "optional metric name",
      "value": 0.0,
      "status": "pass/warn/fail"
    }
  ],
  "patterns": [
    {
      "name": "pattern name",
      "description": "how the pattern applies to this struct",
      "rationale": "why this pattern was chosen"
    }
  ]
}`)
		sb.WriteString("\n```\n\n")
		sb.WriteString("If you don't have information for a particular field, use an empty array [] or empty string \"\" as appropriate. Do not omit any fields from the schema. Your response should be valid JSON with no other text or explanations outside the JSON object.")
	}

	return sb.String()
}

// BuildFilePrompt creates a prompt for file analysis
func (p *PromptBuilder) BuildFilePrompt(repo *models.Repository, file *models.RepositoryFile, functions []models.RepositoryFunction, symbols []models.RepositorySymbol) string {
	var sb strings.Builder

	// System instruction
	sb.WriteString("You are an expert code analyst capable of understanding Go source code in depth.\n\n")

	// Description of what we want
	sb.WriteString("Please analyze the following Go source file and extract key insights about its purpose, design, and content:\n\n")

	// File details
	sb.WriteString("## File Details\n\n")
	sb.WriteString(fmt.Sprintf("Path: %s\n", file.FilePath))
	sb.WriteString(fmt.Sprintf("Repository: %s\n", repo.Name))

	// File content (limited to avoid exceeding token limits)
	sb.WriteString("## Source Code (limited preview)\n\n")
	code := ""
	if len(code) > 2000 {
		code = code[:2000] + "\n... [content truncated] ...\n"
	}
	sb.WriteString("```go\n")
	sb.WriteString(code)
	sb.WriteString("\n```\n\n")

	// Functions in file
	if len(functions) > 0 {
		sb.WriteString("## Functions\n\n")
		sb.WriteString("This file contains the following functions:\n\n")

		for _, fn := range functions {
			sb.WriteString(fmt.Sprintf("- %s (Line %d)\n", fn.Name, fn.Line))
			sb.WriteString(fmt.Sprintf("Code Block:\n\n%s\n\n", fn.CodeBlock))
		}
		sb.WriteString("\n")
	}

	// Symbols in file
	if len(symbols) > 0 {
		sb.WriteString("## Symbols\n\n")
		sb.WriteString("This file contains the following symbols:\n\n")

		for _, sym := range symbols {
			sb.WriteString(fmt.Sprintf("- %s (Line %d, Type: %s)\n", sym.Name, sym.Line, sym.Kind))
		}
		sb.WriteString("\n")
	}

	// Output format instruction
	if p.useJSONFormat {
		sb.WriteString("## Response Format\n\n")
		sb.WriteString("Please provide your analysis in JSON format conforming exactly to the following schema:\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(`{
  "responsibilities": {
    "main_purpose": "The primary purpose of this file",
    "details": "More detailed explanation of the file's responsibilities"
  },
  "contains": ["list", "of", "key", "components", "in", "the", "file"],
  "dependencies": [
    {
      "name": "dependency name",
      "type": "internal/external",
      "purpose": "why this dependency is needed"
    }
  ],
  "observability": [
    {
      "type": "log/metric/trace",
      "purpose": "why this observability data is emitted"
    }
  ],
  "quality": [
    {
      "category": "type of quality concern",
      "description": "details about the quality issue",
      "metric": "optional metric name",
      "value": 0.0,
      "status": "pass/warn/fail"
    }
  ],
  "patterns": [
    {
      "name": "pattern name",
      "description": "how the pattern is applied in this file",
      "rationale": "why this pattern was chosen"
    }
  ]
}`)
		sb.WriteString("\n```\n\n")
		sb.WriteString("If you don't have information for a particular field, use an empty array [] or empty string \"\" as appropriate. Do not omit any fields from the schema. Your response should be valid JSON with no other text or explanations outside the JSON object.")
	}

	return sb.String()
}

// BuildRepositoryPrompt creates a prompt for repository analysis
func (p *PromptBuilder) BuildRepositoryPrompt(repo *models.Repository, files []models.RepositoryFile, functions []models.RepositoryFunction, symbols []models.RepositorySymbol) string {
	var sb strings.Builder

	// System instruction
	sb.WriteString("You are an expert code analyst capable of understanding large codebases in depth.\n\n")

	// Description of what we want
	sb.WriteString("Please analyze the following Go repository and extract key insights about its architecture, design patterns, and technical characteristics:\n\n")

	// Repository details
	sb.WriteString("## Repository Details\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", repo.Name))
	sb.WriteString(fmt.Sprintf("Files: %d\n", len(files)))
	sb.WriteString(fmt.Sprintf("Functions: %d\n", len(functions)))
	sb.WriteString(fmt.Sprintf("Symbols: %d\n\n", len(symbols)))

	// Directory structure (top level)
	sb.WriteString("## Directory Structure\n\n")
	dirMap := make(map[string]int)
	for _, file := range files {
		parts := strings.Split(file.FilePath, "/")
		if len(parts) > 1 {
			topDir := parts[0]
			dirMap[topDir]++
		}
	}

	for dir, count := range dirMap {
		sb.WriteString(fmt.Sprintf("- %s/ (%d files)\n", dir, count))
	}
	sb.WriteString("\n")

	// Sample files (limited to avoid exceeding token limits)
	sb.WriteString("## Key Files\n\n")
	maxFiles := 10
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("- %s\n", file.FilePath))
	}
	sb.WriteString("\n")

	// Output format instruction
	if p.useJSONFormat {
		sb.WriteString("## Response Format\n\n")
		sb.WriteString("Please provide your analysis in JSON format conforming exactly to the following schema:\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(`{
  "domain": {
    "name": "The domain this repository serves",
    "description": "Description of the problem domain",
    "ontology_uri": "Optional: A URI to an ontology entry"
  },
  "architecture": {
    "pattern": "The architectural pattern used (e.g., 'hexagonal', 'MVC', etc.)",
    "description": "Description of the architecture",
    "strengths": "Strengths of the chosen architecture",
    "weaknesses": "Weaknesses or challenges with the architecture",
    "reason": "Why this architecture was chosen"
  },
  "frameworks": [
    {
      "name": "framework name",
      "version": "version if known",
      "purpose": "how the framework is used"
    }
  ],
  "design_patterns": [
    {
      "name": "pattern name",
      "description": "how the pattern is implemented",
      "location": "where the pattern is implemented",
      "reason": "why this pattern was chosen",
      "applies_to": ["list", "of", "components"]
    }
  ],
  "coding_patterns": [
    {
      "name": "pattern name",
      "description": "description of the coding pattern",
      "rationale": "why this pattern is used",
      "example": "brief example of the pattern"
    }
  ],
  "critical_paths": ["list", "of", "important", "execution", "paths"],
  "tech_debt": [
    {
      "category": "type of technical debt",
      "description": "details about the technical debt",
      "severity": "high/medium/low",
      "metric": "optional metric name",
      "value": 0.0,
      "status": "status assessment"
    }
  ]
}`)
		sb.WriteString("\n```\n\n")
		sb.WriteString("If you don't have information for a particular field, use an empty array [] or empty string \"\" as appropriate. Do not omit any fields from the schema. Your response should be valid JSON with no other text or explanations outside the JSON object.")
	}

	return sb.String()
}
