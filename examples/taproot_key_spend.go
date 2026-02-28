//go:build ignore

// Taproot Key Spend contract
//
// This example demonstrates Taproot introspection: verifying the internal
// key and tapleaf version match expected compile-time constants before
// authorising a spend via BIP-340 signature.
//
// Use case: contracts that must assert they are executing in a specific
// Taproot context — e.g. ensuring the correct internal key was committed
// to the Taproot output and the tapleaf uses the expected version byte.
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const SIG: [u8; 64] = 0x00...00;
//	}
//	mod param {
//	    const EXPECTED_INTERNAL_KEY: u256 = 0x...;
//	    const EXPECTED_TAPLEAF_VERSION: u8 = 0xc0;
//	    const OWNER_PUBKEY: u256 = 0x...;
//	}
//
//	fn main() {
//	    let key: u256 = jet::internal_key();
//	    jet::eq_256(key, param::EXPECTED_INTERNAL_KEY)
//	    let version: u8 = jet::tapleaf_version();
//	    jet::eq_8(version, param::EXPECTED_TAPLEAF_VERSION)
//	    let msg: u256 = jet::sig_all_hash();
//	    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/taproot_key_spend.go
package main

import "simplicity/jet"

// ExpectedInternalKey is the required Taproot internal key for this contract
const ExpectedInternalKey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// ExpectedTapleafVersion is the required tapleaf version byte (0xc0 = standard Tapscript)
const ExpectedTapleafVersion uint8 = 0xc0

// OwnerPubkey is the BIP-340 x-only public key of the contract owner
const OwnerPubkey = 0xdff1d77f2a671c5f36183726db2341be58feae1da2deced843240f7b502ba659

func main() {
	// Witness: owner signature provided at spending time
	var sig [64]byte

	// 1. Assert the Taproot internal key matches the expected commitment
	key := jet.InternalKey()
	jet.Eq256(key, ExpectedInternalKey)

	// 2. Assert the tapleaf version is the standard Tapscript version (0xc0)
	version := jet.TapleafVersion()
	jet.Eq8(version, ExpectedTapleafVersion)

	// 3. Verify the owner's BIP-340 signature
	msg := jet.SigAllHash()
	jet.BIP340Verify(OwnerPubkey, msg, sig)
}
