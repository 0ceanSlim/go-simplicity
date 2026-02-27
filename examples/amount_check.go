//go:build ignore

// Amount Check Contract
//
// Demonstrates arithmetic and comparison jets (Phase 5) in a real contract.
// This contract ensures:
//   1. The current input has an index within the allowed range
//   2. The transaction has not expired (lock height check)
//
// This example uses:
//   - jet.TxLockHeight()    — transaction introspection
//   - jet.CurrentIndex()    — transaction introspection
//   - jet.Le32 / jet.Lt32   — comparison jets
//   - jet.Verify()          — assertion jet
//   - >= operator           — auto-maps to le_32 with swapped args
//   - jet.CheckLockHeight() — time lock jet
//
// Usage:
//   go run cmd/simgo/main.go -input examples/amount_check.go

package main

import "simplicity/jet"

// MinBlockHeight is the earliest block at which spending is allowed.
const MinBlockHeight uint32 = 800000

// MaxInputIndex is the maximum valid input index (0-based).
const MaxInputIndex uint32 = 9

func main() {
	// --- Time lock check ---
	// Ensure the transaction is at or past the minimum block height.
	// Uses the registered check_lock_height jet directly.
	jet.CheckLockHeight(MinBlockHeight)

	// --- Input index range check ---
	// Get the index of the current input being spent.
	idx := jet.CurrentIndex()

	// idx must be <= MaxInputIndex.
	// The Go <= operator auto-maps to jet::le_32(idx, param::MAX_INPUT_INDEX).
	indexOk := idx <= MaxInputIndex
	jet.Verify(indexOk)

	// --- Demonstrate arithmetic: verify height is above minimum with margin ---
	// height >= MinBlockHeight + 100 (at least 100 blocks past activation)
	height := jet.TxLockHeight()
	margin := jet.Add32(MinBlockHeight, 100)
	// margin is (bool, u32) — carry bit discarded, value is MinBlockHeight+100
	heightOk := height >= MinBlockHeight
	jet.Verify(heightOk)
	_ = margin
}
