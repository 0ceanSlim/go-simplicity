# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Phase 5 Complete

**Working contract types:** P2PK, HTLC, 2-of-3 Multisig, arithmetic/comparison contracts.

See [ROADMAP.md](ROADMAP.md) for the full development plan.

---

## What Works

- **Boolean functions** with pattern matching
- **Witness/parameter separation** — runtime vs compile-time values
- **Control flow** — `if/else` converted to match expressions
- **Function composition** with boolean parameters
- **Types**: `bool`, `uint8`, `uint16`, `uint32`, `uint64`, `u256`, `Ctx8`
- **Hex literals** with automatic type inference (`0x...` → `u8`/`u16`/`u32`/`u64`/`u128`/`u256`)
- **Fixed-size arrays**: `[T; N]` including `[u8; 64]` for signatures
- **Sum types**: `Either[L, R]` and `Option[T]` with match expression generation
- **Tuple types** with destructuring support
- **Struct witnesses** with field access
- **Option pattern structs**: `IsSome bool` + `Value T` auto-detects as `Option<T>`
- **Counter accumulation**: Multiple match expressions compile to counting logic
- **Arithmetic operators**: `+`, `-`, `*`, `/`, `%` auto-map to the correct `add_N`/`subtract_N`/etc. jet
- **Comparison operators**: `<`, `<=`, `>`, `>=`, `==` auto-map to `lt_N`/`le_N`/`eq_N` jets (args swapped where needed)
- **Bitwise operators**: `&`, `|`, `^` auto-map to `and_N`/`or_N`/`xor_N` jets
- **Carry-bit destructuring**: arithmetic jets returning `(bool, uN)` emit `let (_, v): (bool, u32) = jet::add_32(...)`

---

## Quick Start

```bash
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
go build -o build/simgo cmd/simgo/main.go

./build/simgo -input examples/p2pk.go
./build/simgo -input examples/htlc.go
./build/simgo -input examples/multisig.go
./build/simgo -input examples/amount_check.go

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
        jet.BIP340Verify(RecipientPubkey, jet.SigAllHash(), w.RecipientSig)
    } else {
        jet.BIP340Verify(SenderPubkey, jet.SigAllHash(), w.SenderSig)
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
    validCount := 0

    if sig0.IsSome { jet.BIP340Verify(AlicePubkey, msg, sig0.Value);   validCount++ }
    if sig1.IsSome { jet.BIP340Verify(BobPubkey, msg, sig1.Value);     validCount++ }
    if sig2.IsSome { jet.BIP340Verify(CharliePubkey, msg, sig2.Value); validCount++ }

    jet.Verify(jet.Le32(2, validCount))
}
```

**Generated SimplicityHL:**

```rust
mod witness {
    const SIG0: Option<[u8; 64]> = None;
    const SIG1: Option<[u8; 64]> = None;
    const SIG2: Option<[u8; 64]> = None;
}
mod param {
    const ALICE_PUBKEY: u256 = ...;
    const BOB_PUBKEY: u256 = ...;
    const CHARLIE_PUBKEY: u256 = ...;
}

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

### Amount Check — Arithmetic & Comparison Jets (Phase 5)

```go
package main

import "simplicity/jet"

const MinBlockHeight uint32 = 800000
const MaxInputIndex  uint32 = 9

func main() {
    jet.CheckLockHeight(MinBlockHeight)

    idx     := jet.CurrentIndex()
    indexOk := idx <= MaxInputIndex   // auto-maps to jet::le_32
    jet.Verify(indexOk)

    height   := jet.TxLockHeight()
    heightOk := height >= MinBlockHeight  // auto-maps to jet::le_32 with swapped args
    jet.Verify(heightOk)
}
```

**Generated SimplicityHL:**

```rust
mod witness {}
mod param {
    const MIN_BLOCK_HEIGHT: u32 = 800000;
    const MAX_INPUT_INDEX: u32 = 9;
}

fn main() {
    jet::check_lock_height(param::MIN_BLOCK_HEIGHT)
    let idx: u32 = jet::current_index();
    let index_ok: bool = jet::le_32(idx, param::MAX_INPUT_INDEX);
    jet::verify(index_ok)
    let height: u32 = jet::tx_lock_height();
    let height_ok: bool = jet::le_32(param::MIN_BLOCK_HEIGHT, height);
    jet::verify(height_ok)
}
```

---

## Available Jets

### Signature & Hash

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.BIP340Verify(pubkey, msg, sig)` | `jet::bip_0340_verify` | — |
| `jet.SigAllHash()` | `jet::sig_all_hash` | `u256` |
| `jet.SHA256Init()` | `jet::sha_256_ctx_8_init` | `Ctx8` |
| `jet.SHA256Add32(ctx, data)` | `jet::sha_256_ctx_8_add_32` | `Ctx8` |
| `jet.SHA256Finalize(ctx)` | `jet::sha_256_ctx_8_finalize` | `u256` |

### Arithmetic

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.Add32(a, b)` | `jet::add_32` | `(bool, u32)` |
| `jet.Subtract32(a, b)` | `jet::subtract_32` | `(bool, u32)` |
| `jet.Multiply32(a, b)` | `jet::multiply_32` | `u64` |
| `jet.Divide32(a, b)` | `jet::divide_32` | `u32` |
| `jet.Modulo32(a, b)` | `jet::modulo_32` | `u32` |
| *(8/16/64 variants also registered)* | | |

> Arithmetic operators `+`, `-`, `*`, `/`, `%` in `main()` auto-map to the correct jet based on operand width.

### Comparison

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.Lt32(a, b)` | `jet::lt_32` | `bool` |
| `jet.Le32(a, b)` | `jet::le_32` | `bool` |
| `jet.Eq32(a, b)` | `jet::eq_32` | `bool` |
| `jet.Eq256(a, b)` | `jet::eq_256` | `bool` |
| *(8/16/64 variants also registered)* | | |

> Operators `<`, `<=`, `>`, `>=`, `==` auto-map to the correct jet. `>` and `>=` swap arguments since no `gt`/`ge` jet exists.

### Time Locks

| Go Function | Simplicity Jet |
|-------------|----------------|
| `jet.CheckLockHeight(h)` | `jet::check_lock_height` |
| `jet.CheckLockTime(t)` | `jet::check_lock_time` |
| `jet.TxLockHeight()` | `jet::tx_lock_height` |
| `jet.TxLockTime()` | `jet::tx_lock_time` |
| `jet.TxIsFinal()` | `jet::tx_is_final` |
| `jet.CheckLockDistance(d)` | `jet::check_lock_distance` |
| `jet.CheckLockDuration(d)` | `jet::check_lock_duration` |

### Transaction Introspection

| Go Function | Simplicity Jet | Returns |
|-------------|----------------|---------|
| `jet.CurrentIndex()` | `jet::current_index` | `u32` |
| `jet.NumInputs()` | `jet::num_inputs` | `u32` |
| `jet.NumOutputs()` | `jet::num_outputs` | `u32` |
| `jet.Version()` | `jet::version` | `u32` |
| `jet.TransactionId()` | `jet::transaction_id` | `u256` |
| `jet.CurrentPrevOutpoint()` | `jet::current_prev_outpoint` | `(u256, u32)` |
| `jet.CurrentScriptHash()` | `jet::current_script_hash` | `u256` |
| `jet.InternalKey()` | `jet::internal_key` | `u256` |
| `jet.GenesisBlockHash()` | `jet::genesis_block_hash` | `u256` |

### Utilities

| Go Function | Simplicity Jet |
|-------------|----------------|
| `jet.Verify(cond)` | `jet::verify` |

---

## Sum Types

### Either[L, R]

Use a struct with `IsLeft bool` — auto-detected as `Either<L, R>`:

```go
type HTLCWitness struct {
    IsLeft       bool
    Preimage     [32]byte   // Left arm data
    RecipientSig [64]byte
    SenderSig    [64]byte   // Right arm data
}
// → Either<([u8; 32], [u8; 64]), [u8; 64]>
```

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
├── jets/           # Jet registry (80+ jets)
├── transpiler/     # Core Go → SimplicityHL AST walker
│   ├── transpiler.go   # Main logic (~1500 lines)
│   ├── patterns.go     # Either/Option match extraction
│   └── arrays.go       # Fixed-size arrays + loop unrolling
├── types/          # Type mapping (Go → Simplicity)
│   ├── types.go
│   └── either.go
└── testkeys/       # BIP-340 spec test vectors
examples/           # Contract examples
tests/              # 46 tests across all phases
```

---

## Not Supported

- Dynamic arrays / slices
- Maps, channels, goroutines
- Interfaces
- Recursive function calls
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
