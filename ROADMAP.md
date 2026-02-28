# go-simplicity Release Roadmap

> Current state: Phase 9 complete — Four advanced example contracts (vault, oracle price, relative timelock, Taproot key spend), 5 new tests, README and ROADMAP updated.
> Goal: Fully functional transpiler covering all mainstream Bitcoin/Elements contract patterns.

---

## Gap Analysis (Post Phase 6)

### Jet Coverage
Currently 91 jets registered across all categories:
- **Arithmetic**: add/subtract/multiply (8/16/32/64), divide/modulo (32/64) — ✓ complete
- **Comparison**: lt/le/eq (8/16/32/64) — ✓ complete
- **Bitwise logic**: and/or/xor/complement (8/16/32/64), left/right shift — ✓ complete
- **Hash variants**: SHA256 add variants (1–64 bytes), sha_256_block, sha_256_iv — ✓ complete
- **Time locks**: check_lock_time/height/distance/duration, tx_is_final, tx_lock_* — ✓ complete
- **Transaction introspection**: num_inputs/outputs, output_amount, output_script_hash, current_sequence, input_prev_outpoint, version, transaction_id, internal_key, tapleaf_version, tappath, script_cmr, and more — ✓ complete

### Transpiler Logic
- Arithmetic/comparison/bitwise operators (`+`, `-`, `*`, `/`, `%`, `<`, `<=`, `==`, `&`, `|`, `^`) **mapped to jet calls** ✓
- No multi-path Either (3+ spending paths) — deferred to Phase 9

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

## ✅ Phase 7 — Helper Functions & Switch Dispatch (Complete)

**Goal**: Fix the three biggest transpiler stubs so real multi-function contracts work.

### ✅ 7.1 — Fix `analyzeFunctionBody`
Linear jet-call sequences in helper functions now transpile correctly. Parameter names resolve as bare identifiers and are substituted at call sites. If/else inside helpers deferred to Phase 8.

### ✅ 7.2 — Fix `evaluateCallExpr` for User-Defined Functions
User-defined helper calls are looked up in `t.functions` and their bodies inlined at call sites using word-boundary parameter substitution.

### ✅ 7.3 — Implement `analyzeSwitchAsMatch`
`switch { case w.IsLeft: ... case !w.IsLeft: ... }` → SimplicityHL `match witness::W { Left(data) => { ... }, Right(sig) => { ... } }`. Mirrors `analyzeIfAsMatch` using `extractSumTypeCondition` on each case clause.

### 7.4 — Multi-Path Spending (3+ Either arms) — Deferred to Phase 10
Support nested `Either<A, Either<B, C>>` for contracts with 3+ spending conditions.

### ✅ 7.5 — Tests
- `TestHelperFunctionBody` — linear helper transpilation
- `TestSwitchMatchGeneration` — switch → match expression
- `TestInlinedHelperCall` — helper inlined at switch arm call site
- `TestExampleHTLCHelper` — end-to-end example with switch + helper inlining

---

## ✅ Phase 8 — SHA-256 Variant Jets (Complete)

**Goal**: Full SHA-256 family for hashing data of any size.

### ✅ 8.1 — Register Remaining Hash Jets
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

### ✅ 8.2 — Auto-Select SHA256AddN
When `jet.SHA256Add(ctx, data)` is written, the transpiler auto-selects the correctly-sized variant based on the Go type of `data` (`[16]byte` → `add_16`, `[64]byte` → `add_64`, `u256` → `add_32`, etc.) at transpile time. Intercepts are in both `analyzeMainFunction` and `evaluateJetCall`.

### ✅ 8.3 — Example: Double SHA-256
`double_sha256.go` — computes SHA256(SHA256(preimage)) and verifies against a stored hash. Demonstrates SHA-256 chaining and `SHA256Add` auto-select resolving to `sha_256_ctx_8_add_32`.

---

## ✅ Phase 9 — Advanced Examples & Documentation (Complete)

**Goal**: Demonstrate production-grade contract patterns; update all docs.

### ✅ Contracts Added
1. `vault.go` — Vault with hot/cold key spend paths and timelock (Either + CheckLockHeight + OutputScriptHash)
2. `oracle_price.go` — Oracle-gated spend: oracle signature authorises (Left), owner emergency withdrawal (Right)
3. `relative_timelock.go` — CSV-style relative timelock using `CheckLockDistance` (block-based)
4. `taproot_key_spend.go` — Taproot introspection: `InternalKey` + `TapleafVersion` before BIP-340 verify

### ✅ Documentation Updates
- README status updated to Phase 9 Complete
- Quick start: 4 new `./build/simgo` commands
- 4 new example sections with Go snippets and generated SimplicityHL
- Example and test counts updated (14 examples, ~60 tests)
- Not Supported: 3+ spending paths deferred to Phase 10

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
| **7** — Helper Functions / Control Flow | High | High | ✅ Complete |
| **8** — SHA-256 Variants | Medium | Low | ✅ Complete |
| **9** — Advanced Examples | Medium | Medium | ✅ Complete |
| **10** — Release Quality | Medium | Medium | Last |

**Recommended start**: Phase 6 — jets are registered, transpiler just needs the timelock/introspection call patterns and real atomic swap example.
