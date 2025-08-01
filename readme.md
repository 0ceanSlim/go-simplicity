# go-simplicity

A Go to Simplicity transpiler that enables developers to write smart contracts in Go and compile them to Simplicity bytecode for blockchain applications.

## Overview

Simplicity is a typed, combinator-based, functional language designed for blockchain applications that provides formal verification capabilities and static resource analysis. This project bridges the gap between Go's familiar syntax and Simplicity's powerful guarantees.

## Features

- **Go to SimplicityHL Transpilation**: Convert Go code to SimplicityHL (Rust-like syntax)
- **Type Safety**: Automatic mapping of Go types to Simplicity types
- **Static Analysis**: Validates Go code against Simplicity's constraints
- **Bitcoin Integration**: Support for Bitcoin-specific types and operations
- **Example Contracts**: Includes atomic swaps, DEX orders, and basic validation

## Installation

```bash
go install github.com/yourusername/go-simplicity/cmd/simgo@latest
```

Or clone and build:

```bash
git clone https://github.com/yourusername/go-simplicity.git
cd go-simplicity
go build -o simgo cmd/simgo/main.go
```

## Quick Start

### 1. Write Go Code

```go
package main

func ValidateAmount(amount uint64) bool {
    return amount > 0
}

func CalculateFee(amount uint64, rate uint64) uint64 {
    return (amount * rate) / 10000
}

func main() {
    var amount uint64 = 1000
    var rate uint64 = 25 // 0.25%
    result := ValidateAmount(amount)
    if !result {
        return
    }
}
```

### 2. Compile to SimplicityHL

```bash
simgo -input examples/basic_swap.go -output basic_swap.shl
```

### 3. Generated SimplicityHL

```rust
// Generated from Go source by go-simplicity compiler

fn validate_amount(amount: u64) -> bool {
    (amount > 0)
}

fn calculate_fee(amount: u64, rate: u64) -> u64 {
    ((amount * rate) / 10000)
}

fn main() {
    let amount: u64 = 1000;
    let rate: u64 = 25;
    let result = validate_amount(amount);
    match result {
        true => {
        },
        false => {
            ()
        },
    }
}
```

## Supported Go Features

### ✅ Supported

- Basic types: `bool`, `uint8`, `uint16`, `uint32`, `uint64`
- Fixed-size arrays: `[32]byte`, `[4]uint64`
- Functions with parameters and return values
- Basic arithmetic: `+`, `-`, `*`, `/`
- Comparisons: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Boolean logic: `&&`, `||`, `!`
- If/else statements
- Variable declarations and assignments
- Constants

### ❌ Not Supported (Simplicity Limitations)

- Loops (`for`, `range`)
- Slices (`[]T`)
- Maps (`map[K]V`)
- Channels (`chan T`)
- Goroutines (`go`)
- Interfaces
- Pointers
- Recursion
- Dynamic memory allocation

## Type Mapping

| Go Type         | Simplicity Type | Description             |
| --------------- | --------------- | ----------------------- |
| `bool`          | `bool`          | Boolean values          |
| `uint8`, `byte` | `u8`            | 8-bit unsigned integer  |
| `uint16`        | `u16`           | 16-bit unsigned integer |
| `uint32`        | `u32`           | 32-bit unsigned integer |
| `uint64`        | `u64`           | 64-bit unsigned integer |
| `[N]T`          | `[T; N]`        | Fixed-size array        |
| `struct{...}`   | `(T1, T2, ...)` | Tuple                   |

## Bitcoin Integration

When importing `"simplicity/bitcoin"`, additional types become available:

```go
import "simplicity/bitcoin"

func CheckSignature(pubkey bitcoin.Pubkey, sig bitcoin.Signature, msg bitcoin.Hash) bool {
    // This would use Simplicity jets for ECDSA verification
    return bitcoin.VerifySignature(pubkey, sig, msg)
}
```

## Examples

The `examples/` directory contains several example contracts:

- `basic_swap.go`: Simple amount validation and fee calculation
- `atomic_swap.go`: Hash time-locked contract (HTLC)
- `dex_order.go`: Decentralized exchange order matching

## Command Line Usage

```bash
simgo [options]

Options:
  -input string
        Input Go source file (required)
  -output string
        Output SimplicityHL file (default: stdout)
  -target string
        Target format: simplicityhl, simplicity (default "simplicityhl")
  -debug
        Enable debug output
```

## Development Status

This is an early-stage project implementing the basic transpilation pipeline:

- ✅ **Phase 1**: Basic Go to SimplicityHL transpilation
- 🚧 **Phase 2**: Bitcoin-specific types and jets integration
- 📋 **Phase 3**: Advanced optimization and tooling

## Architecture

```
cmd/simgo/          # Compiler binary
pkg/
├── compiler/       # Main compilation logic
├── parser/         # Go AST parsing (uses go/parser)
├── transpiler/     # Go → SimplicityHL conversion
└── types/          # Type system mapping
examples/           # Example contracts
tests/              # Test suite
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

### Phase 1: Core Transpiler (Current)

- [x] Basic Go parsing
- [x] Type mapping
- [x] SimplicityHL generation
- [x] Function transpilation
- [ ] Comprehensive testing

### Phase 2: Smart Contract Features

- [ ] Bitcoin-specific types (Hash, Pubkey, Signature)
- [ ] Simplicity jets integration
- [ ] Asset handling (L-BTC, L-USDT)
- [ ] Error handling and assertions

### Phase 3: Advanced Features

- [ ] Optimization passes
- [ ] VS Code extension
- [ ] Debug support
- [ ] Direct Simplicity compilation (bypass SimplicityHL)

## Related Projects

- [Simplicity](https://github.com/BlockstreamResearch/simplicity) - The core Simplicity language
- [SimplicityHL](https://github.com/BlockstreamResearch/SimplicityHL) - High-level Simplicity frontend
- [Elements](https://github.com/ElementsProject/elements) - Bitcoin sidechain with Simplicity support

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf) - Original research paper
- [Simplicity Docs](https://simplicity.readthedocs.io/) - Language documentation
- [Blockstream Research](https://blockstream.com/research/) - Latest developments

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Russell O'Connor and the Blockstream Research team for creating Simplicity
- The Go team for excellent parsing tools in the `go/ast` package
- The Bitcoin development community for inspiration

---

**Note**: This is experimental software. Do not use in production without thorough testing and auditing.
