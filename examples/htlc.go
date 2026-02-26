//go:build ignore

// HTLC (Hash Time Locked Contract) Example
//
// This example demonstrates a Hash Time Locked Contract that:
// 1. Allows the recipient to claim funds by revealing a preimage and signature (Left path)
// 2. Allows the sender to reclaim funds after timeout with just a signature (Right path)
//
// The witness is an Either type:
// - Left: (preimage, recipient_signature) for successful completion
// - Right: sender_signature for cancellation/timeout
//
// Expected SimplicityHL output:
//
//   mod witness {
//       const HTLC_WITNESS: Either<(u256, [u8; 64]), [u8; 64]> = /* witness */;
//   }
//   mod param {
//       const RECIPIENT_PUBKEY: u256 = 0x...;
//       const SENDER_PUBKEY: u256 = 0x...;
//       const HASH_LOCK: u256 = 0x...;
//   }
//
//   fn main() {
//       match witness::HTLC_WITNESS {
//           Left(data) => {
//               let (preimage, sig): (u256, [u8; 64]) = data;
//               let hash: u256 = jet::sha_256(preimage);
//               jet::eq_256(hash, param::HASH_LOCK);
//               jet::bip_0340_verify((param::RECIPIENT_PUBKEY, jet::sig_all_hash()), sig)
//           },
//           Right(sig) => {
//               jet::bip_0340_verify((param::SENDER_PUBKEY, jet::sig_all_hash()), sig)
//           },
//       }
//   }
//
// Usage:
//   go run cmd/simgo/main.go -input examples/htlc.go

package main

import "simplicity/jet"

// RecipientPubkey is the BIP-340 x-only public key for the recipient (Alice)
const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// SenderPubkey is the BIP-340 x-only public key for the sender (Bob)
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4

// HashLock is the SHA-256 hash that must be revealed to claim funds
const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

// HTLCWitness represents the witness data for the HTLC
// IsLeft=true: Complete path with preimage and recipient signature
// IsLeft=false: Cancel path with sender signature
type HTLCWitness struct {
	IsLeft bool
	// Left path: preimage (32 bytes) + signature (64 bytes)
	Preimage [32]byte
	RecipientSig [64]byte
	// Right path: sender signature (64 bytes)
	SenderSig [64]byte
}

func main() {
	var w HTLCWitness

	if w.IsLeft {
		// Complete path: recipient claims with preimage + signature
		// 1. Hash the preimage
		hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), w.Preimage))
		// 2. Verify hash matches the lock
		jet.Eq256(hash, HashLock)
		// 3. Verify recipient's signature
		msg := jet.SigAllHash()
		jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
	} else {
		// Cancel path: sender reclaims with signature (after timeout)
		msg := jet.SigAllHash()
		jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
	}
}
