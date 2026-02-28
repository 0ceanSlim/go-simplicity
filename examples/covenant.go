//go:build ignore

// Covenant contract
//
// This example demonstrates a simple covenant that:
// 1. Verifies that output 0's script hash matches a known expected value
// 2. Verifies the owner's BIP-340 signature
//
// This enforces that funds can only be sent to a specific script,
// preventing arbitrary redirection.
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const SIG: [u8; 64] = 0x00...00;
//	}
//	mod param {
//	    const EXPECTED_SCRIPT_HASH: u256 = 0x...;
//	    const OWNER_PUBKEY: u256 = 0x...;
//	    const OUTPUT_INDEX: u32 = 0;
//	}
//
//	fn main() {
//	    let hash: u256 = jet::output_script_hash(param::OUTPUT_INDEX);
//	    jet::eq_256(hash, param::EXPECTED_SCRIPT_HASH)
//	    let msg: u256 = jet::sig_all_hash();
//	    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/covenant.go
package main

import "simplicity/jet"

// ExpectedScriptHash is the SHA-256 hash of the required output script
const ExpectedScriptHash = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

// OwnerPubkey is the BIP-340 x-only public key of the covenant owner
const OwnerPubkey = 0xf9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9

// OutputIndex is the index of the output whose script hash is checked
const OutputIndex uint32 = 0

func main() {
	// Witness: owner signature provided at spending time
	var sig [64]byte

	// 1. Retrieve the script hash of output 0
	hash := jet.OutputScriptHash(OutputIndex)
	// 2. Enforce the output script matches the expected covenant target
	jet.Eq256(hash, ExpectedScriptHash)
	// 3. Verify the owner's signature
	msg := jet.SigAllHash()
	jet.BIP340Verify(OwnerPubkey, msg, sig)
}
