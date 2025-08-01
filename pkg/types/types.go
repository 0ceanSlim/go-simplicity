package types

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// TypeMapper maps Go types to Simplicity types
type TypeMapper struct {
	builtinTypes map[string]string
}

// NewTypeMapper creates a new type mapper
func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		builtinTypes: map[string]string{
			// Go basic types -> Simplicity types
			"bool":   "bool",
			"uint8":  "u8",
			"uint16": "u16",
			"uint32": "u32",
			"uint64": "u64",
			"byte":   "u8",

			// Bitcoin-specific types (when imported)
			"Hash":      "u256",
			"Address":   "u256",
			"Pubkey":    "u256",
			"Signature": "[u8; 64]",
		},
	}
}

// MapGoType converts a Go type to its Simplicity equivalent
func (tm *TypeMapper) MapGoType(goType ast.Expr) (string, error) {
	switch t := goType.(type) {
	case *ast.Ident:
		return tm.mapIdentType(t)
	case *ast.ArrayType:
		return tm.mapArrayType(t)
	case *ast.StructType:
		return tm.mapStructType(t)
	case *ast.SelectorExpr:
		return tm.mapSelectorType(t)
	default:
		return "", fmt.Errorf("unsupported Go type: %T", goType)
	}
}

func (tm *TypeMapper) mapIdentType(ident *ast.Ident) (string, error) {
	if simplicityType, exists := tm.builtinTypes[ident.Name]; exists {
		return simplicityType, nil
	}

	// For custom types, return as-is (they should be defined elsewhere)
	return ident.Name, nil
}

func (tm *TypeMapper) mapArrayType(arrayType *ast.ArrayType) (string, error) {
	// Get element type
	elemType, err := tm.MapGoType(arrayType.Elt)
	if err != nil {
		return "", fmt.Errorf("failed to map array element type: %w", err)
	}

	// Get array length
	if arrayType.Len == nil {
		return "", fmt.Errorf("slices are not supported, use fixed-size arrays")
	}

	length, err := tm.evaluateArrayLength(arrayType.Len)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate array length: %w", err)
	}

	return fmt.Sprintf("[%s; %d]", elemType, length), nil
}

func (tm *TypeMapper) mapStructType(structType *ast.StructType) (string, error) {
	// Simplicity doesn't have structs, so we convert to tuples
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		return "()", nil
	}

	var fieldTypes []string
	for _, field := range structType.Fields.List {
		fieldType, err := tm.MapGoType(field.Type)
		if err != nil {
			return "", fmt.Errorf("failed to map struct field type: %w", err)
		}

		// If field has multiple names, add the type for each
		if len(field.Names) == 0 {
			// Anonymous field
			fieldTypes = append(fieldTypes, fieldType)
		} else {
			for range field.Names {
				fieldTypes = append(fieldTypes, fieldType)
			}
		}
	}

	if len(fieldTypes) == 1 {
		return fmt.Sprintf("(%s,)", fieldTypes[0]), nil // Single-element tuple
	}

	return fmt.Sprintf("(%s)", strings.Join(fieldTypes, ", ")), nil
}

func (tm *TypeMapper) mapSelectorType(sel *ast.SelectorExpr) (string, error) {
	// Handle qualified types like bitcoin.Hash
	if ident, ok := sel.X.(*ast.Ident); ok {
		qualifiedName := fmt.Sprintf("%s.%s", ident.Name, sel.Sel.Name)

		// Handle bitcoin package types
		if ident.Name == "bitcoin" {
			switch sel.Sel.Name {
			case "Hash":
				return "u256", nil
			case "Address":
				return "u256", nil
			case "Pubkey":
				return "u256", nil
			case "Signature":
				return "[u8; 64]", nil
			case "Amount":
				return "u64", nil
			default:
				return "", fmt.Errorf("unsupported bitcoin type: %s", sel.Sel.Name)
			}
		}

		return "", fmt.Errorf("unsupported qualified type: %s", qualifiedName)
	}

	return "", fmt.Errorf("unsupported selector expression")
}

func (tm *TypeMapper) evaluateArrayLength(expr ast.Expr) (int, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.INT {
			return strconv.Atoi(e.Value)
		}
	case *ast.Ident:
		// For now, we don't support const evaluation
		return 0, fmt.Errorf("array length must be a literal integer")
	}

	return 0, fmt.Errorf("unsupported array length expression: %T", expr)
}

// IsSupported checks if a Go type is supported in Simplicity
func (tm *TypeMapper) IsSupported(goType ast.Expr) bool {
	_, err := tm.MapGoType(goType)
	return err == nil
}

// GetBitSize returns the bit size for a Simplicity type
func (tm *TypeMapper) GetBitSize(simplicityType string) int {
	switch simplicityType {
	case "bool", "u1":
		return 1
	case "u2":
		return 2
	case "u4":
		return 4
	case "u8":
		return 8
	case "u16":
		return 16
	case "u32":
		return 32
	case "u64":
		return 64
	case "u128":
		return 128
	case "u256":
		return 256
	case "()":
		return 0
	default:
		// For arrays like [u8; 32], we need to parse
		if strings.HasPrefix(simplicityType, "[") && strings.Contains(simplicityType, ";") {
			// Extract element type and count
			parts := strings.Split(simplicityType[1:len(simplicityType)-1], ";")
			if len(parts) == 2 {
				elemType := strings.TrimSpace(parts[0])
				count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err == nil {
					return tm.GetBitSize(elemType) * count
				}
			}
		}
		return 0 // Unknown
	}
}

// SupportedTypes returns a list of all supported Go types
func (tm *TypeMapper) SupportedTypes() []string {
	var types []string
	for goType := range tm.builtinTypes {
		types = append(types, goType)
	}
	return types
}
