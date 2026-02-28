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

// TestPhase5ArithmeticJetRegistry verifies all Phase 5 arithmetic, comparison,
// and bitwise logic jets are registered with correct Simplicity names and types.
func TestPhase5ArithmeticJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
		returnType     string
	}{
		// Arithmetic — carry/borrow return types
		{"Add8", "add_8", "(bool, u8)"},
		{"Add16", "add_16", "(bool, u16)"},
		{"Add32", "add_32", "(bool, u32)"},
		{"Add64", "add_64", "(bool, u64)"},
		{"Subtract8", "subtract_8", "(bool, u8)"},
		{"Subtract16", "subtract_16", "(bool, u16)"},
		{"Subtract32", "subtract_32", "(bool, u32)"},
		{"Subtract64", "subtract_64", "(bool, u64)"},
		// Multiply returns double-width
		{"Multiply8", "multiply_8", "u16"},
		{"Multiply16", "multiply_16", "u32"},
		{"Multiply32", "multiply_32", "u64"},
		{"Multiply64", "multiply_64", "u128"},
		// Divide and modulo — same width as input
		{"Divide32", "divide_32", "u32"},
		{"Divide64", "divide_64", "u64"},
		{"Modulo32", "modulo_32", "u32"},
		{"Modulo64", "modulo_64", "u64"},
		// Comparisons — all return bool
		{"Lt8", "lt_8", "bool"},
		{"Lt16", "lt_16", "bool"},
		{"Lt32", "lt_32", "bool"},
		{"Lt64", "lt_64", "bool"},
		{"Le8", "le_8", "bool"},
		{"Le16", "le_16", "bool"},
		{"Le64", "le_64", "bool"},
		{"Eq8", "eq_8", "bool"},
		{"Eq16", "eq_16", "bool"},
		{"Eq64", "eq_64", "bool"},
		// Bitwise logic
		{"And32", "and_32", "u32"},
		{"Or32", "or_32", "u32"},
		{"Xor32", "xor_32", "u32"},
		{"Complement32", "complement_32", "u32"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Phase 5 jet %s should be registered", tc.goName)
				return
			}
			if info.SimplicityName != tc.simplicityName {
				t.Errorf("%s: expected SimplicityName %q, got %q", tc.goName, tc.simplicityName, info.SimplicityName)
			}
			if info.ReturnType != tc.returnType {
				t.Errorf("%s: expected ReturnType %q, got %q", tc.goName, tc.returnType, info.ReturnType)
			}
		})
	}
}

// TestPhase5TimeLockJetRegistry verifies time lock jets are registered.
func TestPhase5TimeLockJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
	}{
		{"CheckLockTime", "check_lock_time"},
		{"TxIsFinal", "tx_is_final"},
		{"TxLockHeight", "tx_lock_height"},
		{"TxLockTime", "tx_lock_time"},
		{"CheckLockDistance", "check_lock_distance"},
		{"CheckLockDuration", "check_lock_duration"},
		{"TxLockDistance", "tx_lock_distance"},
		{"TxLockDuration", "tx_lock_duration"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			_, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Time lock jet %s should be registered", tc.goName)
			}
		})
	}
}

// TestPhase5TxIntrospectionJetRegistry verifies transaction introspection jets are registered.
func TestPhase5TxIntrospectionJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
	}{
		{"NumInputs", "num_inputs"},
		{"NumOutputs", "num_outputs"},
		{"InputPrevOutpoint", "input_prev_outpoint"},
		{"OutputScriptHash", "output_script_hash"},
		{"InputScriptHash", "input_script_hash"},
		{"CurrentSequence", "current_sequence"},
		{"Version", "version"},
		{"TransactionId", "transaction_id"},
		{"GenesisBlockHash", "genesis_block_hash"},
		{"InternalKey", "internal_key"},
		{"TapleafVersion", "tapleaf_version"},
		{"Tappath", "tappath"},
		{"ScriptCmr", "script_cmr"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			_, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Transaction introspection jet %s should be registered", tc.goName)
			}
		})
	}
}

// TestPhase5SHA256VariantRegistry verifies additional SHA-256 add jets are registered.
func TestPhase5SHA256VariantRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	variants := []string{
		"SHA256Add1", "SHA256Add2", "SHA256Add4", "SHA256Add8",
		"SHA256Add16", "SHA256Add64",
	}

	for _, name := range variants {
		t.Run(name, func(t *testing.T) {
			info, found := registry.Lookup(name)
			if !found {
				t.Errorf("SHA-256 variant jet %s should be registered", name)
				return
			}
			if info.ReturnType != "Ctx8" {
				t.Errorf("%s: expected ReturnType Ctx8, got %s", name, info.ReturnType)
			}
		})
	}
}

// TestPhase5ArithmeticJetCallGeneration verifies arithmetic jet calls compile to correct SimplicityHL.
func TestPhase5ArithmeticJetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	result := jet.Add32(100, 200)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::add_32") {
		t.Errorf("expected jet::add_32 in output, got:\n%s", out)
	}
}

// TestPhase5ComparisonJetCallGeneration verifies comparison jet calls compile correctly.
func TestPhase5ComparisonJetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	ok := jet.Lt32(100, 200)
	jet.Verify(ok)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::lt_32") {
		t.Errorf("expected jet::lt_32 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "jet::verify") {
		t.Errorf("expected jet::verify in output, got:\n%s", out)
	}
}

// TestPhase5TimeLockJetCallGeneration verifies time lock jet calls compile correctly.
func TestPhase5TimeLockJetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const MinHeight uint32 = 800000

func main() {
	jet.CheckLockHeight(MinHeight)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::check_lock_height") {
		t.Errorf("expected jet::check_lock_height in output, got:\n%s", out)
	}
	if !strings.Contains(out, "800000") {
		t.Errorf("expected 800000 constant in output, got:\n%s", out)
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

// TestPhase8SHA256VariantRegistry verifies that the Phase 8 SHA-256 variant jets are registered.
func TestPhase8SHA256VariantRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	variants := []struct {
		goName         string
		simplicityName string
	}{
		{"SHA256Add128", "sha_256_ctx_8_add_128"},
		{"SHA256Add256", "sha_256_ctx_8_add_256"},
		{"SHA256Add512", "sha_256_ctx_8_add_512"},
		{"SHA256Block", "sha_256_block"},
		{"SHA256IV", "sha_256_iv"},
	}

	for _, v := range variants {
		t.Run(v.goName, func(t *testing.T) {
			info, found := registry.Lookup(v.goName)
			if !found {
				t.Errorf("Phase 8 jet %s should be registered", v.goName)
				return
			}
			if info.SimplicityName != v.simplicityName {
				t.Errorf("%s: expected SimplicityName %q, got %q", v.goName, v.simplicityName, info.SimplicityName)
			}
		})
	}
}

// TestPhase9TaprootJetRegistry verifies that Taproot-specific introspection jets
// are registered with correct Simplicity names.
func TestPhase9TaprootJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
	}{
		{"InternalKey", "internal_key"},
		{"TapleafVersion", "tapleaf_version"},
		{"Tappath", "tappath"},
		{"ScriptCmr", "script_cmr"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Taproot jet %s should be registered", tc.goName)
				return
			}
			if info.SimplicityName != tc.simplicityName {
				t.Errorf("%s: expected SimplicityName %q, got %q", tc.goName, tc.simplicityName, info.SimplicityName)
			}
		})
	}
}

// TestSHA256AutoSelect verifies that jet.SHA256Add auto-selects the correct variant
// based on the argument type at transpile time.
func TestSHA256AutoSelect(t *testing.T) {
	// A 64-byte witness variable should produce sha_256_ctx_8_add_64
	source64 := `
package main

import "simplicity/jet"

func main() {
	var block [64]byte
	ctx := jet.SHA256Add(jet.SHA256Init(), block)
	hash := jet.SHA256Finalize(ctx)
	jet.Eq256(hash, hash)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out64, err := c.Compile(source64, "test.go")
	if err != nil {
		t.Fatalf("compile (64-byte) failed: %v", err)
	}
	if !strings.Contains(out64, "sha_256_ctx_8_add_64") {
		t.Errorf("64-byte witness: expected sha_256_ctx_8_add_64, got:\n%s", out64)
	}

	// A 32-byte witness variable should produce sha_256_ctx_8_add_32
	source32 := `
package main

import "simplicity/jet"

func main() {
	var data [32]byte
	ctx := jet.SHA256Add(jet.SHA256Init(), data)
	hash := jet.SHA256Finalize(ctx)
	jet.Eq256(hash, hash)
}
`
	out32, err := c.Compile(source32, "test.go")
	if err != nil {
		t.Fatalf("compile (32-byte) failed: %v", err)
	}
	if !strings.Contains(out32, "sha_256_ctx_8_add_32") {
		t.Errorf("32-byte witness: expected sha_256_ctx_8_add_32, got:\n%s", out32)
	}
}
