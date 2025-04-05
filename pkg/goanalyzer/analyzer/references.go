package analyzer

import (
	"go/ast"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// analyzeReferences analyzes references to symbols
func (a *Analyzer) analyzeReferences(file *ast.File, filePath string, analysis *models.FileAnalysis) {
	// Track all identifiers used in the file
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			pos := a.fset.Position(node.Pos())
			symbol := node.Name

			// Skip if it's too short (likely a single letter variable)
			if len(symbol) <= 1 {
				return true
			}

			// Look for the symbol in our symbol table
			var found bool
			for qualifiedName, sym := range a.symbolTable {
				parts := strings.Split(qualifiedName, ".")
				simpleName := parts[len(parts)-1]

				if simpleName == symbol || qualifiedName == symbol {
					// Determine reference type
					refType := "usage"

					// If this is where the symbol is defined
					if sym.Position.File == filePath && sym.Position.Line == pos.Line && sym.Position.Column == pos.Column {
						refType = "declaration"
					}

					// Create reference info
					ref := models.ReferenceInfo{
						Symbol:   qualifiedName,
						Path:     filePath,
						RefType:  refType,
						Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					}

					// Add to references
					analysis.References = append(analysis.References, ref)

					// Also add to references map for lookup
					a.references[qualifiedName] = append(a.references[qualifiedName], ref)

					found = true
				}
			}

			// If not found but it's capitalized, it might be an exported symbol from another package
			if !found && len(symbol) > 0 && symbol[0] >= 'A' && symbol[0] <= 'Z' {
				// We would need imports and package info to resolve this properly
				// For now, just add as an unresolved reference
				ref := models.ReferenceInfo{
					Symbol:   symbol,
					Path:     filePath,
					RefType:  "usage",
					Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
				}

				analysis.References = append(analysis.References, ref)
			}
		}
		return true
	})
}
