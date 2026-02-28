package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
)

// loadExample reads an example file, stripping the //go:build ignore tag so it
// can be compiled by the test compiler.
func loadExample(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read example %s: %v", path, err)
	}
	return strings.Replace(string(data), "//go:build ignore\n", "", 1)
}

func compileExample(t *testing.T, path string) string {
	t.Helper()
	src := loadExample(t, path)
	c := compiler.New(compiler.Config{Target: "simplicityhl"})
	out, err := c.Compile(src, path)
	if err != nil {
		t.Fatalf("compile %s: %v", path, err)
	}
	return out
}

// assertNoInvalidWitness fails the test if any witness value is still the invalid placeholder.
func assertNoInvalidWitness(t *testing.T, name, out string) {
	t.Helper()
	if strings.Contains(out, "/* witness */") {
		t.Errorf("%s: output contains invalid SimplicityHL syntax '/* witness */':\n%s", name, out)
	}
}

// TestExampleP2PK verifies the P2PK example compiles to correct SimplicityHL.
func TestExampleP2PK(t *testing.T) {
	out := compileExample(t, "../examples/p2pk.go")
	assertNoInvalidWitness(t, "p2pk", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"witness module", "mod witness {"},
		{"param module", "mod param {"},
		{"sig witness type", "const SIG: [u8; 64]"},
		{"sig witness zero value", "0x" + strings.Repeat("00", 64)},
		{"alice pubkey param", "ALICE_PUBKEY: u256"},
		{"main function", "fn main()"},
		{"sig_all_hash jet", "jet::sig_all_hash()"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		{"tuple argument", "(param::ALICE_PUBKEY, msg)"},
		{"sig witness in verify", "witness::SIG"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("p2pk: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// SHA-256 old names must NOT appear
	for _, bad := range []string{"sha_256_iv", "sha_256_block", "sha_256_finalize"} {
		if strings.Contains(out, bad) {
			t.Errorf("p2pk: found outdated jet name %q in output", bad)
		}
	}
}

// TestExampleHTLC verifies the HTLC example compiles to correct SimplicityHL.
func TestExampleHTLC(t *testing.T) {
	out := compileExample(t, "../examples/htlc.go")
	assertNoInvalidWitness(t, "htlc", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"Either witness type", "Either<([u8; 32], [u8; 64]), [u8; 64]>"},
		{"Either witness Left placeholder", "Left((0x" + strings.Repeat("00", 32) + ", 0x" + strings.Repeat("00", 64) + "))"},
		{"recipient pubkey param", "RECIPIENT_PUBKEY: u256"},
		{"sender pubkey param", "SENDER_PUBKEY: u256"},
		{"hash lock param", "HASH_LOCK: u256"},
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right(sig)"},
		{"destructuring", "let (preimage, recipient_sig):"},
		{"sha_256_ctx_8_init jet", "jet::sha_256_ctx_8_init()"},
		{"sha_256_ctx_8_add_32 jet", "jet::sha_256_ctx_8_add_32("},
		{"sha_256_ctx_8_finalize jet", "jet::sha_256_ctx_8_finalize("},
		{"eq_256 jet", "jet::eq_256("},
		{"bound variable preimage used", "preimage"},
		{"bound variable recipient_sig used", "recipient_sig"},
		{"bound variable sig used in Right", "sig)"},
		{"no witness field in Left arm body", "recipient_sig)"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("htlc: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// Old SHA-256 jet names must NOT appear
	for _, bad := range []string{"sha_256_iv", "sha_256_block", "sha_256_finalize("} {
		if strings.Contains(out, bad) {
			t.Errorf("htlc: found outdated jet name %q in output", bad)
		}
	}

	// witness::W.field must NOT appear inside match arm bodies (bound vars should be used)
	if strings.Contains(out, "witness::W.preimage") {
		t.Error("htlc: Left arm should use 'preimage' not 'witness::W.preimage'")
	}
	if strings.Contains(out, "witness::W.recipient_sig") {
		t.Error("htlc: Left arm should use 'recipient_sig' not 'witness::W.recipient_sig'")
	}
	if strings.Contains(out, "witness::W.sender_sig") {
		t.Error("htlc: Right arm should use 'sig' not 'witness::W.sender_sig'")
	}
}

// TestExampleP2PKTestable verifies the testable P2PK example compiles with real BIP-340 test vectors.
// The generated output should be immediately executable in the SimplicityHL playground
// with no manual witness substitution required.
func TestExampleP2PKTestable(t *testing.T) {
	out := compileExample(t, "../examples/testable/p2pk_testable.go")

	// Witness module must be empty — the sig is a param constant, not a runtime witness.
	if strings.Contains(out, "const SIG:") {
		t.Error("p2pk_testable: should not have a runtime SIG witness; sig is a param constant")
	}

	checks := []struct {
		desc    string
		present string
	}{
		{"param module", "mod param {"},
		// Real BIP-340 test vector #0 pubkey (NOT a fake zero pubkey)
		{"real alice pubkey", "0xf9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9"},
		// Real test message (all zeros is actually the BIP-340 test vector message)
		{"test message param", "TEST_MSG: u256"},
		// Real BIP-340 test vector #0 signature (starts with e907831f...)
		{"real alice sig", "0xe907831f80848d1069a5371b402410364bdf1c5f8307b0084c55f1ce2dca8215"},
		// Sig is in param (compile-time), not witness (runtime)
		{"sig as param constant", "ALICE_TEST_SIG: [u8; 64]"},
		{"main function", "fn main()"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		// Sig referenced via param:: (not witness::)
		{"sig from param module", "param::ALICE_TEST_SIG"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("p2pk_testable: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// Must NOT have zero-value placeholder signature
	zeroSig := "0x" + strings.Repeat("00", 64)
	if strings.Contains(out, "ALICE_TEST_SIG: [u8; 64] = "+zeroSig) {
		t.Error("p2pk_testable: ALICE_TEST_SIG should be the real BIP-340 test vector, not all zeros")
	}

	// Must NOT use sig_all_hash (testable uses a fixed message constant instead)
	if strings.Contains(out, "sig_all_hash") {
		t.Error("p2pk_testable: should not use sig_all_hash; testable examples use fixed message constants")
	}
}

// TestExampleHTLCTestable verifies the testable HTLC example compiles with real BIP-340 test vectors.
func TestExampleHTLCTestable(t *testing.T) {
	out := compileExample(t, "../examples/testable/htlc_testable.go")

	checks := []struct {
		desc    string
		present string
	}{
		{"param module", "mod param {"},
		// Real BIP-340 test vector pubkeys
		{"real alice pubkey", "0xf9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9"},
		{"real bob pubkey", "0xdff1d77f2a671c5f36183726db2341be58feae1da2deced843240f7b502ba659"},
		// Verified SHA-256 hash lock (SHA-256 of 32 zero bytes)
		{"verified hashlock", "0x66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925"},
		// Real BIP-340 test vector signatures in params
		{"real alice test sig", "0xe907831f80848d1069a5371b402410364bdf1c5f8307b0084c55f1ce2dca8215"},
		{"real bob test sig", "0x6896bd60eeae296db48a229ff71dfe071bde413e6d43f917dc8dcf8c78de3341"},
		// Either match structure
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right("},
		// SHA-256 chain in Left arm
		{"sha_256_ctx_8_init jet", "jet::sha_256_ctx_8_init()"},
		{"sha_256_ctx_8_add_32 jet", "jet::sha_256_ctx_8_add_32("},
		{"sha_256_ctx_8_finalize jet", "jet::sha_256_ctx_8_finalize("},
		{"eq_256 jet", "jet::eq_256("},
		// BIP-340 verify in both arms
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		// Test sigs referenced via param::
		{"alice sig via param", "param::ALICE_TEST_SIG"},
		{"bob sig via param", "param::BOB_TEST_SIG"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("htlc_testable: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// Must NOT use sig_all_hash
	if strings.Contains(out, "sig_all_hash") {
		t.Error("htlc_testable: should not use sig_all_hash; testable examples use fixed message constants")
	}
}

// TestExampleAtomicSwap verifies the atomic swap example compiles to correct SimplicityHL.
func TestExampleAtomicSwap(t *testing.T) {
	out := compileExample(t, "../examples/atomic_swap.go")
	assertNoInvalidWitness(t, "atomic_swap", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"Either witness type", "Either<([u8; 32], [u8; 64]), [u8; 64]>"},
		{"Either witness Left placeholder", "Left((0x" + strings.Repeat("00", 32) + ", 0x" + strings.Repeat("00", 64) + "))"},
		{"recipient pubkey param", "RECIPIENT_PUBKEY: u256"},
		{"sender pubkey param", "SENDER_PUBKEY: u256"},
		{"hash lock param", "HASH_LOCK: u256"},
		{"min refund height param", "MIN_REFUND_HEIGHT: u32"},
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right(sig)"},
		{"destructuring", "let (preimage, recipient_sig):"},
		{"sha_256_ctx_8_init jet", "jet::sha_256_ctx_8_init()"},
		{"sha_256_ctx_8_add_32 jet", "jet::sha_256_ctx_8_add_32("},
		{"sha_256_ctx_8_finalize jet", "jet::sha_256_ctx_8_finalize("},
		{"eq_256 jet", "jet::eq_256("},
		{"check_lock_height jet", "jet::check_lock_height("},
		{"min refund height in check", "param::MIN_REFUND_HEIGHT"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("atomic_swap: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// witness::W.field must NOT appear inside match arm bodies (bound vars should be used)
	if strings.Contains(out, "witness::W.preimage") {
		t.Error("atomic_swap: Left arm should use 'preimage' not 'witness::W.preimage'")
	}
	if strings.Contains(out, "witness::W.recipient_sig") {
		t.Error("atomic_swap: Left arm should use 'recipient_sig' not 'witness::W.recipient_sig'")
	}
	if strings.Contains(out, "witness::W.sender_sig") {
		t.Error("atomic_swap: Right arm should use 'sig' not 'witness::W.sender_sig'")
	}
}

// TestExampleCovenant verifies the covenant example compiles to correct SimplicityHL.
func TestExampleCovenant(t *testing.T) {
	out := compileExample(t, "../examples/covenant.go")
	assertNoInvalidWitness(t, "covenant", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"sig witness type", "const SIG: [u8; 64]"},
		{"sig witness zero value", "0x" + strings.Repeat("00", 64)},
		{"expected script hash param", "EXPECTED_SCRIPT_HASH: u256"},
		{"owner pubkey param", "OWNER_PUBKEY: u256"},
		{"output index param", "OUTPUT_INDEX: u32"},
		{"main function", "fn main()"},
		{"output_script_hash jet", "jet::output_script_hash("},
		{"output index in call", "param::OUTPUT_INDEX"},
		{"eq_256 jet", "jet::eq_256("},
		{"sig_all_hash jet", "jet::sig_all_hash()"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		{"sig witness in verify", "witness::SIG"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("covenant: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}
}

// TestExampleHTLCHelper verifies the HTLC helper function example compiles to correct SimplicityHL.
// It exercises both analyzeSwitchAsMatch (switch {} dispatch) and helper function inlining.
func TestExampleHTLCHelper(t *testing.T) {
	out := compileExample(t, "../examples/htlc_helper.go")
	assertNoInvalidWitness(t, "htlc_helper", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"helper function declaration", "fn verify_hashlock("},
		{"sha_256_ctx_8_finalize in helper/Left arm", "jet::sha_256_ctx_8_finalize("},
		{"sha_256_ctx_8_add_32 in helper/Left arm", "jet::sha_256_ctx_8_add_32("},
		{"sha_256_ctx_8_init in helper/Left arm", "jet::sha_256_ctx_8_init()"},
		{"eq_256 in helper/Left arm", "jet::eq_256("},
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right(sig)"},
		{"destructuring", "let (preimage, recipient_sig):"},
		{"check_lock_height in Right arm", "jet::check_lock_height(param::MIN_REFUND_HEIGHT)"},
		{"bip_0340_verify in both arms", "jet::bip_0340_verify("},
		{"sig_all_hash jet", "jet::sig_all_hash()"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("htlc_helper: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// witness::W.field must NOT appear inside match arm bodies (bound vars should be used)
	if strings.Contains(out, "witness::W.preimage") {
		t.Error("htlc_helper: Left arm should use 'preimage' not 'witness::W.preimage'")
	}
	if strings.Contains(out, "witness::W.recipient_sig") {
		t.Error("htlc_helper: Left arm should use 'recipient_sig' not 'witness::W.recipient_sig'")
	}
	if strings.Contains(out, "witness::W.sender_sig") {
		t.Error("htlc_helper: Right arm should use 'sig' not 'witness::W.sender_sig'")
	}
}

// TestExampleMultisig verifies the 2-of-3 multisig example compiles to correct SimplicityHL.
func TestExampleMultisig(t *testing.T) {
	out := compileExample(t, "../examples/multisig.go")
	assertNoInvalidWitness(t, "multisig", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"sig0 Option witness type", "const SIG0: Option<[u8; 64]>"},
		{"sig1 Option witness type", "const SIG1: Option<[u8; 64]>"},
		{"sig2 Option witness type", "const SIG2: Option<[u8; 64]>"},
		{"sig0 None placeholder", "SIG0: Option<[u8; 64]> = None"},
		{"sig1 None placeholder", "SIG1: Option<[u8; 64]> = None"},
		{"sig2 None placeholder", "SIG2: Option<[u8; 64]> = None"},
		{"alice pubkey param", "ALICE_PUBKEY: u256"},
		{"bob pubkey param", "BOB_PUBKEY: u256"},
		{"charlie pubkey param", "CHARLIE_PUBKEY: u256"},
		{"counter count_0", "let count_0: u32 ="},
		{"counter count_1", "let count_1: u32 = count_0 +"},
		{"counter count_2", "let count_2: u32 = count_1 +"},
		{"Some arm", "Some(sig) => {"},
		{"None arm no braces", "None => 0,"},
		{"bip_0340_verify with semicolon", "jet::bip_0340_verify((param::ALICE_PUBKEY, msg), sig);"},
		{"return value 1", "1"},
		{"final verify", "jet::verify(jet::le_32(2, count_2))"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("multisig: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// None arms must NOT use block braces
	if strings.Contains(out, "None => {") {
		t.Error("multisig: None arm should not use block braces")
	}
}

// TestExampleVault verifies the vault contract compiles to correct SimplicityHL.
// Demonstrates 2-arm Either with CheckLockHeight + OutputScriptHash in the cold-key arm.
func TestExampleVault(t *testing.T) {
	out := compileExample(t, "../examples/vault.go")
	assertNoInvalidWitness(t, "vault", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"Either witness type", "Either<[u8; 64], [u8; 64]>"},
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right(sig)"},
		{"check_lock_height in Right arm", "jet::check_lock_height(param::COLD_KEY_UNLOCK)"},
		{"output_script_hash in Right arm", "jet::output_script_hash("},
		{"eq_256 in Right arm", "jet::eq_256("},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("vault: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// bip_0340_verify must appear in both arms — check count >= 2
	count := strings.Count(out, "jet::bip_0340_verify(")
	if count < 2 {
		t.Errorf("vault: expected jet::bip_0340_verify in both arms, found %d occurrence(s)", count)
	}

	// Bound variables must be used — raw field access must NOT appear in arm bodies
	if strings.Contains(out, "witness::W.hot_key_sig") {
		t.Error("vault: Left arm should use bound variable, not 'witness::W.hot_key_sig'")
	}
	if strings.Contains(out, "witness::W.cold_key_sig") {
		t.Error("vault: Right arm should use bound 'sig', not 'witness::W.cold_key_sig'")
	}
}

// TestExampleOraclePrice verifies the oracle price contract compiles to correct SimplicityHL.
// Demonstrates 2-arm Either where both arms use BIP-340 verify with different pubkeys.
func TestExampleOraclePrice(t *testing.T) {
	out := compileExample(t, "../examples/oracle_price.go")
	assertNoInvalidWitness(t, "oracle_price", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"Either witness type", "Either<[u8; 64], [u8; 64]>"},
		{"match expression", "match witness::W {"},
		{"Left arm", "Left(data)"},
		{"Right arm", "Right(sig)"},
		{"oracle pubkey verify in Left arm", "jet::bip_0340_verify((param::ORACLE_PUBKEY, msg), data)"},
		{"owner pubkey verify in Right arm", "jet::bip_0340_verify((param::OWNER_PUBKEY, msg), sig)"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("oracle_price: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// Bound variables must be used — raw field access must NOT appear in arm bodies
	if strings.Contains(out, "witness::W.oracle_sig") {
		t.Error("oracle_price: Left arm should use bound variable, not 'witness::W.oracle_sig'")
	}
	if strings.Contains(out, "witness::W.owner_sig") {
		t.Error("oracle_price: Right arm should use bound 'sig', not 'witness::W.owner_sig'")
	}
}

// TestExampleRelativeTimelock verifies the relative timelock example compiles to correct SimplicityHL.
// Demonstrates linear CheckLockDistance usage for CSV-style relative timelocks.
func TestExampleRelativeTimelock(t *testing.T) {
	out := compileExample(t, "../examples/relative_timelock.go")
	assertNoInvalidWitness(t, "relative_timelock", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"sig witness type", "const SIG: [u8; 64]"},
		{"check_lock_distance jet", "jet::check_lock_distance(param::RELATIVE_LOCK_BLOCKS)"},
		{"sig_all_hash jet", "jet::sig_all_hash()"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		{"sender pubkey param", "SENDER_PUBKEY: u256"},
		{"relative lock blocks param", "RELATIVE_LOCK_BLOCKS: u16"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("relative_timelock: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}
}

// TestExampleTaprootKeySpend verifies the taproot key spend example compiles to correct SimplicityHL.
// Demonstrates InternalKey + TapleafVersion introspection before signature verification.
func TestExampleTaprootKeySpend(t *testing.T) {
	out := compileExample(t, "../examples/taproot_key_spend.go")
	assertNoInvalidWitness(t, "taproot_key_spend", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"sig witness type", "const SIG: [u8; 64]"},
		{"internal_key jet", "jet::internal_key()"},
		{"eq_256 for key check", "jet::eq_256(key, param::EXPECTED_INTERNAL_KEY)"},
		{"tapleaf_version jet", "jet::tapleaf_version()"},
		{"eq_8 for version check", "jet::eq_8(version, param::EXPECTED_TAPLEAF_VERSION)"},
		{"bip_0340_verify jet", "jet::bip_0340_verify("},
		{"expected internal key param", "EXPECTED_INTERNAL_KEY: u256"},
		{"expected tapleaf version param", "EXPECTED_TAPLEAF_VERSION: u8"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("taproot_key_spend: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}
}

// TestExampleDoubleSHA256 verifies the double-SHA256 example compiles to correct SimplicityHL,
// including SHA256Add auto-select resolving to sha_256_ctx_8_add_32.
func TestExampleDoubleSHA256(t *testing.T) {
	out := compileExample(t, "../examples/double_sha256.go")
	assertNoInvalidWitness(t, "double_sha256", out)

	checks := []struct {
		desc    string
		present string
	}{
		{"preimage witness", "const PREIMAGE: [u8; 32]"},
		{"sig witness", "const SIG: [u8; 64]"},
		{"hash_lock param", "HASH_LOCK: u256"},
		{"owner_pubkey param", "OWNER_PUBKEY: u256"},
		{"main function", "fn main()"},
		{"sha_256_ctx_8_init", "jet::sha_256_ctx_8_init()"},
		{"sha_256_ctx_8_add_32 (auto-selected)", "jet::sha_256_ctx_8_add_32("},
		{"sha_256_ctx_8_finalize", "jet::sha_256_ctx_8_finalize("},
		{"inner_hash let binding", "let inner_hash:"},
		{"outer_hash let binding", "let outer_hash:"},
		{"eq_256", "jet::eq_256("},
		{"sig_all_hash", "jet::sig_all_hash()"},
		{"bip_0340_verify", "jet::bip_0340_verify("},
		{"witness::PREIMAGE", "witness::PREIMAGE"},
		{"witness::SIG", "witness::SIG"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.present) {
			t.Errorf("double_sha256: expected %s — missing %q\nfull output:\n%s", c.desc, c.present, out)
		}
	}

	// auto-select must NOT produce a 64-byte variant for this example
	if strings.Contains(out, "sha_256_ctx_8_add_64") {
		t.Error("double_sha256: unexpected sha_256_ctx_8_add_64 — should be add_32 for [32]byte input")
	}
}
