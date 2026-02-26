# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Phase 3 Complete - HTLC Contracts

This transpiler now supports **HTLC (Hash Time Locked Contracts)** with sum types (`Either`, `Option`), pattern matching via `if/else`, and struct-based witness data. Building toward full Simplicity support.

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
- **Jet functions**: `jet.BIP340Verify`, `jet.SigAllHash`, SHA-256 operations, comparisons

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
    const SIG: [u8; 64] = /* witness */;
}
mod param {
    const ALICE_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4e5e4f7e8d9a0b1c2d3e4f5a6b7c8d9e0;
}

fn main() {
    let msg: u256 = jet::sig_all_hash();
    jet::bip_0340_verify((param::ALICE_PUBKEY, msg), witness::SIG)
}
```

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
    const W: HTLCWitness = /* witness */;
}
mod param {
    const RECIPIENT_PUBKEY: u256 = 0x9bef8d556d80e43ae7e0becb3f7de6b4...;
    const SENDER_PUBKEY: u256 = 0xe37d58a1aae4ba05c9b2d3e4f5a6b7c8...;
    const HASH_LOCK: u256 = 0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6...;
}

fn main() {
    match witness::W {
        Left => {
            let hash = jet::sha_256_finalize(jet::sha_256_block(jet::sha_256_iv(), witness::W.preimage));
            jet::eq_256(hash, param::HASH_LOCK)
            jet::bip_0340_verify(param::RECIPIENT_PUBKEY, jet::sig_all_hash(), witness::W.recipient_sig)
        },
        Right => {
            jet::bip_0340_verify(param::SENDER_PUBKEY, jet::sig_all_hash(), witness::W.sender_sig)
        }
    }
}
```

## Available Jet Functions

| Go Function | Simplicity Jet | Description |
|-------------|----------------|-------------|
| `jet.BIP340Verify(pubkey, msg, sig)` | `jet::bip_0340_verify` | Schnorr signature verification |
| `jet.SigAllHash()` | `jet::sig_all_hash` | Transaction sighash |
| `jet.SHA256Init()` | `jet::sha_256_iv` | SHA-256 initialization |
| `jet.SHA256Add32(ctx, data)` | `jet::sha_256_block` | Add block to hash |
| `jet.SHA256Finalize(ctx)` | `jet::sha_256_finalize` | Finalize hash |
| `jet.Eq256(a, b)` | `jet::eq_256` | 256-bit equality |
| `jet.Eq32(a, b)` | `jet::eq_32` | 32-bit equality |
| `jet.Le32(a, b)` | `jet::le_32` | 32-bit less-or-equal |
| `jet.Verify(cond)` | `jet::verify` | Assert condition |
| `jet.CurrentIndex()` | `jet::current_index` | Current input index |
| `jet.LockTime()` | `jet::lock_time` | Transaction locktime |

## Sum Types

### Either[L, R]

Represents a choice between two types. Used for contracts with alternative spending paths.

```go
// Go syntax
type Witness Either[CompleteData, CancelData]

// Maps to SimplicityHL
Either<CompleteData, CancelData>
```

### Option[T]

Represents an optional value. Used for optional signatures in multisig.

```go
// Go syntax
type MaybeSig Option[[64]byte]

// Maps to SimplicityHL
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

### Phase 4: Arrays & Multisig (NEXT)

**Goal:** Generate multisig contracts

- [ ] Fixed array iteration support
- [ ] Array counting and validation
- [ ] Option array handling `[Option<T>; N]`
- [ ] Counter-based logic
- [ ] Complex validation functions

**Target:** Successfully transpile 2-of-3 multisig contract

### Phase 5: Advanced Contracts

**Goal:** Generate vault, inheritance, and complex contracts

- [ ] Recursive covenant patterns
- [ ] Time-based logic with `jet::lock_time`
- [ ] Oracle data validation
- [ ] Multi-path spending conditions

## Installation & Usage

```bash
# Install
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
make build

# Compile P2PK contract
./build/simgo -input examples/p2pk.go -output p2pk.shl

# Compile HTLC contract
./build/simgo -input examples/htlc.go -output htlc.shl

# Run tests
make test
```

## Current Examples

```bash
# P2PK contract with signature verification
./build/simgo -input examples/p2pk.go

# HTLC contract with Left/Right spending paths
./build/simgo -input examples/htlc.go

# Simple boolean validation
./build/simgo -input examples/simple_logic.go

# Basic swap validation
./build/simgo -input examples/basic_swap.go
```

## Architecture

```
cmd/simgo/          # Compiler binary
pkg/
├── compiler/       # Validation and orchestration
├── jets/           # Jet registry (BIP340, SHA256, etc.)
├── transpiler/     # Core Go -> SimplicityHL conversion
│   ├── transpiler.go  # Main transpiler logic
│   └── patterns.go    # Pattern matching extraction
└── types/          # Type mapping system
    ├── types.go       # Basic type mapping
    └── either.go      # Sum type definitions
examples/           # Contract examples
tests/              # Test suite (28 tests)
```

## Contributing

### Current Focus: Phase 4 - Arrays & Multisig

1. **Array iteration** - For checking multiple signatures
2. **Option arrays** - `[Option<Sig>; 3]` for multisig
3. **Counter logic** - Count valid signatures
4. **Validation functions** - Complex multi-party verification

### Key Files

- `pkg/jets/jets.go` - Jet function registry
- `pkg/types/types.go` - Type system extensions
- `pkg/types/either.go` - Sum type definitions
- `pkg/transpiler/transpiler.go` - Pattern matching and code generation
- `pkg/transpiler/patterns.go` - Match expression extraction
- `examples/` - Real-world contract examples
- `tests/` - Comprehensive test coverage

## Why This Approach Works

1. **Compile-time Safety** - Pre-compute complex operations
2. **Resource Bounds** - Static analysis of all operations
3. **Formal Verification** - Pure pattern matching is provable
4. **Bitcoin Compatible** - Maps directly to Simplicity semantics

## Test Coverage

```
=== Phase 3 Tests ===
TestEitherTypeParsing    - Either<L, R> parsing
TestOptionTypeParsing    - Option<T> parsing
TestTupleTypeParsing     - Tuple type parsing
TestSumTypeDetection     - Sum type identification
TestEitherWitness        - Either witness declaration
TestOptionWitness        - Option witness declaration
TestMatchExpression      - Match generation from if/else
TestSimpleHTLCStructure  - HTLC contract structure
TestMatchArmGeneration   - Match arm code generation

=== Phase 2 Tests ===
TestJetRegistry          - Jet function registry
TestHexTypeInference     - Automatic hex -> u256/u64/etc
TestHexLiteral           - Hex constant parsing
TestJetSigAllHash        - Transaction hash jet
TestJetBIP340Verify      - Signature verification jet
TestP2PKContract         - Full P2PK integration
TestJetCallValidation    - Compiler validation
TestCtx8Type             - SHA-256 context type

=== Core Tests ===
TestBasicFunction        - Function transpilation
TestSimpleValidation     - Boolean validation
TestUnsupportedFeatures  - Error handling
TestSimpleConstants      - Constant extraction
TestBooleanLogic         - Boolean functions
TestWorkingExample       - Basic swap contract
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf)
- [SimplicityHL Documentation](https://github.com/BlockstreamResearch/SimplicityHL)
- [Elements Project](https://github.com/ElementsProject/elements)
