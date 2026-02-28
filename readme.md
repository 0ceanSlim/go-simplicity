# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Phase 9 Complete

**Working contract types:** P2PK, HTLC, Atomic Swap, Covenant, 2-of-3 Multisig, arithmetic/comparison contracts, helper-function contracts with switch dispatch, double-SHA256 hashing, vault (hot/cold key), oracle-gated spend, relative timelock (CSV), Taproot key-spend introspection.

See [ROADMAP.md](ROADMAP.md) for the full development plan.

---

## What Works

- **Boolean functions** with pattern matching
- **Witness/parameter separation** — runtime vs compile-time values
- **Control flow** — `if/else` and `switch {}` converted to match expressions
- **Helper functions** — linear jet-call sequences transpiled and inlined at call sites
- **Function composition** with boolean parameters
- **Types**: `bool`, `uint8`, `uint16`, `uint32`, `uint64`, `u256`, `Ctx8`
- **Hex literals** with automatic type inference (`0x...` → `u8`/`u16`/`u32`/`u64`/`u128`/`u256`)
- **Fixed-size arrays**: `[T; N]` including `[u8; 32]` for preimages, `[u8; 64]` for signatures
- **Sum types**: `Either[L, R]` and `Option[T]` with match expression generation
- **Tuple types** with destructuring support
- **Struct witnesses** with field access
- **Option pattern structs**: `IsSome bool` + `Value T` auto-detects as `Option<T>`
- **Counter accumulation**: Multiple match expressions compile to counting logic
- **Arithmetic operators**: `+`, `-`, `*`, `/`, `%` auto-map to the correct `add_N`/`subtract_N`/etc. jet
- **Comparison operators**: `<`, `<=`, `>`, `>=`, `==` auto-map to `lt_N`/`le_N`/`eq_N` jets (args swapped where needed)
- **Bitwise operators**: `&`, `|`, `^` auto-map to `and_N`/`or_N`/`xor_N` jets
- **Carry-bit destructuring**: arithmetic jets returning `(bool, uN)` emit `let (_, v): (bool, u32) = jet::add_32(...)`
- **Time locks**: `CheckLockHeight`, `CheckLockTime`, `CheckLockDistance`, `CheckLockDuration` and read-only variants
- **Transaction introspection**: `OutputScriptHash`, `NumInputs`, `NumOutputs`, `Version`, `TransactionId`, and more
- **SHA256Add auto-select**: write `jet.SHA256Add(ctx, data)` and the transpiler picks the right `sha_256_ctx_8_add_N` from the Go argument type at transpile time

---

## Quick Start

```bash
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
go build -o build/simgo cmd/simgo/main.go

./build/simgo -input examples/p2pk.go
./build/simgo -input examples/htlc.go
./build/simgo -input examples/atomic_swap.go
./build/simgo -input examples/covenant.go
./build/simgo -input examples/multisig.go
./build/simgo -input examples/htlc_helper.go
./build/simgo -input examples/double_sha256.go
./build/simgo -input examples/vault.go
./build/simgo -input examples/oracle_price.go
./build/simgo -input examples/relative_timelock.go
./build/simgo -input examples/taproot_key_spend.go

go test ./...
```

---

## Examples

### P2PK — Pay to Public Key

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

> **Playground-ready version:** `examples/testable/p2pk_testable.go` uses a real BIP-340 test vector from `pkg/testkeys` — no manual substitution needed.

---

### HTLC — Hash Time Lock Contract

```go
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4...
const SenderPubkey    = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8...
const HashLock        = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6...

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
mod param {
    const RECIPIENT_PUBKEY: u256 = 0x9bef...;
    const SENDER_PUBKEY: u256 = 0xe37d...;
    const HASH_LOCK: u256 = 0xa1b2...;
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

---

### Atomic Swap — Timelock Refund (Phase 6)

Uses `CheckLockHeight` for the refund path. Full source: `examples/atomic_swap.go`.

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

---

### Covenant — Output Script Hash Enforcement (Phase 6)

Verifies that a specific output's script hash matches an expected value. Full source: `examples/covenant.go`.

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

---

### HTLC with Helper Function — Switch Dispatch (Phase 7)

Demonstrates helper function declaration, inlining at call sites, and `switch {}` as sum-type dispatch. Full source: `examples/htlc_helper.go`.

```go
func verifyHashlock(preimage [32]byte) {
    hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), preimage))
    jet.Eq256(hash, HashLock)
}

func main() {
    var w AtomicSwapWitness
    switch {
    case w.IsLeft:
        verifyHashlock(w.Preimage)   // inlined into Left arm
        msg := jet.SigAllHash()
        jet.BIP340Verify(RecipientPubkey, msg, w.RecipientSig)
    case !w.IsLeft:
        jet.CheckLockHeight(MinRefundHeight)
        msg := jet.SigAllHash()
        jet.BIP340Verify(SenderPubkey, msg, w.SenderSig)
    }
}
```

**Generated SimplicityHL:**

```rust
fn verify_hashlock(preimage: [u8; 32]) {
    let hash = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), preimage));
    jet::eq_256(hash, param::HASH_LOCK)
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
            jet::check_lock_height(param::MIN_REFUND_HEIGHT)
            let msg = jet::sig_all_hash();
            jet::bip_0340_verify((param::SENDER_PUBKEY, msg), sig)
        }
    }
}
```

The helper body is both emitted as `fn verify_hashlock` and inlined at its call site with parameters substituted.

---

### Double SHA-256 — SHA256Add Auto-Select (Phase 8)

Demonstrates `jet.SHA256Add` auto-select and SHA-256 chaining. Full source: `examples/double_sha256.go`.

```go
func main() {
    var preimage [32]byte
    var sig [64]byte

    // First pass: SHA256(preimage) — auto-selects sha_256_ctx_8_add_32
    innerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), preimage))

    // Second pass: SHA256(innerHash) — u256 return type → also add_32
    outerHash := jet.SHA256Finalize(jet.SHA256Add(jet.SHA256Init(), innerHash))

    jet.Eq256(outerHash, HashLock)
    msg := jet.SigAllHash()
    jet.BIP340Verify(OwnerPubkey, msg, sig)
}
```

**Generated SimplicityHL:**

```rust
fn main() {
    let inner_hash: u256 = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), witness::PREIMAGE));
    let outer_hash: u256 = jet::sha_256_ctx_8_finalize(jet::sha_256_ctx_8_add_32(jet::sha_256_ctx_8_init(), inner_hash));
    jet::eq_256(outer_hash, param::HASH_LOCK)
    let msg: u256 = jet::sig_all_hash();
    jet::bip_0340_verify((param::OWNER_PUBKEY, msg), witness::SIG)
}
```

Both `jet.SHA256Add` calls resolve to `sha_256_ctx_8_add_32` — the first because `preimage` is `[32]byte`, the second because `SHA256Finalize` returns `u256` (32 bytes).

---

### Vault — Hot/Cold Key Spending (Phase 9)

Two-arm vault: immediate hot-key spend or timelocked cold-key recovery with output script enforcement. Full source: `examples/vault.go`.

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

---

### Oracle Price — Trusted Oracle Authorisation (Phase 9)

Two-arm contract: a trusted oracle's BIP-340 signature authorises the spend (Left), or the owner withdraws directly without oracle involvement (Right). Full source: `examples/oracle_price.go`.

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

---

### Relative Timelock — CSV-Style (Phase 9)

Linear contract enforcing a minimum number of blocks since the funding UTXO was confirmed (`CheckLockDistance`). Full source: `examples/relative_timelock.go`.

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

---

### Taproot Key Spend — Internal Key + Tapleaf Introspection (Phase 9)

Linear contract verifying the Taproot internal key and tapleaf version match expected constants before authorising via BIP-340. Full source: `examples/taproot_key_spend.go`.

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

---

### Multisig — 2-of-3

```go
package main

import "simplicity/jet"

const AlicePubkey   = 0x9bef8d556d80e43ae7e0becb3f7de6b4...
const BobPubkey     = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8...
const CharliePubkey = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6...

type OptionalSig struct {
    IsSome bool
    Value  [64]byte
}

func main() {
    var sig0, sig1, sig2 OptionalSig
    msg := jet.SigAllHash()

    if sig0.IsSome { jet.BIP340Verify(AlicePubkey, msg, sig0.Value);   validCount++ }
    if sig1.IsSome { jet.BIP340Verify(BobPubkey, msg, sig1.Value);     validCount++ }
    if sig2.IsSome { jet.BIP340Verify(CharliePubkey, msg, sig2.Value); validCount++ }

    jet.Verify(jet.Le32(2, validCount))
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

---

## Available Jets

### Signature

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.BIP340Verify(pubkey, msg, sig)` | `jet::bip_0340_verify` | — |
| `jet.SigAllHash()` | `jet::sig_all_hash` | `u256` |

### Hash — SHA-256

| Go Function | Simplicity Jet | Input | Returns |
|-------------|----------------|-------|---------|
| `jet.SHA256Init()` | `jet::sha_256_ctx_8_init` | — | `Ctx8` |
| `jet.SHA256Add(ctx, data)` | *(auto-selected)* | any supported size | `Ctx8` |
| `jet.SHA256Add1(ctx, data)` | `jet::sha_256_ctx_8_add_1` | `u8` | `Ctx8` |
| `jet.SHA256Add2(ctx, data)` | `jet::sha_256_ctx_8_add_2` | `[u8; 2]` | `Ctx8` |
| `jet.SHA256Add4(ctx, data)` | `jet::sha_256_ctx_8_add_4` | `[u8; 4]` | `Ctx8` |
| `jet.SHA256Add8(ctx, data)` | `jet::sha_256_ctx_8_add_8` | `[u8; 8]` | `Ctx8` |
| `jet.SHA256Add16(ctx, data)` | `jet::sha_256_ctx_8_add_16` | `[u8; 16]` | `Ctx8` |
| `jet.SHA256Add32(ctx, data)` | `jet::sha_256_ctx_8_add_32` | `[u8; 32]` | `Ctx8` |
| `jet.SHA256Add64(ctx, data)` | `jet::sha_256_ctx_8_add_64` | `[u8; 64]` | `Ctx8` |
| `jet.SHA256Add128(ctx, data)` | `jet::sha_256_ctx_8_add_128` | `[u8; 128]` | `Ctx8` |
| `jet.SHA256Add256(ctx, data)` | `jet::sha_256_ctx_8_add_256` | `[u8; 256]` | `Ctx8` |
| `jet.SHA256Add512(ctx, data)` | `jet::sha_256_ctx_8_add_512` | `[u8; 512]` | `Ctx8` |
| `jet.SHA256Finalize(ctx)` | `jet::sha_256_ctx_8_finalize` | `Ctx8` | `u256` |
| `jet.SHA256Block(iv, block)` | `jet::sha_256_block` | `u256, [u8; 64]` | `u256` |
| `jet.SHA256IV()` | `jet::sha_256_iv` | — | `u256` |

> `jet.SHA256Add(ctx, data)` is a convenience alias — the transpiler resolves it to the correctly-sized `sha_256_ctx_8_add_N` jet based on the Go type of `data` at transpile time.

### Equality

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.Eq256(a, b)` | `jet::eq_256` | `bool` |
| `jet.Eq32(a, b)` | `jet::eq_32` | `bool` |
| `jet.Eq8(a, b)` | `jet::eq_8` | `bool` |
| *(16/64 variants also registered)* | | |

### Arithmetic

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.Add32(a, b)` | `jet::add_32` | `(bool, u32)` |
| `jet.Subtract32(a, b)` | `jet::subtract_32` | `(bool, u32)` |
| `jet.Multiply32(a, b)` | `jet::multiply_32` | `u64` |
| `jet.Divide32(a, b)` | `jet::divide_32` | `u32` |
| `jet.Modulo32(a, b)` | `jet::modulo_32` | `u32` |
| *(8/16/64 variants also registered)* | | |

> Operators `+`, `-`, `*`, `/`, `%` in `main()` auto-map to the correct jet based on operand width.

### Comparison

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.Lt32(a, b)` | `jet::lt_32` | `bool` |
| `jet.Le32(a, b)` | `jet::le_32` | `bool` |
| *(8/16/64 variants also registered)* | | |

> Operators `<`, `<=`, `>`, `>=`, `==` auto-map to the correct jet. `>` and `>=` swap arguments since no `gt`/`ge` jet exists.

### Bitwise

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.And32(a, b)` | `jet::and_32` | `u32` |
| `jet.Or32(a, b)` | `jet::or_32` | `u32` |
| `jet.Xor32(a, b)` | `jet::xor_32` | `u32` |
| `jet.Complement32(a)` | `jet::complement_32` | `u32` |
| *(8/16/64 variants also registered)* | | |

### Time Locks

| Go Function | Simplicity Jet |
|-------------|----------------|
| `jet.CheckLockHeight(h)` | `jet::check_lock_height` |
| `jet.CheckLockTime(t)` | `jet::check_lock_time` |
| `jet.CheckLockDistance(d)` | `jet::check_lock_distance` |
| `jet.CheckLockDuration(d)` | `jet::check_lock_duration` |
| `jet.TxLockHeight()` | `jet::tx_lock_height` |
| `jet.TxLockTime()` | `jet::tx_lock_time` |
| `jet.TxLockDistance()` | `jet::tx_lock_distance` |
| `jet.TxLockDuration()` | `jet::tx_lock_duration` |
| `jet.TxIsFinal()` | `jet::tx_is_final` |

### Transaction Introspection

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.NumInputs()` | `jet::num_inputs` | `u32` |
| `jet.NumOutputs()` | `jet::num_outputs` | `u32` |
| `jet.Version()` | `jet::version` | `u32` |
| `jet.TransactionId()` | `jet::transaction_id` | `u256` |
| `jet.OutputScriptHash(idx)` | `jet::output_script_hash` | `u256` |
| `jet.InputScriptHash(idx)` | `jet::input_script_hash` | `u256` |
| `jet.InputPrevOutpoint(idx)` | `jet::input_prev_outpoint` | `(u256, u32)` |
| `jet.CurrentSequence()` | `jet::current_sequence` | `u32` |
| `jet.InternalKey()` | `jet::internal_key` | `u256` |
| `jet.TapleafVersion()` | `jet::tapleaf_version` | `u8` |
| `jet.Tappath()` | `jet::tappath` | `u256` |
| `jet.ScriptCmr()` | `jet::script_cmr` | `u256` |
| `jet.GenesisBlockHash()` | `jet::genesis_block_hash` | `u256` |
| `jet.CurrentAmount()` | `jet::current_amount` | — |

### Utilities

| Go Function | Simplicity Jet |
|-------------|----------------|
| `jet.Verify(cond)` | `jet::verify` |

---

## Sum Types

### Either[L, R]

Use a struct with `IsLeft bool` — auto-detected as `Either<L, R>`. Works with both `if/else` and `switch {}`:

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

Use a struct with `IsSome bool` + `Value T` — auto-detected as `Option<T>`:

```go
type OptionalSig struct {
    IsSome bool
    Value  [64]byte
}
// → Option<[u8; 64]>
```

---

## Helper Functions (Phase 7)

Helper functions with linear jet-call bodies are fully transpiled. The body is both emitted as a named function and **inlined** at every call site with parameters substituted:

```go
func verifyHashlock(preimage [32]byte) {
    hash := jet.SHA256Finalize(jet.SHA256Add32(jet.SHA256Init(), preimage))
    jet.Eq256(hash, HashLock)
}
```

Restrictions for Phase 7:
- Body must be a linear sequence of jet calls and assignments (no `if/else` — deferred to Phase 8)
- No recursion
- Constants must be declared before helper functions in the source file

---

## Testable Examples

`examples/testable/` contains variants that use real BIP-340 spec test vectors from `pkg/testkeys` — output is immediately paste-able into the [SimplicityHL playground](https://www.wpsoftware.net/elements-playground/) without manual substitution.

```bash
./build/simgo -input examples/testable/p2pk_testable.go
./build/simgo -input examples/testable/htlc_testable.go
./build/simgo -input examples/testable/timelock_check_testable.go
./build/simgo -input examples/testable/arithmetic_test.go
```

---

## Architecture

```
cmd/simgo/          # CLI binary
pkg/
├── compiler/       # Validation and orchestration
├── jets/           # Jet registry (91+ jets)
├── transpiler/     # Core Go → SimplicityHL AST walker
│   ├── transpiler.go   # Main logic: analysis, code generation, helper inlining
│   ├── patterns.go     # Either/Option match extraction, switch dispatch
│   └── arrays.go       # Fixed-size arrays + loop unrolling
├── types/          # Type mapping (Go → Simplicity)
│   ├── types.go
│   └── either.go
└── testkeys/       # BIP-340 spec test vectors
examples/           # Contract examples (14 files + 4 testable)
tests/              # 60 tests across all phases
```

---

## Not Supported

- Dynamic arrays / slices
- Maps, channels, goroutines
- Interfaces
- Recursive function calls
- `if/else` inside helper functions (deferred)
- 3+ spending paths / nested Either (Phase 10)
- String types
- Imports other than `simplicity/jet`

---

## License

MIT — see [LICENSE](LICENSE) for details.

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf)
- [SimplicityHL Documentation](https://github.com/BlockstreamResearch/SimplicityHL)
- [SimplicityHL Jet Reference](https://docs.rs/simplicityhl-as-rust/latest/simplicityhl_as_rust/jet/index.html)
- [Elements Project](https://github.com/ElementsProject/elements)
- [Liquid Network](https://liquid.net)
