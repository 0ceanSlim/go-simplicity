//go:build ignore

// Timelock check using arithmetic jets (test/demo only).
package main

import "simplicity/jet"

const MinHeight uint32 = 800000

func main() {
	height := jet.TxLockHeight()
	ok := jet.Le32(MinHeight, height)
	jet.Verify(ok)
}
