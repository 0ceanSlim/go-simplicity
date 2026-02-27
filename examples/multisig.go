//go:build ignore

// Multisig (2-of-3) Example
//
// This example demonstrates a 2-of-3 multisig contract that:
// 1. Accepts up to 3 optional signatures
// 2. Verifies each provided signature against its corresponding pubkey
// 3. Requires at least 2 valid signatures to succeed
//
// The witness contains optional signatures - Some(sig) for provided, None for absent
//
// Expected SimplicityHL output:
//
//   mod witness {
//       const SIG_0: Option<[u8; 64]> = /* witness */;
//       const SIG_1: Option<[u8; 64]> = /* witness */;
//       const SIG_2: Option<[u8; 64]> = /* witness */;
//   }
//   mod param {
//       const ALICE_PUBKEY: u256 = 0x...;
//       const BOB_PUBKEY: u256 = 0x...;
//       const CHARLIE_PUBKEY: u256 = 0x...;
//   }
//
//   fn main() {
//       let msg: u256 = jet::sig_all_hash();
//
//       // Unrolled signature checks with counter accumulation
//       let count_0: u32 = match witness::SIG_0 {
//           Some(sig) => { jet::bip_0340_verify((param::ALICE_PUBKEY, msg), sig); 1 },
//           None => 0,
//       };
//       let count_1: u32 = count_0 + match witness::SIG_1 {
//           Some(sig) => { jet::bip_0340_verify((param::BOB_PUBKEY, msg), sig); 1 },
//           None => 0,
//       };
//       let count_2: u32 = count_1 + match witness::SIG_2 {
//           Some(sig) => { jet::bip_0340_verify((param::CHARLIE_PUBKEY, msg), sig); 1 },
//           None => 0,
//       };
//
//       // Require at least 2 valid signatures
//       jet::verify(jet::le_32(2, count_2))
//   }
//
// Usage:
//   go run cmd/simgo/main.go -input examples/multisig.go

package main

import "simplicity/jet"

// Pubkeys for the 3 signers
const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const BobPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const CharliePubkey = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

// OptionalSig represents an optional signature
type OptionalSig struct {
	IsSome bool
	Value  [64]byte
}

func main() {
	// Three optional signatures
	var sig0 OptionalSig
	var sig1 OptionalSig
	var sig2 OptionalSig

	// Message to be signed
	msg := jet.SigAllHash()

	// Track valid signature count
	validCount := 0

	// Check signature 0 (Alice)
	if sig0.IsSome {
		jet.BIP340Verify(AlicePubkey, msg, sig0.Value)
		validCount++
	}

	// Check signature 1 (Bob)
	if sig1.IsSome {
		jet.BIP340Verify(BobPubkey, msg, sig1.Value)
		validCount++
	}

	// Check signature 2 (Charlie)
	if sig2.IsSome {
		jet.BIP340Verify(CharliePubkey, msg, sig2.Value)
		validCount++
	}

	// Require at least 2 valid signatures
	jet.Verify(jet.Le32(2, validCount))
}
