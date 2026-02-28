package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// UnrolledLoop represents a for loop that has been unrolled
type UnrolledLoop struct {
	IndexVar   string
	Iterations int
	BodyStmts  [][]string // Body statements for each iteration
}

// evaluateIndexExpr handles array indexing like arr[i] or arr[0]
func (t *Transpiler) evaluateIndexExpr(expr *ast.IndexExpr) (string, error) {
	// Get the array expression
	arrayExpr, err := t.evaluateExpression(expr.X)
	if err != nil {
		return "", err
	}

	// Get the index
	indexExpr, err := t.evaluateExpression(expr.Index)
	if err != nil {
		return "", err
	}

	// Check if index is a literal number
	if idx, err := strconv.Atoi(indexExpr); err == nil {
		// Use bracket notation for literal indices
		return fmt.Sprintf("%s[%d]", arrayExpr, idx), nil
	}

	// For variable indices, use bracket notation
	return fmt.Sprintf("%s[%s]", arrayExpr, indexExpr), nil
}

// unrollForLoop converts a bounded for loop into unrolled statements
func (t *Transpiler) unrollForLoop(forStmt *ast.ForStmt) (*UnrolledLoop, error) {
	unrolled := &UnrolledLoop{}

	// Extract loop bounds from: for i := 0; i < N; i++
	// Init: i := 0
	if assignStmt, ok := forStmt.Init.(*ast.AssignStmt); ok {
		if len(assignStmt.Lhs) == 1 {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
				unrolled.IndexVar = ident.Name
			}
		}
	}

	// Cond: i < N
	if binaryExpr, ok := forStmt.Cond.(*ast.BinaryExpr); ok {
		if binaryExpr.Op == token.LSS {
			// Get the upper bound
			if lit, ok := binaryExpr.Y.(*ast.BasicLit); ok {
				if lit.Kind == token.INT {
					n, err := strconv.Atoi(lit.Value)
					if err != nil {
						return nil, fmt.Errorf("invalid loop bound: %s", lit.Value)
					}
					unrolled.Iterations = n
				}
			}
		}
	}

	if unrolled.Iterations == 0 {
		return nil, fmt.Errorf("could not determine loop bounds")
	}

	// Unroll the body for each iteration
	for i := 0; i < unrolled.Iterations; i++ {
		var iterStmts []string
		for _, stmt := range forStmt.Body.List {
			stmtStr, err := t.analyzeStatementWithIndex(stmt, unrolled.IndexVar, i)
			if err != nil {
				return nil, err
			}
			if stmtStr != "" {
				iterStmts = append(iterStmts, stmtStr)
			}
		}
		unrolled.BodyStmts = append(unrolled.BodyStmts, iterStmts)
	}

	return unrolled, nil
}

// analyzeStatementWithIndex analyzes a statement, substituting index variable
func (t *Transpiler) analyzeStatementWithIndex(stmt ast.Stmt, indexVar string, indexVal int) (string, error) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		return t.analyzeIfStmtWithIndex(s, indexVar, indexVal)
	case *ast.ExprStmt:
		return t.analyzeExprStmtWithIndex(s, indexVar, indexVal)
	case *ast.AssignStmt:
		return t.analyzeAssignStmtWithIndex(s, indexVar, indexVal)
	case *ast.IncDecStmt:
		return t.analyzeIncDecStmtWithIndex(s, indexVar, indexVal)
	default:
		return "", nil
	}
}

// analyzeIfStmtWithIndex handles if statements in unrolled loops
func (t *Transpiler) analyzeIfStmtWithIndex(ifStmt *ast.IfStmt, indexVar string, indexVal int) (string, error) {
	var sb strings.Builder

	// Check condition - likely checking Option.IsSome
	condStr, pattern := t.extractConditionWithIndex(ifStmt.Cond, indexVar, indexVal)
	if condStr == "" {
		return "", nil
	}

	// Generate match expression for Option
	sb.WriteString(fmt.Sprintf("match %s {\n", condStr))

	// Some branch (the if body)
	sb.WriteString("        Some(sig) => {\n")
	for _, bodyStmt := range ifStmt.Body.List {
		stmtStr, err := t.analyzeStatementWithIndex(bodyStmt, indexVar, indexVal)
		if err != nil {
			return "", err
		}
		if stmtStr != "" {
			sb.WriteString(fmt.Sprintf("            %s\n", stmtStr))
		}
	}
	if pattern == "Some" {
		sb.WriteString("            1\n") // Return 1 for valid signature
	}
	sb.WriteString("        },\n")

	// None branch
	sb.WriteString("        None => 0,\n")
	sb.WriteString("    }")

	return sb.String(), nil
}

// extractConditionWithIndex extracts condition info with index substitution
func (t *Transpiler) extractConditionWithIndex(cond ast.Expr, _ string, indexVal int) (string, string) {
	if sel, ok := cond.(*ast.SelectorExpr); ok {
		// sigs[i].IsSome -> witness::SIGS[indexVal]
		if idx, ok := sel.X.(*ast.IndexExpr); ok {
			arrayName, _ := t.evaluateExpression(idx.X)
			// Resolve to witness reference
			arrayName = t.resolveArrayRef(arrayName)
			fieldName := sel.Sel.Name

			if fieldName == "IsSome" {
				return fmt.Sprintf("%s[%d]", arrayName, indexVal), "Some"
			}
		}
	}
	return "", ""
}

// resolveArrayRef resolves an array name to its proper reference
func (t *Transpiler) resolveArrayRef(name string) string {
	snakeName := strings.ToUpper(t.toSnakeCase(name))
	for _, w := range t.witnessValues {
		if strings.ToUpper(w.Name) == snakeName {
			return fmt.Sprintf("witness::%s", snakeName)
		}
	}
	return t.toSnakeCase(name)
}

// analyzeExprStmtWithIndex handles expression statements with index substitution
func (t *Transpiler) analyzeExprStmtWithIndex(stmt *ast.ExprStmt, indexVar string, indexVal int) (string, error) {
	if callExpr, ok := stmt.X.(*ast.CallExpr); ok {
		return t.analyzeCallExprWithIndex(callExpr, indexVar, indexVal)
	}
	return "", nil
}

// analyzeCallExprWithIndex handles call expressions with index substitution
func (t *Transpiler) analyzeCallExprWithIndex(callExpr *ast.CallExpr, indexVar string, indexVal int) (string, error) {
	// Check for jet calls
	if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "jet" {
			jetName := sel.Sel.Name
			jetInfo, found := t.jetRegistry.Lookup(jetName)
			if !found {
				return "", fmt.Errorf("unknown jet: %s", jetName)
			}

			// Evaluate arguments with index substitution
			var args []string
			for _, arg := range callExpr.Args {
				argStr, err := t.evaluateExprWithIndex(arg, indexVar, indexVal)
				if err != nil {
					return "", err
				}
				args = append(args, argStr)
			}

			if len(args) == 0 {
				return fmt.Sprintf("jet::%s()", jetInfo.SimplicityName), nil
			}

			// Format BIP340Verify specially
			if jetName == "BIP340Verify" && len(args) == 3 {
				return fmt.Sprintf("jet::%s((%s, %s), %s)", jetInfo.SimplicityName, args[0], args[1], args[2]), nil
			}

			return fmt.Sprintf("jet::%s(%s)", jetInfo.SimplicityName, strings.Join(args, ", ")), nil
		}
	}
	return "", nil
}

// evaluateExprWithIndex evaluates an expression with index substitution
func (t *Transpiler) evaluateExprWithIndex(expr ast.Expr, indexVar string, indexVal int) (string, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == indexVar {
			return strconv.Itoa(indexVal), nil
		}
		return t.evaluateJetArg(e)
	case *ast.IndexExpr:
		// array[i] with i substituted
		arrayExpr, err := t.evaluateExpression(e.X)
		if err != nil {
			return "", err
		}
		arrayExpr = t.resolveArrayRef(arrayExpr)

		// Check if index is the loop variable
		if ident, ok := e.Index.(*ast.Ident); ok && ident.Name == indexVar {
			return fmt.Sprintf("%s[%d]", arrayExpr, indexVal), nil
		}

		indexExpr, err := t.evaluateExpression(e.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s[%s]", arrayExpr, indexExpr), nil
	case *ast.SelectorExpr:
		// Handle sigs[i].Value -> witness::SIGS[indexVal]
		if idx, ok := e.X.(*ast.IndexExpr); ok {
			arrayName, _ := t.evaluateExpression(idx.X)
			arrayName = t.resolveArrayRef(arrayName)

			// Check if index is the loop variable
			if ident, ok := idx.Index.(*ast.Ident); ok && ident.Name == indexVar {
				// For .Value on Option, just return the indexed element (inside match it becomes 'sig')
				if e.Sel.Name == "Value" {
					return "sig", nil // Inside match Some(sig) branch
				}
				return fmt.Sprintf("%s[%d].%s", arrayName, indexVal, t.toSnakeCase(e.Sel.Name)), nil
			}
		}
		return t.evaluateJetArg(e)
	case *ast.CallExpr:
		return t.analyzeCallExprWithIndex(e, indexVar, indexVal)
	default:
		return t.evaluateExpression(expr)
	}
}

// analyzeAssignStmtWithIndex handles assignments with index substitution
func (t *Transpiler) analyzeAssignStmtWithIndex(stmt *ast.AssignStmt, indexVar string, indexVal int) (string, error) {
	if len(stmt.Lhs) == 1 && len(stmt.Rhs) == 1 {
		lhs := ""
		if ident, ok := stmt.Lhs[0].(*ast.Ident); ok {
			lhs = t.toSnakeCase(ident.Name)
		}

		rhs, err := t.evaluateExprWithIndex(stmt.Rhs[0], indexVar, indexVal)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("let %s = %s;", lhs, rhs), nil
	}
	return "", nil
}

// analyzeIncDecStmtWithIndex handles increment/decrement statements
func (t *Transpiler) analyzeIncDecStmtWithIndex(stmt *ast.IncDecStmt, _ string, _ int) (string, error) {
	// validCount++ becomes part of accumulation logic
	if ident, ok := stmt.X.(*ast.Ident); ok {
		varName := t.toSnakeCase(ident.Name)
		if stmt.Tok == token.INC {
			// This will be handled by counter accumulation
			return fmt.Sprintf("// %s++ (accumulated)", varName), nil
		}
	}
	return "", nil
}
