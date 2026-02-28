package transpiler

import (
	"fmt"
	"go/ast"
	"strings"
)

// MatchCase represents a single case in a type switch or match expression
type MatchCase struct {
	Pattern   string   // "Left", "Right", "Some", "None"
	VarName   string   // The variable name bound in the case
	VarType   string   // The type of the bound variable
	BodyStmts []string // Statements in the case body
}

// MatchExpression represents a complete match/type-switch
type MatchExpression struct {
	Scrutinee     string      // The expression being matched
	ScrutineeType string      // The type of the scrutinee (e.g., "Either<u256, [u8; 64]>")
	Cases         []MatchCase // The cases
}

// analyzeTypeSwitchStmt extracts pattern matching info from a Go type switch
func (t *Transpiler) analyzeTypeSwitchStmt(stmt *ast.TypeSwitchStmt) (*MatchExpression, error) {
	match := &MatchExpression{}

	// Extract the scrutinee from the assign statement
	// switch v := expr.(type) { ... }
	if assign, ok := stmt.Assign.(*ast.AssignStmt); ok {
		if len(assign.Rhs) == 1 {
			if typeAssert, ok := assign.Rhs[0].(*ast.TypeAssertExpr); ok {
				scrutinee, err := t.evaluateExpression(typeAssert.X)
				if err != nil {
					return nil, err
				}
				match.Scrutinee = scrutinee
			}
		}
	}

	// Process each case clause
	for _, stmt := range stmt.Body.List {
		if caseClause, ok := stmt.(*ast.CaseClause); ok {
			matchCase, err := t.analyzeCaseClause(caseClause)
			if err != nil {
				return nil, err
			}
			if matchCase != nil {
				match.Cases = append(match.Cases, *matchCase)
			}
		}
	}

	return match, nil
}

// analyzeCaseClause extracts a single case from a type switch
func (t *Transpiler) analyzeCaseClause(clause *ast.CaseClause) (*MatchCase, error) {
	mc := &MatchCase{}

	// Get the pattern from the case type
	if len(clause.List) > 0 {
		// case Left: or case Right:
		if ident, ok := clause.List[0].(*ast.Ident); ok {
			mc.Pattern = ident.Name
		}
	} else {
		// default case
		mc.Pattern = "_"
	}

	// Process the body statements
	for _, bodyStmt := range clause.Body {
		stmtStr, err := t.analyzeStatement(bodyStmt)
		if err != nil {
			return nil, err
		}
		if stmtStr != "" {
			mc.BodyStmts = append(mc.BodyStmts, stmtStr)
		}
	}

	return mc, nil
}

// analyzeStatement converts a Go statement to SimplicityHL
func (t *Transpiler) analyzeStatement(stmt ast.Stmt) (string, error) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return t.analyzeAssignStmt(s)
	case *ast.ExprStmt:
		return t.analyzeExprStmt(s)
	case *ast.ReturnStmt:
		return t.analyzeReturnStmt(s)
	case *ast.DeclStmt:
		return t.analyzeDeclStmt(s)
	default:
		return "", nil
	}
}

// analyzeAssignStmt converts assignment statements
func (t *Transpiler) analyzeAssignStmt(stmt *ast.AssignStmt) (string, error) {
	if len(stmt.Lhs) == 1 && len(stmt.Rhs) == 1 {
		lhs := ""
		if ident, ok := stmt.Lhs[0].(*ast.Ident); ok {
			lhs = t.toSnakeCase(ident.Name)
		}

		// Check if this is a jet call
		if callExpr, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if selIdent, ok := sel.X.(*ast.Ident); ok && selIdent.Name == "jet" {
					jetCall, err := t.evaluateJetCall(sel.Sel.Name, callExpr.Args)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("let %s = %s;", lhs, jetCall), nil
				}
			}
		}

		rhs, err := t.evaluateExpression(stmt.Rhs[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("let %s = %s;", lhs, rhs), nil
	}

	// Handle tuple destructuring: a, b := tuple
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
		var names []string
		for _, l := range stmt.Lhs {
			if ident, ok := l.(*ast.Ident); ok {
				names = append(names, t.toSnakeCase(ident.Name))
			}
		}
		rhs, err := t.evaluateExpression(stmt.Rhs[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("let (%s) = %s;", strings.Join(names, ", "), rhs), nil
	}

	return "", nil
}

// analyzeExprStmt converts expression statements (like jet calls)
func (t *Transpiler) analyzeExprStmt(stmt *ast.ExprStmt) (string, error) {
	if callExpr, ok := stmt.X.(*ast.CallExpr); ok {
		// jet.X(...) selector calls
		if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if selIdent, ok := sel.X.(*ast.Ident); ok && selIdent.Name == "jet" {
				jetCall, err := t.evaluateJetCall(sel.Sel.Name, callExpr.Args)
				if err != nil {
					return "", err
				}
				return jetCall, nil
			}
		}
		// User-defined function calls (e.g., verifyHashlock(...))
		if _, ok := callExpr.Fun.(*ast.Ident); ok {
			return t.evaluateCallExpr(callExpr)
		}
	}
	return "", nil
}

// analyzeReturnStmt converts return statements
func (t *Transpiler) analyzeReturnStmt(stmt *ast.ReturnStmt) (string, error) {
	if len(stmt.Results) == 0 {
		return "", nil
	}
	result, err := t.evaluateExpression(stmt.Results[0])
	if err != nil {
		return "", err
	}
	return result, nil
}

// analyzeDeclStmt handles variable declarations
func (t *Transpiler) analyzeDeclStmt(stmt *ast.DeclStmt) (string, error) {
	// For now, skip declarations inside match arms
	// They're usually handled as witness values
	return "", nil
}

// generateMatchExpression generates SimplicityHL match expression
func (t *Transpiler) generateMatchExpression(match *MatchExpression, indent string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%smatch %s {\n", indent, match.Scrutinee))

	for i, mc := range match.Cases {
		// Generate the pattern
		pattern := mc.Pattern
		if mc.VarName != "" {
			pattern = fmt.Sprintf("%s(%s)", mc.Pattern, mc.VarName)
		}

		sb.WriteString(fmt.Sprintf("%s    %s => {\n", indent, pattern))

		// Generate body statements (each entry may span multiple lines)
		for _, bodyStmt := range mc.BodyStmts {
			for _, line := range strings.Split(bodyStmt, "\n") {
				if strings.TrimSpace(line) != "" {
					sb.WriteString(fmt.Sprintf("%s        %s\n", indent, line))
				}
			}
		}

		sb.WriteString(fmt.Sprintf("%s    }", indent))
		if i < len(match.Cases)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%s}", indent))
	return sb.String()
}

// SwitchCase represents a case in an if/switch on sum types
type SwitchCase struct {
	Condition string   // The condition (e.g., "w.IsLeft")
	Pattern   string   // "Left" or "Right" for Either, "Some" or "None" for Option
	VarExpr   string   // Expression to get the value (e.g., "w.Left")
	Body      []string // Body statements
}

// analyzeSumTypeSwitch analyzes if/else chains that act as pattern matching
func (t *Transpiler) analyzeSumTypeSwitch(ifStmt *ast.IfStmt, scrutinee string) (*MatchExpression, error) {
	match := &MatchExpression{
		Scrutinee: scrutinee,
	}

	// Analyze the if branch
	leftCase, err := t.analyzeIfBranch(ifStmt, "Left")
	if err != nil {
		return nil, err
	}
	if leftCase != nil {
		match.Cases = append(match.Cases, *leftCase)
	}

	// Analyze the else branch
	if ifStmt.Else != nil {
		if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			rightCase := &MatchCase{
				Pattern: "Right",
			}
			for _, stmt := range elseBlock.List {
				stmtStr, err := t.analyzeStatement(stmt)
				if err != nil {
					return nil, err
				}
				if stmtStr != "" {
					rightCase.BodyStmts = append(rightCase.BodyStmts, stmtStr)
				}
			}
			match.Cases = append(match.Cases, *rightCase)
		}
	}

	return match, nil
}

// analyzeIfBranch extracts pattern info from an if branch
func (t *Transpiler) analyzeIfBranch(ifStmt *ast.IfStmt, defaultPattern string) (*MatchCase, error) {
	mc := &MatchCase{
		Pattern: defaultPattern,
	}

	// Analyze the condition to determine pattern
	if binary, ok := ifStmt.Cond.(*ast.BinaryExpr); ok {
		// Check for w.IsLeft or similar
		if sel, ok := binary.X.(*ast.SelectorExpr); ok {
			fieldName := sel.Sel.Name
			if fieldName == "IsLeft" {
				mc.Pattern = "Left"
			} else if fieldName == "IsSome" {
				mc.Pattern = "Some"
			}
		}
	} else if sel, ok := ifStmt.Cond.(*ast.SelectorExpr); ok {
		fieldName := sel.Sel.Name
		if fieldName == "IsLeft" {
			mc.Pattern = "Left"
		} else if fieldName == "IsSome" {
			mc.Pattern = "Some"
		}
	}

	// Analyze the body
	for _, stmt := range ifStmt.Body.List {
		stmtStr, err := t.analyzeStatement(stmt)
		if err != nil {
			return nil, err
		}
		if stmtStr != "" {
			mc.BodyStmts = append(mc.BodyStmts, stmtStr)
		}
	}

	return mc, nil
}
