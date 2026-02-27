//go:build ignore

// HTLC Testable Example
//
// This is the PLAYGROUND-EXECUTABLE version of the HTLC contract.
// It uses BIP-340 test vectors and a known SHA-256 preimage/hash pair
// so the generated SimplicityHL is immediately executable.
//
// KEY DIFFERENCES from examples/htlc.go (production):
//   1. Uses fixed known messages instead of jet.SigAllHash()
//   2. Signatures are param constants (compile-time) rather than runtime witnesses
//   3. This is ONLY for verifying the transpiler output executes correctly
//
// Test credentials:
//   Alice (recipient) — BIP-340 vector #0:
//     pubkey:  F9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9
//     msg:     0000000000000000000000000000000000000000000000000000000000000000
//     sig:     E907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0
//
//   Bob (sender) — BIP-340 vector #1:
//     pubkey:  DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659
//     msg:     243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89
//     sig:     6896BD60EEAE296DB48A229FF71DFE071BDE413E6D43F917DC8DCF8C78DE33418906D11AC976ABCCB20B091292BFF4EA897EFCB639EA871CFA95F6DE339E4B0A
//
//   HTLC hashlock:
//     preimage: 0x0000000000000000000000000000000000000000000000000000000000000000
//     hashlock: 0x66687AADF862BD776C8FC18B8E9F8E20089714856EE233B3902A591D0D5F2925
//              (SHA-256 of 32 zero bytes — verified with Go crypto/sha256)
//
// Usage:
//   go run cmd/simgo/main.go -input examples/testable/htlc_testable.go

package main

import "simplicity/jet"

// RecipientPubkey is Alice's test x-only public key (BIP-340 vector #0).
const RecipientPubkey = 0xF9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9

// SenderPubkey is Bob's test x-only public key (BIP-340 vector #1).
const SenderPubkey = 0xDFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659

// HashLock is SHA-256(0x000...000) — the SHA-256 of 32 zero bytes.
// Verified: Go crypto/sha256.Sum256(make([]byte, 32))
const HashLock = 0x66687AADF862BD776C8FC18B8E9F8E20089714856EE233B3902A591D0D5F2925

// RecipientTestMsg is the fixed message Alice signs (BIP-340 vector #0 message).
// In production, use jet.SigAllHash() instead.
const RecipientTestMsg = 0x0000000000000000000000000000000000000000000000000000000000000000

// SenderTestMsg is the fixed message Bob signs (BIP-340 vector #1 message).
// In production, use jet.SigAllHash() instead.
const SenderTestMsg = 0x243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89

// AliceTestSig is Alice's pre-computed signature over RecipientTestMsg.
const AliceTestSig = 0xE907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0

// BobTestSig is Bob's pre-computed signature over SenderTestMsg.
const BobTestSig = 0x6896BD60EEAE296DB48A229FF71DFE071BDE413E6D43F917DC8DCF8C78DE33418906D11AC976ABCCB20B091292BFF4EA897EFCB639EA871CFA95F6DE339E4B0A

// HTLCWitness represents the witness data for the testable HTLC.
// IsLeft=true:  Complete path — reveal preimage and get recipient's sig.
// IsLeft=false: Cancel path  — sender reclaims with signature.
type HTLCWitness struct {
	IsLeft bool
	// Left path: preimage (all zeros matches HashLock) + recipient signature
	Preimage     [32]byte
	RecipientSig [64]byte
	// Right path: sender signature
	SenderSig [64]byte
}

func main() {
	var w HTLCWitness

	if w.IsLeft {
		// Complete path: verify preimage hashes to HashLock, then verify sig.
		hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), w.Preimage))
		jet.Eq256(hash, HashLock)
		jet.BIP340Verify(RecipientPubkey, RecipientTestMsg, AliceTestSig)
	} else {
		// Cancel path: sender reclaims.
		jet.BIP340Verify(SenderPubkey, SenderTestMsg, BobTestSig)
	}
}
