# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Phase 2 Complete - P2PK Contracts

This transpiler now supports **real P2PK (Pay-to-Public-Key) contracts** with hex literals and Simplicity jet functions. The foundation is solid and expanding toward full Simplicity support.

## What Works Right Now

### Currently Supported Go -> SimplicityHL

- **Boolean functions** with pattern matching
- **Compile-time expression evaluation** (constants, arithmetic, comparisons)
- **Witness/parameter separation** (runtime vs compile-time values)
- **Simple control flow** (if/else converted to pattern matching)
- **Function composition** with boolean parameters
- **Basic types**: `bool`, `uint8`, `uint16`, `uint32`, `uint64`, `u256`, `Ctx8`
- **Hex literals** with automatic type inference (0x... -> u8/u16/u32/u64/u128/u256)
- **Fixed-size arrays**: `[T; N]` including `[u8; 64]` for signatures
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

## Target: Real Simplicity Contracts

The goal is to generate contracts like these working Simplicity examples:

<details>
<summary>P2PK (Pay to Public Key) - NOW WORKING</summary>

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

</details>

<details>
<summary>2-of-3 Multisig (Phase 4)</summary>

```rust
mod witness {
    const SIGNATURES_2_OF_3: [Option<[u8; 64]>; 3] = [Some(0xd95c15407cda...), None, Some(0x6f0854f1bb0d...)];
}
mod param {
    const ALICE_PUBLIC_KEY: u256 = 0x9bef8d556d80e43a...;
    const BOB_PUBLIC_KEY: u256 = 0xe37d58a1aae4ba05...;
    const CHARLIE_PUBLIC_KEY: u256 = 0x688466442a134ee3...;
}
fn main() {
    check2of3multisig([param::ALICE_PUBLIC_KEY, param::BOB_PUBLIC_KEY, param::CHARLIE_PUBLIC_KEY], witness::SIGNATURES_2_OF_3);
}
```

</details>

<details>
<summary>HTLC (Hash Time Locked Contract) (Phase 3)</summary>

```rust
mod witness {
    const COMPLETE_OR_CANCEL: Either<(u256, [u8; 64]),[u8; 64]> = Left((0x9bf49a6a0755f953..., 0xddd1b8079208208e...));
}
fn main() {
    match witness::COMPLETE_OR_CANCEL {
        Left(preimage_and_sig: (u256, Signature)) => {
            let (preimage, recipient_sig): (u256, Signature) = preimage_and_sig;
            complete_spend(preimage, recipient_sig);
        },
        Right(sender_sig: Signature) => cancel_spend(sender_sig),
    }
}
```

</details>

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

### Phase 3: Sum Types & Pattern Matching (NEXT)

**Goal:** Generate HTLC and conditional contracts

- [ ] Implement `Either<A, B>` sum types
- [ ] Add `Option<T>` support
- [ ] Create `match` expression generation
- [ ] Handle tuple destructuring
- [ ] Support complex witness data patterns

**Target:** Successfully transpile HTLC contract with Left/Right paths

### Phase 4: Arrays & Multisig

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

# Run tests
make test
```

## Current Examples

```bash
# P2PK contract with signature verification
./build/simgo -input examples/p2pk.go

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
└── types/          # Type mapping system
examples/           # Contract examples
tests/              # Test suite (17 tests)
```

## Contributing

### Current Focus: Phase 3 - Sum Types

1. **Either<A, B>** - Critical for HTLC contracts
2. **Option<T>** - For optional signatures in multisig
3. **Match expressions** - Beyond boolean logic
4. **Tuple destructuring** - For complex witness data

### Key Files

- `pkg/jets/jets.go` - Jet function registry
- `pkg/types/types.go` - Type system extensions
- `pkg/transpiler/transpiler.go` - Pattern matching and code generation
- `examples/` - Real-world contract examples
- `tests/` - Comprehensive test coverage

## Why This Approach Works

1. **Compile-time Safety** - Pre-compute complex operations
2. **Resource Bounds** - Static analysis of all operations
3. **Formal Verification** - Pure pattern matching is provable
4. **Bitcoin Compatible** - Maps directly to Simplicity semantics

## Test Coverage

```
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
