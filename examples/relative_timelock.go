//go:build ignore

// Relative Timelock contract
//
// This example demonstrates a CSV-style relative timelock using
// CheckLockDistance (block-based). The funds can only be spent after
// a minimum number of blocks have been mined since the funding transaction
// was confirmed.
//
// Variants (uncomment to switch):
//
//	jet.CheckLockDistance(RelativeLockBlocks)   // block-based (used below)
//	jet.CheckLockDuration(RelativeLockSeconds)  // time-based (512-second units)
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const SIG: [u8; 64] = 0x00...00;
//	}
//	mod param {
//	    const SENDER_PUBKEY: u256 = 0x...;
//	    const RELATIVE_LOCK_BLOCKS: u16 = 10;
//	}
//
//	fn main() {
//	    jet::check_lock_distance(param::RELATIVE_LOCK_BLOCKS)
//	    let msg: u256 = jet::sig_all_hash();
//	    jet::bip_0340_verify((param::SENDER_PUBKEY, msg), witness::SIG)
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/relative_timelock.go
package main

import "simplicity/jet"

// SenderPubkey is the BIP-340 x-only public key of the funds sender
const SenderPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// RelativeLockBlocks is the minimum number of blocks that must have been
// mined since the funding transaction (CSV-style relative timelock)
const RelativeLockBlocks uint16 = 10

func main() {
	// Witness: sender signature provided at spending time
	var sig [64]byte

	// 1. Enforce the relative timelock — at least RelativeLockBlocks mined
	//    since the funding UTXO was confirmed.
	//    For time-based: jet.CheckLockDuration(RelativeLockSeconds)
	jet.CheckLockDistance(RelativeLockBlocks)

	// 2. Verify the sender's signature
	msg := jet.SigAllHash()
	jet.BIP340Verify(SenderPubkey, msg, sig)
}
