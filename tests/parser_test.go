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

	expected := `// Generated from Go source by go-simplicity compiler

fn Add(a: u32, b: u32) -> u32 {
    (a + b)
}

fn main() {
    let result = Add(40, 2);
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

	// Normalize whitespace for comparison
	if normalizeWhitespace(result) != normalizeWhitespace(expected) {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestBooleanLogic(t *testing.T) {
	source := `
package main

func ValidateAmount(amount uint64) bool {
    return amount > 0
}

func main() {
    var amount uint64 = 1000
    valid := ValidateAmount(amount)
    if valid {
        // success
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

	// Check that the result contains expected patterns
	if !contains(result, "fn ValidateAmount(amount: u64) -> bool") {
		t.Error("Function signature not correctly transpiled")
	}

	if !contains(result, "(amount > 0)") {
		t.Error("Boolean expression not correctly transpiled")
	}

	if !contains(result, "match valid") {
		t.Error("If statement not correctly transpiled to match")
	}
}

func TestArrayTypes(t *testing.T) {
	source := `
package main

func ProcessHash(hash [32]byte) bool {
    var zero [32]byte
    for i := 0; i < 32; i++ {
        if hash[i] != zero[i] {
            return true
        }
    }
    return false
}
`

	c := compiler.New(compiler.Config{
		Target: "simplicityhl",
		Debug:  false,
	})

	_, err := c.Compile(source, "test.go")
	// This should fail because loops are not supported
	if err == nil {
		t.Error("Expected compilation to fail due to loop usage")
	}

	if !contains(err.Error(), "loops are not supported") {
		t.Errorf("Expected error about loops, got: %v", err)
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

func TestTypeMapping(t *testing.T) {
	source := `
package main

type Hash [32]byte
type Amount uint64

func ProcessTransaction(hash Hash, amount Amount) bool {
    return amount > 0
}

func main() {
    var h Hash
    var a Amount = 1000
    result := ProcessTransaction(h, a)
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

	// Check type aliases are generated
	if !contains(result, "type Hash = [u8; 32];") {
		t.Error("Hash type alias not correctly generated")
	}

	if !contains(result, "type Amount = u64;") {
		t.Error("Amount type alias not correctly generated")
	}
}

func TestComplexExpression(t *testing.T) {
	source := `
package main

func CalculateFee(amount uint64, rate uint64) uint64 {
    return (amount * rate) / 10000
}

func main() {
    fee := CalculateFee(1000, 25)
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

	// Check that complex arithmetic is properly parenthesized
	if !contains(result, "((amount * rate) / 10000)") {
		t.Error("Complex arithmetic expression not correctly transpiled")
	}
}

func TestVariableDeclarations(t *testing.T) {
	source := `
package main

func main() {
    var amount uint64 = 1000
    const fee uint64 = 100
    rate := 25
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

	// Check variable declarations
	if !contains(result, "let amount: u64 = 1000;") {
		t.Error("Variable declaration not correctly transpiled")
	}

	if !contains(result, "let fee: u64 = 100;") {
		t.Error("Constant declaration not correctly transpiled")
	}

	if !contains(result, "let rate = 25;") {
		t.Error("Type inference assignment not correctly transpiled")
	}
}

func TestFunctionCalls(t *testing.T) {
	source := `
package main

func Add(a uint32, b uint32) uint32 {
    return a + b
}

func main() {
    x := Add(10, 20)
    y := Add(x, 5)
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

	// Check function calls
	if !contains(result, "let x = Add(10, 20);") {
		t.Error("Function call not correctly transpiled")
	}

	if !contains(result, "let y = Add(x, 5);") {
		t.Error("Nested function call not correctly transpiled")
	}
}

func TestConditionals(t *testing.T) {
	source := `
package main

func main() {
    amount := 1000
    if amount > 0 {
        return
    } else {
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

	// Check conditional structure
	if !contains(result, "match (amount > 0)") {
		t.Error("If condition not correctly transpiled to match")
	}

	if !contains(result, "true => {") {
		t.Error("True branch not correctly generated")
	}

	if !contains(result, "false => {") {
		t.Error("False branch not correctly generated")
	}
}

// Helper functions

func normalizeWhitespace(s string) string {
	// Remove leading/trailing whitespace and normalize internal whitespace
	lines := strings.Split(s, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return strings.Join(normalized, "\n")
}

func contains(text, substring string) bool {
	return strings.Contains(text, substring)
}
