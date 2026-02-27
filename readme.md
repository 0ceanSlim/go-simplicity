# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Phase 4 Complete - Multisig Contracts

This transpiler now supports **2-of-3 Multisig Contracts** with Option arrays, counter accumulation, and pattern matching. Building toward full Simplicity support for production contracts.

## What Works Right Now

### Currently Supported Go -> SimplicityHL

- **Boolean functions** with pattern matching
- **Compile-time expression evaluation** (constants, arithmetic, comparisons)
- **Witness/parameter separation** (runtime vs compile-time values)
- **Control flow** (if/else converted to match expressions)
- **Function composition** with boolean parameters
- **Basic types**: `bool`, `uint8`, `uint16`, `uint32`, `uint64`, `u256`, `Ctx8`
- **Hex literals** with automatic type inference (0x... -> u8/u16/u32/u64/u128/u256)
- **Fixed-size arrays**: `[T; N]` including `[u8; 64]` for signatures
- **Sum types**: `Either[L, R]` and `Option[T]` with match expression generation
- **Tuple types**: `(A, B, C)` with destructuring support
- **Struct witnesses**: Custom witness structs with field access
- **Option pattern structs**: Structs with `IsSome bool` + `Value T` auto-detect as `Option<T>`
- **Counter accumulation**: Multiple match expressions compile to counting logic
- **Jet functions**: `jet.BIP340Verify`, `jet.SigAllHash`, SHA-256 operations, comparisons, `jet.Verify`

### Example: Working P2PK Contract

**Input Go Code:**

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
    const SIG: [u8; 64] = 0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000;
}
mod param {
    const ALICE_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
}

fn main() {
    let msg: u256 = jet::sig_all_hash();
    jet::bip_0340_verify((param::ALICE_PUBKEY, msg), witness::SIG)
}
```

> **Note:** Witness values are initialized to zero-value placeholders. Replace `SIG` with a real BIP-340 Schnorr signature before execution — the zero placeholder is syntactically valid but will fail signature verification at runtime.

### Example: Working HTLC Contract

**Input Go Code:**

```go
package main

import "simplicity/jet"

const RecipientPubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4...
const SenderPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8...
const HashLock = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6...

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
    const W: Either<([u8; 32], [u8; 64]), [u8; 64]> = Left((0x0000000000000000000000000000000000000000000000000000000000000000, 0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000));
}
mod param {
    const RECIPIENT_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
    const SENDER_PUBKEY: u256 = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4;
    const HASH_LOCK: u256 = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2;
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

### Example: Working 2-of-3 Multisig Contract

**Input Go Code:**

```go
package main

import "simplicity/jet"

const AlicePubkey = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0
const BobPubkey = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4
const CharliePubkey = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2

type OptionalSig struct {
    IsSome bool
    Value  [64]byte
}

func main() {
    var sig0, sig1, sig2 OptionalSig
    msg := jet.SigAllHash()
    validCount := 0

    if sig0.IsSome {
        jet.BIP340Verify(AlicePubkey, msg, sig0.Value)
        validCount++
    }
    if sig1.IsSome {
        jet.BIP340Verify(BobPubkey, msg, sig1.Value)
        validCount++
    }
    if sig2.IsSome {
        jet.BIP340Verify(CharliePubkey, msg, sig2.Value)
        validCount++
    }

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
    const ALICE_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
    const BOB_PUBKEY: u256 = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4;
    const CHARLIE_PUBKEY: u256 = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2;
}

fn main() {
    let msg: u256 = jet::sig_all_hash();

    // Signature verification with counter accumulation
    let count_0: u32 =
        match witness::SIG0 {
            Some(sig) => {
                jet::bip_0340_verify((param::ALICE_PUBKEY, msg), sig);
                1
            },
            None => 0,
        };
    let count_1: u32 = count_0 +
        match witness::SIG1 {
            Some(sig) => {
                jet::bip_0340_verify((param::BOB_PUBKEY, msg), sig);
                1
            },
            None => 0,
        };
    let count_2: u32 = count_1 +
        match witness::SIG2 {
            Some(sig) => {
                jet::bip_0340_verify((param::CHARLIE_PUBKEY, msg), sig);
                1
            },
            None => 0,
        };

    // Require at least 2 valid signatures
    jet::verify(jet::le_32(2, count_2))
}
```

## Available Jet Functions

| Go Function | Simplicity Jet | Description |
|-------------|----------------|-------------|
| `jet.BIP340Verify(pubkey, msg, sig)` | `jet::bip_0340_verify` | Schnorr signature verification |
| `jet.SigAllHash()` | `jet::sig_all_hash` | Transaction sighash |
| `jet.SHA256Init()` | `jet::sha_256_ctx_8_init` | SHA-256 initialization |
| `jet.SHA256Add32(ctx, data)` | `jet::sha_256_ctx_8_add_32` | Add 32-byte block to hash |
| `jet.SHA256Finalize(ctx)` | `jet::sha_256_ctx_8_finalize` | Finalize hash |
| `jet.CheckLockHeight(h)` | `jet::check_lock_height` | Timelock height check |
| `jet.Eq256(a, b)` | `jet::eq_256` | 256-bit equality |
| `jet.Eq32(a, b)` | `jet::eq_32` | 32-bit equality |
| `jet.Le32(a, b)` | `jet::le_32` | 32-bit less-or-equal |
| `jet.Verify(cond)` | `jet::verify` | Assert condition |
| `jet.CurrentIndex()` | `jet::current_index` | Current input index |
| `jet.LockTime()` | `jet::lock_time` | Transaction locktime |
| `jet.CurrentPrevOutpoint()` | `jet::current_prev_outpoint` | Current input outpoint |
| `jet.CurrentScriptHash()` | `jet::current_script_hash` | Current script hash |

## Sum Types

### Either[L, R]

Represents a choice between two types. Used for contracts with alternative spending paths.

```go
// Go syntax - struct with IsLeft pattern
type Witness struct {
    IsLeft bool
    Left   CompleteData
    Right  CancelData
}

// Auto-detects and maps to SimplicityHL
Either<CompleteData, CancelData>
```

### Option[T]

Represents an optional value. Used for optional signatures in multisig.

```go
// Go syntax - struct with IsSome pattern
type OptionalSig struct {
    IsSome bool
    Value  [64]byte
}

// Auto-detects and maps to SimplicityHL
Option<[u8; 64]>
```

## Development Roadmap

### Phase 2: Core Bitcoin Types - COMPLETE

**Goal:** Generate basic P2PK contracts

- [x] Add `u256` type support
- [x] Implement `[u8; 64]` arrays for signatures
- [x] Create Bitcoin type aliases (`Pubkey`, `Signature`, `Hash`)
- [x] Add jet function support (`jet::bip_0340_verify`, `jet::sig_all_hash`, etc.)
- [x] Hex literal support with automatic type inference
- [x] Test with real P2PK contract generation

**Result:** Successfully transpiling P2PK contracts from Go to working SimplicityHL

### Phase 3: Sum Types & Pattern Matching - COMPLETE

**Goal:** Generate HTLC and conditional contracts

- [x] Implement `Either<A, B>` sum types
- [x] Add `Option<T>` support
- [x] Create `match` expression generation from if/else
- [x] Handle tuple type parsing and generation
- [x] Support struct-based witness data with field access

**Result:** Successfully transpiling HTLC contracts with Left/Right spending paths

### Phase 4: Arrays & Multisig - COMPLETE

**Goal:** Generate multisig contracts

- [x] Option pattern struct detection (`IsSome bool` + `Value T` -> `Option<T>`)
- [x] Multiple match expression handling
- [x] Counter accumulation logic
- [x] `jet::verify` and `jet::le_32` integration
- [x] Bounded for loop detection (compiler validation)

**Result:** Successfully transpiling 2-of-3 multisig contracts with counter-based validation

### Phase 5: Advanced Contracts (NEXT)

**Goal:** Generate vault, inheritance, and complex contracts

- [ ] Recursive covenant patterns
- [ ] Time-based logic with `jet::lock_time`
- [ ] Oracle data validation
- [ ] Multi-path spending conditions
- [ ] Introspection jets for UTXO inspection

## Bug Fixes

Tracked bugs from comparing generated output against the SimplicityHL playground.

- [x] **BUG-1**: SHA-256 jet names wrong — `sha_256_iv` / `sha_256_block` / `sha_256_finalize` should be `sha_256_ctx_8_init` / `sha_256_ctx_8_add_32` / `sha_256_ctx_8_finalize`
- [x] **BUG-2**: Missing `CheckLockHeight` jet — `jet::check_lock_height` not in registry, needed for timelock contracts
- [x] **BUG-3**: HTLC Left arm uses `witness::W.field` instead of bound variable — match arms must destructure `data` and use local field names
- [x] **BUG-4**: HTLC Right arm uses `witness::W.sender_sig` instead of bound `sig` — `varBase` was not passed to the else branch
- [x] **BUG-5**: `None` match arm wraps value in unnecessary block — `None => { 0 }` should be `None => 0,`
- [x] **BUG-6**: Jet calls in `Some` match arm body missing semicolons — statements before a return value need `;`
- [x] **BUG-7**: Witness values use `/* witness */` which is not valid SimplicityHL syntax — each type needs a zero-value placeholder (`None` for Option, `Left(...)` for Either, `0x00...` for byte arrays)

## Known Limitations & Future Work

### Current Limitations

1. **No for loop unrolling** - Bounded loops are detected but not yet fully unrolled
2. **Single witness struct** - Multiple independent structs may need manual adjustment
3. **No arithmetic jets** - Addition, multiplication not yet mapped
4. **No introspection beyond basics** - Limited UTXO introspection jets
5. **No helper function generation** - All logic emitted inline in `fn main()`; the playground idiom of extracting sub-functions is not yet supported
6. **No type aliases** - Raw types emitted (`u256`, `[u8; 64]`) rather than named aliases like `Pubkey`, `Signature`

### Recommended Patterns

- Use struct-based Option types with `IsSome bool` + `Value T` fields
- Use struct-based Either types with `IsLeft bool` + `Left L` + `Right R` fields
- Define pubkeys as hex constants (`const Pubkey = 0x...`)
- Use `jet.SigAllHash()` for transaction message hashing
- Use explicit `jet.Verify()` for final assertions

### Not Yet Supported

- Dynamic arrays or slices
- String types
- Maps/dictionaries
- Recursive function calls
- Closures or function variables
- Interface types
- Goroutines or channels
- Import statements (except `simplicity/jet`)

## Installation & Usage

```bash
# Install
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
go build -o build/simgo cmd/simgo/main.go

# Compile P2PK contract
./build/simgo -input examples/p2pk.go

# Compile HTLC contract
./build/simgo -input examples/htlc.go

# Compile Multisig contract
./build/simgo -input examples/multisig.go

# Run tests
go test ./...
```

## Current Examples

```bash
# P2PK contract with signature verification
./build/simgo -input examples/p2pk.go

# HTLC contract with Left/Right spending paths
./build/simgo -input examples/htlc.go

# 2-of-3 Multisig with optional signatures
./build/simgo -input examples/multisig.go

# Simple boolean validation
./build/simgo -input examples/simple_logic.go
```

## Architecture

```
cmd/simgo/          # Compiler binary
pkg/
├── compiler/       # Validation and orchestration
├── jets/           # Jet registry (BIP340, SHA256, etc.)
├── transpiler/     # Core Go -> SimplicityHL conversion
│   ├── transpiler.go  # Main transpiler logic
│   ├── patterns.go    # Pattern matching extraction
│   └── arrays.go      # Array handling and loop unrolling
└── types/          # Type mapping system
    ├── types.go       # Basic type mapping
    └── either.go      # Sum type definitions
examples/           # Contract examples (p2pk, htlc, multisig)
tests/              # Test suite (41 tests)
```

## Test Coverage

```
=== Example Integration Tests (3 tests) ===
TestExampleP2PK            - Compiles p2pk.go, validates witness/param/jet structure
TestExampleHTLC            - Compiles htlc.go, validates bound vars, destructuring, SHA-256 jets
TestExampleMultisig        - Compiles multisig.go, validates None arms, semicolons, counters

=== Phase 4 Tests (10 tests) ===
TestArrayTypeParsing       - Array type parsing
TestOptionArrayDetection   - Option sum type detection
TestSimpleArrayDeclaration - Array witness declaration
TestOptionArrayDeclaration - Option array declaration
TestArrayConstant          - Multiple pubkey constants
TestBoundedForLoopAllowed  - For loop compiler validation
TestArrayIndexing          - Array index access
TestMultiplePubkeyConstants- Multi-pubkey param generation
TestSimpleMultisigStructure- Multisig contract structure
TestArrayUnrollFramework   - Loop unroll framework

=== Phase 3 Tests (10 tests) ===
TestEitherTypeParsing      - Either<L, R> parsing
TestOptionTypeParsing      - Option<T> parsing
TestTupleTypeParsing       - Tuple type parsing
TestSumTypeDetection       - Sum type identification
TestEitherWitness          - Either witness declaration
TestOptionWitness          - Option witness declaration
TestMatchExpression        - Match generation from if/else
TestSimpleHTLCStructure    - HTLC contract structure
TestGoGenericEitherType    - Go generic type support
TestMatchArmGeneration     - Match arm code generation

=== Phase 2 Tests (8 tests) ===
TestJetRegistry            - Jet function registry
TestHexTypeInference       - Automatic hex -> u256/u64/etc
TestHexLiteral             - Hex constant parsing
TestJetSigAllHash          - Transaction hash jet
TestJetBIP340Verify        - Signature verification jet
TestP2PKContract           - Full P2PK integration
TestJetCallValidation      - Compiler validation
TestCtx8Type               - SHA-256 context type

=== Core Tests (10 tests) ===
TestBasicFunction          - Function transpilation
TestSimpleValidation       - Boolean validation
TestUnsupportedFeatures    - Error handling
TestSimpleConstants        - Constant extraction
TestBooleanLogic           - Boolean functions
TestWorkingExample         - Basic swap contract
... and more
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf)
- [SimplicityHL Documentation](https://github.com/BlockstreamResearch/SimplicityHL)
- [Elements Project](https://github.com/ElementsProject/elements)
- [Liquid Network](https://liquid.net)
