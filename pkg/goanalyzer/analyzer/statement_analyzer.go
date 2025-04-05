package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"cred.com/hack25/backend/pkg/goanalyzer/models"
)

// AnalyzeStatements processes a list of statements and extracts meaningful information
func (a *Analyzer) AnalyzeStatements(stmts []ast.Stmt, filePath string) []models.StatementInfo {
	var result []models.StatementInfo

	for _, stmt := range stmts {
		info := a.analyzeStatement(stmt, filePath)
		if info != nil {
			result = append(result, *info)
		}
	}

	return result
}

// analyzeStatement analyzes a single statement
func (a *Analyzer) analyzeStatement(stmt ast.Stmt, filePath string) *models.StatementInfo {
	if stmt == nil {
		return nil
	}

	pos := a.fset.Position(stmt.Pos())
	stmtPos := models.Position{
		File:   filePath,
		Line:   pos.Line,
		Column: pos.Column,
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return &models.StatementInfo{
			Type:      "assignment",
			Text:      formatAssignment(a, s),
			Position:  stmtPos,
			Variables: extractVariables(a, s.Lhs),
			Calls:     extractCalls(a, s.Rhs),
		}

	case *ast.BlockStmt:
		info := &models.StatementInfo{
			Type:     "block",
			Text:     "{...}",
			Position: stmtPos,
		}

		for _, blockStmt := range s.List {
			subInfo := a.analyzeStatement(blockStmt, filePath)
			if subInfo != nil {
				info.SubStatements = append(info.SubStatements, *subInfo)
			}
		}

		return info

	case *ast.BranchStmt:
		branchType := ""
		switch s.Tok {
		case token.BREAK:
			branchType = "break"
		case token.CONTINUE:
			branchType = "continue"
		case token.GOTO:
			branchType = "goto"
		case token.FALLTHROUGH:
			branchType = "fallthrough"
		}

		var label string
		if s.Label != nil {
			label = s.Label.Name
		}

		return &models.StatementInfo{
			Type:     branchType,
			Text:     fmt.Sprintf("%s %s", branchType, label),
			Position: stmtPos,
		}

	case *ast.DeclStmt:
		// Handle declarations within blocks
		if genDecl, ok := s.Decl.(*ast.GenDecl); ok {
			switch genDecl.Tok {
			case token.VAR:
				vars := []string{}
				for _, spec := range genDecl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							vars = append(vars, name.Name)
						}
					}
				}

				return &models.StatementInfo{
					Type:      "variable_declaration",
					Text:      fmt.Sprintf("var %s", strings.Join(vars, ", ")),
					Position:  stmtPos,
					Variables: vars,
				}

			case token.CONST:
				consts := []string{}
				for _, spec := range genDecl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							consts = append(consts, name.Name)
						}
					}
				}

				return &models.StatementInfo{
					Type:      "constant_declaration",
					Text:      fmt.Sprintf("const %s", strings.Join(consts, ", ")),
					Position:  stmtPos,
					Variables: consts,
				}
			}
		}

		return &models.StatementInfo{
			Type:     "declaration",
			Text:     "declaration",
			Position: stmtPos,
		}

	case *ast.ExprStmt:
		// Most common case: function calls
		if call, ok := s.X.(*ast.CallExpr); ok {
			callInfo := &models.StatementInfo{
				Type:     "function_call",
				Text:     a.formatNode(call),
				Position: stmtPos,
				Calls:    []string{a.formatNode(call.Fun)},
			}

			return callInfo
		}

		return &models.StatementInfo{
			Type:     "expression",
			Text:     a.formatNode(s.X),
			Position: stmtPos,
		}

	case *ast.ForStmt:
		info := &models.StatementInfo{
			Type:     "for_loop",
			Text:     "for ...",
			Position: stmtPos,
		}

		// Extract condition if present
		if s.Cond != nil {
			info.Conditions = append(info.Conditions, a.formatNode(s.Cond))
		}

		// Extract initialization variables
		if init, ok := s.Init.(*ast.AssignStmt); ok {
			info.Variables = extractVariables(a, init.Lhs)
		}

		// Process body
		if s.Body != nil {
			blockInfo := a.analyzeStatement(s.Body, filePath)
			if blockInfo != nil {
				info.SubStatements = blockInfo.SubStatements
			}
		}

		return info

	case *ast.IfStmt:
		info := &models.StatementInfo{
			Type:     "if_statement",
			Text:     "if ...",
			Position: stmtPos,
		}

		// Extract condition
		if s.Cond != nil {
			info.Conditions = append(info.Conditions, a.formatNode(s.Cond))
		}

		// Extract initialization variables
		if init, ok := s.Init.(*ast.AssignStmt); ok {
			info.Variables = extractVariables(a, init.Lhs)
		}

		// Process then clause
		if s.Body != nil {
			thenInfo := a.analyzeStatement(s.Body, filePath)
			if thenInfo != nil {
				info.SubStatements = append(info.SubStatements, *thenInfo)
			}
		}

		// Process else clause
		if s.Else != nil {
			elseInfo := a.analyzeStatement(s.Else, filePath)
			if elseInfo != nil {
				elseInfo.Type = "else_clause"
				info.SubStatements = append(info.SubStatements, *elseInfo)
			}
		}

		return info

	case *ast.IncDecStmt:
		var op string
		switch s.Tok {
		case token.INC:
			op = "++"
		case token.DEC:
			op = "--"
		}

		return &models.StatementInfo{
			Type:      "inc_dec",
			Text:      a.formatNode(s.X) + op,
			Position:  stmtPos,
			Variables: []string{a.formatNode(s.X)},
		}

	case *ast.RangeStmt:
		info := &models.StatementInfo{
			Type:     "range_loop",
			Text:     "for ... range ...",
			Position: stmtPos,
		}

		// Extract key and value variables
		if s.Key != nil {
			info.Variables = append(info.Variables, a.formatNode(s.Key))
		}

		if s.Value != nil {
			info.Variables = append(info.Variables, a.formatNode(s.Value))
		}

		// Extract range expression
		if s.X != nil {
			info.Conditions = append(info.Conditions, "range "+a.formatNode(s.X))
		}

		// Process body
		if s.Body != nil {
			blockInfo := a.analyzeStatement(s.Body, filePath)
			if blockInfo != nil {
				info.SubStatements = blockInfo.SubStatements
			}
		}

		return info

	case *ast.ReturnStmt:
		info := &models.StatementInfo{
			Type:     "return",
			Text:     "return",
			Position: stmtPos,
		}

		// Extract returned values
		var returnVals []string
		for _, result := range s.Results {
			returnVals = append(returnVals, a.formatNode(result))

			// Check for function calls in return statement
			if call, ok := result.(*ast.CallExpr); ok {
				info.Calls = append(info.Calls, a.formatNode(call.Fun))
			}
		}

		if len(returnVals) > 0 {
			info.Text = "return " + strings.Join(returnVals, ", ")
		}

		return info

	case *ast.SwitchStmt:
		info := &models.StatementInfo{
			Type:     "switch",
			Text:     "switch ...",
			Position: stmtPos,
		}

		// Extract tag expression
		if s.Tag != nil {
			info.Conditions = append(info.Conditions, a.formatNode(s.Tag))
		}

		// Extract initialization variables
		if init, ok := s.Init.(*ast.AssignStmt); ok {
			info.Variables = extractVariables(a, init.Lhs)
		}

		// Process body
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if caseClause, ok := clause.(*ast.CaseClause); ok {
					caseInfo := &models.StatementInfo{
						Type: "case",
						Position: models.Position{
							File:   filePath,
							Line:   a.fset.Position(caseClause.Pos()).Line,
							Column: a.fset.Position(caseClause.Pos()).Column,
						},
					}

					// Extract case expressions
					var caseExprs []string
					for _, expr := range caseClause.List {
						caseExprs = append(caseExprs, a.formatNode(expr))
					}

					if len(caseExprs) > 0 {
						caseInfo.Text = "case " + strings.Join(caseExprs, ", ")
						caseInfo.Conditions = caseExprs
					} else {
						caseInfo.Text = "default"
						caseInfo.Type = "default"
					}

					// Extract case body
					for _, stmt := range caseClause.Body {
						stmtInfo := a.analyzeStatement(stmt, filePath)
						if stmtInfo != nil {
							caseInfo.SubStatements = append(caseInfo.SubStatements, *stmtInfo)
						}
					}

					info.SubStatements = append(info.SubStatements, *caseInfo)
				}
			}
		}

		return info

	default:
		return &models.StatementInfo{
			Type:     fmt.Sprintf("%T", stmt),
			Text:     fmt.Sprintf("<%T>", stmt),
			Position: stmtPos,
		}
	}
}

// extractVariables extracts variable names from a list of expressions
func extractVariables(a *Analyzer, exprs []ast.Expr) []string {
	var vars []string

	for _, expr := range exprs {
		switch e := expr.(type) {
		case *ast.Ident:
			vars = append(vars, e.Name)
		case *ast.SelectorExpr:
			vars = append(vars, a.formatNode(e))
		}
	}

	return vars
}

// extractCalls extracts function calls from a list of expressions
func extractCalls(a *Analyzer, exprs []ast.Expr) []string {
	var calls []string

	for _, expr := range exprs {
		if call, ok := expr.(*ast.CallExpr); ok {
			calls = append(calls, a.formatNode(call.Fun))
		}
	}

	return calls
}

// formatAssignment formats an assignment statement in a readable form
func formatAssignment(a *Analyzer, stmt *ast.AssignStmt) string {
	var lhs, rhs []string

	for _, expr := range stmt.Lhs {
		lhs = append(lhs, a.formatNode(expr))
	}

	for _, expr := range stmt.Rhs {
		rhs = append(rhs, a.formatNode(expr))
	}

	var op string
	switch stmt.Tok {
	case token.ASSIGN:
		op = "="
	case token.ADD_ASSIGN:
		op = "+="
	case token.SUB_ASSIGN:
		op = "-="
	case token.MUL_ASSIGN:
		op = "*="
	case token.QUO_ASSIGN:
		op = "/="
	case token.REM_ASSIGN:
		op = "%="
	case token.AND_ASSIGN:
		op = "&="
	case token.OR_ASSIGN:
		op = "|="
	case token.XOR_ASSIGN:
		op = "^="
	case token.SHL_ASSIGN:
		op = "<<="
	case token.SHR_ASSIGN:
		op = ">>="
	case token.AND_NOT_ASSIGN:
		op = "&^="
	case token.DEFINE:
		op = ":="
	default:
		op = "?"
	}

	return strings.Join(lhs, ", ") + " " + op + " " + strings.Join(rhs, ", ")
}
