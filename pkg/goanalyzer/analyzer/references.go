package analyzer

import (
	"fmt"
	"go/ast"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
	"github.com/sirupsen/logrus"
)

// analyzeReferences analyzes references to symbols
func (a *Analyzer) analyzeReferences(file *ast.File, filePath string, analysis *models.FileAnalysis) {
	// Create a map to track assignment targets
	assignmentTargets := make(map[string]bool)
	positionRefs := make(map[string][]models.ReferenceInfo)

	// Build the parent map for AST nodes
	parentMap := make(map[ast.Node]ast.Node)
	v := &parentTrackingVisitor{parentMap: parentMap}
	ast.Walk(v, file)

	// First pass: identify assignment targets
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			// Mark all identifiers on the left-hand side as assignment targets
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					pos := a.fset.Position(id.Pos())
					key := id.Name + "_" + filePath + "_" + pos.String()
					assignmentTargets[key] = true
				}
			}
		case *ast.SelectorExpr:
			// Handle package-qualified identifiers
			if x, ok := node.X.(*ast.Ident); ok {
				// Check if this selector is part of an assignment
				parent := getParentAssignStmt(n, parentMap)
				if parent != nil {
					// Check if this is on the LHS of the assignment
					for _, lhs := range parent.Lhs {
						if sameNode(lhs, n) {
							pos := a.fset.Position(node.Sel.Pos())
							key := x.Name + "." + node.Sel.Name + "_" + filePath + "_" + pos.String()
							assignmentTargets[key] = true
						}
					}
				}
			}
		}
		return true
	})

	a.log().WithFields(logrus.Fields{
		"assignmentTargets": assignmentTargets,
		"parentMap":         parentMap,
		"symbolTable":       a.symbolTable,
	}).Debug("Identified assignment targets")
	// Second pass: track all identifiers used in the file
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			pos := a.fset.Position(node.Pos())
			symbol := node.Name

			// Skip if it's too short (likely a single letter variable)
			if len(symbol) <= 1 {
				return true
			}

			// Skip if it's a reserved keyword or a built-in type
			if isReservedOrBuiltin(symbol) {
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
					} else {
						// Check if this is a modification (assignment target)
						key := symbol + "_" + filePath + "_" + pos.String()
						if assignmentTargets[key] {
							refType = "modification"
						}
					}
					_ = refType

					// Create reference info
					// ref := models.ReferenceInfo{
					// 	Symbol:   qualifiedName,
					// 	Path:     filePath,
					// 	RefType:  refType,
					// 	Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					// }
					// posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
					// if _, ok := positionRefs[posKey]; ok {
					// 	break
					// }
					// positionRefs[posKey] = append(positionRefs[posKey], ref)

					// // Add to references
					// analysis.References = append(analysis.References, ref)

					// // Also add to references map for lookup
					// a.references[qualifiedName] = append(a.references[qualifiedName], ref)
					// a.log().WithFields(logrus.Fields{
					// 	"symbol":        symbol,
					// 	"qualifiedName": qualifiedName,
					// 	"sym":           sym,
					// 	"simpleName":    simpleName,
					// 	"refType":       refType,
					// }).Debug("Found reference in symbol table :", ref)
					// found = true
				}
			}

			// If not found but it's capitalized, it might be an exported symbol from another package
			if !found && len(symbol) > 0 && symbol[0] >= 'A' && symbol[0] <= 'Z' {
				// Try to resolve it using import information
				// resolvedSymbol := a.tryResolveExportedSymbol(symbol, file, filePath)

				// Determine reference type
				// refType := "usage"
				// key := symbol + "_" + filePath + "_" + pos.String()
				// if assignmentTargets[key] {
				// 	refType = "modification"
				// }

				// ref := models.ReferenceInfo{
				// 	Symbol:   resolvedSymbol,
				// 	Path:     filePath,
				// 	RefType:  refType,
				// 	Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
				// }

				// posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
				// if _, ok := positionRefs[posKey]; ok {
				// 	break
				// }
				// positionRefs[posKey] = []models.ReferenceInfo{}
				// positionRefs[posKey] = append(positionRefs[posKey], ref)

				// a.log().WithFields(logrus.Fields{
				// 	"symbol":         symbol,
				// 	"resolvedSymbol": resolvedSymbol,
				// 	"key":            key,
				// 	"refType":        refType,
				// }).Debug("Found reference in CAP symbol table :", ref)
				// analysis.References = append(analysis.References, ref)
			}
		case *ast.SelectorExpr:
			// Handle package-qualified references (pkg.Symbol)
			if x, ok := node.X.(*ast.Ident); ok {
				pkgName := x.Name
				symbolName := node.Sel.Name
				pos := a.fset.Position(node.Sel.Pos())

				// Skip if it's a field access on a variable (not a package)
				if !a.isPackage(pkgName, file) && !a.isImportAlias(pkgName, file) {
					return true
				}

				// Try to resolve the import path
				importPath := a.resolveImportPath(pkgName, file)
				qualifiedName := pkgName + "." + symbolName
				if importPath != "" {
					qualifiedName = importPath + "." + symbolName
				}

				// Determine reference type
				refType := "usage"
				key := pkgName + "." + symbolName + "_" + filePath + "_" + pos.String()
				if assignmentTargets[key] {
					refType = "modification"
				}

				ref := models.ReferenceInfo{
					Symbol:   qualifiedName,
					Path:     filePath,
					RefType:  refType,
					Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
				}

				posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
				if _, ok := positionRefs[posKey]; ok {
					break
				}
				positionRefs[posKey] = []models.ReferenceInfo{}
				positionRefs[posKey] = append(positionRefs[posKey], ref)

				a.log().WithFields(logrus.Fields{
					"symbolName":    symbolName,
					"importPath":    importPath,
					"pkgName":       pkgName,
					"qualifiedName": qualifiedName,
					"refType":       refType,
				}).Debug("Found reference in symbol table :", ref)
				analysis.References = append(analysis.References, ref)
			}
		case *ast.CallExpr:
			// Handle function calls
			switch fun := node.Fun.(type) {
			case *ast.Ident:
				// Direct function call (e.g., foo())
				pos := a.fset.Position(fun.Pos())
				symbol := fun.Name

				// Skip if it's a built-in function
				if isReservedOrBuiltin(symbol) {
					return true
				}

				// Look for the function in our symbol table
				var found bool
				var bestMatch string
				for qualifiedName, _ := range a.symbolTable {
					parts := strings.Split(qualifiedName, ".")
					simpleName := parts[len(parts)-1]

					if simpleName == symbol || qualifiedName == symbol {
						bestMatch = qualifiedName
						found = true
						break
					}
				}

				if found {
					// Create reference info - function calls are always usage
					ref := models.ReferenceInfo{
						Symbol:   bestMatch,
						Path:     filePath,
						RefType:  "usage",
						Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					}

					posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
					if _, ok := positionRefs[posKey]; ok {
						break
					}
					positionRefs[posKey] = []models.ReferenceInfo{}
					positionRefs[posKey] = append(positionRefs[posKey], ref)

					a.log().WithFields(logrus.Fields{
						"symbol":    symbol,
						"bestMatch": bestMatch,
						"filePath":  filePath,
						"pos":       pos.String(),
					}).Debug("Found reference in CallExpr table :", ref)

					analysis.References = append(analysis.References, ref)
					// Also add to references map for lookup
					a.references[bestMatch] = append(a.references[bestMatch], ref)
				} else if len(symbol) > 0 && symbol[0] >= 'A' && symbol[0] <= 'Z' {
					// If not found but it's capitalized, it might be an exported function from another package
					// Try to resolve it using import information
					resolvedSymbol := a.tryResolveExportedSymbol(symbol, file, filePath)

					ref := models.ReferenceInfo{
						Symbol:   resolvedSymbol,
						Path:     filePath,
						RefType:  "usage",
						Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					}
					posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
					if _, ok := positionRefs[posKey]; ok {
						break
					}
					positionRefs[posKey] = []models.ReferenceInfo{}
					positionRefs[posKey] = append(positionRefs[posKey], ref)

					a.log().WithFields(logrus.Fields{
						"symbol":         symbol,
						"resolvedSymbol": resolvedSymbol,
						"filePath":       filePath,
						"pos":            pos.String(),
					}).Debug("Found reference in CallExpr CAP table :", ref)
					analysis.References = append(analysis.References, ref)
				}

			case *ast.SelectorExpr:
				// Method call (e.g., x.foo())
				if x, ok := fun.X.(*ast.Ident); ok {
					pkgName := x.Name
					methodName := fun.Sel.Name
					pos := a.fset.Position(fun.Sel.Pos())

					// Try to determine if this is a package-qualified function call or a method call
					if a.isPackage(pkgName, file) || a.isImportAlias(pkgName, file) {
						// It's a package-qualified function call (e.g., fmt.Println())
						importPath := a.resolveImportPath(pkgName, file)
						qualifiedName := pkgName + "." + methodName
						if importPath != "" {
							qualifiedName = importPath + "." + methodName
						}
						_ = qualifiedName

						// Only add one reference for package-qualified function call
						// ref := models.ReferenceInfo{
						// 	Symbol:   qualifiedName,
						// 	Path:     filePath,
						// 	RefType:  "usage",
						// 	Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
						// }

						// posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
						// if _, ok := positionRefs[posKey]; ok {
						// 	break
						// }
						// positionRefs[posKey] = []models.ReferenceInfo{}
						// positionRefs[posKey] = append(positionRefs[posKey], ref)

						// a.log().WithFields(logrus.Fields{
						// 	"symbol":     qualifiedName,
						// 	"pkgName":    pkgName,
						// 	"methodName": methodName,
						// 	"filePath":   filePath,
						// 	"pos":        pos.String(),
						// }).Debug("Found reference in SelectorExpr table :", ref)
						// analysis.References = append(analysis.References, ref)
					} else {
						// It's likely a method call on a variable
						// Here we need to find the type of 'x' to fully qualify the method
						// For simplicity, we'll just record it as a reference to x.methodName
						qualifiedName := pkgName + "." + methodName

						// Add only one reference for method call
						ref := models.ReferenceInfo{
							Symbol:   qualifiedName,
							Path:     filePath,
							RefType:  "usage",
							Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
						}

						posKey := fmt.Sprintf("%d:%d", pos.Line, pos.Column)
						if _, ok := positionRefs[posKey]; ok {
							break
						}
						positionRefs[posKey] = []models.ReferenceInfo{}
						positionRefs[posKey] = append(positionRefs[posKey], ref)

						a.log().WithFields(logrus.Fields{
							"symbol":     qualifiedName,
							"pkgName":    pkgName,
							"methodName": methodName,
							"filePath":   filePath,
							"pos":        pos.String(),
						}).Debug("Found reference in SelectorExpr method call on variable :", ref)
						analysis.References = append(analysis.References, ref)
					}
				}
			}
		}
		return true
	})
}

// tryResolveExportedSymbol attempts to resolve an exported symbol to its package
func (a *Analyzer) tryResolveExportedSymbol(symbol string, file *ast.File, filePath string) string {
	// Default to just the symbol name if we can't resolve it
	return symbol
}

// isPackage checks if an identifier is a package name
func (a *Analyzer) isPackage(name string, file *ast.File) bool {
	// Check if it's in the imports
	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == name {
			return true
		}

		// Extract the package name from the import path
		importPath := strings.Trim(imp.Path.Value, `"`)
		parts := strings.Split(importPath, "/")
		pkgName := parts[len(parts)-1]
		if pkgName == name {
			return true
		}
	}

	return false
}

// isImportAlias checks if a name is an import alias
func (a *Analyzer) isImportAlias(name string, file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == name {
			return true
		}
	}

	return false
}

// resolveImportPath resolves a package name to its import path
func (a *Analyzer) resolveImportPath(pkgName string, file *ast.File) string {
	for _, imp := range file.Imports {
		// Check explicit aliases
		if imp.Name != nil && imp.Name.Name == pkgName {
			return strings.Trim(imp.Path.Value, `"`)
		}

		// Check if the last part of the import path matches
		importPath := strings.Trim(imp.Path.Value, `"`)
		parts := strings.Split(importPath, "/")
		lastPart := parts[len(parts)-1]
		if lastPart == pkgName {
			return importPath
		}
	}

	return ""
}

// We need a custom visitor to track parent-child relationships
type parentTrackingVisitor struct {
	parentMap     map[ast.Node]ast.Node
	currentParent ast.Node
}

// Visit implements the ast.Visitor interface
func (v *parentTrackingVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Store the parent for this node
	if v.currentParent != nil {
		v.parentMap[node] = v.currentParent
	}

	// Use this node as the parent for its children
	v.currentParent = node

	return &parentTrackingVisitor{
		parentMap:     v.parentMap,
		currentParent: v.currentParent,
	}
}

// findParentNode looks for a parent node of a specific type
func findParentNode[T ast.Node](parentMap map[ast.Node]ast.Node, node ast.Node) T {
	var zero T
	current := node
	for {
		parent, exists := parentMap[current]
		if !exists {
			return zero
		}

		if typed, ok := parent.(T); ok {
			return typed
		}

		current = parent
	}
}

// getParentAssignStmt finds the parent assignment statement for a node if it exists
func getParentAssignStmt(node ast.Node, parentMap map[ast.Node]ast.Node) *ast.AssignStmt {
	return findParentNode[*ast.AssignStmt](parentMap, node)
}

// sameNode checks if two nodes are the same
func sameNode(a, b ast.Node) bool {
	return a.Pos() == b.Pos() && a.End() == b.End()
}

// isReservedOrBuiltin checks if a word is a Go reserved keyword or built-in type
func isReservedOrBuiltin(word string) bool {
	// Go keywords
	keywords := map[string]bool{
		"break": true, "default": true, "func": true, "interface": true,
		"select": true, "case": true, "defer": true, "go": true, "map": true,
		"struct": true, "chan": true, "else": true, "goto": true, "package": true,
		"switch": true, "const": true, "fallthrough": true, "if": true,
		"range": true, "type": true, "continue": true, "for": true,
		"import": true, "return": true, "var": true,
	}

	// Go built-in types
	builtins := map[string]bool{
		"bool": true, "byte": true, "complex64": true, "complex128": true,
		"error": true, "float32": true, "float64": true, "int": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"rune": true, "string": true, "uint": true, "uint8": true,
		"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		// Built-in functions
		"append": true, "cap": true, "close": true, "complex": true,
		"copy": true, "delete": true, "imag": true, "len": true,
		"make": true, "new": true, "panic": true, "print": true,
		"println": true, "real": true, "recover": true,
	}

	return keywords[word] || builtins[word]
}
