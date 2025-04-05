package models

import (
	"time"
)

// FunctionCall represents a call from one function to another
type FunctionCall struct {
	ID            int64     `json:"id" db:"id"`
	CallerID      int64     `json:"caller_id" db:"caller_id"`
	CalleeName    string    `json:"callee_name" db:"callee_name"`
	CalleePackage string    `json:"callee_package" db:"callee_package"`
	CalleeID      *int64    `json:"callee_id,omitempty" db:"callee_id"`
	Line          int       `json:"line" db:"line"`
	Parameters    string    `json:"parameters" db:"parameters"` // JSON string
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// FunctionReference represents a reference to a function in the code
type FunctionReference struct {
	ID             int64     `json:"id" db:"id"`
	FunctionID     int64     `json:"function_id" db:"function_id"`
	ReferenceType  string    `json:"reference_type" db:"reference_type"` // "declaration", "usage", "modification"
	FileID         int64     `json:"file_id" db:"file_id"`
	Line           int       `json:"line" db:"line"`
	ColumnPosition int       `json:"column_position" db:"column_position"`
	Context        string    `json:"context" db:"context"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// FunctionStatement represents a statement within a function
type FunctionStatement struct {
	ID                int64                `json:"id" db:"id"`
	FunctionID        int64                `json:"function_id" db:"function_id"`
	StatementType     string               `json:"statement_type" db:"statement_type"` // "if", "for", "switch", etc.
	Text              string               `json:"text" db:"text"`
	Line              int                  `json:"line" db:"line"`
	Conditions        string               `json:"conditions" db:"conditions"` // JSON string
	Variables         string               `json:"variables" db:"variables"`   // JSON string
	Calls             string               `json:"calls" db:"calls"`           // JSON string
	ParentStatementID *int64               `json:"parent_statement_id,omitempty" db:"parent_statement_id"`
	CreatedAt         time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at" db:"updated_at"`
	Children          []FunctionStatement `json:"children,omitempty" db:"-"` // Nested statements, not stored in DB
	
	// Temporary fields used during conversion, not stored in DB
	FunctionIndex    int  `json:"-" db:"-"`  // Index in the functions slice for lookup
	ParentIndex      int  `json:"-" db:"-"`  // Index in the statements slice for parent lookup
}

// SymbolReference represents a reference to a symbol in the code
type SymbolReference struct {
	ID             int64     `json:"id" db:"id"`
	SymbolID       int64     `json:"symbol_id" db:"symbol_id"`
	ReferenceType  string    `json:"reference_type" db:"reference_type"` // "declaration", "usage", "modification"
	FileID         int64     `json:"file_id" db:"file_id"`
	Line           int       `json:"line" db:"line"`
	ColumnPosition int       `json:"column_position" db:"column_position"`
	Context        string    `json:"context" db:"context"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ExtendedRepositoryFunction includes the base RepositoryFunction with its related entities
type ExtendedRepositoryFunction struct {
	Function   RepositoryFunction  `json:"function"`
	Calls      []FunctionCall      `json:"calls,omitempty"`
	References []FunctionReference `json:"references,omitempty"`
	Statements []FunctionStatement `json:"statements,omitempty"`
}

// ExtendedRepositorySymbol includes the base RepositorySymbol with its related entities
type ExtendedRepositorySymbol struct {
	Symbol     RepositorySymbol  `json:"symbol"`
	References []SymbolReference `json:"references,omitempty"`
}

// ExtendedGetIndexResponse includes the base response with extended function and symbol information
type ExtendedGetIndexResponse struct {
	Repository *Repository                  `json:"repository"`
	Files      []RepositoryFile             `json:"files,omitempty"`
	Functions  []ExtendedRepositoryFunction `json:"functions,omitempty"`
	Symbols    []ExtendedRepositorySymbol   `json:"symbols,omitempty"`
}
