package structured

import (
	"encoding/json"
)

// SchemaBuilder provides methods to build JSON schemas for structured output formats
type SchemaBuilder struct{}

// NewSchemaBuilder creates a new SchemaBuilder instance
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{}
}

// FunctionInsightJSONSchema returns the JSON schema for function insights
func (b *SchemaBuilder) FunctionInsightJSONSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"intent": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"problem": map[string]interface{}{"type": "string"},
					"goal":    map[string]interface{}{"type": "string"},
					"result":  map[string]interface{}{"type": "string"},
				},
				"required": []string{"problem", "goal", "result"},
			},
			"params": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":    map[string]interface{}{"type": "string"},
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "type", "purpose"},
				},
			},
			"returns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"type", "purpose"},
				},
			},
			"network": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"protocol": map[string]interface{}{"type": "string"},
						"endpoint": map[string]interface{}{"type": "string"},
						"purpose":  map[string]interface{}{"type": "string"},
					},
					"required": []string{"protocol", "endpoint", "purpose"},
				},
			},
			"database": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"engine":  map[string]interface{}{"type": "string"},
						"action":  map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"engine", "action", "purpose"},
				},
			},
			"object_store": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"provider":    map[string]interface{}{"type": "string"},
						"bucket":      map[string]interface{}{"type": "string"},
						"action":      map[string]interface{}{"type": "string"},
						"key_pattern": map[string]interface{}{"type": "string"},
						"purpose":     map[string]interface{}{"type": "string"},
					},
					"required": []string{"provider", "bucket", "action", "purpose"},
				},
			},
			"compute": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
			"observability": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"type", "purpose"},
				},
			},
			"quality": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
			"frameworks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "purpose"},
				},
			},
			"patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "description"},
				},
			},
			"related": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"notes": map[string]interface{}{"type": "string"},
		},
		"required": []string{"intent", "params", "returns"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}

// SymbolInsightJSONSchema returns the JSON schema for symbol insights
func (b *SchemaBuilder) SymbolInsightJSONSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"concept": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":      map[string]interface{}{"type": "string"},
					"name":        map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"ontology_uri": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"domain", "name", "description"},
			},
			"decision": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"problem":      map[string]interface{}{"type": "string"},
					"rationale":    map[string]interface{}{"type": "string"},
					"alternatives": map[string]interface{}{"type": "string"},
				},
				"required": []string{"problem", "rationale", "alternatives"},
			},
			"used_by": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"rationale":   map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "description"},
				},
			},
			"quality": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"metric":      map[string]interface{}{"type": "string"},
						"value":       map[string]interface{}{"type": "number"},
						"status":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
		},
		"required": []string{"concept", "decision", "used_by"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}

// StructInsightJSONSchema returns the JSON schema for struct insights
func (b *SchemaBuilder) StructInsightJSONSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"concept": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":      map[string]interface{}{"type": "string"},
					"name":        map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"ontology_uri": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"domain", "name", "description"},
			},
			"fields": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":    map[string]interface{}{"type": "string"},
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "type", "purpose"},
				},
			},
			"relations": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pattern":     map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
					},
					"required": []string{"pattern", "description"},
				},
			},
			"persistence": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"engine":   map[string]interface{}{"type": "string"},
					"table":    map[string]interface{}{"type": "string"},
					"strategy": map[string]interface{}{"type": "string"},
				},
				"required": []string{"engine", "table", "strategy"},
			},
			"observability": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"type", "purpose"},
				},
			},
			"quality": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"metric":      map[string]interface{}{"type": "string"},
						"value":       map[string]interface{}{"type": "number"},
						"status":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
			"patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"rationale":   map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "description"},
				},
			},
		},
		"required": []string{"concept", "fields"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}

// FileInsightJSONSchema returns the JSON schema for file insights
func (b *SchemaBuilder) FileInsightJSONSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"responsibilities": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"main_purpose": map[string]interface{}{"type": "string"},
					"details":      map[string]interface{}{"type": "string"},
				},
				"required": []string{"main_purpose", "details"},
			},
			"contains": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"dependencies": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":    map[string]interface{}{"type": "string"},
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "type", "purpose"},
				},
			},
			"observability": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":    map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"type", "purpose"},
				},
			},
			"quality": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"metric":      map[string]interface{}{"type": "string"},
						"value":       map[string]interface{}{"type": "number"},
						"status":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
			"patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"rationale":   map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "description"},
				},
			},
		},
		"required": []string{"responsibilities", "contains", "dependencies"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}

// RepositoryInsightJSONSchema returns the JSON schema for repository insights
func (b *SchemaBuilder) RepositoryInsightJSONSchema() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"domain": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"ontology_uri": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"name", "description"},
			},
			"architecture": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern":    map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"strengths":   map[string]interface{}{"type": "string"},
					"weaknesses":  map[string]interface{}{"type": "string"},
					"reason":      map[string]interface{}{"type": "string"},
				},
				"required": []string{"pattern", "description", "strengths", "weaknesses"},
			},
			"frameworks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":    map[string]interface{}{"type": "string"},
						"version": map[string]interface{}{"type": "string"},
						"purpose": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "purpose"},
				},
			},
			"design_patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"location":    map[string]interface{}{"type": "string"},
						"reason":      map[string]interface{}{"type": "string"},
						"applies_to": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required": []string{"name", "description"},
				},
			},
			"coding_patterns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":        map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"rationale":   map[string]interface{}{"type": "string"},
						"example":     map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "description"},
				},
			},
			"critical_paths": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"tech_debt": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":    map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"severity":    map[string]interface{}{"type": "string"},
						"metric":      map[string]interface{}{"type": "string"},
						"value":       map[string]interface{}{"type": "number"},
						"status":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"category", "description"},
				},
			},
		},
		"required": []string{"domain", "architecture", "frameworks"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}
