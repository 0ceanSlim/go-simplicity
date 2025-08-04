# go-simplicity

A Go to SimplicityHL transpiler that converts Go smart contract logic into Simplicity bytecode for Bitcoin and Elements sidechains.

## Project Status: Early Stage - Boolean Pattern Matching

This transpiler currently handles **boolean-based smart contract patterns** and provides a foundation for full Simplicity support. While not yet feature-complete, it demonstrates a novel approach to blockchain smart contracts through compile-time evaluation and pure pattern matching.

## What Works Right Now

### ✅ Currently Supported Go → SimplicityHL

- **Boolean functions** with pattern matching
- **Compile-time expression evaluation** (constants, arithmetic, comparisons)
- **Witness/parameter separation** (runtime vs compile-time values)
- **Simple control flow** (if/else converted to pattern matching)
- **Function composition** with boolean parameters
- **Basic types**: `bool`, `uint32`, `uint64`

### Example: Working P2PK Contract

**Input Go Code:**

```go
package main

func ValidateP2PK(signatureValid bool) bool {
    return signatureValid
}

func main() {
    signatureValid := true
    result := ValidateP2PK(signatureValid)

    if !result {
        return
    }
}
```

**Generated SimplicityHL:**

```rust
mod witness {
    const SIGNATURE_VALID: bool = true;
}

mod param {
}

fn validate_p2pk(signature_valid: bool) -> bool {
    signature_valid
}

fn main() {
    assert!(validate_p2pk(witness::SIGNATURE_VALID));
}
```

## Target: Real Simplicity Contracts

The goal is to generate contracts like these working Simplicity examples:

<details>
<summary>P2PK (Pay to Public Key)</summary>

```rust
mod witness {
    const ALICE_SIGNATURE: [u8; 64] = 0x9a3a093012693c1d...;
}
mod param {
    const ALICE_PUBLIC_KEY: u256 = 0x9bef8d556d80e43a...;
}
fn main() {
    jet::bip_0340_verify((param::ALICE_PUBLIC_KEY, jet::sig_all_hash()), witness::ALICE_SIGNATURE)
}
```

</details>

<details>
<summary>2-of-3 Multisig</summary>

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
<summary>HTLC (Hash Time Locked Contract)</summary>

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

## Critical Gap Analysis

### ❌ **Missing for Real Simplicity Contracts:**

1. **Bitcoin Types**

   - `u256`, `[u8; 64]`, `Pubkey`, `Signature`
   - Type aliases for Bitcoin primitives

2. **Complex Types**

   - `Either<A, B>` sum types
   - `Option<T>` nullable types
   - Fixed arrays `[T; N]`
   - Tuples `(A, B, C)`

3. **Pattern Matching**

   - `match` expressions on non-boolean types
   - Destructuring assignments
   - Complex case analysis

4. **Jet Functions**

   - `jet::bip_0340_verify()` - signature verification
   - `jet::sha_256_ctx_8_*()` - hashing functions
   - `jet::sig_all_hash()` - transaction hashing
   - `jet::eq_256()`, `jet::le_32()` - comparisons

5. **Advanced Control Flow**
   - Multiple return paths
   - Complex witness data handling
   - Recursive data structures

## Development Roadmap

### 🎯 **Phase 2: Core Bitcoin Types (CRITICAL)**

**Goal:** Generate basic P2PK and P2PKH contracts

- [ ] Add `u256` type support
- [ ] Implement `[u8; 64]` arrays
- [ ] Create Bitcoin type aliases (`Pubkey`, `Signature`, `Hash`)
- [ ] Add jet function stubs (`jet::bip_0340_verify`, etc.)
- [ ] Test with real P2PK contract generation

**Target:** Successfully transpile P2PK contract from Go to working SimplicityHL

### 🎯 **Phase 3: Sum Types & Pattern Matching**

**Goal:** Generate HTLC and conditional contracts

- [ ] Implement `Either<A, B>` types
- [ ] Add `Option<T>` support
- [ ] Create `match` expression generation
- [ ] Handle tuple destructuring
- [ ] Support complex witness data patterns

**Target:** Successfully transpile HTLC contract with Left/Right paths

### 🎯 **Phase 4: Arrays & Multisig**

**Goal:** Generate multisig contracts

- [ ] Fixed array support `[T; N]`
- [ ] Array iteration and counting
- [ ] Option array handling `[Option<T>; N]`
- [ ] Counter-based logic
- [ ] Complex validation functions

**Target:** Successfully transpile 2-of-3 multisig contract

### 🎯 **Phase 5: Advanced Contracts**

**Goal:** Generate vault, inheritance, and complex contracts

- [ ] Recursive covenant patterns
- [ ] Time-based logic
- [ ] Oracle data validation
- [ ] Multi-path spending conditions

## Installation & Usage

```bash
# Install
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
make build

# Test current capabilities
./build/simgo -input examples/basic_swap.go -output basic_swap.shl

# Run tests
make test
```

## Current Examples That Work

```bash
# Simple boolean validation
./build/simgo -input examples/simple_logic.go

# Basic payment validation
./build/simgo -input examples/simple_payment.go

# Multisig (boolean-only version)
./build/simgo -input examples/simple_multisig.go
```

## Contributing to Bridge the Gap

### 🚨 **High Priority Issues**

1. **Bitcoin Types** - Implement `u256`, `[u8; 64]` in type mapper
2. **Jet Functions** - Add stubs for cryptographic operations
3. **Either Types** - Critical for real contracts
4. **Pattern Matching** - Beyond boolean logic

### 📝 **How to Contribute**

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

**Key Areas:**

- `pkg/types/` - Type system extensions
- `pkg/transpiler/` - Pattern matching and code generation
- `examples/` - Real-world contract examples
- `tests/` - Comprehensive test coverage

## Architecture

```
cmd/simgo/          # Compiler binary
pkg/
├── compiler/       # Validation and orchestration
├── transpiler/     # Core Go → SimplicityHL conversion
└── types/          # Type mapping (NEEDS EXTENSION)
examples/           # Contract examples (BASIC ONLY)
tests/              # Test suite
```

## Why This Approach Works

1. **Compile-time Safety** - Pre-compute complex operations
2. **Resource Bounds** - Static analysis of all operations
3. **Formal Verification** - Pure pattern matching is provable
4. **Bitcoin Compatible** - Maps directly to Simplicity semantics

## Limitations & Honest Assessment

### Current Reality

- **Boolean-heavy contracts only**
- **No cryptographic operations**
- **Limited type system**
- **Simple pattern matching only**

### But the Foundation is Solid

- **Correct transpilation strategy**
- **Working witness/parameter separation**
- **Extensible architecture**
- **Clear path to full Simplicity support**

## Next Milestone: First Real Bitcoin Contract

**Goal:** Generate a working P2PK contract that validates signatures

**Success Criteria:**

```rust
// Generated from Go input
mod witness {
    const ALICE_SIGNATURE: [u8; 64] = /* actual signature */;
}
mod param {
    const ALICE_PUBLIC_KEY: u256 = /* actual pubkey */;
}
fn main() {
    jet::bip_0340_verify((param::ALICE_PUBLIC_KEY, jet::sig_all_hash()), witness::ALICE_SIGNATURE)
}
```

**Required Additions:**

1. `u256` and `[u8; 64]` types
2. Jet function generation
3. Bitcoin type aliases
4. Constant hex literal support

---

**This is experimental software with a clear development path. The boolean pattern matching foundation is solid, but significant work remains to support real Bitcoin contracts.**

## License

MIT License - see [LICENSE](LICENSE) for details.

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf)
- [SimplicityHL Documentation](https://github.com/BlockstreamResearch/SimplicityHL)
- [Elements Project](https://github.com/ElementsProject/elements)
