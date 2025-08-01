package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/0ceanslim/go-simplicity/pkg/types"
)

// Transpiler converts Go AST to SimplicityHL
type Transpiler struct {
	typeMapper *types.TypeMapper
	output     strings.Builder
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{
		typeMapper: types.NewTypeMapper(),
	}
}

// ToSimplicityHL transpiles Go AST to SimplicityHL code
func (t *Transpiler) ToSimplicityHL(file *ast.File, fset *token.FileSet) (string, error) {
	t.output.Reset()

	// Add file header comment
	t.writeLine("// Generated from Go source by go-simplicity compiler")
	t.writeLine("")

	// Process package declaration (skip for now)

	// Process imports (handle simplicity-specific imports)
	for _, imp := range file.Imports {
		if err := t.processImport(imp); err != nil {
			return "", err
		}
	}

	// Process type declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if err := t.processGenDecl(d); err != nil {
				return "", err
			}
		case *ast.FuncDecl:
			if err := t.processFuncDecl(d); err != nil {
				return "", err
			}
		}
	}

	return t.output.String(), nil
}

func (t *Transpiler) processImport(imp *ast.ImportSpec) error {
	if imp.Path.Value == `"simplicity/bitcoin"` {
		// Handle bitcoin-specific imports
		t.writeLine("// Bitcoin primitives available via jets")
		return nil
	}
	return nil
}

func (t *Transpiler) processGenDecl(decl *ast.GenDecl) error {
	switch decl.Tok {
	case token.TYPE:
		return t.processTypeDecl(decl)
	case token.VAR:
		return t.processVarDecl(decl)
	case token.CONST:
		return t.processConstDecl(decl)
	}
	return nil
}

func (t *Transpiler) processTypeDecl(decl *ast.GenDecl) error {
	for _, spec := range decl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			simplicityType, err := t.typeMapper.MapGoType(typeSpec.Type)
			if err != nil {
				return fmt.Errorf("failed to map type %s: %w", typeSpec.Name.Name, err)
			}
			t.writeLine(fmt.Sprintf("type %s = %s;", typeSpec.Name.Name, simplicityType))
		}
	}
	return nil
}

func (t *Transpiler) processVarDecl(decl *ast.GenDecl) error {
	// Variables in Simplicity are immutable, so we generate let statements
	for _, spec := range decl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				simplicityType, err := t.typeMapper.MapGoType(valueSpec.Type)
				if err != nil {
					return fmt.Errorf("failed to map type for variable %s: %w", name.Name, err)
				}

				var value string
				if i < len(valueSpec.Values) {
					value, err = t.transpileExpr(valueSpec.Values[i])
					if err != nil {
						return fmt.Errorf("failed to transpile value for %s: %w", name.Name, err)
					}
				} else {
					value = t.getDefaultValue(simplicityType)
				}

				t.writeLine(fmt.Sprintf("let %s: %s = %s;", name.Name, simplicityType, value))
			}
		}
	}
	return nil
}

func (t *Transpiler) processConstDecl(decl *ast.GenDecl) error {
	// Constants are similar to variables in SimplicityHL
	return t.processVarDecl(decl)
}

func (t *Transpiler) processFuncDecl(decl *ast.FuncDecl) error {
	if decl.Name.Name == "main" {
		return t.processMainFunc(decl)
	}

	// Build function signature
	sig, err := t.buildFuncSignature(decl)
	if err != nil {
		return fmt.Errorf("failed to build signature for function %s: %w", decl.Name.Name, err)
	}

	t.writeLine(fmt.Sprintf("fn %s%s {", decl.Name.Name, sig))

	// Process function body
	if decl.Body != nil {
		if err := t.processBlockStmt(decl.Body, 1); err != nil {
			return fmt.Errorf("failed to process body of function %s: %w", decl.Name.Name, err)
		}
	}

	t.writeLine("}")
	t.writeLine("")

	return nil
}

func (t *Transpiler) processMainFunc(decl *ast.FuncDecl) error {
	t.writeLine("fn main() {")

	if decl.Body != nil {
		if err := t.processBlockStmt(decl.Body, 1); err != nil {
			return fmt.Errorf("failed to process main function body: %w", err)
		}
	}

	t.writeLine("}")
	return nil
}

func (t *Transpiler) buildFuncSignature(decl *ast.FuncDecl) (string, error) {
	var params []string
	var returnType string = "()"

	// Process parameters
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			simplicityType, err := t.typeMapper.MapGoType(field.Type)
			if err != nil {
				return "", err
			}

			for _, name := range field.Names {
				params = append(params, fmt.Sprintf("%s: %s", name.Name, simplicityType))
			}
		}
	}

	// Process return type
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		if len(decl.Type.Results.List) == 1 {
			rt, err := t.typeMapper.MapGoType(decl.Type.Results.List[0].Type)
			if err != nil {
				return "", err
			}
			returnType = rt
		} else {
			// Multiple return values become tuples
			var returnTypes []string
			for _, field := range decl.Type.Results.List {
				rt, err := t.typeMapper.MapGoType(field.Type)
				if err != nil {
					return "", err
				}
				returnTypes = append(returnTypes, rt)
			}
			returnType = fmt.Sprintf("(%s)", strings.Join(returnTypes, ", "))
		}
	}

	paramStr := strings.Join(params, ", ")
	return fmt.Sprintf("(%s) -> %s", paramStr, returnType), nil
}

func (t *Transpiler) processBlockStmt(block *ast.BlockStmt, indent int) error {
	for _, stmt := range block.List {
		if err := t.processStmt(stmt, indent); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transpiler) processStmt(stmt ast.Stmt, indent int) error {
	switch s := stmt.(type) {
	case *ast.DeclStmt:
		return t.processDeclStmt(s, indent)
	case *ast.ExprStmt:
		return t.processExprStmt(s, indent)
	case *ast.AssignStmt:
		return t.processAssignStmt(s, indent)
	case *ast.ReturnStmt:
		return t.processReturnStmt(s, indent)
	case *ast.IfStmt:
		return t.processIfStmt(s, indent)
	}
	return nil
}

func (t *Transpiler) processDeclStmt(stmt *ast.DeclStmt, indent int) error {
	// Handle local declarations
	return nil
}

func (t *Transpiler) processExprStmt(stmt *ast.ExprStmt, indent int) error {
	expr, err := t.transpileExpr(stmt.X)
	if err != nil {
		return err
	}
	t.writeIndented(fmt.Sprintf("%s;", expr), indent)
	return nil
}

func (t *Transpiler) processAssignStmt(stmt *ast.AssignStmt, indent int) error {
	// In SimplicityHL, assignments are let statements
	if len(stmt.Lhs) == 1 && len(stmt.Rhs) == 1 {
		if ident, ok := stmt.Lhs[0].(*ast.Ident); ok {
			value, err := t.transpileExpr(stmt.Rhs[0])
			if err != nil {
				return err
			}
			// For now, we'll assume the type can be inferred
			t.writeIndented(fmt.Sprintf("let %s = %s;", ident.Name, value), indent)
		}
	}
	return nil
}

func (t *Transpiler) processReturnStmt(stmt *ast.ReturnStmt, indent int) error {
	if len(stmt.Results) == 0 {
		t.writeIndented("()", indent)
	} else if len(stmt.Results) == 1 {
		expr, err := t.transpileExpr(stmt.Results[0])
		if err != nil {
			return err
		}
		t.writeIndented(expr, indent)
	} else {
		// Multiple returns become tuples
		var exprs []string
		for _, result := range stmt.Results {
			expr, err := t.transpileExpr(result)
			if err != nil {
				return err
			}
			exprs = append(exprs, expr)
		}
		t.writeIndented(fmt.Sprintf("(%s)", strings.Join(exprs, ", ")), indent)
	}
	return nil
}

func (t *Transpiler) processIfStmt(stmt *ast.IfStmt, indent int) error {
	cond, err := t.transpileExpr(stmt.Cond)
	if err != nil {
		return err
	}

	t.writeIndented(fmt.Sprintf("match %s {", cond), indent)
	t.writeIndented("true => {", indent+1)

	if stmt.Body != nil {
		if err := t.processBlockStmt(stmt.Body, indent+2); err != nil {
			return err
		}
	}

	t.writeIndented("},", indent+1)
	t.writeIndented("false => {", indent+1)

	if stmt.Else != nil {
		if elseBlock, ok := stmt.Else.(*ast.BlockStmt); ok {
			if err := t.processBlockStmt(elseBlock, indent+2); err != nil {
				return err
			}
		}
	}

	t.writeIndented("},", indent+1)
	t.writeIndented("}", indent)

	return nil
}

func (t *Transpiler) transpileExpr(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value, nil
	case *ast.Ident:
		return e.Name, nil
	case *ast.BinaryExpr:
		return t.transpileBinaryExpr(e)
	case *ast.CallExpr:
		return t.transpileCallExpr(e)
	case *ast.UnaryExpr:
		return t.transpileUnaryExpr(e)
	}
	return "", fmt.Errorf("unsupported expression type: %T", expr)
}

func (t *Transpiler) transpileBinaryExpr(expr *ast.BinaryExpr) (string, error) {
	left, err := t.transpileExpr(expr.X)
	if err != nil {
		return "", err
	}

	right, err := t.transpileExpr(expr.Y)
	if err != nil {
		return "", err
	}

	var op string
	switch expr.Op {
	case token.ADD:
		op = "+"
	case token.SUB:
		op = "-"
	case token.MUL:
		op = "*"
	case token.QUO:
		op = "/"
	case token.EQL:
		op = "=="
	case token.NEQ:
		op = "!="
	case token.LSS:
		op = "<"
	case token.GTR:
		op = ">"
	case token.LEQ:
		op = "<="
	case token.GEQ:
		op = ">="
	case token.LAND:
		op = "&&"
	case token.LOR:
		op = "||"
	default:
		return "", fmt.Errorf("unsupported binary operator: %s", expr.Op)
	}

	return fmt.Sprintf("(%s %s %s)", left, op, right), nil
}

func (t *Transpiler) transpileCallExpr(expr *ast.CallExpr) (string, error) {
	// Handle function calls
	if ident, ok := expr.Fun.(*ast.Ident); ok {
		var args []string
		for _, arg := range expr.Args {
			argStr, err := t.transpileExpr(arg)
			if err != nil {
				return "", err
			}
			args = append(args, argStr)
		}
		return fmt.Sprintf("%s(%s)", ident.Name, strings.Join(args, ", ")), nil
	}
	return "", fmt.Errorf("unsupported call expression")
}

func (t *Transpiler) transpileUnaryExpr(expr *ast.UnaryExpr) (string, error) {
	operand, err := t.transpileExpr(expr.X)
	if err != nil {
		return "", err
	}

	switch expr.Op {
	case token.NOT:
		return fmt.Sprintf("!%s", operand), nil
	case token.SUB:
		return fmt.Sprintf("-%s", operand), nil
	}

	return "", fmt.Errorf("unsupported unary operator: %s", expr.Op)
}

func (t *Transpiler) getDefaultValue(typeName string) string {
	switch typeName {
	case "bool":
		return "false"
	case "u8", "u16", "u32", "u64", "u128", "u256":
		return "0"
	default:
		return "0" // fallback
	}
}

func (t *Transpiler) writeLine(line string) {
	t.output.WriteString(line)
	t.output.WriteString("\n")
}

func (t *Transpiler) writeIndented(line string, indent int) {
	for i := 0; i < indent; i++ {
		t.output.WriteString("    ")
	}
	t.output.WriteString(line)
	t.output.WriteString("\n")
}
