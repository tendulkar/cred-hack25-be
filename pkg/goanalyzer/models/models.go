package models

import "go/ast"

// Symbol represents a Go symbol such as a variable, function, or type
type Symbol struct {
	Name             string          `json:"name"`
	Kind             string          `json:"kind"`
	Type             string          `json:"type,omitempty"`
	Value            string          `json:"value,omitempty"`
	Exported         bool            `json:"exported"`
	Position         Position        `json:"position"`
	Comments         string          `json:"comments,omitempty"`   // Comments associated with the symbol
	Parameters       []Symbol        `json:"parameters,omitempty"`
	Results          []Symbol        `json:"results,omitempty"`
	Fields           []Symbol        `json:"fields,omitempty"`
	Methods          []string        `json:"methods,omitempty"`
	Receiver         string          `json:"receiver,omitempty"`
	CodeBlock        string          `json:"code_block,omitempty"`
	ASTNode          ast.Node        `json:"-"`                   // The AST node for this symbol
	Statements       []ast.Stmt      `json:"-"`                   // List of statements for functions/methods
	Declarations     []ast.Decl      `json:"-"`                   // List of declarations
	Expression       ast.Expr        `json:"-"`                   // For expressions
	StatementAnalysis []StatementInfo `json:"statement_analysis,omitempty"` // Detailed analysis of statements
}

// StatementInfo represents an analyzed statement with meaning
type StatementInfo struct {
	Type          string          `json:"type"`     // "if", "for", "switch", "return", etc.
	Text          string          `json:"text"`     // Text representation
	Position      Position        `json:"position"` // Position in source
	Conditions    []string        `json:"conditions,omitempty"`
	Variables     []string        `json:"variables,omitempty"`
	Calls         []string        `json:"calls,omitempty"`
	SubStatements []StatementInfo `json:"sub_statements,omitempty"`
}

// Position represents the position of a symbol in a file
type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// CallInfo represents information about a function call
type CallInfo struct {
	Caller     string   `json:"caller"`
	CallerPath string   `json:"caller_path"`
	Callee     string   `json:"callee"`
	CalleePath string   `json:"callee_path,omitempty"`
	Position   Position `json:"position"`
	Parameters []string `json:"parameters,omitempty"`
}

// ReferenceInfo represents a reference to a symbol
type ReferenceInfo struct {
	Symbol   string   `json:"symbol"`
	Path     string   `json:"path"`
	RefType  string   `json:"ref_type"` // "declaration", "usage", "modification"
	Position Position `json:"position"`
}

// FileAnalysis represents the analysis of a single file
type FileAnalysis struct {
	FilePath   string          `json:"file_path"`
	Package    string          `json:"package"`
	Imports    []Symbol        `json:"imports"`
	Constants  []Symbol        `json:"constants"`
	Variables  []Symbol        `json:"variables"`
	Types      []Symbol        `json:"types"`
	Functions  []Symbol        `json:"functions"`
	Structs    []Symbol        `json:"structs"`
	Interfaces []Symbol        `json:"interfaces"`
	Calls      []CallInfo      `json:"calls"`
	References []ReferenceInfo `json:"references"`
}

// PackageAnalysis represents the analysis of a package
type PackageAnalysis struct {
	PackagePath string         `json:"package_path"`
	PackageName string         `json:"package_name"`
	Files       []FileAnalysis `json:"files"`
}
