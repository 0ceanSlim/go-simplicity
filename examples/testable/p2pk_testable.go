//go:build ignore

// P2PK Testable Example
//
// This is the PLAYGROUND-EXECUTABLE version of the P2PK contract.
// It uses BIP-340 test vector #0 from the Bitcoin specification so that
// the generated SimplicityHL is immediately executable in the playground
// with no manual witness substitution needed.
//
// KEY DIFFERENCES from examples/p2pk.go (production):
//   1. Uses a FIXED known message instead of jet.SigAllHash()
//      (sig_all_hash binds to a real transaction — we use a constant for testability)
//   2. The signature is a param constant (compile-time known) rather than a
//      runtime witness — valid for execution testing, not for production
//
// BIP-340 Test Vector #0 (secret key = 0x000...0003):
//   pubkey:  F9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9
//   message: 0000000000000000000000000000000000000000000000000000000000000000
//   sig:     E907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0
//
// Expected SimplicityHL output (paste directly into playground — it WILL execute):
//
//   mod witness {
//   }
//   mod param {
//       const ALICE_PUBKEY: u256 = 0xF9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9;
//       const TEST_MSG: u256 = 0x0000000000000000000000000000000000000000000000000000000000000000;
//       const ALICE_TEST_SIG: [u8; 64] = 0xE907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0;
//   }
//
//   fn main() {
//       jet::bip_0340_verify((param::ALICE_PUBKEY, param::TEST_MSG), param::ALICE_TEST_SIG)
//   }
//
// Usage:
//   go run cmd/simgo/main.go -input examples/testable/p2pk_testable.go

package main

import "simplicity/jet"

// AlicePubkey is the BIP-340 test vector #0 x-only public key.
// Corresponding secret key: 0x0000000000000000000000000000000000000000000000000000000000000003
const AlicePubkey = 0xF9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9

// TestMsg is the fixed 32-byte message from BIP-340 test vector #0.
// In a production P2PK contract, use jet.SigAllHash() instead to bind
// the signature to the spending transaction.
const TestMsg = 0x0000000000000000000000000000000000000000000000000000000000000000

// AliceTestSig is the pre-computed BIP-340 Schnorr signature by Alice over TestMsg.
// This is a COMPILE-TIME test constant (in mod param) rather than a runtime
// witness — valid only for execution testing.
const AliceTestSig = 0xE907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0

func main() {
	// Verify Alice's pre-computed test signature against the known test message.
	// This executes correctly in the playground because AliceTestSig is a valid
	// BIP-340 signature (from the Bitcoin specification test vectors).
	jet.BIP340Verify(AlicePubkey, TestMsg, AliceTestSig)
}
