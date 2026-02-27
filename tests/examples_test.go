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
