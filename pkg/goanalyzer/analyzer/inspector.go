package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// extractCommentText extracts the text from a comment group
func (a *Analyzer) extractCommentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return cg.Text()
}

// analyzeFile analyzes a single file
func (a *Analyzer) analyzeFile(file *ast.File, filePath string) *models.FileAnalysis {
	analysis := &models.FileAnalysis{
		FilePath: filePath,
		Package:  file.Name.Name,
	}

	a.log().Info("Analyzing file", "file", filePath, "imports", file.Imports)
	// Extract imports
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		a.log().Info("Import", "path", path, "name", imp.Name)
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			// Extract the package name from the import path
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		pos := a.fset.Position(imp.Pos())
		analysis.Imports = append(analysis.Imports, models.Symbol{
			Name:     name,
			Kind:     "import",
			Value:    path,
			Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
		})
	}
	a.log().Info("Extracted imports", "file", filePath, "imports", analysis.Imports)

	// Walk through the AST and extract symbols
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.CONST {
				// Extract constants
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						// Extract comments for this value spec
						comments := a.extractCommentText(valueSpec.Doc)
						if comments == "" {
							// If no doc comment on the spec itself, try to get comment from the parent GenDecl
							comments = a.extractCommentText(node.Doc)
						}
						
						for i, name := range valueSpec.Names {
							pos := a.fset.Position(name.Pos())

							var value string
							if i < len(valueSpec.Values) {
								value = a.formatNode(valueSpec.Values[i])
							}

							var typeStr string
							if valueSpec.Type != nil {
								typeStr = a.formatNode(valueSpec.Type)
							} else if i < len(valueSpec.Values) {
								typeStr = "inferred"
							}

							symbol := models.Symbol{
								Name:     name.Name,
								Kind:     "constant",
								Type:     typeStr,
								Value:    value,
								Exported: name.IsExported(),
								Comments: comments,
								Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
							}

							analysis.Constants = append(analysis.Constants, symbol)

							// Add to symbol table
							qualifiedName := fmt.Sprintf("%s.%s", file.Name.Name, name.Name)
							a.symbolTable[qualifiedName] = symbol
						}
					}
				}
			} else if node.Tok == token.VAR {
				// Extract variables
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						// Extract comments for this value spec
						comments := a.extractCommentText(valueSpec.Doc)
						if comments == "" {
							// If no doc comment on the spec itself, try to get comment from the parent GenDecl
							comments = a.extractCommentText(node.Doc)
						}
						
						for i, name := range valueSpec.Names {
							pos := a.fset.Position(name.Pos())

							var value string
							if i < len(valueSpec.Values) {
								value = a.formatNode(valueSpec.Values[i])
							}

							var typeStr string
							if valueSpec.Type != nil {
								typeStr = a.formatNode(valueSpec.Type)
							} else if i < len(valueSpec.Values) {
								typeStr = "inferred"
							}

							symbol := models.Symbol{
								Name:     name.Name,
								Kind:     "variable",
								Type:     typeStr,
								Value:    value,
								Exported: name.IsExported(),
								Comments: comments,
								Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
							}

							analysis.Variables = append(analysis.Variables, symbol)

							// Add to symbol table
							qualifiedName := fmt.Sprintf("%s.%s", file.Name.Name, name.Name)
							a.symbolTable[qualifiedName] = symbol
						}
					}
				}
			} else if node.Tok == token.TYPE {
				// Extract types
				for _, spec := range node.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						pos := a.fset.Position(typeSpec.Pos())
						
						// Extract comments for this type spec
						comments := a.extractCommentText(typeSpec.Doc)
						if comments == "" {
							// If no doc comment on the spec itself, try to get comment from the parent GenDecl
							comments = a.extractCommentText(node.Doc)
						}

						typeSymbol := models.Symbol{
							Name:     typeSpec.Name.Name,
							Kind:     "type",
							Exported: typeSpec.Name.IsExported(),
							Comments: comments,
							Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
						}

						// Check if it's a struct or interface type
						switch typeNode := typeSpec.Type.(type) {
						case *ast.StructType:
							structSymbol := models.Symbol{
								Name:     typeSpec.Name.Name,
								Kind:     "struct",
								Exported: typeSpec.Name.IsExported(),
								Comments: comments,
								Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
							}

							// Extract fields
							if typeNode.Fields != nil {
								for _, field := range typeNode.Fields.List {
									// Extract field comments
									fieldComments := a.extractCommentText(field.Doc)
									for _, name := range field.Names {
										fieldPos := a.fset.Position(name.Pos())
										structSymbol.Fields = append(structSymbol.Fields, models.Symbol{
											Name:     name.Name,
											Kind:     "field",
											Type:     a.formatNode(field.Type),
											Exported: name.IsExported(),
											Comments: fieldComments,
											Position: models.Position{File: filePath, Line: fieldPos.Line, Column: fieldPos.Column},
										})
									}

									// Handle embedded fields (no name)
									if len(field.Names) == 0 {
										fieldPos := a.fset.Position(field.Pos())
										fieldType := a.formatNode(field.Type)

										// For embedded fields, the name is the type name
										fieldName := fieldType
										// Handle pointer embedded fields like *Type
										fieldName = strings.TrimPrefix(fieldName, "*")
										// Handle package qualified types like pkg.Type
										if parts := strings.Split(fieldName, "."); len(parts) > 1 {
											fieldName = parts[len(parts)-1]
										}

										structSymbol.Fields = append(structSymbol.Fields, models.Symbol{
											Name:     fieldName,
											Kind:     "embedded field",
											Type:     fieldType,
											Exported: true, // Embedding is usually for exported fields
											Comments: fieldComments,
											Position: models.Position{File: filePath, Line: fieldPos.Line, Column: fieldPos.Column},
										})
									}
								}
							}

							analysis.Structs = append(analysis.Structs, structSymbol)

							// Add to symbol table
							qualifiedName := fmt.Sprintf("%s.%s", file.Name.Name, typeSpec.Name.Name)
							a.symbolTable[qualifiedName] = structSymbol

						case *ast.InterfaceType:
							interfaceSymbol := models.Symbol{
								Name:     typeSpec.Name.Name,
								Kind:     "interface",
								Exported: typeSpec.Name.IsExported(),
								Comments: comments,
								Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
							}

							// Extract methods
							if typeNode.Methods != nil {
								for _, method := range typeNode.Methods.List {
									// Extract method comments
									methodComments := a.extractCommentText(method.Doc)
									
									for _, name := range method.Names {
										pos := a.fset.Position(name.Pos())

										methodSymbol := models.Symbol{
											Name:     name.Name,
											Kind:     "method",
											Exported: name.IsExported(),
											Comments: methodComments,
											Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
										}

										// If method has function type, extract parameters and results
										if funcType, ok := method.Type.(*ast.FuncType); ok {
											methodSymbol.Parameters = a.extractFuncParams(funcType, filePath)
											methodSymbol.Results = a.extractFuncResults(funcType, filePath)
										}

										interfaceSymbol.Methods = append(interfaceSymbol.Methods, methodSymbol.Name)
									}

									// Handle embedded interfaces (no name)
									if len(method.Names) == 0 {
										methodType := a.formatNode(method.Type)

										// For embedded interfaces, we just add the name
										interfaceSymbol.Methods = append(interfaceSymbol.Methods, methodType)
									}
								}
							}

							analysis.Interfaces = append(analysis.Interfaces, interfaceSymbol)

							// Add to symbol table
							qualifiedName := fmt.Sprintf("%s.%s", file.Name.Name, typeSpec.Name.Name)
							a.symbolTable[qualifiedName] = interfaceSymbol

						default:
							// Simple type alias
							typeSymbol.Type = a.formatNode(typeSpec.Type)
							analysis.Types = append(analysis.Types, typeSymbol)

							// Add to symbol table
							qualifiedName := fmt.Sprintf("%s.%s", file.Name.Name, typeSpec.Name.Name)
							a.symbolTable[qualifiedName] = typeSymbol
						}
					}
				}
			}

		case *ast.FuncDecl:
			pos := a.fset.Position(node.Pos())
			
			// Extract comments for this function
			comments := a.extractCommentText(node.Doc)

			funcSymbol := models.Symbol{
				Name:     node.Name.Name,
				Kind:     "function",
				Exported: node.Name.IsExported(),
				Comments: comments,
				Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
			}

			// Extract parameters and results
			if node.Type != nil {
				funcSymbol.Parameters = a.extractFuncParams(node.Type, filePath)
				funcSymbol.Results = a.extractFuncResults(node.Type, filePath)
			}

			// Check if this is a method
			if node.Recv != nil && len(node.Recv.List) > 0 {
				funcSymbol.Kind = "method"
				funcSymbol.Receiver = a.formatNode(node.Recv.List[0].Type)
			}

			analysis.Functions = append(analysis.Functions, funcSymbol)

			// Add to symbol table
			var qualifiedName string
			if funcSymbol.Kind == "method" {
				// For methods, include the receiver type
				receiverType := funcSymbol.Receiver
				if strings.HasPrefix(receiverType, "*") {
					receiverType = receiverType[1:] // Remove pointer for qualification
				}
				qualifiedName = fmt.Sprintf("%s.%s.%s", file.Name.Name, receiverType, node.Name.Name)
			} else {
				qualifiedName = fmt.Sprintf("%s.%s", file.Name.Name, node.Name.Name)
			}
			a.symbolTable[qualifiedName] = funcSymbol
		}

		return true
	})

	return analysis
}
