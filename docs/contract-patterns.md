# Contract Patterns & Examples

Full worked examples of each contract pattern supported by go-simplicity. Each section shows the Go source and the SimplicityHL output produced by the transpiler.

---

## P2PK — Pay to Public Key

```go
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0

func main() {
    var sig [64]byte
    msg := jet.SigAllHash()
    jet.BIP340Verify(AlicePubkey, msg, sig)
}
```

**Generated SimplicityHL:**

```rust
mod witness {
    const SIG: [u8; 64] = 0x0000...0000;
}
mod param {
    const ALICE_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
}

fn main() {
    let msg: u256 = jet::sig_all_hash();
    jet::bip_0340_verify((param::ALICE_PUBKEY, msg), witness::SIG)
}
```

Source: `examples/p2pk.go`

---

## HTLC — Hash Time Lock Contract

```go
type HTLCWitness struct {
    IsLeft       bool
    Preimage     [32]byte
    RecipientSig [64]byte
    SenderSig    [64]byte
}

func main() {
    var w HTLCWitness
    if w.IsLeft {
        hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), w.Preimage))
        jet.Eq256(hash, HashLock)
        msg := jet.SigAllHash()
        jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
    } else {
        msg := jet.SigAllHash()
        jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
    }
}
```

**Generated SimplicityHL:**

```rust
mod witness {
    const W: Either<([u8; 32], [u8; 64]), [u8; 64]> = Left((0x0000...0000, 0x0000...0000));
}

fn main() {
    match witness::W {
        Left(data) => {
            let (preimage, recipient_sig): ([u8; 32], [u8; 64]) = data;
            let hash = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), preimage));
            jet::eq_256(hash, param::HASH_LOCK)
            let msg = jet::sig_all_hash();
            jet::bip_0340_verify((param::RECIPIENT_PUBKEY, msg), recipient_sig)
        },
        Right(sig) => {
            let msg = jet::sig_all_hash();
            jet::bip_0340_verify((param::SENDER_PUBKEY, msg), sig)
        }
    }
}
```

Source: `examples/htlc.go`

---

## Atomic Swap — Timelock Refund

Left path: Alice claims with SHA-256 preimage + signature. Right path: Bob refunds after a block height with signature.

```go
func main() {
    var w AtomicSwapWitness
    if w.IsLeft {
        hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), w.Preimage))
        jet.Eq256(hash, HashLock)
        msg := jet.SigAllHash()
        jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
    } else {
        jet.CheckLockHeight(MinRefundHeight)
        msg := jet.SigAllHash()
        jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
    }
}
```

**Generated SimplicityHL (Right arm):**

```rust
Right(sig) => {
    jet::check_lock_height(param::MIN_REFUND_HEIGHT)
    let msg = jet::sig_all_hash();
    jet::bip_0340_verify((param::SENDER_PUBKEY, msg), sig)
}
```

Source: `examples/atomic_swap.go`

---

## Covenant — Output Script Hash Enforcement

Verifies that a specific output's script hash matches a known value, preventing arbitrary redirection of funds.

```go
func main() {
    var sig [64]byte
    hash := jet.OutputScriptHash(OutputIndex)
    jet.Eq256(hash, ExpectedScriptHash)
    msg := jet.SigAllHash()
    jet.BIP340Verify(OwnerPubkey, msg, sig)
}
```

**Generated SimplicityHL:**

```rust
fn main() {
    let hash = jet::output_script_hash(param::OUTPUT_INDEX);
    jet::eq_256(hash, param::EXPECTED_SCRIPT_HASH)
    let msg = jet::sig_all_hash();
    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
}
```

Source: `examples/covenant.go`

---

## Vault — Hot/Cold Key Spending

Immediate hot-key spend (Left) or timelocked cold-key recovery with output script enforcement (Right).

```go
func main() {
    var w VaultWitness
    if w.IsLeft {
        msg := jet.SigAllHash()
        jet.BIP340Verify(HotKeyPubkey, msg, w.HotKeySig)
    } else {
        jet.CheckLockHeight(ColdKeyUnlock)
        scriptHash := jet.OutputScriptHash(VaultOutputIndex)
        jet.Eq256(scriptHash, VaultScript)
        msg := jet.SigAllHash()
        jet.BIP340Verify(ColdKeyPubkey, msg, w.ColdKeySig)
    }
}
```

**Generated SimplicityHL (Right arm):**

```rust
Right(sig) => {
    jet::check_lock_height(param::COLD_KEY_UNLOCK)
    let script_hash = jet::output_script_hash(param::VAULT_OUTPUT_INDEX);
    jet::eq_256(script_hash, param::VAULT_SCRIPT)
    let msg = jet::sig_all_hash();
    jet::bip_0340_verify((param::COLD_KEY_PUBKEY, msg), sig)
}
```

Source: `examples/vault.go`

---

## Oracle Price — Trusted Oracle Authorisation

Oracle's BIP-340 signature authorises the spend (Left); owner can withdraw directly without the oracle (Right).

```go
func main() {
    var w OracleWitness
    if w.IsLeft {
        msg := jet.SigAllHash()
        jet.BIP340Verify(OraclePubkey, msg, w.OracleSig)
    } else {
        msg := jet.SigAllHash()
        jet.BIP340Verify(OwnerPubkey, msg, w.OwnerSig)
    }
}
```

**Generated SimplicityHL:**

```rust
match witness::W {
    Left(data) => {
        let msg = jet::sig_all_hash();
        jet::bip_0340_verify((param::ORACLE_PUBKEY, msg), data)
    },
    Right(sig) => {
        let msg = jet::sig_all_hash();
        jet::bip_0340_verify((param::OWNER_PUBKEY, msg), sig)
    }
}
```

Source: `examples/oracle_price.go`

---

## Relative Timelock — CSV-Style

Enforces a minimum number of blocks since the funding UTXO was confirmed (`check_lock_distance`).

```go
const RelativeLockBlocks uint16 = 10

func main() {
    var sig [64]byte
    jet.CheckLockDistance(RelativeLockBlocks)
    msg := jet.SigAllHash()
    jet.BIP340Verify(SenderPubkey, msg, sig)
}
```

**Generated SimplicityHL:**

```rust
fn main() {
    jet::check_lock_distance(param::RELATIVE_LOCK_BLOCKS)
    let msg = jet::sig_all_hash();
    jet::bip_0340_verify((param::SENDER_PUBKEY, msg), witness::SIG)
}
```

For time-based (512-second units): use `jet.CheckLockDuration(n)` → `jet::check_lock_duration`.

Source: `examples/relative_timelock.go`

---

## Taproot Key Spend — Internal Key Introspection

Asserts the Taproot internal key and tapleaf version match expected constants before authorising.

```go
const ExpectedTapleafVersion uint8 = 0xc0

func main() {
    var sig [64]byte
    key := jet.InternalKey()
    jet.Eq256(key, ExpectedInternalKey)
    version := jet.TapleafVersion()
    jet.Eq8(version, ExpectedTapleafVersion)
    msg := jet.SigAllHash()
    jet.BIP340Verify(OwnerPubkey, msg, sig)
}
```

**Generated SimplicityHL:**

```rust
fn main() {
    let key = jet::internal_key();
    jet::eq_256(key, param::EXPECTED_INTERNAL_KEY)
    let version = jet::tapleaf_version();
    jet::eq_8(version, param::EXPECTED_TAPLEAF_VERSION)
    let msg = jet::sig_all_hash();
    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
}
```

Source: `examples/taproot_key_spend.go`

---

## HTLC with Helper Function — Switch Dispatch

Helper function inlining + `switch {}` as sum-type dispatch.

```go
func verifyHashlock(preimage [32]byte) {
    hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), preimage))
    jet.Eq256(hash, HashLock)
}

func main() {
    var w AtomicSwapWitness
    switch {
    case w.IsLeft:
        verifyHashlock(w.Preimage)
        msg := jet.SigAllHash()
        jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
    case !w.IsLeft:
        jet.CheckLockHeight(MinRefundHeight)
        msg := jet.SigAllHash()
        jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
    }
}
```

The helper body is emitted as `fn verify_hashlock` and inlined at its call site.

Source: `examples/htlc_helper.go`

---

## Multisig — 2-of-3

Uses `Option<[u8; 64]>` witnesses (struct with `IsSome bool` + `Value [64]byte`) and a counter accumulation pattern.

```go
func main() {
    var sig0, sig1, sig2 OptionalSig
    msg := jet.SigAllHash()
    // each sig match contributes 0 or 1 to a running count
    jet.Verify(jet.Le32(2, count_2))  // require at least 2 valid sigs
}
```

**Generated SimplicityHL:**

```rust
fn main() {
    let msg: u256 = jet::sig_all_hash();
    let count_0: u32 = match witness::SIG0 {
        Some(sig) => { jet::bip_0340_verify((param::ALICE_PUBKEY, msg), sig); 1 },
        None => 0,
    };
    let count_1: u32 = count_0 + match witness::SIG1 { ... };
    let count_2: u32 = count_1 + match witness::SIG2 { ... };
    jet::verify(jet::le_32(2, count_2))
}
```

Source: `examples/multisig.go`

---

## Double SHA-256 — SHA256Add Auto-Select

`jet.SHA256Add(ctx, data)` auto-selects the correct `sha_256_ctx_8_add_N` variant based on the Go type of `data`.

```go
func main() {
    var preimage [32]byte
    var sig [64]byte
    innerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), preimage))
    outerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), innerHash))
    jet.Eq256(outerHash, HashLock)
    msg := jet.SigAllHash()
    jet.BIP340Verify(OwnerPubkey, msg, sig)
}
```

Both `SHA256Add` calls resolve to `sha_256_ctx_8_add_32` — the first because `preimage` is `[32]byte`, the second because `SHA256Finalize` returns `u256` (32 bytes).

Source: `examples/double_sha256.go`

---

## Sum Types

### Either[L, R]

Use a struct with `IsLeft bool` — auto-detected as `Either<L, R>`. Works with `if/else` and `switch {}`:

```go
type HTLCWitness struct {
    IsLeft       bool
    Preimage     [32]byte   // Left arm data
    RecipientSig [64]byte
    SenderSig    [64]byte   // Right arm data
}
// → Either<([u8; 32], [u8; 64]), [u8; 64]>
```

Both `if w.IsLeft { ... } else { ... }` and `switch { case w.IsLeft: ... case !w.IsLeft: ... }` generate the same match expression.

### Option[T]

Use a struct with `IsSome bool` + `Value T`:

```go
type OptionalSig struct {
    IsSome bool
    Value  [64]byte
}
// → Option<[u8; 64]>
```

---

## Testable Examples

`examples/testable/` contains variants that embed real BIP-340 spec test vectors from `pkg/testkeys`. The output is immediately paste-able into the [SimplicityHL playground](https://www.wpsoftware.net/elements-playground/) with no manual substitution needed.

```bash
./build/simgo -input examples/testable/p2pk_testable.go
./build/simgo -input examples/testable/htlc_testable.go
./build/simgo -input examples/testable/timelock_check_testable.go
./build/simgo -input examples/testable/arithmetic_test.go
```
