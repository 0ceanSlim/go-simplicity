//go:build ignore

// AMM Pool Invariant Contract
//
// Minimal AMM constant-product invariant check — standalone, no anchor deps.
// Tests that simgo emits unwrap/unwrap_right and borrow-arithmetic instead of
// match/le_128 helpers that produce CASE nodes (anti-DOS weight check failures).
//
// This example exercises:
//   - jet.InputAmount(i)  — amountPairOpt (Bug 1: match → unwrap)
//   - jet.OutputAmount(i) — amountPairOpt (Bug 1: match → unwrap)
//   - jet.OutputAsset(i)  — assetOpt      (Bug 1: match → unwrap_right)
//   - jet.Verify(jet.Le128(kOld, kNew)) — inline borrow-arithmetic (Bug 2+3)
//
// Usage:
//   go run cmd/simgo/main.go -input examples/amm_pool.go

package main

import "simplicity/jet"

// Pool input/output indices.
const PoolInputA uint32 = 0
const PoolInputB uint32 = 1
const PoolOutputA uint32 = 0
const PoolOutputB uint32 = 1

func main() {
	// Read reserves from pool input UTXOs.
	reserve0 := jet.InputAmount(PoolInputA)
	reserve1 := jet.InputAmount(PoolInputB)

	// Read output amounts and asset IDs.
	newReserve0 := jet.OutputAmount(PoolOutputA)
	newReserve1 := jet.OutputAmount(PoolOutputB)
	asset0 := jet.OutputAsset(PoolOutputA)
	_ = asset0

	// k-invariant: k_new >= k_old (constant product must not decrease).
	// Le128(kOld, kNew) is emitted as borrow-arithmetic, zero CASE nodes.
	kOld := jet.Multiply64(reserve0, reserve1)
	kNew := jet.Multiply64(newReserve0, newReserve1)
	jet.Verify(jet.Le128(kOld, kNew))
}
