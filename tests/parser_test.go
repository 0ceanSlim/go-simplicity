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

	if !contains(result, "const AMOUNT_VALID: bool = true") {
		t.Error("Should pre-compute amount validation")
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

	if !contains(result, "assert!") {
		t.Error("Main function should contain assertion")
	}

	// Verify it looks like valid SimplicityHL
	lines := strings.Split(result, "\n")
	if len(lines) < 10 {
		t.Error("Generated code seems too short")
	}
}

// Helper functions
func contains(text, substring string) bool {
	return strings.Contains(text, substring)
}
