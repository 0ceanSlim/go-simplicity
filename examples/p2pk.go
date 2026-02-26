//go:build ignore

// P2PK (Pay-to-Public-Key) Contract Example
//
// This example demonstrates a simple P2PK contract that:
// 1. Defines Alice's public key as a compile-time constant
// 2. Gets the transaction message hash using jet::sig_all_hash
// 3. Verifies Alice's BIP-340 Schnorr signature
//
// Expected SimplicityHL output:
//
//   mod witness {
//       const SIG: [u8; 64] = /* witness */;
//   }
//   mod param {
//       const ALICE_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
//   }
//
//   fn main() {
//       let msg: u256 = jet::sig_all_hash();
//       jet::bip_0340_verify((param::ALICE_PUBKEY, msg), witness::SIG)
//   }
//
// Usage:
//   go run cmd/simgo/main.go -input examples/p2pk.go

package main

import "simplicity/jet"

// AlicePubkey is the BIP-340 x-only public key for Alice
// In a real contract, this would be the actual public key
const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
	// Declare signature as witness data (provided at spending time)
	var sig [64]byte

	// Get the transaction sighash using the Simplicity jet
	msg := jet.SigAllHash()

	// Verify the signature against Alice's public key
	// This will fail the script if the signature is invalid
	jet.BIP340Verify(AlicePubkey, msg, sig)
}
