//go:build ignore

// Vault contract
//
// This example demonstrates a vault with two spending paths:
//  1. Left (hot key): immediate spend with hot key signature
//  2. Right (cold key): timelocked recovery — block height check, script hash
//     enforcement, and cold key signature
//
// This pattern protects funds in a vault: routine spends use the hot key,
// while the cold key can only recover funds after a lockup period and only
// to the pre-committed vault recovery script.
//
// The witness is an Either type:
// - Left: hot_key_sig for immediate spend
// - Right: cold_key_sig for timelocked recovery
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const W: Either<[u8; 64], [u8; 64]> = Left(0x00...00);
//	}
//	mod param {
//	    const HOT_KEY_PUBKEY: u256 = 0x...;
//	    const COLD_KEY_PUBKEY: u256 = 0x...;
//	    const COLD_KEY_UNLOCK: u32 = 1000;
//	    const VAULT_SCRIPT: u256 = 0x...;
//	    const VAULT_OUTPUT_INDEX: u32 = 0;
//	}
//
//	fn main() {
//	    match witness::W {
//	        Left(data) => {
//	            let msg: u256 = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::HOT_KEY_PUBKEY, msg), data)
//	        },
//	        Right(sig) => {
//	            jet::check_lock_height(param::COLD_KEY_UNLOCK)
//	            let script_hash: u256 = jet::output_script_hash(param::VAULT_OUTPUT_INDEX);
//	            jet::eq_256(script_hash, param::VAULT_SCRIPT)
//	            let msg: u256 = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::COLD_KEY_PUBKEY, msg), sig)
//	        },
//	    }
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/vault.go
package main

import "simplicity/jet"

// HotKeyPubkey is the BIP-340 x-only public key for routine (hot) spends
const HotKeyPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// ColdKeyPubkey is the BIP-340 x-only public key for cold key recovery
const ColdKeyPubkey = 0xdff1d77f2a671c5f36183726db2341be58feae1da2deced843240f7b502ba659

// ColdKeyUnlock is the minimum block height at which cold key recovery is allowed
const ColdKeyUnlock uint32 = 1000

// VaultScript is the SHA-256 hash of the required recovery output script
const VaultScript = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

// VaultOutputIndex is the output index whose script hash is enforced during recovery
const VaultOutputIndex uint32 = 0

// VaultWitness represents the witness data for the vault contract.
// IsLeft=true:  hot key path — immediate spend with hot key sig
// IsLeft=false: cold key path — timelocked recovery with script hash check
type VaultWitness struct {
	IsLeft     bool
	HotKeySig  [64]byte // Left: immediate hot key spend
	ColdKeySig [64]byte // Right: timelocked cold key recovery
}

func main() {
	var w VaultWitness

	if w.IsLeft {
		// Hot key path: immediate spend — no conditions other than signature
		msg := jet.SigAllHash()
		jet.BIP340Verify(HotKeyPubkey, msg, w.HotKeySig)
	} else {
		// Cold key path: timelocked recovery with script hash enforcement
		// 1. Enforce the timelock — cold key cannot act before ColdKeyUnlock
		jet.CheckLockHeight(ColdKeyUnlock)
		// 2. Enforce output 0 sends to the pre-committed vault recovery script
		scriptHash := jet.OutputScriptHash(VaultOutputIndex)
		jet.Eq256(scriptHash, VaultScript)
		// 3. Verify cold key signature
		msg := jet.SigAllHash()
		jet.BIP340Verify(ColdKeyPubkey, msg, w.ColdKeySig)
	}
}
