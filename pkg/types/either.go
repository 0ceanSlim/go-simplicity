package types

import (
	"fmt"
	"strings"
)

// SumType represents Either<L, R> or Option<T> types
type SumType struct {
	Kind      SumTypeKind
	LeftType  string // For Either: left type, For Option: the inner type
	RightType string // For Either: right type, For Option: unused
}

// SumTypeKind identifies whether a SumType is Either or Option.
type SumTypeKind int

const (
	// SumTypeEither represents an Either<L, R> type.
	SumTypeEither SumTypeKind = iota
	// SumTypeOption represents an Option<T> type.
	SumTypeOption
)

// ParseSumType parses a string like "Either<u256, [u8; 64]>" or "Option<[u8; 64]>"
func ParseSumType(typeStr string) (*SumType, error) {
	typeStr = strings.TrimSpace(typeStr)

	// Check for Either<L, R>
	if strings.HasPrefix(typeStr, "Either<") && strings.HasSuffix(typeStr, ">") {
		inner := typeStr[7 : len(typeStr)-1] // Remove "Either<" and ">"
		left, right, err := splitTypeParams(inner)
		if err != nil {
			return nil, fmt.Errorf("invalid Either type: %w", err)
		}
		return &SumType{
			Kind:      SumTypeEither,
			LeftType:  left,
			RightType: right,
		}, nil
	}

	// Check for Option<T>
	if strings.HasPrefix(typeStr, "Option<") && strings.HasSuffix(typeStr, ">") {
		inner := typeStr[7 : len(typeStr)-1] // Remove "Option<" and ">"
		return &SumType{
			Kind:     SumTypeOption,
			LeftType: strings.TrimSpace(inner),
		}, nil
	}

	return nil, fmt.Errorf("not a sum type: %s", typeStr)
}

// splitTypeParams splits "A, B" handling nested brackets
func splitTypeParams(params string) (string, string, error) {
	depth := 0
	for i, c := range params {
		switch c {
		case '<', '(', '[':
			depth++
		case '>', ')', ']':
			depth--
		case ',':
			if depth == 0 {
				left := strings.TrimSpace(params[:i])
				right := strings.TrimSpace(params[i+1:])
				return left, right, nil
			}
		}
	}
	return "", "", fmt.Errorf("could not split type parameters: %s", params)
}

// ToSimplicityHL returns the SimplicityHL representation of the sum type
func (st *SumType) ToSimplicityHL() string {
	switch st.Kind {
	case SumTypeEither:
		return fmt.Sprintf("Either<%s, %s>", st.LeftType, st.RightType)
	case SumTypeOption:
		return fmt.Sprintf("Option<%s>", st.LeftType)
	default:
		return "unknown"
	}
}

// IsEither returns true if this is an Either type
func (st *SumType) IsEither() bool {
	return st.Kind == SumTypeEither
}

// IsOption returns true if this is an Option type
func (st *SumType) IsOption() bool {
	return st.Kind == SumTypeOption
}

// IsSumType checks if a type string represents a sum type
func IsSumType(typeStr string) bool {
	return strings.HasPrefix(typeStr, "Either<") || strings.HasPrefix(typeStr, "Option<")
}

// MatchArm represents a single arm of a match expression
type MatchArm struct {
	Pattern string // "Left(data)" or "Right(sig)" or "Some(val)" or "None"
	VarName string // The bound variable name (data, sig, val, etc.)
	VarType string // The type of the bound variable
	Body    string // The body of the match arm
}

// MatchExpr represents a complete match expression
type MatchExpr struct {
	Scrutinee string     // The expression being matched (e.g., "witness::DATA")
	Arms      []MatchArm // The match arms
}

// ToSimplicityHL generates SimplicityHL code for the match expression
func (m *MatchExpr) ToSimplicityHL(indent string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("match %s {\n", m.Scrutinee))

	for i, arm := range m.Arms {
		sb.WriteString(fmt.Sprintf("%s    %s => {\n", indent, arm.Pattern))

		// Add the body lines with proper indentation
		bodyLines := strings.Split(arm.Body, "\n")
		for _, line := range bodyLines {
			if strings.TrimSpace(line) != "" {
				sb.WriteString(fmt.Sprintf("%s        %s\n", indent, strings.TrimSpace(line)))
			}
		}

		sb.WriteString(fmt.Sprintf("%s    }", indent))
		if i < len(m.Arms)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%s}", indent))
	return sb.String()
}

// TupleType represents a tuple like (u256, [u8; 64])
type TupleType struct {
	Elements []string
}

// ParseTupleType parses a string like "(u256, [u8; 64])"
func ParseTupleType(typeStr string) (*TupleType, error) {
	typeStr = strings.TrimSpace(typeStr)

	if !strings.HasPrefix(typeStr, "(") || !strings.HasSuffix(typeStr, ")") {
		return nil, fmt.Errorf("not a tuple type: %s", typeStr)
	}

	inner := typeStr[1 : len(typeStr)-1] // Remove "(" and ")"
	if inner == "" {
		return &TupleType{Elements: []string{}}, nil
	}

	elements, err := splitTupleElements(inner)
	if err != nil {
		return nil, err
	}

	return &TupleType{Elements: elements}, nil
}

// splitTupleElements splits tuple elements handling nested brackets
func splitTupleElements(params string) ([]string, error) {
	var elements []string
	depth := 0
	start := 0

	for i, c := range params {
		switch c {
		case '<', '(', '[':
			depth++
		case '>', ')', ']':
			depth--
		case ',':
			if depth == 0 {
				elements = append(elements, strings.TrimSpace(params[start:i]))
				start = i + 1
			}
		}
	}

	// Add the last element
	last := strings.TrimSpace(params[start:])
	if last != "" {
		elements = append(elements, last)
	}

	return elements, nil
}

// ToSimplicityHL returns the SimplicityHL representation
func (tt *TupleType) ToSimplicityHL() string {
	if len(tt.Elements) == 0 {
		return "()"
	}
	if len(tt.Elements) == 1 {
		return fmt.Sprintf("(%s,)", tt.Elements[0])
	}
	return fmt.Sprintf("(%s)", strings.Join(tt.Elements, ", "))
}
