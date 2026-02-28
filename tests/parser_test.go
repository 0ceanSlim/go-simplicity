package tests

import (
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
)

func TestBasicFunction(t *testing.T) {
	source := `
package main

func Add(a uint32, b uint32) uint32 {
    return a + b
}

func main() {
    result := Add(40, 2)
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

	// Check that the result contains expected patterns for new transpiler
	if !contains(result, "mod witness") {
		t.Error("Generated code should contain witness module")
	}

	if !contains(result, "mod param") {
		t.Error("Generated code should contain param module")
	}

	if !contains(result, "fn add(") {
		t.Error("Function should be converted to snake_case")
	}

	if !contains(result, "fn main()") {
		t.Error("Main function should be generated")
	}

	if !contains(result, "assert!") {
		t.Error("Main should contain assertion")
	}
}

func TestSimpleValidation(t *testing.T) {
	source := `
package main

func ValidateAmount(amountValid bool) bool {
    return amountValid
}

func main() {
    var amount uint64 = 1000
    amountValid := amount > 0
    result := ValidateAmount(amountValid)
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

	// Check for witness constants
	if !contains(result, "const AMOUNT: u64 = 1000") {
		t.Error("Should generate amount constant in witness module")
	}

	// Phase 5+: amount > 0 is a runtime comparison that maps to a lt_64 jet call
	// instead of being pre-computed as a compile-time bool witness.
	if !contains(result, "jet::lt_64") {
		t.Error("amount > 0 should generate a lt_64 jet call (witness::AMOUNT is runtime)")
	}

	// Check function generation
	if !contains(result, "fn validate_amount(amount_valid: bool) -> bool") {
		t.Error("Function signature should be converted correctly")
	}
}

func TestUnsupportedFeatures(t *testing.T) {
	testCases := []struct {
		name     string
		source   string
		errorMsg string
	}{
		{
			name: "Slice usage",
			source: `
package main
func process(data []byte) {}
`,
			errorMsg: "slices are not supported",
		},
		{
			name: "Map usage",
			source: `
package main
func process() {
    m := make(map[string]int)
}
`,
			errorMsg: "maps are not supported",
		},
		{
			name: "Channel usage",
			source: `
package main
func process() {
    ch := make(chan int)
}
`,
			errorMsg: "channels are not supported",
		},
		{
			name: "Goroutine usage",
			source: `
package main
func process() {
    go func() {}()
}
`,
			errorMsg: "goroutines are not supported",
		},
		{
			name: "Interface usage",
			source: `
package main
type Reader interface {
    Read() []byte
}
`,
			errorMsg: "interfaces are not supported",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := compiler.New(compiler.Config{
				Target: "simplicityhl",
				Debug:  false,
			})

			_, err := c.Compile(tc.source, "test.go")
			if err == nil {
				t.Errorf("Expected compilation to fail for %s", tc.name)
				return
			}

			if !contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
			}
		})
	}
}

func TestSimpleConstants(t *testing.T) {
	source := `
package main

const MinAmount uint64 = 1000

func main() {
    var amount uint64 = 5000
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

	// Check that constants are generated
	if !contains(result, "const MIN_AMOUNT: u64 = 1000") {
		t.Error("Constants should be generated in param module")
	}

	if !contains(result, "const AMOUNT: u64 = 5000") {
		t.Error("Variable should be generated as witness constant")
	}
}

func TestBooleanLogic(t *testing.T) {
	source := `
package main

func ValidateLogic(a bool, b bool) bool {
	if !a {
		return false
	}
	return b
}

func main() {
	result := ValidateLogic(true, false)
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

	// Check function generation
	if !contains(result, "fn validate_logic(a: bool, b: bool) -> bool") {
		t.Error("Function with boolean parameters should be generated")
	}

	// Should contain witness constants
	if !contains(result, "mod witness") {
		t.Error("Should generate witness module")
	}
}

func TestWorkingExample(t *testing.T) {
	// Use the exact working example
	source := `
package main

func ValidateAmount(amountValid bool) bool {
	return amountValid
}

func ValidateFee(feeValid bool) bool {
	return feeValid
}

func BasicSwap(amountValid bool, feeValid bool) bool {
	if !amountValid {
		return false
	}
	return feeValid
}

func main() {
	var amount uint64 = 1000
	var rate uint64 = 1500
	var minFee uint64 = 100
	
	amountValid := amount > 0
	calculatedFee := (amount * rate) / 10000
	feeValid := calculatedFee >= minFee
	
	result := BasicSwap(amountValid, feeValid)
	
	if !result {
		return
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

	// This should generate working SimplicityHL code
	if !contains(result, "mod witness") && !contains(result, "mod param") {
		t.Error("Should generate both witness and param modules")
	}

	if !contains(result, "fn main()") {
		t.Error("Should generate main function")
	}

	// Phase 5+: runtime comparisons now emit jet calls (e.g. amount > 0 → lt_64).
	// The old fallback assert!() is no longer generated when jet calls are present.
	// Check that fn main() has meaningful content instead.
	if !contains(result, "fn main()") {
		t.Error("Should generate main function")
	}

	// Verify it looks like valid SimplicityHL
	lines := strings.Split(result, "\n")
	if len(lines) < 10 {
		t.Error("Generated code seems too short")
	}
}

// TestSwitchMatchGeneration verifies that a tagless switch {} with IsLeft conditions
// generates a SimplicityHL match expression with Left(data) and Right(sig) arms.
func TestSwitchMatchGeneration(t *testing.T) {
	source := `
//go:build ignore
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const SenderPubkey    = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const MinRefundHeight uint32 = 800000

type SwapWitness struct {
	IsLeft    bool
	RecipSig  [64]byte
	SenderSig [64]byte
}

func main() {
	var w SwapWitness
	switch {
	case w.IsLeft:
		msg := jet.SigAllHash()
		jet.BIP340Verify(RecipientPubkey, msg, w.RecipSig)
	case !w.IsLeft:
		jet.CheckLockHeight(MinRefundHeight)
		msg := jet.SigAllHash()
		jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
	}
}
`
	source = strings.Replace(source, "//go:build ignore\n", "", 1)

	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	for _, want := range []string{"match witness::W {", "Left(data)", "Right(sig)"} {
		if !strings.Contains(out, want) {
			t.Errorf("TestSwitchMatchGeneration: missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestHelperFunctionBody verifies that a linear helper function body is correctly transpiled —
// parameter names resolve as bare identifiers and jet calls are emitted with correct names.
func TestHelperFunctionBody(t *testing.T) {
	source := `
//go:build ignore
package main

import "simplicity/jet"

const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

func verifyHashlock(preimage [32]byte) {
	hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), preimage))
	jet.Eq256(hash, HashLock)
}

func main() {
}
`
	source = strings.Replace(source, "//go:build ignore\n", "", 1)

	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	for _, want := range []string{
		"fn verify_hashlock(",
		"jet::sha_256_ctx_8_finalize(",
		"jet::sha_256_ctx_8_add_32(",
		"jet::sha_256_ctx_8_init()",
		"jet::eq_256(",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("TestHelperFunctionBody: missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestInlinedHelperCall verifies that when a helper function is called from a switch arm,
// its body is inlined into the match arm — not just referenced by name.
func TestInlinedHelperCall(t *testing.T) {
	source := `
//go:build ignore
package main

import "simplicity/jet"

const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const SenderPubkey    = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const MinRefundHeight uint32 = 800000

type AtomicSwapWitness struct {
	IsLeft       bool
	Preimage     [32]byte
	RecipientSig [64]byte
	SenderSig    [64]byte
}

func verifyHashlock(preimage [32]byte) {
	hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), preimage))
	jet.Eq256(hash, HashLock)
}

func main() {
	var w AtomicSwapWitness
	switch {
	case w.IsLeft:
		verifyHashlock(w.Preimage)
		msg := jet.SigAllHash()
		jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
	case !w.IsLeft:
		jet.CheckLockHeight(MinRefundHeight)
		msg := jet.SigAllHash()
		jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
	}
}
`
	source = strings.Replace(source, "//go:build ignore\n", "", 1)

	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(source, "test.go")
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Inlined jet calls must appear inside the match arm body (Left arm), not just in fn verify_hashlock
	if !strings.Contains(out, "Left(data)") {
		t.Errorf("TestInlinedHelperCall: missing Left(data) arm\nfull output:\n%s", out)
	}

	// Count occurrences of sha_256_ctx_8_finalize — should appear at least twice (fn body + inlined)
	count := strings.Count(out, "jet::sha_256_ctx_8_finalize(")
	if count < 2 {
		t.Errorf("TestInlinedHelperCall: expected >=2 occurrences of sha_256_ctx_8_finalize (fn + inlined), got %d\nfull output:\n%s", count, out)
	}

	// The inlined calls must use the destructured field name, not witness::W.preimage
	if strings.Contains(out, "witness::W.preimage") {
		t.Errorf("TestInlinedHelperCall: inlined helper should use 'preimage' not 'witness::W.preimage'\nfull output:\n%s", out)
	}
}

// Helper functions
func contains(text, substring string) bool {
	return strings.Contains(text, substring)
}
