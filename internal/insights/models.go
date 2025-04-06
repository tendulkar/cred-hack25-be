// Package insights – strongly‑typed knowledge about Go code.
//
// All structs are token‑efficient and LLM‑friendly: one‑line narrative
// + small value objects, no free‑text blobs.
package insights

import "time"

////////////////////////////////////////////////////////////////////////////////
// GENERIC STORY ELEMENTS
////////////////////////////////////////////////////////////////////////////////

// Narrative captures the “why” in three lines.
type Narrative struct {
	Problem string `json:"problem"` // real‑world pain
	Goal    string `json:"goal"`    // what “done” looks like
	Result  string `json:"result"`  // measurable outcome
}

// IOParam gives semantic meaning to a parameter or return value.
type IOParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`              // Go type
	Meaning  string `json:"meaning"`           // human description
	Example  string `json:"example,omitempty"` // JSON / literal
	Optional bool   `json:"optional,omitempty"`
}

// KnowledgeRef links a symbol or struct to a domain concept / ontology.
type KnowledgeRef struct {
	Concept     string `json:"concept"`                // “Invoice”, “User”
	OntologyURI string `json:"ontology_uri,omitempty"` // DBpedia, schema.org …
	Description string `json:"description,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// SOFTWARE‑OBJECT INTERACTIONS
////////////////////////////////////////////////////////////////////////////////

// NetworkCall – outward network interaction (HTTP, gRPC, WebSocket…).
type NetworkCall struct {
	Protocol string `json:"protocol"`         // http | grpc | ws
	Method   string `json:"method,omitempty"` // GET, POST…
	Endpoint string `json:"endpoint"`         // URL / host:port
	Purpose  string `json:"purpose"`          // why we call it
}

// DatabaseOp – SQL / NoSQL interaction.
type DatabaseOp struct {
	Engine  string `json:"engine"` // postgres | mysql | mongo
	Table   string `json:"table,omitempty"`
	Action  string `json:"action"`          // select | insert | update
	Query   string `json:"query,omitempty"` // optional, trimmed
	Purpose string `json:"purpose"`         // business reason
}

// ObjectStoreOp – S3 / GCS / MinIO …
type ObjectStoreOp struct {
	Provider   string `json:"provider"` // s3 | gcs | …
	Bucket     string `json:"bucket"`
	Action     string `json:"action"` // put | get | delete
	KeyPattern string `json:"key_pattern"`
	Purpose    string `json:"purpose"`
}

// ComputeTask – offloaded compute (λ, Cloud Run, k8s Job …).
type ComputeTask struct {
	Service string `json:"service"` // lambda | cloud_run | job
	Trigger string `json:"trigger"` // http | schedule | event
	Purpose string `json:"purpose"`
}

////////////////////////////////////////////////////////////////////////////////
// OBSERVABILITY & QUALITY
////////////////////////////////////////////////////////////////////////////////

// ObservabilityHook – metric, log line, or trace span emitted.
type ObservabilityHook struct {
	Type   string `json:"type"`   // metric | log | trace
	Name   string `json:"name"`   // counter name, log key…
	Detail string `json:"detail"` // tags, log level, span attrs
}

// QualityMetric – objective code quality signal.
type QualityMetric struct {
	Metric    string  `json:"metric"` // coverage | cyclomatic_complexity | lint_errors
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold,omitempty"` // desired
	Status    string  `json:"status"`              // pass | warn | fail
}

////////////////////////////////////////////////////////////////////////////////
// FRAMEWORKS, PATTERNS & ARCHITECTURE
////////////////////////////////////////////////////////////////////////////////

// FrameworkUsage – external lib / framework leveraged.
type FrameworkUsage struct {
	Name    string `json:"name"` // gin, gorm, cobra …
	Version string `json:"version,omitempty"`
	Purpose string `json:"purpose"` // routing, ORM…
}

// CodingPattern – idiomatic Go technique.
type CodingPattern struct {
	Name      string `json:"name"`              // context‑prop, error‑wrap…
	Rationale string `json:"rationale"`         // why we use it
	Example   string `json:"example,omitempty"` // short snippet
}

// DesignPattern – OO / DDD / enterprise pattern in play.
type DesignPattern struct {
	Name      string   `json:"name"` // repository, factory…
	Reason    string   `json:"reason"`
	AppliesTo []string `json:"applies_to"` // funcs/structs
}

// ArchitecturePattern – repo‑level style.
type ArchitecturePattern struct {
	Name   string `json:"name"`   // hexagonal, clean, microservice…
	Reason string `json:"reason"` // why chosen
}

////////////////////////////////////////////////////////////////////////////////
// REQUEST / RESPONSE ENVELOPE
////////////////////////////////////////////////////////////////////////////////

type InsightRequest struct {
	RepositoryID int64       `json:"repository_id"`
	Asset        CodeAssetID `json:"asset"`           // exactly one artefact
	Model        string      `json:"model,omitempty"` // defaults to system model
}

type CodeAssetID struct {
	FileID     int64  `json:"file_id,omitempty"`
	FunctionID int64  `json:"function_id,omitempty"`
	SymbolID   int64  `json:"symbol_id,omitempty"`
	Path       string `json:"path,omitempty"` // fallback
}

////////////////////////////////////////////////////////////////////////////////
// INSIGHT PAYLOADS
////////////////////////////////////////////////////////////////////////////////

// FunctionInsight – deepest granularity.
type FunctionInsight struct {
	Intent        Narrative           `json:"intent"`
	Params        []IOParam           `json:"params"`
	Returns       []IOParam           `json:"returns"`
	Network       []NetworkCall       `json:"network,omitempty"`
	Database      []DatabaseOp        `json:"database,omitempty"`
	ObjectStore   []ObjectStoreOp     `json:"object_store,omitempty"`
	Compute       []ComputeTask       `json:"compute,omitempty"`
	Observability []ObservabilityHook `json:"observability,omitempty"`
	Quality       []QualityMetric     `json:"quality,omitempty"`
	Frameworks    []FrameworkUsage    `json:"frameworks,omitempty"`
	Patterns      []CodingPattern     `json:"patterns,omitempty"`
	Related       []string            `json:"related,omitempty"` // func IDs
	Notes         string              `json:"notes,omitempty"`
}

// SymbolInsight – constant / var / alias semantics.
type SymbolInsight struct {
	Concept  KnowledgeRef    `json:"concept"`
	Decision Narrative       `json:"decision"` // why kept, not removed, etc.
	UsedBy   []string        `json:"used_by"`  // funcs/structs
	Patterns []CodingPattern `json:"patterns,omitempty"`
	Quality  []QualityMetric `json:"quality,omitempty"`
}

// StructInsight – data model view.
type StructInsight struct {
	Concept       KnowledgeRef        `json:"concept"`
	Fields        []IOParam           `json:"fields"`
	Relations     []DesignPattern     `json:"relations,omitempty"` // e.g. Aggregate Root
	Persistence   DatabaseOp          `json:"persistence,omitempty"`
	Observability []ObservabilityHook `json:"observability,omitempty"`
	Quality       []QualityMetric     `json:"quality,omitempty"`
	Patterns      []CodingPattern     `json:"patterns,omitempty"`
}

// FileInsight – map of a source file.
type FileInsight struct {
	Responsibilities Narrative           `json:"responsibilities"`
	Contains         []string            `json:"contains"`     // funcs, structs
	Dependencies     []FrameworkUsage    `json:"dependencies"` // imports
	Observability    []ObservabilityHook `json:"observability,omitempty"`
	Quality          []QualityMetric     `json:"quality,omitempty"`
	Patterns         []CodingPattern     `json:"patterns,omitempty"`
}

// RepositoryInsight – birds‑eye view.
type RepositoryInsight struct {
	Domain         KnowledgeRef        `json:"domain"`
	Architecture   ArchitecturePattern `json:"architecture"`
	Frameworks     []FrameworkUsage    `json:"frameworks,omitempty"`
	DesignPatterns []DesignPattern     `json:"design_patterns,omitempty"`
	CodingPatterns []CodingPattern     `json:"coding_patterns,omitempty"`
	CriticalPaths  []string            `json:"critical_paths,omitempty"` // call chains
	TechDebt       []QualityMetric     `json:"tech_debt,omitempty"`      // reuse QualityMetric
}

////////////////////////////////////////////////////////////////////////////////
// STORAGE WRAPPER
////////////////////////////////////////////////////////////////////////////////

type InsightType string

const (
	FunctionInsightType   InsightType = "function"
	SymbolInsightType     InsightType = "symbol"
	StructInsightType     InsightType = "struct"
	FileInsightType       InsightType = "file"
	RepositoryInsightType InsightType = "repository"
)

type InsightRecord struct {
	ID           int64       `json:"id" db:"id"`
	RepositoryID int64       `json:"repository_id" db:"repository_id"`
	FileID       *int64      `json:"file_id,omitempty" db:"file_id"`
	FunctionID   *int64      `json:"function_id,omitempty" db:"function_id"`
	SymbolID     *int64      `json:"symbol_id,omitempty" db:"symbol_id"`
	Path         string      `json:"path,omitempty" db:"path"`
	Type         InsightType `json:"type" db:"type"`
	Data         string      `json:"data" db:"data"` // raw JSON
	Model        string      `json:"model" db:"model"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

////////////////////////////////////////////////////////////////////////////////
// LLM WRAPPER
////////////////////////////////////////////////////////////////////////////////

type LLMRequest struct {
	Model     string       `json:"model"`
	Messages  []LLMMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

type LLMMessage struct {
	Role string `json:"role"` // system | user | assistant
	Text string `json:"text"`
}

type LLMResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}
