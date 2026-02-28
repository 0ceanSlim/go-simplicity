package tests

import (
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
	"github.com/0ceanslim/go-simplicity/pkg/types"
)

func TestArrayTypeParsing(t *testing.T) {
	tm := types.NewTypeMapper()

	// GetBitSize handles SimplicityHL type strings directly — test array types.
	bitSizeCases := []struct {
		simplicityType string
		expectedBits   int
	}{
		{"[u8; 32]", 256}, // 32 bytes × 8 bits
		{"[u8; 64]", 512}, // 64 bytes × 8 bits
		{"[u8; 1]", 8},    // 1 byte
		{"[u8; 16]", 128}, // 16 bytes × 8 bits
	}

	for _, tc := range bitSizeCases {
		t.Run("bitsize_"+tc.simplicityType, func(t *testing.T) {
			got := tm.GetBitSize(tc.simplicityType)
			if got != tc.expectedBits {
				t.Errorf("GetBitSize(%q) = %d, want %d", tc.simplicityType, got, tc.expectedBits)
			}
		})
	}

	// Verify array types appear correctly in compiled SimplicityHL output.
	compileCases := []struct {
		name string
		decl string // Go variable declaration
		want string // expected SimplicityHL type string in output
	}{
		{"[32]byte witness", "var preimage [32]byte", "[u8; 32]"},
		{"[64]byte witness", "var sig [64]byte", "[u8; 64]"},
	}

	for _, tc := range compileCases {
		t.Run("compile_"+tc.name, func(t *testing.T) {
			src := `package main
import "simplicity/jet"
func main() {
	` + tc.decl + `
	_ = preimage
	msg := jet.SigAllHash()
	jet.BIP340Verify(0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0, msg, sig)
}
`
			// For the sig-only case, adjust source
			if tc.decl == "var sig [64]byte" {
				src = `package main
import "simplicity/jet"
func main() {
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0, msg, sig)
}
`
			} else {
				src = `package main
import "simplicity/jet"
func main() {
	var preimage [32]byte
	var sig [64]byte
	msg := jet.SigAllHash()
	jet.BIP340Verify(0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0, msg, sig)
	_ = preimage
}
`
			}

			c := compiler.New(compiler.Config{Target: "simplicityhl"})
			out, err := c.Compile(src, "test.go")
			if err != nil {
				t.Fatalf("compile failed: %v", err)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, out)
			}
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
