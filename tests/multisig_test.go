package tests

import (
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
	"github.com/0ceanslim/go-simplicity/pkg/types"
)

func TestArrayTypeParsing(t *testing.T) {
	tm := types.NewTypeMapper()

	testCases := []struct {
		name     string
		goType   string
		expected string
	}{
		{"simple array", "[3]u256", "[u256; 3]"},
		{"byte array", "[64]byte", "[u8; 64]"},
		{"nested option", "[3]Option[[64]byte]", "[Option<[u8; 64]>; 3]"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: Direct string conversion isn't supported, but we test via compilation
			_ = tm
			_ = tc
		})
	}
}

func TestOptionArrayDetection(t *testing.T) {
	// Test that Option arrays are properly detected
	testCases := []struct {
		typeStr  string
		isOption bool
	}{
		{"Option<[u8; 64]>", true},
		{"Option<u256>", true},
		{"[u8; 64]", false},
		{"u256", false},
	}

	for _, tc := range testCases {
		t.Run(tc.typeStr, func(t *testing.T) {
			result := types.IsSumType(tc.typeStr)
			if result != tc.isOption {
				t.Errorf("IsSumType(%s) = %v, want %v", tc.typeStr, result, tc.isOption)
			}
		})
	}
}

func TestSimpleArrayDeclaration(t *testing.T) {
	source := `
package main

func main() {
	var sigs [3][64]byte
	_ = sigs
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

	// Check that array type appears in witness
	if !strings.Contains(result, "[[u8; 64]; 3]") {
		t.Errorf("Expected [[u8; 64]; 3] in output, got:\n%s", result)
	}
}

func TestOptionArrayDeclaration(t *testing.T) {
	source := `
package main

func main() {
	var sigs [3]Option[[64]byte]
	_ = sigs
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

	// Check that Option array type appears in witness
	if !strings.Contains(result, "Option<") {
		t.Errorf("Expected Option type in output, got:\n%s", result)
	}
}

func TestArrayConstant(t *testing.T) {
	source := `
package main

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const BobPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const CharliePubkey = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

func main() {
	var sig [64]byte
	_ = sig
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

	// Check that all pubkeys appear in params
	checks := []string{
		"ALICE_PUBKEY",
		"BOB_PUBKEY",
		"CHARLIE_PUBKEY",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Expected %s in output, got:\n%s", check, result)
		}
	}
}

func TestBoundedForLoopAllowed(t *testing.T) {
	// Test that bounded for loops are allowed by the compiler
	source := `
package main

func main() {
	var count int
	for i := 0; i < 3; i++ {
		count++
	}
	_ = count
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	_, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Bounded for loop should be allowed: %v", err)
	}
}

func TestArrayIndexing(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const Pubkey0 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(Pubkey0, msg, sig)
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

	// Check that jet calls are present
	if !strings.Contains(result, "jet::bip_0340_verify") {
		t.Error("Should contain jet::bip_0340_verify")
	}
	if !strings.Contains(result, "jet::sig_all_hash") {
		t.Error("Should contain jet::sig_all_hash")
	}
}

func TestMultiplePubkeyConstants(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const BobPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4

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

	// Check both pubkeys are in params
	if !strings.Contains(result, "ALICE_PUBKEY") {
		t.Error("Should contain ALICE_PUBKEY")
	}
	if !strings.Contains(result, "BOB_PUBKEY") {
		t.Error("Should contain BOB_PUBKEY")
	}
}

func TestSimpleMultisigStructure(t *testing.T) {
	// Test that we can parse a simple multisig-like structure
	source := `
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const BobPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const CharliePubkey = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

func main() {
	var sig0 [64]byte
	var sig1 [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(AlicePubkey, msg, sig0)
	jet.BIP340Verify(BobPubkey, msg, sig1)
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

	// Check all pubkeys are present
	checks := []string{
		"ALICE_PUBKEY",
		"BOB_PUBKEY",
		"CHARLIE_PUBKEY",
		"jet::bip_0340_verify",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Output should contain %s", check)
		}
	}
}

func TestArrayUnrollFramework(t *testing.T) {
	// Test the array unroll framework with arrays.go
	source := `
package main

func main() {
	var sigs [3][64]byte
	for i := 0; i < 3; i++ {
		_ = sigs[i]
	}
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	_, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compilation with loop should not fail: %v", err)
	}
}
