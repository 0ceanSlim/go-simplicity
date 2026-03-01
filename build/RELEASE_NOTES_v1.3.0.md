# go-simplicity v1.3.0

## Boolean if/else support

This release adds `if boolVar { ... } else { ... }` to the transpiler, enabling multi-mode contracts — contracts that behave differently based on a runtime condition computed from transaction data.

---

## What's New

### Boolean if/else

Any variable assigned from a comparison or boolean jet (`Lt64`, `Le64`, `Eq64`, `Lt128`, etc.) can now be used as an `if` condition directly:

```go
isRemoveMode := jet.Lt64(newReserve0, reserve0)

if isRemoveMode {
    // remove-liquidity path: proportional payout checks
} else {
    // swap / add-liquidity path: k-invariant check
}
```

Compiles to:

```
let is_remove_mode: bool = jet::lt_64(new_reserve0, reserve0);
match is_remove_mode {
    true  => {
        let total_supply: u64 = jet::input_amount(param::LP_SUPPLY_INPUT);
        ...
    },
    false => {
        let k_old: u128 = jet::multiply_64(reserve0, reserve1);
        ...
    }
}
```

All jet calls before the `if` are emitted in declaration order (including standalone `jet.Verify(...)` calls). Inside each arm, the full range of statements is supported:

- Jet call assignments: `totalSupply := jet.InputAmount(2)`
- Binary expression assignments: `payout := reserve - newReserve` → `subtract_64`
- Nested jet calls: `jet.Verify(jet.Le128(jet.Multiply64(...), ...))`
- Arm-local variables are tracked so subsequent arm statements can reference them

---

## Transpiler Changes

**`pkg/transpiler/patterns.go`**
- Added `IsBoolMatch bool` to `MatchExpression`

**`pkg/transpiler/transpiler.go`**
- `analyzeIfAsMatch`: extended to detect `if identVar { ... }` where `identVar` resolves to a `bool`-typed jet call variable
- New `analyzeBoolIfElse`: builds a `MatchExpression` with `true` / `false` cases
- New `analyzeArmBodyStmts`: processes arm body statements; arm-local jet variables are appended to `t.jetCalls` temporarily (saved/restored per arm) so subsequent arm statements can reference them
- New `analyzeArmBodyStmt`: handles jet call assignments and binary expression assignments in arm body context
- `generateMainFunction`: new `hasBoolMatch` path emits all top-level jet calls (named and standalone) in declaration order, then the match expression
- `analyzeFunctionBody`: removed early-return stub that blocked `IfStmt`

---

## Tests

2 new test functions in `tests/jet_test.go`:

- `TestBooleanIfElse` — basic boolean if/else with mode detection, verifies `match is_remove_mode`, `true =>`, `false =>`, `lt_64`, `multiply_64`, `le_128`
- `TestBooleanIfElseWithSubtract` — boolean if/else with arithmetic in the true arm, verifies `subtract_64` is emitted inside the match arm

---

## What This Unlocks

The [anchor](https://github.com/0ceanslim/anchor) AMM protocol is now fully compilable. All four contracts use boolean if/else for mode detection:

| Contract | Condition | True arm | False arm |
|----------|-----------|----------|-----------|
| `pool_a.go` | `newReserve0 < reserve0` | proportional payout checks | k-invariant |
| `pool_b.go` | `newReserve1 < reserve1` | proportional payout checks | k-invariant |
| `lp_supply.go` | `totalSupply < newTotalSupply` | LP mint proportionality | LP burn amount check |

`pool_creation.go` requires no branching (straight-line sqrt verification).

---

## Binaries

| File | OS | Arch |
|---|---|---|
| `simgo-windows-amd64.exe` | Windows | x86-64 |
| `simgo-linux-amd64` | Linux | x86-64 |
| `simgo-darwin-arm64` | macOS | Apple Silicon |
| `simgo-darwin-amd64` | macOS | Intel |
