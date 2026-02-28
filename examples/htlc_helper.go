//go:build ignore

// HTLC with helper function and switch dispatch
//
// This example demonstrates:
// 1. A helper function (verifyHashlock) that is both emitted and inlined at its call site.
// 2. A tagless switch {} used as sum-type dispatch instead of if/else.
//
// Expected SimplicityHL output shape:
//
//	mod witness {
//	    const W: Either<([u8; 32], [u8; 64]), [u8; 64]> = Left(...);
//	}
//	mod param {
//	    const RECIPIENT_PUBKEY: u256 = 0x...;
//	    const SENDER_PUBKEY: u256 = 0x...;
//	    const HASH_LOCK: u256 = 0x...;
//	    const MIN_REFUND_HEIGHT: u32 = 800000;
//	}
//
//	fn verify_hashlock(preimage: [u8; 32]) {
//	    let hash = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), preimage));
//	    jet::eq_256(hash, param::HASH_LOCK)
//	}
//
//	fn main() {
//	    match witness::W {
//	        Left(data) => {
//	            let (preimage, recipient_sig): ([u8; 32], [u8; 64]) = data;
//	            let hash = jet::sha_256_ctx_8_finalize(...preimage...);
//	            jet::eq_256(hash, param::HASH_LOCK)
//	            let msg = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::RECIPIENT_PUBKEY, msg), recipient_sig)
//	        },
//	        Right(sig) => {
//	            jet::check_lock_height(param::MIN_REFUND_HEIGHT)
//	            let msg = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::SENDER_PUBKEY, msg), sig)
//	        }
//	    }
//	}
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
const MinRefundHeight uint32 = 800000

// AtomicSwapWitness represents the witness data for the HTLC.
// IsLeft=true:  recipient claim path — preimage + recipient signature
// IsLeft=false: sender refund path — sender signature (after block height)
type AtomicSwapWitness struct {
	IsLeft       bool
	Preimage     [32]byte
	RecipientSig [64]byte
	SenderSig    [64]byte
}

// verifyHashlock checks that preimage hashes to the expected HashLock constant.
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
