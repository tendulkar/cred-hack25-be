package analyzer

import (
	"fmt"
	"go/ast"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// analyzeCallHierarchy analyzes function calls
func (a *Analyzer) analyzeCallHierarchy(file *ast.File, filePath string, analysis *models.FileAnalysis) {
	// Find all function calls
	ast.Inspect(file, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		pos := a.fset.Position(callExpr.Pos())

		// Determine the caller function
		var callerFunc *models.Symbol
		fileContent, ok := a.codeMap[filePath]
		for _, fn := range analysis.Functions {
			fnPos := fn.Position
			if fnPos.File == filePath &&
				fnPos.Line <= pos.Line &&
				(len(fn.CodeBlock) == 0 || (ok && pos.Offset < len(fileContent) &&
					strings.Contains(fn.CodeBlock, fileContent[pos.Offset:pos.Offset+1]))) {
				callerFunc = &fn
				break
			}
		}

		// If we couldn't find the containing function, skip
		if callerFunc == nil {
			return true
		}

		// Determine the callee function name
		var calleeName string
		switch fun := callExpr.Fun.(type) {
		case *ast.Ident:
			calleeName = fun.Name
		case *ast.SelectorExpr:
			if x, ok := fun.X.(*ast.Ident); ok {
				calleeName = x.Name + "." + fun.Sel.Name
			} else {
				calleeName = a.formatNode(fun.X) + "." + fun.Sel.Name
			}
		default:
			calleeName = a.formatNode(callExpr.Fun)
		}

		// Extract parameters
		var params []string
		for _, arg := range callExpr.Args {
			params = append(params, a.formatNode(arg))
		}

		// Create call info
		call := models.CallInfo{
			Caller:     callerFunc.Name,
			CallerPath: filePath,
			Callee:     calleeName,
			Position:   models.Position{File: filePath, Line: pos.Line, Column: pos.Column},
			Parameters: params,
		}

		// Add to calls
		analysis.Calls = append(analysis.Calls, call)

		// Also add to call graph for lookup
		callerKey := fmt.Sprintf("%s:%s", filePath, callerFunc.Name)
		a.callGraph[callerKey] = append(a.callGraph[callerKey], call)

		return true
	})
}
