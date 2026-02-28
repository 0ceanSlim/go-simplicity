//go:build ignore

// Atomic Swap contract
//
// This example demonstrates a real atomic swap using:
// 1. Left path: Alice claims with SHA-256 preimage + BIP-340 signature
// 2. Right path: Bob refunds after block height with BIP-340 signature
//
// The witness is an Either type:
// - Left: (preimage, recipient_signature) for successful completion
// - Right: sender_signature for refund after MinRefundHeight
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const W: Either<([u8; 32], [u8; 64]), [u8; 64]> = Left((...));
//	}
//	mod param {
//	    const RECIPIENT_PUBKEY: u256 = 0x...;
//	    const SENDER_PUBKEY: u256 = 0x...;
//	    const HASH_LOCK: u256 = 0x...;
//	    const MIN_REFUND_HEIGHT: u32 = 800000;
//	}
//
//	fn main() {
//	    match witness::W {
//	        Left(data) => {
//	            let (preimage, recipient_sig): ([u8; 32], [u8; 64]) = data;
//	            let hash = jet::sha_256_ctx_8_finalize(...);
//	            jet::eq_256(hash, param::HASH_LOCK)
//	            let msg = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::RECIPIENT_PUBKEY, msg), recipient_sig)
//	        },
//	        Right(sig) => {
//	            jet::check_lock_height(param::MIN_REFUND_HEIGHT)
//	            let msg = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::SENDER_PUBKEY, msg), sig)
//	        },
//	    }
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/atomic_swap.go
package main

import "simplicity/jet"

// RecipientPubkey is the BIP-340 x-only public key for the recipient (Alice)
const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// SenderPubkey is the BIP-340 x-only public key for the sender (Bob)
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4

// HashLock is the SHA-256 hash that must be revealed to claim funds
const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

// MinRefundHeight is the minimum block height at which Bob may refund
const MinRefundHeight uint32 = 800000

// AtomicSwapWitness represents the witness data for the atomic swap
// IsLeft=true:  Alice's claim path — preimage + recipient signature
// IsLeft=false: Bob's refund path — sender signature (after block height)
type AtomicSwapWitness struct {
	IsLeft bool
	// Left path: preimage (32 bytes) + Alice's signature (64 bytes)
	Preimage     [32]byte
	RecipientSig [64]byte
	// Right path: Bob's signature (64 bytes)
	SenderSig [64]byte
}

func main() {
	var w AtomicSwapWitness

	if w.IsLeft {
		// Claim path: Alice reveals preimage and provides signature
		// 1. Hash the preimage
		hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), w.Preimage))
		// 2. Verify hash matches the lock
		jet.Eq256(hash, HashLock)
		// 3. Verify Alice's signature
		msg := jet.SigAllHash()
		jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
	} else {
		// Refund path: Bob reclaims after the lock height with signature
		jet.CheckLockHeight(MinRefundHeight)
		msg := jet.SigAllHash()
		jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
	}
}
