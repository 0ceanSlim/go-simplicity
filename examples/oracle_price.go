//go:build ignore

// Oracle Price contract
//
// This example demonstrates an oracle-gated spending pattern with two paths:
// 1. Left (oracle): a trusted oracle's BIP-340 signature authorises the spend
// 2. Right (owner): emergency owner withdrawal without oracle involvement
//
// Use case: a contract where a trusted price oracle must attest that
// transaction conditions are met (e.g. asset price is acceptable),
// with an emergency withdrawal path reserved for the owner.
//
// The witness is an Either type:
// - Left: oracle_sig — oracle attests the transaction is acceptable
// - Right: owner_sig — emergency owner withdrawal
//
// Expected SimplicityHL output:
//
//	mod witness {
//	    const W: Either<[u8; 64], [u8; 64]> = Left(0x00...00);
//	}
//	mod param {
//	    const ORACLE_PUBKEY: u256 = 0x...;
//	    const OWNER_PUBKEY: u256 = 0x...;
//	}
//
//	fn main() {
//	    match witness::W {
//	        Left(data) => {
//	            let msg: u256 = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::ORACLE_PUBKEY, msg), data)
//	        },
//	        Right(sig) => {
//	            let msg: u256 = jet::sig_all_hash();
//	            jet::bip_0340_verify((param::OWNER_PUBKEY, msg), sig)
//	        },
//	    }
//	}
//
// Usage:
//
//	go run cmd/simgo/main.go -input examples/oracle_price.go
package main

import "simplicity/jet"

// OraclePubkey is the BIP-340 x-only public key of the trusted price oracle
const OraclePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

// OwnerPubkey is the BIP-340 x-only public key of the contract owner
const OwnerPubkey = 0xdff1d77f2a671c5f36183726db2341be58feae1da2deced843240f7b502ba659

// OracleWitness represents the witness data for the oracle price contract.
// IsLeft=true:  oracle path — oracle signature authorises the transaction
// IsLeft=false: owner path — emergency owner withdrawal
type OracleWitness struct {
	IsLeft    bool
	OracleSig [64]byte // Left: oracle attests transaction is acceptable
	OwnerSig  [64]byte // Right: emergency owner withdrawal
}

func main() {
	var w OracleWitness

	if w.IsLeft {
		// Oracle path: the oracle's signature authorises this transaction.
		// The oracle is trusted to only sign transactions meeting price conditions.
		msg := jet.SigAllHash()
		jet.BIP340Verify(OraclePubkey, msg, w.OracleSig)
	} else {
		// Owner emergency path: owner bypasses the oracle for emergency withdrawal
		msg := jet.SigAllHash()
		jet.BIP340Verify(OwnerPubkey, msg, w.OwnerSig)
	}
}
