package tests

import (
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
	"github.com/0ceanslim/go-simplicity/pkg/jets"
	"github.com/0ceanslim/go-simplicity/pkg/types"
)

func TestJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	// Test that core jets are registered
	testCases := []struct {
		goName         string
		simplicityName string
	}{
		{"BIP340Verify", "bip_0340_verify"},
		{"SigAllHash", "sig_all_hash"},
		{"SHA256Init", "sha_256_ctx_8_init"},
		{"Eq256", "eq_256"},
		{"Le32", "le_32"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Jet %s should be registered", tc.goName)
				return
			}
			if info.SimplicityName != tc.simplicityName {
				t.Errorf("Expected Simplicity name %s, got %s", tc.simplicityName, info.SimplicityName)
			}
		})
	}
}

func TestHexTypeInference(t *testing.T) {
	tm := types.NewTypeMapper()

	testCases := []struct {
		hexValue     string
		expectedType string
	}{
		{"0x12", "u8"},
		{"0x1234", "u16"},
		{"0x12345678", "u32"},
		{"0x1234567890abcdef", "u64"},
		{"0x1234567890abcdef1234567890abcdef", "u128"},
		{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "u256"},
	}

	for _, tc := range testCases {
		t.Run(tc.hexValue, func(t *testing.T) {
			result := tm.InferHexType(tc.hexValue)
			if result != tc.expectedType {
				t.Errorf("InferHexType(%s) = %s, want %s", tc.hexValue, result, tc.expectedType)
			}
		})
	}
}

func TestHexLiteral(t *testing.T) {
	source := `
package main

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
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

	// Check that hex literal is preserved with lowercase normalization
	if !strings.Contains(result, "0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0") {
		t.Error("Hex literal should be preserved in output")
	}

	// Check that constant is in param module
	if !strings.Contains(result, "ALICE_PUBKEY") {
		t.Error("Constant name should be converted to UPPER_SNAKE_CASE")
	}
}

func TestJetSigAllHash(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	msg := jet.SigAllHash()
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

	// Check that jet call is generated
	if !strings.Contains(result, "jet::sig_all_hash()") {
		t.Error("Should generate jet::sig_all_hash() call")
	}

	// Check that let binding is generated
	if !strings.Contains(result, "let msg:") {
		t.Error("Should generate let binding for msg")
	}
}

func TestJetBIP340Verify(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(AlicePubkey, msg, sig)
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

	// Check that BIP340Verify jet call is generated
	if !strings.Contains(result, "jet::bip_0340_verify") {
		t.Error("Should generate jet::bip_0340_verify() call")
	}

	// Check that witness declaration for sig is generated
	if !strings.Contains(result, "SIG") && !strings.Contains(result, "[u8; 64]") {
		t.Error("Should generate witness for signature")
	}
}

func TestP2PKContract(t *testing.T) {
	// Full P2PK contract test
	source := `
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(AlicePubkey, msg, sig)
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

	// Verify all key components are present
	checks := []struct {
		name     string
		contains string
	}{
		{"witness module", "mod witness"},
		{"param module", "mod param"},
		{"main function", "fn main()"},
		{"pubkey constant", "ALICE_PUBKEY"},
		{"sig_all_hash jet", "jet::sig_all_hash"},
		{"bip_0340_verify jet", "jet::bip_0340_verify"},
	}

	for _, check := range checks {
		if !strings.Contains(result, check.contains) {
			t.Errorf("Output should contain %s (%s)", check.name, check.contains)
		}
	}
}

func TestJetCallValidation(t *testing.T) {
	// Test that jet calls don't trigger unsupported feature errors
	source := `
package main

import "simplicity/jet"

func main() {
	hash := jet.SigAllHash()
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	_, err := c.Compile(source, "test.go")
	if err != nil {
		t.Errorf("Jet calls should be allowed: %v", err)
	}
}

func TestCtx8Type(t *testing.T) {
	tm := types.NewTypeMapper()

	// Test that Ctx8 is a supported type
	supported := tm.SupportedTypes()
	found := false
	for _, typ := range supported {
		if typ == "Ctx8" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Ctx8 should be a supported type")
	}
}
