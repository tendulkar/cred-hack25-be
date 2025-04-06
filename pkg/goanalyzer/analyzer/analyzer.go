package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Analyzer is the main analyzer struct
type Analyzer struct {
	fset        *token.FileSet
	codeMap     map[string]string
	fileMap     map[string]*ast.File
	callGraph   map[string][]models.CallInfo
	references  map[string][]models.ReferenceInfo
	symbolTable map[string]models.Symbol
}

// New creates a new code analyzer
func New() *Analyzer {
	return &Analyzer{
		fset:        token.NewFileSet(),
		codeMap:     make(map[string]string),
		fileMap:     make(map[string]*ast.File),
		callGraph:   make(map[string][]models.CallInfo),
		references:  make(map[string][]models.ReferenceInfo),
		symbolTable: make(map[string]models.Symbol),
	}
}

func (a *Analyzer) log() *logrus.Entry {
	return logger.Log.WithField("component", "code-analyzer")
}

// AnalyzeFile analyzes a single Go file
func (a *Analyzer) AnalyzeFile(filePath string) (*models.FileAnalysis, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Store the code content for extracting code blocks later
	a.codeMap[filePath] = string(content)

	// Parse the file
	file, err := parser.ParseFile(a.fset, filePath, content, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %w", err)
	}

	// Store the file AST for later reference analysis
	a.fileMap[filePath] = file

	// Analyze the file
	analysis := a.analyzeFile(file, filePath)

	// Extract code blocks for functions
	a.extractCodeBlocks(file, filePath, analysis)

	// Analyze call hierarchy
	a.analyzeCallHierarchy(file, filePath, analysis)

	// Analyze references
	a.analyzeReferences(file, filePath, analysis)

	return analysis, nil
}

// AnalyzeDirectory analyzes all Go files in a directory
func (a *Analyzer) AnalyzeDirectory(dirPath string) ([]models.FileAnalysis, error) {
	var results []models.FileAnalysis

	// First pass: load all files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Just load and parse the file, don't analyze yet
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil // continue with next file
			}

			// Store the code content
			a.codeMap[path] = string(content)

			// Parse the file
			file, err := parser.ParseFile(a.fset, path, content, parser.AllErrors|parser.ParseComments)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", path, err)
				return nil // continue with next file
			}

			// Store the file AST
			a.fileMap[path] = file
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Second pass: analyze all files
	for path, file := range a.fileMap {
		analysis := a.analyzeFile(file, path)

		// Extract code blocks for functions
		a.extractCodeBlocks(file, path, analysis)

		results = append(results, *analysis)
	}

	// Third pass: analyze call hierarchy and references
	for i, analysis := range results {
		path := analysis.FilePath
		file := a.fileMap[path]

		// Analyze call hierarchy
		a.analyzeCallHierarchy(file, path, &results[i])

		// Analyze references
		a.analyzeReferences(file, path, &results[i])
	}

	return results, nil
}

// extractCodeBlocks extracts the code block and AST nodes for functions
func (a *Analyzer) extractCodeBlocks(file *ast.File, filePath string, analysis *models.FileAnalysis) {
	fileContent, ok := a.codeMap[filePath]
	if !ok {
		return // Skip if we don't have the file content
	}

	// Extract code blocks for functions
	for i, fn := range analysis.Functions {
		// Find the function in the AST
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Check if this is the right function
			if funcDecl.Name.Name == fn.Name {
				// For methods, check the receiver
				if fn.Kind == "method" {
					if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
						return true
					}
					receiverType := a.formatNode(funcDecl.Recv.List[0].Type)
					if receiverType != fn.Receiver {
						return true
					}
				} else if funcDecl.Recv != nil {
					return true
				}

				// Get the function's position
				start := a.fset.Position(funcDecl.Pos())
				end := a.fset.Position(funcDecl.End())

				// Store the AST node and statements
				analysis.Functions[i].ASTNode = funcDecl
				if funcDecl.Body != nil {
					analysis.Functions[i].Statements = funcDecl.Body.List
				}

				// Extract the code block as text
				if start.Offset < end.Offset && end.Offset <= len(fileContent) {
					codeBlock := fileContent[start.Offset:end.Offset]
					analysis.Functions[i].CodeBlock = codeBlock
				}

				// Analyze statements for a detailed breakdown
				if funcDecl.Body != nil {
					stmtAnalysis := a.AnalyzeStatements(funcDecl.Body.List, filePath)
					// Store the statement analysis in the symbol for later persistence
					analysis.Functions[i].StatementAnalysis = stmtAnalysis
				}

				return false
			}
			return true
		})
	}
}

// GetCallHierarchy returns the call hierarchy for a specific function
func (a *Analyzer) GetCallHierarchy(filePath, funcName string) []models.CallInfo {
	key := fmt.Sprintf("%s:%s", filePath, funcName)
	return a.callGraph[key]
}

// GetReferences returns all references to a symbol
func (a *Analyzer) GetReferences(symbolName string) []models.ReferenceInfo {
	return a.references[symbolName]
}

// GetSymbol returns a symbol by name
func (a *Analyzer) GetSymbol(symbolName string) (models.Symbol, bool) {
	sym, ok := a.symbolTable[symbolName]
	return sym, ok
}
