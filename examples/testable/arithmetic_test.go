//go:build ignore

// Test for arithmetic operator → jet call transpilation.
package main

import "simplicity/jet"

const MinAmount uint32 = 1000
const Fee uint32 = 50

func main() {
	// Subtraction: subtract_32 with carry-bit destructuring
	amount := jet.Subtract32(2000, Fee)
	// Comparison: lt_32 via Go < operator — should use runtime values
	height := jet.TxLockHeight()
	ok := jet.Le32(MinAmount, height)
	jet.Verify(ok)
	// Add32 — should generate (bool, u32) destructuring
	total := jet.Add32(MinAmount, Fee)
	_ = total
	_ = amount
}
