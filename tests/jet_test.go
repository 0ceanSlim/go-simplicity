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
	if !strings.Contains(out, "assert!") {
		t.Errorf("expected assert! in output, got:\n%s", out)
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

// TestElementsAmountJetRegistry verifies all Elements amount and issuance jets
// are registered with correct Simplicity names and return types.
func TestElementsAmountJetRegistry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
		returnType     string
	}{
		// Output amount jets
		{"OutputAsset", "output_asset", "u256"},
		{"OutputAmount", "output_amount", "u64"},
		// Input amount jets
		{"InputAsset", "input_asset", "u256"},
		{"InputAmount", "input_amount", "u64"},
		// Current input jets
		{"CurrentAsset", "current_asset", "u256"},
		{"CurrentAmount", "current_amount", "u64"},
		// Issuance jets
		{"IssuanceAssetAmount", "issuance_asset_amount", "u64"},
		{"IssuanceTokenAmount", "issuance_token_amount", "u64"},
		{"NewIssuanceContract", "new_issuance_contract", "u256"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("Elements jet %s should be registered", tc.goName)
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

// TestElementsAmountJetCallGeneration verifies Elements amount jets compile to correct SimplicityHL.
func TestElementsAmountJetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	asset := jet.OutputAsset(0)
	amount := jet.OutputAmount(0)
	_ = asset
	_ = amount
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::output_asset") {
		t.Errorf("expected jet::output_asset in output, got:\n%s", out)
	}
	if !strings.Contains(out, "jet::output_amount") {
		t.Errorf("expected jet::output_amount in output, got:\n%s", out)
	}
}

// TestElementsIssuanceJetCallGeneration verifies issuance jets compile to correct SimplicityHL.
func TestElementsIssuanceJetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	lpMinted := jet.IssuanceAssetAmount(0)
	_ = lpMinted
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::issuance_asset_amount") {
		t.Errorf("expected jet::issuance_asset_amount in output, got:\n%s", out)
	}
}

// TestV12Add128Subtract128Registry verifies that v1.2 Add128 and Subtract128 jets
// are registered with correct Simplicity names and carry-tuple return types.
func TestV12Add128Subtract128Registry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
		returnType     string
	}{
		{"Add128", "add_128", "(bool, u128)"},
		{"Subtract128", "subtract_128", "(bool, u128)"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("v1.2 jet %s should be registered", tc.goName)
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

// TestV12Le128Registry verifies Le128, Lt128, and Eq128 are registered correctly.
func TestV12Le128Registry(t *testing.T) {
	registry := jets.NewRegistry()

	testCases := []struct {
		goName         string
		simplicityName string
	}{
		{"Le128", "le_128"},
		{"Lt128", "lt_128"},
		{"Eq128", "eq_128"},
	}

	for _, tc := range testCases {
		t.Run(tc.goName, func(t *testing.T) {
			info, found := registry.Lookup(tc.goName)
			if !found {
				t.Errorf("jet %s should be registered", tc.goName)
				return
			}
			if info.SimplicityName != tc.simplicityName {
				t.Errorf("%s: expected SimplicityName %q, got %q", tc.goName, tc.simplicityName, info.SimplicityName)
			}
			if info.ReturnType != "bool" {
				t.Errorf("%s: expected ReturnType %q, got %q", tc.goName, "bool", info.ReturnType)
			}
		})
	}
}

// TestV12Add128JetCallGeneration verifies Add128 compiles to the correct SimplicityHL binding.
// The carry bit must be discarded with the (_, varName) destructuring pattern.
func TestV12Add128JetCallGeneration(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	sum := jet.Add128(1, 2)
	_ = sum
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(out, "jet::add_128") {
		t.Errorf("expected jet::add_128 in output, got:\n%s", out)
	}
	// Carry bit must be discarded via tuple destructuring
	if !strings.Contains(out, "(_, sum)") {
		t.Errorf("expected carry-discard pattern '(_, sum)' in output, got:\n%s", out)
	}
}

// TestV12AMMInvariantCompilation verifies the full AMM invariant pattern compiles correctly.
// This exercises: multiply_64 (→u128), le_128, eq_256, output_amount, output_asset,
// current_amount, input_amount, output_script_hash, current_script_hash.
// It also tests Bug B1 fix: add/subtract results used as subsequent multiply inputs
// must produce multiply_64, not multiply_32.
func TestV12AMMInvariantCompilation(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const Asset0 = 0x25b251070e29ca19043cf33ccd7324e2ddab03ecc4ae0b5e77c4fc0e5cf6c95a
const Asset1 = 0xce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2

const PoolInputB uint32 = 1
const PoolOutputA uint32 = 0
const PoolOutputB uint32 = 1

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.InputAmount(PoolInputB)
	newReserve0 := jet.OutputAmount(PoolOutputA)
	newReserve1 := jet.OutputAmount(PoolOutputB)
	asset0Out := jet.OutputAsset(PoolOutputA)
	asset1Out := jet.OutputAsset(PoolOutputB)
	jet.Verify(asset0Out == Asset0)
	jet.Verify(asset1Out == Asset1)
	kOld := jet.Multiply64(reserve0, reserve1)
	kNew := jet.Multiply64(newReserve0, newReserve1)
	jet.Verify(jet.Le128(kOld, kNew))
	newScriptA := jet.OutputScriptHash(PoolOutputA)
	myScript := jet.CurrentScriptHash()
	jet.Verify(newScriptA == myScript)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("AMM pool contract compile failed: %v", err)
	}

	checks := []struct {
		name     string
		contains string
	}{
		{"current_amount", "jet::current_amount"},
		{"input_amount", "jet::input_amount"},
		{"output_amount", "jet::output_amount"},
		{"output_asset", "jet::output_asset"},
		{"eq_256 for asset check", "jet::eq_256"},
		{"multiply_64 for kOld", "jet::multiply_64"},
		{"k-invariant borrow arithmetic", "jet::full_subtract_64("},
		{"k-invariant unwrap_left", "unwrap_left::<()>(<bool>::into("},
		{"output_script_hash", "jet::output_script_hash"},
		{"current_script_hash", "jet::current_script_hash"},
		{"verify", "assert!"},
	}

	for _, check := range checks {
		if !strings.Contains(out, check.contains) {
			t.Errorf("AMM output missing %s (%q):\n%s", check.name, check.contains, out)
		}
	}

	// le_128 is now inlined as borrow-arithmetic for top-level verify — no helper emitted.
	if strings.Contains(out, "fn le_128(") {
		t.Errorf("AMM output should not emit fn le_128 helper (inlined as borrow-arithmetic):\n%s", out)
	}
}

// TestV12CarryTupleUnwrapBugB1 verifies that a variable produced by add_64/subtract_64
// (which returns "(bool, u64)") can be used as an argument to multiply_64 without
// silently emitting multiply_32.  This tests the Bug B1 fix in inferExprType.
func TestV12CarryTupleUnwrapBugB1(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	a := jet.Add64(10, 20)
	b := jet.Add64(30, 40)
	product := jet.Multiply64(a, b)
	_ = product
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if strings.Contains(out, "jet::multiply_32") {
		t.Errorf("Bug B1 regression: got multiply_32 instead of multiply_64 for u64 operands:\n%s", out)
	}
	if !strings.Contains(out, "jet::multiply_64") {
		t.Errorf("expected jet::multiply_64 for u64 operands, got:\n%s", out)
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

func TestBooleanIfElse(t *testing.T) {
	// Test basic boolean if/else compilation (the mode-detection pattern used in AMM contracts).
	source := `
package main

import "simplicity/jet"

const PoolInputB uint32 = 1
const PoolOutputA uint32 = 0
const PoolOutputB uint32 = 1

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.InputAmount(PoolInputB)
	newReserve0 := jet.OutputAmount(PoolOutputA)
	newReserve1 := jet.OutputAmount(PoolOutputB)

	isRemoveMode := jet.Lt64(newReserve0, reserve0)

	if isRemoveMode {
		kOld := jet.Multiply64(reserve0, reserve1)
		kNew := jet.Multiply64(newReserve0, newReserve1)
		jet.Verify(jet.Le128(kOld, kNew))
	} else {
		kOld := jet.Multiply64(reserve0, reserve1)
		kNew := jet.Multiply64(newReserve0, newReserve1)
		jet.Verify(jet.Le128(kOld, kNew))
	}
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Boolean if/else compilation failed: %v", err)
	}

	// Should generate a match expression on the bool variable
	if !strings.Contains(result, "match is_remove_mode") {
		t.Errorf("Expected 'match is_remove_mode', got:\n%s", result)
	}

	// Should have true and false branches
	if !strings.Contains(result, "true =>") {
		t.Errorf("Expected 'true =>' branch, got:\n%s", result)
	}
	if !strings.Contains(result, "false =>") {
		t.Errorf("Expected 'false =>' branch, got:\n%s", result)
	}

	// Should contain lt_64 for mode detection
	if !strings.Contains(result, "jet::lt_64") {
		t.Errorf("Expected 'jet::lt_64', got:\n%s", result)
	}

	// Should contain multiply_64 and le_128 inside match arms
	if !strings.Contains(result, "jet::multiply_64") {
		t.Errorf("Expected 'jet::multiply_64', got:\n%s", result)
	}
	if !strings.Contains(result, "le_128(") {
		t.Errorf("Expected 'le_128(' helper call, got:\n%s", result)
	}
	if !strings.Contains(result, "fn le_128(") {
		t.Errorf("Expected 'fn le_128(' helper function definition, got:\n%s", result)
	}
}

func TestBooleanIfElseWithSubtract(t *testing.T) {
	// Test boolean if/else with arithmetic (subtract) in the true arm.
	source := `
package main

import "simplicity/jet"

const PoolInputB uint32 = 1
const PoolOutputA uint32 = 0
const PoolOutputB uint32 = 1
const LpSupplyInput uint32 = 2

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.InputAmount(PoolInputB)
	newReserve0 := jet.OutputAmount(PoolOutputA)
	newReserve1 := jet.OutputAmount(PoolOutputB)

	isRemoveMode := jet.Lt64(newReserve0, reserve0)

	if isRemoveMode {
		totalSupply := jet.InputAmount(LpSupplyInput)
		payout0 := reserve0 - newReserve0
		lhsFloor := jet.Multiply64(payout0, totalSupply)
		rhsFloor := jet.Multiply64(totalSupply, reserve0)
		jet.Verify(jet.Le128(lhsFloor, rhsFloor))
	} else {
		kOld := jet.Multiply64(reserve0, reserve1)
		kNew := jet.Multiply64(newReserve0, newReserve1)
		jet.Verify(jet.Le128(kOld, kNew))
	}
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Boolean if/else with subtract failed: %v", err)
	}

	if !strings.Contains(result, "match is_remove_mode") {
		t.Errorf("Expected match on is_remove_mode, got:\n%s", result)
	}
	if !strings.Contains(result, "jet::subtract_64") {
		t.Errorf("Expected jet::subtract_64 in true arm, got:\n%s", result)
	}
	if !strings.Contains(result, "jet::multiply_64") {
		t.Errorf("Expected jet::multiply_64 in arms, got:\n%s", result)
	}
}

// ── Bug 1 fix: Liquid jet unwrap emission ────────────────────────────────────

// TestLiquidJetUnwrapEmission verifies that amountPairOpt, assetOpt, and
// optScalarU256 jets emit unwrap/unwrap_right instead of match with assert!(false).
func TestLiquidJetUnwrapEmission(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const Out0 uint32 = 0
const In1  uint32 = 1

func main() {
	amt  := jet.OutputAmount(Out0)
	_    = amt
	amt2 := jet.InputAmount(In1)
	_    = amt2
	asst := jet.OutputAsset(Out0)
	_    = asst
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Must use unwrap() for Option unwrapping.
	if !strings.Contains(result, "unwrap(") {
		t.Errorf("Expected 'unwrap(' in output, got:\n%s", result)
	}
	// Must use unwrap_right for Either extraction.
	if !strings.Contains(result, "unwrap_right::<(u1, u256)>(") {
		t.Errorf("Expected 'unwrap_right::<(u1, u256)>(' in output, got:\n%s", result)
	}
	// Must NOT use match-based false-arm pattern.
	if strings.Contains(result, "None => { assert!(false)") {
		t.Errorf("Bug 1 regression: found 'None => { assert!(false)' in output:\n%s", result)
	}
	if strings.Contains(result, "Left(x:") && strings.Contains(result, "assert!(false)") {
		t.Errorf("Bug 1 regression: found Left(x: ... assert!(false) in output:\n%s", result)
	}
}

// ── Bug 2+3 fix: le_128 / lt_128 in Verify context ──────────────────────────

// TestLe128VerifyExpansion verifies that a top-level jet.Verify(jet.Le128(a,b))
// is emitted as borrow-arithmetic (no CASE nodes) instead of calling le_128().
func TestLe128VerifyExpansion(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.CurrentAmount()
	kOld := jet.Multiply64(reserve0, reserve1)
	newR0 := jet.CurrentAmount()
	newR1 := jet.CurrentAmount()
	kNew := jet.Multiply64(newR0, newR1)
	jet.Verify(jet.Le128(kOld, kNew))
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result, "jet::full_subtract_64(") {
		t.Errorf("Expected 'jet::full_subtract_64(' in output, got:\n%s", result)
	}
	if !strings.Contains(result, "unwrap_left::<()>(<bool>::into(") {
		t.Errorf("Expected 'unwrap_left::<()>(<bool>::into(' in output, got:\n%s", result)
	}
	// Helper function must NOT be emitted when le_128 only used in verify.
	if strings.Contains(result, "fn le_128(") {
		t.Errorf("Bug 2 regression: 'fn le_128(' helper should not be emitted for top-level verify, got:\n%s", result)
	}
}

// TestLt128VerifyExpansion verifies the same for lt_128.
func TestLt128VerifyExpansion(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	a := jet.CurrentAmount()
	b := jet.CurrentAmount()
	kOld := jet.Multiply64(a, b)
	kNew := jet.Multiply64(a, b)
	jet.Verify(jet.Lt128(kOld, kNew))
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result, "jet::full_subtract_64(") {
		t.Errorf("Expected 'jet::full_subtract_64(' in output, got:\n%s", result)
	}
	if !strings.Contains(result, "unwrap_left::<()>(<bool>::into(") {
		t.Errorf("Expected 'unwrap_left::<()>(<bool>::into(' in output, got:\n%s", result)
	}
	if strings.Contains(result, "fn lt_128(") {
		t.Errorf("Bug 2 regression: 'fn lt_128(' helper should not be emitted for top-level verify, got:\n%s", result)
	}
}

// TestLe128HelperStillEmittedForNonVerify verifies that when le_128 is stored
// as a value (not directly in Verify), the helper function IS still emitted.
func TestLe128HelperStillEmittedForNonVerify(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	a := jet.CurrentAmount()
	b := jet.CurrentAmount()
	kOld := jet.Multiply64(a, b)
	kNew := jet.Multiply64(a, b)
	ok := jet.Le128(kOld, kNew)
	jet.Verify(ok)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// le_128 is stored as a value, so the helper must be emitted.
	if !strings.Contains(result, "fn le_128(") {
		t.Errorf("Expected 'fn le_128(' helper for non-verify use, got:\n%s", result)
	}
}

// TestEq128VerifyExpansion verifies that a top-level jet.Verify(jet.Eq128(a,b))
// is inlined as two assert!(jet::eq_64(...)) calls with no CASE nodes.
func TestEq128VerifyExpansion(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.CurrentAmount()
	deposit0 := jet.CurrentAmount()
	deposit1 := jet.CurrentAmount()
	prod1 := jet.Multiply64(deposit0, reserve1)
	prod2 := jet.Multiply64(deposit1, reserve0)
	jet.Verify(jet.Eq128(prod1, prod2))
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result, "assert!(jet::eq_64(") {
		t.Errorf("Expected 'assert!(jet::eq_64(' in output, got:\n%s", result)
	}
	// Helper function must NOT be emitted when eq_128 only used in top-level verify.
	if strings.Contains(result, "fn eq_128(") {
		t.Errorf("'fn eq_128(' helper should not be emitted for top-level verify, got:\n%s", result)
	}
	// Must not contain any match expression (CASE nodes).
	if strings.Contains(result, "match ") {
		t.Errorf("Output must not contain 'match' (CASE nodes), got:\n%s", result)
	}
}

// TestEq128HelperStillEmittedForNonVerify verifies that when eq_128 is stored
// as a value (not directly in Verify), the helper function IS still emitted.
func TestEq128HelperStillEmittedForNonVerify(t *testing.T) {
	source := `
package main

import "simplicity/jet"

func main() {
	a := jet.CurrentAmount()
	b := jet.CurrentAmount()
	prod1 := jet.Multiply64(a, b)
	prod2 := jet.Multiply64(a, b)
	ok := jet.Eq128(prod1, prod2)
	jet.Verify(ok)
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// eq_128 is stored as a value, so the helper must be emitted.
	if !strings.Contains(result, "fn eq_128(") {
		t.Errorf("Expected 'fn eq_128(' helper for non-verify use, got:\n%s", result)
	}
}

// TestFeeAdjustedLe128Verify verifies that jet.Verify(jet.FeeAdjustedLe128(...))
// is inlined as multiply/add/subtract arithmetic (no CASE nodes, no helper function).
func TestFeeAdjustedLe128Verify(t *testing.T) {
	source := `
package main

import "simplicity/jet"

const FeeNum  uint64 = 997
const FeeDen  uint64 = 1000
const FeeDiff uint64 = 3

func main() {
	reserve0 := jet.CurrentAmount()
	reserve1 := jet.InputAmount(0)
	newReserve0 := jet.OutputAmount(0)
	newReserve1 := jet.OutputAmount(1)
	jet.Verify(jet.FeeAdjustedLe128(reserve0, newReserve0, FeeNum, FeeDiff, FeeDen, newReserve1, reserve1))
}
`
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	result, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Must contain multiply_64 (for fee-adjusted products)
	if !strings.Contains(result, "jet::multiply_64(") {
		t.Errorf("Expected 'jet::multiply_64(' in output, got:\n%s", result)
	}
	// Must contain the overflow guard
	if !strings.Contains(result, "assert!(jet::eq_64(") {
		t.Errorf("Expected 'assert!(jet::eq_64(' in output, got:\n%s", result)
	}
	// Must contain borrow-arithmetic for the final >= comparison
	if !strings.Contains(result, "unwrap_left::<()>(<bool>::into(") {
		t.Errorf("Expected 'unwrap_left::<()>(<bool>::into(' in output, got:\n%s", result)
	}
	// Must NOT emit a helper function (always inlined)
	if strings.Contains(result, "fn fee_adjusted_le_128(") {
		t.Errorf("fee_adjusted_le_128 should be inlined, not emitted as a helper function, got:\n%s", result)
	}
	// Must NOT contain match expressions (no CASE nodes)
	if strings.Contains(result, "match ") {
		t.Errorf("Expected no 'match ' in output (no CASE nodes), got:\n%s", result)
	}
	// Must NOT emit le_128 helper (fee_adjusted contains le_128 as substring but is different)
	if strings.Contains(result, "fn le_128(") {
		t.Errorf("'fn le_128(' helper should not be emitted for fee_adjusted_le_128, got:\n%s", result)
	}
	// Must use add_64 for u128 sum (add_128 jet does not exist in Simplicity)
	if !strings.Contains(result, "jet::add_64(") {
		t.Errorf("Expected 'jet::add_64(' for u128 sum in output, got:\n%s", result)
	}
	// Must NOT emit add_128 (jet does not exist in the Simplicity protocol)
	if strings.Contains(result, "jet::add_128(") {
		t.Errorf("'jet::add_128' does not exist in Simplicity and must not appear in output, got:\n%s", result)
	}
}
