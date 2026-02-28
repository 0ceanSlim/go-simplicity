# go-simplicity

[![CI](https://github.com/0ceanslim/go-simplicity/actions/workflows/ci.yml/badge.svg)](https://github.com/0ceanslim/go-simplicity/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/0ceanslim/go-simplicity)](https://goreportcard.com/report/github.com/0ceanslim/go-simplicity)
[![Go Reference](https://pkg.go.dev/badge/github.com/0ceanslim/go-simplicity.svg)](https://pkg.go.dev/github.com/0ceanslim/go-simplicity)
[![GitHub release](https://img.shields.io/github/v/release/0ceanslim/go-simplicity)](https://github.com/0ceanslim/go-simplicity/releases)
[![Downloads](https://img.shields.io/github/downloads/0ceanslim/go-simplicity/total)](https://github.com/0ceanslim/go-simplicity/releases)
[![Stars](https://img.shields.io/github/stars/0ceanslim/go-simplicity)](https://github.com/0ceanslim/go-simplicity/stargazers)
[![Go Version](https://img.shields.io/github/go-mod/go-version/0ceanslim/go-simplicity)](https://pkg.go.dev/github.com/0ceanslim/go-simplicity)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/0ceanslim/go-simplicity)

A Go-to-SimplicityHL transpiler for writing Bitcoin/Elements smart contracts in Go.

Write contract logic in idiomatic Go. The transpiler converts it to [SimplicityHL](https://github.com/BlockstreamResearch/SimplicityHL) — a typed, functional intermediate language that compiles to Simplicity bytecode for Bitcoin and Elements sidechains.

---

## Quick Start

```bash
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
make build

./build/simgo -input examples/p2pk.go
./build/simgo -input examples/htlc.go
./build/simgo -input examples/vault.go

make test
```

---

## What It Does

- **Transpiles Go contracts** — `go run cmd/simgo/main.go -input contract.go` prints ready-to-use SimplicityHL
- **Witness/parameter separation** — `var sig [64]byte` → `mod witness`; `const Pubkey = 0x...` → `mod param`
- **Sum types** — struct with `IsLeft bool` → `Either<L, R>`; struct with `IsSome bool` + `Value T` → `Option<T>`
- **Match expression generation** — `if w.IsLeft { ... } else { ... }` and `switch { case w.IsLeft: ... }` both compile to `match witness::W { Left(...) => { ... }, Right(...) => { ... } }`
- **Helper function inlining** — user-defined helper functions are emitted as named functions and inlined at call sites
- **Operator mapping** — `+`, `-`, `*`, `/`, `%`, `<`, `<=`, `==`, `&`, `|`, `^` auto-map to the correct `add_N`/`subtract_N`/`lt_N`/`and_N`/etc. jet based on operand width
- **SHA256Add auto-select** — `jet.SHA256Add(ctx, data)` resolves to the correctly-sized `sha_256_ctx_8_add_N` variant at transpile time
- **102 jets registered** across signature, hash, arithmetic, comparison, bitwise, time lock, transaction introspection, and Elements amount/issuance categories

---

## Supported Contract Patterns

| Pattern | Example | Key Jets |
|---------|---------|----------|
| Pay to public key | `examples/p2pk.go` | `sig_all_hash`, `bip_0340_verify` |
| Hash time lock (HTLC) | `examples/htlc.go` | `sha_256_ctx_8_*`, `eq_256`, `bip_0340_verify` |
| Absolute timelock | `examples/atomic_swap.go` | `check_lock_height`, `bip_0340_verify` |
| Relative timelock (CSV) | `examples/relative_timelock.go` | `check_lock_distance` |
| Covenant | `examples/covenant.go` | `output_script_hash`, `eq_256` |
| Vault (hot/cold key) | `examples/vault.go` | `check_lock_height`, `output_script_hash` |
| Oracle-gated spend | `examples/oracle_price.go` | `bip_0340_verify` (two pubkeys) |
| Taproot key spend | `examples/taproot_key_spend.go` | `internal_key`, `tapleaf_version` |
| 2-of-3 multisig | `examples/multisig.go` | `Option<[u8; 64]>`, counter accumulation |
| Helper functions | `examples/htlc_helper.go` | switch dispatch + inlining |
| Double SHA-256 | `examples/double_sha256.go` | `SHA256Add` auto-select |

See [docs/contract-patterns.md](docs/contract-patterns.md) for Go source and generated SimplicityHL for each pattern.

---

## Available Jets

102 jets registered. Quick reference by category:

| Category | Jets |
|----------|------|
| **Signature** | `bip_0340_verify`, `sig_all_hash` |
| **SHA-256** | `sha_256_ctx_8_init/add_N/finalize`, `sha_256_block`, `sha_256_iv` (N = 1–512) |
| **Arithmetic** | `add/subtract/multiply_N`, `divide/modulo_32/64` (N = 8/16/32/64); `add_128`, `subtract_128` |
| **Comparison** | `lt/le/eq_N` (N = 8/16/32/64/128/256) |
| **Bitwise** | `and/or/xor/complement_N` (N = 8/16/32/64) |
| **Time locks** | `check_lock_height/time/distance/duration`, `tx_lock_*`, `tx_is_final` |
| **Tx introspection** | `num_inputs/outputs`, `output_script_hash`, `input_prev_outpoint`, `version`, `transaction_id`, `internal_key`, `tapleaf_version`, `tappath`, `script_cmr`, and more |
| **Elements amounts** | `output_asset`, `output_amount`, `input_asset`, `input_amount`, `current_asset`, `current_amount` |
| **Elements issuance** | `issuance_asset_amount`, `issuance_token_amount`, `new_issuance_contract` |
| **Utility** | `verify` |

Run `simgo -list-jets` to print the full list.

---

## Types

| Go | SimplicityHL |
|----|-------------|
| `bool` | `bool` |
| `uint8` | `u8` |
| `uint16` | `u16` |
| `uint32` | `u32` |
| `uint64` | `u64` |
| `0x...` hex literal | `u8`/`u16`/`u32`/`u64`/`u128`/`u256` (inferred from length) |
| `[N]byte` | `[u8; N]` |
| struct with `IsLeft bool` | `Either<L, R>` |
| struct with `IsSome bool` + `Value T` | `Option<T>` |

---

## Architecture

```
cmd/simgo/          # CLI binary (-input, -output, -target, -debug, -list-jets, -version)
pkg/
├── compiler/       # Validation and orchestration
├── jets/           # Jet registry (102 jets)
├── transpiler/     # Core Go → SimplicityHL AST walker
│   ├── transpiler.go   # Analysis, code generation, helper inlining
│   ├── patterns.go     # Either/Option match extraction, switch dispatch
│   └── arrays.go       # Fixed-size arrays
├── types/          # Type mapping (Go → Simplicity)
└── testkeys/       # BIP-340 spec test vectors
examples/           # 16 contract examples + 4 testable variants
tests/              # 60 tests
```

---

## Not Supported

- Dynamic arrays, slices, maps, channels, goroutines, interfaces
- 3+ spending paths / nested `Either` (two arms only)
- Recursive function calls
- `if/else` inside helper function bodies
- Imports other than `simplicity/jet`
- String types

---

## Documentation

- [Contract patterns & examples](docs/contract-patterns.md)

---

## Resources

- [Simplicity Paper](https://blockstream.com/simplicity.pdf)
- [SimplicityHL repository](https://github.com/BlockstreamResearch/SimplicityHL)
- [Elements Project](https://github.com/ElementsProject/elements)

---

## License

MIT — see [license](license) for details.
