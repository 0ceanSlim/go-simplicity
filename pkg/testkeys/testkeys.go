// Package testkeys provides hardcoded BIP-340 test vectors and SHA-256 test
// pairs for verifying that generated SimplicityHL code is executable.
//
// These values are taken directly from the BIP-340 specification:
// https://github.com/bitcoin/bips/blob/master/bip-0340/test-vectors.csv
//
// All signing test vectors (result = TRUE) are cryptographically proven
// correct by the Bitcoin specification review process.
//
// WHY THIS EXISTS:
// The go-simplicity transpiler generates `mod witness { ... }` blocks with
// zero-value placeholder signatures (0x000...000). A zero value is NOT a valid
// BIP-340 Schnorr signature — it fails at runtime in any Simplicity evaluator
// or playground. These test vectors let you build contracts that actually
// execute, useful for:
//   - Playground verification (paste output directly and it runs)
//   - CI end-to-end testing against a Simplicity evaluator
//   - Confirming generated code structure is semantically correct
//
// FOR PRODUCTION: never use these keys. Sign the real sig_all_hash output
// with your own private key.
//
// NOTE ON sig_all_hash:
// In production contracts, BIP-340 signatures cover the transaction sighash
// returned by jet::sig_all_hash(). In the testable examples, we instead use
// a FIXED known message constant so the pre-computed signature is valid. This
// trades transaction-binding for executability — acceptable for testing, not
// for production.
package testkeys

// BIP-340 Test Vector 0
//
// Source: bip-0340/test-vectors.csv row 0 (result=TRUE)
// Secret key: 0000000000000000000000000000000000000000000000000000000000000003
const (
	Vector0Pubkey = "F9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9"
	Vector0Msg    = "0000000000000000000000000000000000000000000000000000000000000000"
	Vector0Sig    = "E907831F80848D1069A5371B402410364BDF1C5F8307B0084C55F1CE2DCA821525F66A4A85EA8B71E482A74F382D2CE5EBEEE8FDB2172F477DF4900D310536C0"
)

// BIP-340 Test Vector 1
//
// Source: bip-0340/test-vectors.csv row 1 (result=TRUE)
// Secret key: B7E151628AED2A6ABF7158809CF4F3C762E7160F38B4DA56A784D9045190CFEF
const (
	Vector1Pubkey = "DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"
	Vector1Msg    = "243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"
	Vector1Sig    = "6896BD60EEAE296DB48A229FF71DFE071BDE413E6D43F917DC8DCF8C78DE33418906D11AC976ABCCB20B091292BFF4EA897EFCB639EA871CFA95F6DE339E4B0A"
)

// BIP-340 Test Vector 2
//
// Source: bip-0340/test-vectors.csv row 2 (result=TRUE)
// Secret key: C90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B14E5C9
const (
	Vector2Pubkey = "DD308AFEC5777E13121FA72B9CC1B7CC0139715309B086C960E18FD969774EB8"
	Vector2Msg    = "7E2D58D8B3BCDF1ABADEC7829054F90DDA9805AAB56C77333024B9D0A508B75C"
	Vector2Sig    = "5831AAEED7B44BB74E5EAB94BA9D4294C49BCF2A60728D8B4C200F50DD313C1BAB745879A5AD954A72C45A91C3A51D3C7ADEA98D82F8481E0E1E03674A6F3FB7"
)

// Convenience aliases using "Alice" and "Bob" role names.
//
// Alice = Vector0 (signs the all-zero 32-byte message).
// Bob   = Vector1 (signs a non-trivial message).
const (
	// AlicePubkey is Alice's x-only BIP-340 public key.
	AlicePubkey = Vector0Pubkey
	// AliceMsg is the 32-byte message that AliceSig proves knowledge of.
	AliceMsg = Vector0Msg
	// AliceSig is a valid BIP-340 Schnorr signature by Alice over AliceMsg.
	AliceSig = Vector0Sig

	// BobPubkey is Bob's x-only BIP-340 public key.
	BobPubkey = Vector1Pubkey
	// BobMsg is the 32-byte message that BobSig proves knowledge of.
	BobMsg = Vector1Msg
	// BobSig is a valid BIP-340 Schnorr signature by Bob over BobMsg.
	BobSig = Vector1Sig
)

// SHA-256 test pairs for HTLC hashlock testing.
//
// Verified with Go's crypto/sha256 package.
const (
	// PreimageAllZero is 32 zero bytes — the simplest possible preimage.
	PreimageAllZero = "0000000000000000000000000000000000000000000000000000000000000000"

	// SHA256OfAllZero is SHA-256(0x000...000, 32 bytes).
	// Use this as the HashLock constant when the witness preimage is PreimageAllZero.
	// Verified: Go sha256.Sum256(make([]byte, 32)) == 66687AAD...
	SHA256OfAllZero = "66687AADF862BD776C8FC18B8E9F8E20089714856EE233B3902A591D0D5F2925"
)

// WithHexPrefix returns a test vector constant with a 0x prefix, ready for
// use as a Go hex literal or in generated SimplicityHL output.
func WithHexPrefix(hex string) string {
	return "0x" + hex
}
