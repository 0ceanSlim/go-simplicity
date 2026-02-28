//go:build ignore

// Double SHA-256 contract
//
// This example demonstrates:
//  1. SHA-256 chaining: SHA256(SHA256(preimage)) compared against a stored hash.
//  2. SHA256Add auto-select: jet.SHA256Add resolves to sha_256_ctx_8_add_32 at
//     transpile time because the argument type is [32]byte / u256.
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const PREIMAGE: [u8; 32] = 0x00...00;
//	    const SIG: [u8; 64] = 0x00...00;
//	}
//	mod param {
//	    const HASH_LOCK: u256 = 0x...;
//	    const OWNER_PUBKEY: u256 = 0x...;
//	}
//
//	fn main() {
//	    let inner_hash: u256 = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), witness::PREIMAGE));
//	    let outer_hash: u256 = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), inner_hash));
//	    jet::eq_256(outer_hash, param::HASH_LOCK)
//	    let msg: u256 = jet::sig_all_hash();
//	    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/double_sha256.go
package main

import "simplicity/jet"

// HashLock is the double-SHA256 of the expected preimage.
const HashLock = 0xb472a266d0bd89c13706a4132ccfb16f7c3b9fcbe4de92ac37d421b7a0cb7e22

// OwnerPubkey is the BIP-340 x-only public key of the contract owner.
const OwnerPubkey = 0xf9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9

func main() {
	// Witness: 32-byte preimage and owner signature provided at spending time
	var preimage [32]byte
	var sig [64]byte

	// First SHA-256 pass: SHA256(preimage)
	innerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), preimage))

	// Second SHA-256 pass: SHA256(innerHash)
	// jet.SHA256Add auto-selects sha_256_ctx_8_add_32 because innerHash is u256.
	outerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), innerHash))

	// Verify the double-SHA256 result matches the stored hash lock
	jet.Eq256(outerHash, HashLock)

	// Verify the owner's signature
	msg := jet.SigAllHash()
	jet.BIP340Verify(OwnerPubkey, msg, sig)
}
