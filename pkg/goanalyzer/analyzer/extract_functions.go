package analyzer

import (
	"go/ast"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// extractFuncParams extracts function parameters
func (a *Analyzer) extractFuncParams(funcType *ast.FuncType, filePath string) []models.Symbol {
	var params []models.Symbol

	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			paramType := a.formatNode(param.Type)

			// Named parameters
			if len(param.Names) > 0 {
				for _, name := range param.Names {
					pos := a.fset.Position(name.Pos())
					params = append(params, models.Symbol{
						Name:     name.Name,
						Kind:     "parameter",
						Type:     paramType,
						Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					})
				}
			} else {
				// Unnamed parameters
				pos := a.fset.Position(param.Pos())
				params = append(params, models.Symbol{
					Name:     "",
					Kind:     "parameter",
					Type:     paramType,
					Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
				})
			}
		}
	}

	return params
}

// extractFuncResults extracts function results
func (a *Analyzer) extractFuncResults(funcType *ast.FuncType, filePath string) []models.Symbol {
	var results []models.Symbol

	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			resultType := a.formatNode(result.Type)

			// Named results
			if len(result.Names) > 0 {
				for _, name := range result.Names {
					pos := a.fset.Position(name.Pos())
					results = append(results, models.Symbol{
						Name:     name.Name,
						Kind:     "result",
						Type:     resultType,
						Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
					})
				}
			} else {
				// Unnamed results
				pos := a.fset.Position(result.Pos())
				results = append(results, models.Symbol{
					Name:     "",
					Kind:     "result",
					Type:     resultType,
					Position: models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
				})
			}
		}
	}

	return results
}
