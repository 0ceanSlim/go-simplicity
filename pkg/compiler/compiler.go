// Package compiler validates Go source and orchestrates transpilation to
// SimplicityHL or raw Simplicity bytecode.
package compiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/0ceanslim/go-simplicity/pkg/transpiler"
)

// Config holds compiler configuration
type Config struct {
	Target string // "simplicityhl" or "simplicity"
	Debug  bool
}

// Compiler represents the Go to Simplicity compiler
type Compiler struct {
	config     Config
	fset       *token.FileSet
	transpiler *transpiler.Transpiler
}

// New creates a new compiler instance
func New(config Config) *Compiler {
	return &Compiler{
		config:     config,
		fset:       token.NewFileSet(),
		transpiler: transpiler.New(),
	}
}

// Compile compiles Go source code to the target format
func (c *Compiler) Compile(source, filename string) (string, error) {
	// Parse Go source
	file, err := parser.ParseFile(c.fset, filename, source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse Go source: %w", err)
	}

	if c.config.Debug {
		fmt.Printf("Parsed AST for %s\n", filename)
		ast.Print(c.fset, file)
	}

	// Validate that the Go code is compatible with Simplicity
	if err := c.validateGoCode(file); err != nil {
		return "", fmt.Errorf("go code validation failed: %w", err)
	}

	// Transpile to target format
	switch c.config.Target {
	case "simplicityhl":
		return c.transpiler.ToSimplicityHL(file)
	case "simplicity":
		return "", fmt.Errorf("direct Simplicity compilation not yet implemented")
	default:
		return "", fmt.Errorf("unsupported target: %s", c.config.Target)
	}
}

// validateGoCode checks if the Go code uses only supported features
func (c *Compiler) validateGoCode(file *ast.File) error {
	validator := &goValidator{
		errors: []string{},
	}

	ast.Inspect(file, validator.visit)

	if len(validator.errors) > 0 {
		return fmt.Errorf("unsupported Go features detected:\n%s", strings.Join(validator.errors, "\n"))
	}

	return nil
}

type goValidator struct {
	errors []string
}

func (v *goValidator) visit(n ast.Node) bool {
	switch node := n.(type) {
	case *ast.ForStmt:
		if !v.isBoundedForLoop(node) {
			v.errors = append(v.errors, "unbounded loops are not supported in Simplicity (use bounded for loops like 'for i := 0; i < N; i++')")
			return false
		}
		return true
	case *ast.RangeStmt:
		v.errors = append(v.errors, "range loops are not supported in Simplicity")
		return false
	case *ast.GoStmt:
		v.errors = append(v.errors, "goroutines are not supported in Simplicity")
		return false
	case *ast.ChanType:
		v.errors = append(v.errors, "channels are not supported in Simplicity")
		return false
	case *ast.InterfaceType:
		v.errors = append(v.errors, "interfaces are not supported in Simplicity")
		return false
	case *ast.ArrayType:
		if node.Len == nil {
			v.errors = append(v.errors, "slices are not supported, use fixed-size arrays")
			return false
		}
	case *ast.MapType:
		v.errors = append(v.errors, "maps are not supported in Simplicity")
		return false
	case *ast.CallExpr:
		return v.visitCallExpr(node)
	case *ast.TypeSpec:
		if _, ok := node.Type.(*ast.InterfaceType); ok {
			v.errors = append(v.errors, "interfaces are not supported in Simplicity")
		}
	}
	return true
}

// visitCallExpr validates a call expression node.
// jet.X() calls are always allowed; make() calls are checked for unsupported types.
func (v *goValidator) visitCallExpr(node *ast.CallExpr) bool {
	if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "jet" {
			return true
		}
	}
	if ident, ok := node.Fun.(*ast.Ident); ok && ident.Name == "make" {
		v.validateMakeArgs(node.Args)
	}
	return true
}

// validateMakeArgs checks the first argument of a make() call for unsupported types.
func (v *goValidator) validateMakeArgs(args []ast.Expr) {
	if len(args) == 0 {
		return
	}
	switch t := args[0].(type) {
	case *ast.MapType:
		v.errors = append(v.errors, "maps are not supported in Simplicity")
	case *ast.ChanType:
		v.errors = append(v.errors, "channels are not supported in Simplicity")
	case *ast.ArrayType:
		if t.Len == nil {
			v.errors = append(v.errors, "slices are not supported, use fixed-size arrays")
		}
	}
}

// isBoundedForLoop checks if a for loop has compile-time bounds.
// Accepted pattern: for i := 0; i < N; i++ where N is an integer literal.
func (v *goValidator) isBoundedForLoop(forStmt *ast.ForStmt) bool {
	if forStmt.Init == nil || forStmt.Cond == nil || forStmt.Post == nil {
		return false
	}
	return validLoopInit(forStmt.Init) && validLoopCond(forStmt.Cond) && validLoopPost(forStmt.Post)
}

// validLoopInit checks that the loop initialiser is: i := 0
func validLoopInit(init ast.Stmt) bool {
	assign, ok := init.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return false
	}
	lit, ok := assign.Rhs[0].(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "0"
}

// validLoopCond checks that the loop condition is: i < N or i <= N where N is a literal.
func validLoopCond(cond ast.Expr) bool {
	binary, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if binary.Op != token.LSS && binary.Op != token.LEQ {
		return false
	}
	_, ok = binary.Y.(*ast.BasicLit)
	return ok
}

// validLoopPost checks that the loop post statement is: i++
func validLoopPost(post ast.Stmt) bool {
	inc, ok := post.(*ast.IncDecStmt)
	return ok && inc.Tok == token.INC
}
