# Contributing to go-simplicity

Contributions are welcome — bug reports, new contract examples, transpiler improvements, and documentation fixes.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- `make` (optional but recommended)

### Setup

```bash
git clone https://github.com/0ceanslim/go-simplicity.git
cd go-simplicity
make dev-setup   # installs golangci-lint
make test        # verify everything works
make build       # build the simgo binary
```

### Project Structure

```
cmd/simgo/          # CLI binary
pkg/
├── compiler/       # Validation and orchestration
├── jets/           # Jet registry (91 jets)
├── transpiler/     # Core Go → SimplicityHL AST walker
│   ├── transpiler.go
│   ├── patterns.go
│   └── arrays.go
├── types/          # Type mapping
└── testkeys/       # BIP-340 test vectors
examples/           # Contract examples (14 files + 4 testable)
tests/              # Test suite (60 tests)
docs/               # Extended documentation
```

## Development Workflow

1. Fork the repository and create a branch from `main`
2. Make your changes and add tests
3. Run the checks locally:

   ```bash
   make fmt      # format
   make test     # tests must pass
   make build    # binary must build
   make ci       # full CI check (fmt + tests + mod verify)
   ```

4. Open a pull request

## Coding Standards

- Follow standard Go formatting — `gofmt -s` must produce no output
- Match the style of the surrounding code
- Add tests for new transpiler behaviour — see `tests/` for patterns
- New jet registrations go in `pkg/jets/jets.go`
- New contract examples go in `examples/` with a `//go:build ignore` tag
- New example tests go in `tests/examples_test.go`

## Adding a New Jet

1. Register it in `pkg/jets/jets.go`:

   ```go
   r.jets["MyJet"] = JetInfo{
       GoName:         "MyJet",
       SimplicityName: "my_jet",
       ParamTypes:     []string{"u32"},
       ReturnType:     "()",
   }
   ```

2. Add a registry test in `tests/jet_test.go`
3. Add a usage example in `examples/` if it enables a new contract pattern
4. Add an end-to-end compile test in `tests/examples_test.go`

## Adding a Contract Example

1. Create `examples/your_contract.go` with `//go:build ignore` as the first line
2. Add a `TestExampleYourContract` function to `tests/examples_test.go`
3. Document the pattern in `docs/contract-patterns.md`
4. Add the example to the `examples` target in the `makefile`

## Commit Messages

Use conventional commits:

```
feat(transpiler): add support for N-of-M multisig patterns
fix(jets): correct return type for tapleaf_version
docs(examples): add vault contract example
test(jet_test): add Phase 9 taproot jet registry checks
```

## Issues and Feature Requests

Use [GitHub Issues](https://github.com/0ceanslim/go-simplicity/issues) for bug reports and feature requests. Please include:

- Go source that demonstrates the problem
- Expected SimplicityHL output
- Actual output or error message
