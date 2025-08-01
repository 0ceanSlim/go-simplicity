# Contributing to go-simplicity

We welcome contributions to the go-simplicity project! This document provides guidelines for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.20 or later
- Git
- Make (optional, but recommended)

### Setting up the Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:

   ```bash
   git clone https://github.com/yourusername/go-simplicity.git
   cd go-simplicity
   ```

3. **Set up the development environment**:

   ```bash
   make dev-setup
   ```

4. **Verify the setup**:
   ```bash
   make test
   make build
   ```

### Project Structure

```
go-simplicity/
├── cmd/simgo/          # Main compiler binary
├── pkg/
│   ├── compiler/       # Core compilation logic
│   ├── transpiler/     # Go to SimplicityHL conversion
│   └── types/          # Type system mapping
├── examples/           # Example Go contracts
├── tests/              # Test files
├── docs/               # Documentation
└── scripts/            # Build and utility scripts
```

## Development Process

### Branching Strategy

- `main`: Stable, production-ready code
- `develop`: Integration branch for features
- `feature/*`: Feature development branches
- `bugfix/*`: Bug fix branches
- `hotfix/*`: Critical fixes for production

### Workflow

1. Create a feature branch from `develop`:

   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards

3. Write or update tests for your changes

4. Ensure all tests pass:

   ```bash
   make test
   make lint
   ```

5. Commit your changes with a descriptive commit message

6. Push to your fork and create a pull request

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Write comprehensive comments for public APIs
- Keep functions focused and reasonably sized
- Handle errors explicitly

### Code Organization

- Group related functionality in packages
- Use interfaces to define contracts
- Minimize dependencies between packages
- Follow the principle of least privilege

### Example Code Style

```go
// MapGoType converts a Go AST type to its Simplicity equivalent.
// It returns an error if the Go type is not supported in Simplicity.
func (tm *TypeMapper) MapGoType(goType ast.Expr) (string, error) {
    switch t := goType.(type) {
    case *ast.Ident:
        return tm.mapIdentType(t)
    case *ast.ArrayType:
        return tm.mapArrayType(t)
    default:
        return "", fmt.Errorf("unsupported Go type: %T", goType)
    }
}
```

### Commit Messages

Follow the conventional commit format:

```
type(scope): brief description

Longer description if needed

Fixes #123
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Build process or auxiliary tool changes

Examples:

- `feat(transpiler): add support for struct types`
- `fix(parser): handle empty function bodies correctly`
- `docs(readme): update installation instructions`

## Testing

### Writing Tests

- Write unit tests for all new functionality
- Use table-driven tests for multiple test cases
- Include both positive and negative test cases
- Mock external dependencies

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test input",
            expected: "expected output",
            wantErr:  false,
        },
        {
            name:    "invalid input",
            input:   "bad input",
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tc.input)

            if tc.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if result != tc.expected {
                t.Errorf("expected %q, got %q", tc.expected, result)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/transpiler/

# Run benchmarks
make bench
```

## Documentation

### Code Documentation

- Document all public functions, types, and constants
- Use complete sentences in comments
- Include examples for complex functions
- Document any non-obvious behavior

### User Documentation

- Update README.md for user-facing changes
- Add examples to the examples/ directory
- Update API documentation when interfaces change

### Documentation Style

```go
// TypeMapper maps Go types to their Simplicity equivalents.
// It maintains a registry of built-in type mappings and provides
// methods for converting Go AST types to Simplicity type strings.
type TypeMapper struct {
    builtinTypes map[string]string
}

// MapGoType converts a Go AST type expression to its Simplicity equivalent.
// It returns the Simplicity type as a string, or an error if the Go type
// is not supported.
//
// Example:
//   mapper := NewTypeMapper()
//   simplicityType, err := mapper.MapGoType(goASTType)
//   if err != nil {
//       return fmt.Errorf("unsupported type: %w", err)
//   }
func (tm *TypeMapper) MapGoType(goType ast.Expr) (string, error) {
    // implementation
}
```

## Submitting Changes

### Pull Request Process

1. **Ensure your branch is up to date** with the target branch
2. **Run all checks** locally:
   ```bash
   make ci
   ```
3. **Create a pull request** with:
   - Clear title and description
   - Reference to any related issues
   - Screenshots for UI changes (if applicable)
   - Breaking change notes (if applicable)

### Pull Request Template

When creating a pull request, include:

- **What**: Brief description of the change
- **Why**: Motivation and context
- **How**: Technical details of the implementation
- **Testing**: How the change was tested
- **Breaking Changes**: Any breaking changes and migration notes

### Review Process

- All PRs require at least one review
- Address reviewer feedback promptly
- Keep the PR focused and reasonably sized
- Rebase and squash commits before merging

## Additional Guidelines

### Performance Considerations

- Profile code for performance-critical paths
- Consider memory allocations in hot paths
- Use benchmarks to validate performance improvements

### Security Considerations

- Validate all inputs, especially from untrusted sources
- Be careful with file system operations
- Follow secure coding practices

### Compatibility

- Maintain backward compatibility when possible
- Document breaking changes clearly
- Provide migration guides for major changes

## Getting Help

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub discussions for questions and brainstorming
- **Chat**: Join our community chat (link to be added)

## Recognition

Contributors will be recognized in the project documentation and release notes. We appreciate all forms of contribution, from code to documentation to bug reports!

Thank you for contributing to go-simplicity!
