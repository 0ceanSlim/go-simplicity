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
		return "", fmt.Errorf("Go code validation failed: %w", err)
	}

	// Transpile to target format
	switch c.config.Target {
	case "simplicityhl":
		return c.transpiler.ToSimplicityHL(file, c.fset)
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
	case *ast.ForStmt, *ast.RangeStmt:
		v.errors = append(v.errors, "loops are not supported in Simplicity")
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
		// Check for make() calls
		if ident, ok := node.Fun.(*ast.Ident); ok && ident.Name == "make" {
			if len(node.Args) > 0 {
				switch node.Args[0].(type) {
				case *ast.MapType:
					v.errors = append(v.errors, "maps are not supported in Simplicity")
				case *ast.ChanType:
					v.errors = append(v.errors, "channels are not supported in Simplicity")
				case *ast.ArrayType:
					if arrType, ok := node.Args[0].(*ast.ArrayType); ok && arrType.Len == nil {
						v.errors = append(v.errors, "slices are not supported, use fixed-size arrays")
					}
				}
			}
		}
	case *ast.TypeSpec:
		// Check for interface types in type declarations
		if _, ok := node.Type.(*ast.InterfaceType); ok {
			v.errors = append(v.errors, "interfaces are not supported in Simplicity")
		}
	}
	return true
}
