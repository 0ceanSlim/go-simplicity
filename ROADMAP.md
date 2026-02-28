# go-simplicity Release Roadmap

> Current state: Phase 6 complete — Real atomic swap & covenant examples, `check_lock_height` and `output_script_hash` transpilation verified end-to-end.
> Goal: Fully functional transpiler covering all mainstream Bitcoin/Elements contract patterns.

---

## Gap Analysis (Post Phase 5)

### Jet Coverage
Currently 86 jets registered across all categories:
- **Arithmetic**: add/subtract/multiply (8/16/32/64), divide/modulo (32/64) — ✓ complete
- **Comparison**: lt/le/eq (8/16/32/64) — ✓ complete
- **Bitwise logic**: and/or/xor/complement (8/16/32/64), left/right shift — ✓ complete
- **Hash variants**: SHA256 add variants (1–64 bytes), sha_256_block, sha_256_iv — ✓ complete
- **Time locks**: check_lock_time/height/distance/duration, tx_is_final, tx_lock_* — ✓ complete
- **Transaction introspection**: num_inputs/outputs, output_amount, output_script_hash, current_sequence, input_prev_outpoint, version, transaction_id, internal_key, tapleaf_version, tappath, script_cmr, and more — ✓ complete

### Transpiler Logic
- `analyzeFunctionBody` is a stub — helper functions are never actually transpiled, always returns `"true"`
- `evaluateCallExpr` hardcoded: non-jet function calls always return `"true"`
- `analyzeSwitchAsMatch` is a stub — `switch {}` statements silently ignored
- Arithmetic/comparison/bitwise operators (`+`, `-`, `*`, `/`, `%`, `<`, `<=`, `==`, `&`, `|`, `^`) **mapped to jet calls** ✓
- No multi-path Either (3+ spending paths)

### Examples
- `atomic_swap.go` uses no jets — placeholder only
- No covenant/script-hash contract example

---

## ✅ Phase 5 — Arithmetic & Logic Jets (Complete)

**Goal**: Enable numeric arithmetic and bitwise operations in contracts.

### 5.1 — Register Arithmetic Jets
Add to `pkg/jets/jets.go`:
- `Add8/16/32/64` → `add_8/16/32/64`
- `Subtract8/16/32/64` → `subtract_8/16/32/64`
- `Multiply8/16/32/64` → `multiply_8/16/32/64`
- `Divide32/64` → `divide_32/64`
- `Modulo32/64` → `modulo_32/64`
- `Lt8/16/32/64` → `lt_8/16/32/64` (strict less-than)
- `Le8/16/64` → `le_8/16/64` (le_32 already registered)
- `Eq8/16/64` → `eq_8/16/64` (eq_32/256 already registered)

### 5.2 — Register Bitwise Logic Jets
- `And8/16/32/64` → `and_8/16/32/64`
- `Or8/16/32/64` → `or_8/16/32/64`
- `Xor8/16/32/64` → `xor_8/16/32/64`
- `Complement8/16/32/64` → `complement_8/16/32/64`
- `LeftShift32`, `RightShift32` → `left_shift_32`, `right_shift_32`

### 5.3 — Operator-to-Jet Transpilation
In the transpiler, map binary expressions inside `main()` to jet calls:
- `a + b` (uint32) → `let result: u32 = jet::add_32(a, b);`
- `a - b` → `jet::subtract_32(a, b)`
- `a * b` → `jet::multiply_32(a, b)`
- `a < b` → `jet::lt_32(a, b)` → emit as `jet::verify(jet::lt_32(...))`
- `a & b` → `jet::and_32(a, b)`
- Type-width selection based on Go type annotation

### 5.4 — Tests & Example
- Unit tests for all arithmetic jet registrations
- Example: `amount_check.go` — verifies output amount meets a minimum using `add_32` + `le_32`

---

## ✅ Phase 6 — Complete Time Locks & Transaction Introspection (Complete)

**Goal**: Full Bitcoin/Elements primitive coverage for real contract patterns.

### 6.1 — Register Remaining Time Lock Jets
- `CheckLockTime` → `check_lock_time`
- `TxIsFinal` → `tx_is_final`
- `TxLockHeight` → `tx_lock_height`
- `TxLockTime` → `tx_lock_time`
- `CheckLockDistance` → `check_lock_distance`
- `CheckLockDuration` → `check_lock_duration`
- `TxLockDistance` → `tx_lock_distance`
- `TxLockDuration` → `tx_lock_duration`

### 6.2 — Register Transaction Introspection Jets
Core Bitcoin subset:
- `NumInputs` → `num_inputs` (returns u32)
- `NumOutputs` → `num_outputs` (returns u32)
- `InputPrevOutpoint` → `input_prev_outpoint` (u32 index → (u256, u32))
- `OutputScriptHash` → `output_script_hash` (u32 index → u256)
- `InputScriptHash` → `input_script_hash` (u32 index → u256)
- `CurrentSequence` → `current_sequence` (→ u32)
- `CurrentAnnexHash` → `current_annex_hash` (→ Option<u256>)
- `Version` → `version` (→ u32)
- `TransactionId` → `transaction_id` (→ u256)
- `GenesisBlockHash` → `genesis_block_hash` (→ u256)
- `InternalKey` → `internal_key` (→ u256)
- `TapleafVersion` → `tapleaf_version` (→ u8)
- `Tappath` → `tappath` (→ u256)
- `ScriptCmr` → `script_cmr` (→ u256)
- `CurrentAmount` → `current_amount`
- `OutputAmount` → `output_amount` (u32 index → ...)

### 6.3 — Transpiler: Timelock Pattern Generation
Support `jet.CheckLockHeight(height)` and `jet.CheckLockTime(t)` as standalone statements that emit directly as `jet::check_lock_height(param::MIN_HEIGHT)`.

### 6.4 — Example: Real Atomic Swap ✅
Rewrite `atomic_swap.go` using actual jets:
- Either path: Left = hashlock (SHA-256 preimage + BIP340), Right = timelock refund (CheckLockHeight + BIP340)

### 6.5 — Example: Covenant Contract ✅
`covenant.go` — verifies the script hash of an output matches a known hash:
- Uses `output_script_hash`, `eq_256`, `verify`

---

## Phase 7 — Helper Functions & Advanced Control Flow

**Goal**: Fix the two biggest transpiler stubs so real multi-function contracts work.

### 7.1 — Fix `analyzeFunctionBody`
Currently returns `"true"` for any non-trivial function body. Needs to handle:
- `if/else` chains → nested match or sequential jet::verify calls
- Multi-statement bodies with local variable assignments
- Calls to other user-defined functions (recursion-free inlining)
- Boolean return via jet results

### 7.2 — Fix `evaluateCallExpr` for User-Defined Functions
When `main()` calls a helper function, inline its transpiled body rather than returning `"true"`. Track function bodies during analysis, substitute at call sites.

### 7.3 — Implement `analyzeSwitchAsMatch`
`switch w.field { case X: ... default: ... }` → SimplicityHL `match` block.

### 7.4 — Multi-Path Spending (3+ Either arms)
Support nested `Either<A, Either<B, C>>` for contracts with 3+ spending conditions. Example: 3-path contract with hashlock / timelock / multisig.

### 7.5 — Tests
- Full test suite for helper function transpilation
- Switch-statement → match expression tests
- 3-path Either generation test

---

## Phase 8 — SHA-256 Variant Jets

**Goal**: Full SHA-256 family for hashing data of any size.

### 8.1 — Register Remaining Hash Jets
- `SHA256Add1` → `sha_256_ctx_8_add_1`
- `SHA256Add2` → `sha_256_ctx_8_add_2`
- `SHA256Add4` → `sha_256_ctx_8_add_4`
- `SHA256Add8` → `sha_256_ctx_8_add_8`
- `SHA256Add16` → `sha_256_ctx_8_add_16`
- `SHA256Add64` → `sha_256_ctx_8_add_64`
- `SHA256Add128` → `sha_256_ctx_8_add_128`
- `SHA256Add256` → `sha_256_ctx_8_add_256`
- `SHA256Add512` → `sha_256_ctx_8_add_512`
- `SHA256Block` → `sha_256_block`
- `SHA256IV` → `sha_256_iv`

### 8.2 — Auto-Select SHA256AddN
When a `SHA256Add*` call is made, auto-select the correctly-sized variant based on the Go type of the argument (`[16]byte` → `add_16`, `[64]byte` → `add_64`, etc.) rather than requiring the user to manually pick the right function.

### 8.3 — Example: Double SHA-256
`double_sha256.go` — computes SHA256(SHA256(preimage)) and verifies against a stored hash. Demonstrates SHA-256 chaining with multiple add variants.

---

## Phase 9 — Advanced Examples & Documentation

**Goal**: Demonstrate production-grade contract patterns; update all docs.

### Contracts to Add
1. `vault.go` — Vault with hot/cold key spend paths and timelock (combines multisig + timelock + script hash)
2. `oracle_price.go` — Oracle-signed price assertion: verify oracle signature, then check amount against price
3. `relative_timelock.go` — CSV-style relative timelock using `tx_lock_distance`/`tx_lock_duration`
4. `taproot_key_spend.go` — Demonstrates `internal_key` + `tapleaf_version` introspection

### Documentation Updates
- Update readme status to reflect all supported jets/patterns
- Add complete jet reference table to readme
- Add section: "What contract patterns are possible?"
- Add troubleshooting guide for common errors

---

## Phase 10 — Release Quality

**Goal**: Production-quality v1.0.0 release.

### 10.1 — Fix CI & Linter
- Uncomment linter job in `.github/workflows/ci.yml`
- Fix any linter errors uncovered

### 10.2 — Fill Stub Tests
- `TestArrayTypeParsing` — add real assertions
- `TestSimpleMultisigStructure` — verify Charlie's constant emits correctly
- Add integration test for each new example contract

### 10.3 — CLI Improvements
- Add `--list-jets` flag to print all registered jets and their Simplicity names
- Add `--version` flag
- Better error messages with line numbers from the Go source

### 10.4 — Error Quality
- Propagate `token.Pos` through analysis so errors cite the Go source line
- Suggest closest jet name on unknown jet error (`did you mean BIP340Verify?`)
- Detect and reject unsupported operator combinations with clear messages

---

## Priority Order

| Phase | Value | Effort | Status |
|-------|-------|--------|--------|
| **5** — Arithmetic & Logic Jets | High | Medium | ✅ Complete |
| **6** — Time Locks + Tx Introspection | High | Medium | ✅ Complete |
| **7** — Helper Functions / Control Flow | High | High | After 6 |
| **8** — SHA-256 Variants | Medium | Low | Parallel w/ 7 |
| **9** — Advanced Examples | Medium | Medium | After 7 |
| **10** — Release Quality | Medium | Medium | Last |

**Recommended start**: Phase 6 — jets are registered, transpiler just needs the timelock/introspection call patterns and real atomic swap example.
