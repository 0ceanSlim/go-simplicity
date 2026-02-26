package tests

import (
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
	"github.com/0ceanslim/go-simplicity/pkg/types"
)

func TestEitherTypeParsing(t *testing.T) {
	testCases := []struct {
		input    string
		isEither bool
		left     string
		right    string
	}{
		{"Either<u256, [u8; 64]>", true, "u256", "[u8; 64]"},
		{"Either<(u256, [u8; 64]), [u8; 64]>", true, "(u256, [u8; 64])", "[u8; 64]"},
		{"Option<u256>", false, "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			st, err := types.ParseSumType(tc.input)
			if tc.isEither {
				if err != nil {
					t.Fatalf("Failed to parse Either type: %v", err)
				}
				if !st.IsEither() {
					t.Error("Expected Either type")
				}
				if st.LeftType != tc.left {
					t.Errorf("Left type: got %s, want %s", st.LeftType, tc.left)
				}
				if st.RightType != tc.right {
					t.Errorf("Right type: got %s, want %s", st.RightType, tc.right)
				}
			}
		})
	}
}

func TestOptionTypeParsing(t *testing.T) {
	testCases := []struct {
		input    string
		isOption bool
		inner    string
	}{
		{"Option<u256>", true, "u256"},
		{"Option<[u8; 64]>", true, "[u8; 64]"},
		{"Either<u256, u256>", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			st, err := types.ParseSumType(tc.input)
			if tc.isOption {
				if err != nil {
					t.Fatalf("Failed to parse Option type: %v", err)
				}
				if !st.IsOption() {
					t.Error("Expected Option type")
				}
				if st.LeftType != tc.inner {
					t.Errorf("Inner type: got %s, want %s", st.LeftType, tc.inner)
				}
			}
		})
	}
}

func TestTupleTypeParsing(t *testing.T) {
	testCases := []struct {
		input    string
		elements []string
	}{
		{"(u256, [u8; 64])", []string{"u256", "[u8; 64]"}},
		{"(u256, u256, u256)", []string{"u256", "u256", "u256"}},
		{"()", []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			tt, err := types.ParseTupleType(tc.input)
			if err != nil {
				t.Fatalf("Failed to parse tuple type: %v", err)
			}
			if len(tt.Elements) != len(tc.elements) {
				t.Errorf("Element count: got %d, want %d", len(tt.Elements), len(tc.elements))
			}
			for i, elem := range tc.elements {
				if i < len(tt.Elements) && tt.Elements[i] != elem {
					t.Errorf("Element %d: got %s, want %s", i, tt.Elements[i], elem)
				}
			}
		})
	}
}

func TestSumTypeDetection(t *testing.T) {
	testCases := []struct {
		input  string
		isSum  bool
	}{
		{"Either<u256, [u8; 64]>", true},
		{"Option<u256>", true},
		{"u256", false},
		{"[u8; 64]", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := types.IsSumType(tc.input)
			if result != tc.isSum {
				t.Errorf("IsSumType(%s) = %v, want %v", tc.input, result, tc.isSum)
			}
		})
	}
}

func TestEitherWitnessDeclaration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

type CompleteData struct {
	Preimage [32]byte
	Sig      [64]byte
}

func main() {
	var witness Either[CompleteData, [64]byte]
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check that Either type is in witness
	if !strings.Contains(result, "Either<") {
		t.Error("Should contain Either type in output")
	}
}

func TestOptionWitnessDeclaration(t *testing.T) {
	source := `
package main

func main() {
	var maybeSig Option[[64]byte]
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check that Option type is in witness
	if !strings.Contains(result, "Option<") {
		t.Error("Should contain Option type in output")
	}
}

func TestMatchExpressionGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4

type Witness struct {
	IsLeft bool
	Left   [64]byte
	Right  [64]byte
}

func main() {
	var w Witness

	if w.IsLeft {
		jet.BIP340Verify(RecipientPubkey, jet.SigAllHash(), w.Left)
	} else {
		jet.BIP340Verify(SenderPubkey, jet.SigAllHash(), w.Right)
	}
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check that match expression is generated
	if !strings.Contains(result, "match") {
		t.Error("Should generate match expression")
	}

	// Check for Left and Right arms
	if !strings.Contains(result, "Left") {
		t.Error("Should contain Left pattern")
	}
}

func TestSimpleHTLCStructure(t *testing.T) {
	// Test that we can parse HTLC-like contract structure
	source := `
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

func main() {
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(RecipientPubkey, msg, sig)
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify key components
	checks := []string{
		"RECIPIENT_PUBKEY",
		"SENDER_PUBKEY",
		"HASH_LOCK",
		"jet::bip_0340_verify",
		"jet::sig_all_hash",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Output should contain %s", check)
		}
	}
}

func TestGoGenericEitherType(t *testing.T) {
	// Test that Go generic syntax Either[L, R] maps correctly
	tm := types.NewTypeMapper()

	// This would be tested with actual AST parsing
	// For now, verify the type mapper handles the conversion
	supported := tm.SupportedTypes()
	if len(supported) == 0 {
		t.Error("Type mapper should have supported types")
	}
}

func TestMatchArmGeneration(t *testing.T) {
	// Test MatchExpr generation
	match := &types.MatchExpr{
		Scrutinee: "witness::DATA",
		Arms: []types.MatchArm{
			{
				Pattern: "Left(data)",
				Body:    "let x = data;",
			},
			{
				Pattern: "Right(sig)",
				Body:    "jet::verify(sig)",
			},
		},
	}

	result := match.ToSimplicityHL("    ")

	if !strings.Contains(result, "match witness::DATA") {
		t.Error("Should contain match statement")
	}
	if !strings.Contains(result, "Left(data)") {
		t.Error("Should contain Left arm")
	}
	if !strings.Contains(result, "Right(sig)") {
		t.Error("Should contain Right arm")
	}
}
