# Contributing to digital.vasic.lazy

Thank you for taking the time to contribute. This document covers everything needed to get a
development environment running, submit changes, and meet the quality bar required for merge.

---

## Prerequisites

| Tool | Minimum version | Notes |
|------|-----------------|-------|
| Go | 1.25 | `go version` to check |
| git | 2.x | Any recent version |
| golangci-lint | 1.57+ | Optional but recommended for local lint |

No external services, databases, or environment variables are required. The module has no
runtime dependencies beyond the Go standard library.

---

## Development Workflow

### 1. Fork and clone

Fork the repository on GitHub, then clone your fork:

```bash
git clone https://github.com/<your-username>/Lazy.git
cd Lazy
```

### 2. Create a feature branch

Branch names should be short and descriptive, prefixed by type:

```bash
git checkout -b feat/batch-lazy-loading
git checkout -b fix/reset-race-condition
git checkout -b docs/api-reference-examples
```

### 3. Make changes

All production code lives in `pkg/lazy/lazy.go`. Tests live in the same package directory.
There are no sub-packages or build tags to manage.

### 4. Run the test suite

Always run tests with the race detector before committing:

```bash
go test ./... -count=1 -race
```

For a quick sanity check without the race detector:

```bash
go test ./... -count=1
```

For benchmarks:

```bash
go test -bench=. -benchmem ./...
```

### 5. Run static analysis

```bash
go vet ./...
```

If `golangci-lint` is installed:

```bash
golangci-lint run ./...
```

### 6. Commit your changes

Follow the Conventional Commits format described below, then push to your fork and open a pull
request against the `main` branch of the upstream repository.

---

## Code Standards

### Formatting

All Go source files must be formatted with `gofmt` before committing. Most editors do this
automatically on save.

```bash
gofmt -w ./...
```

### Static analysis

The code must pass `go vet ./...` with zero warnings. No exceptions.

### Line length

Keep lines at or below **100 characters**. This applies to both source files and test files.
Comments are included in this limit.

### Naming conventions

| Scope | Convention | Example |
|-------|------------|---------|
| Exported identifiers | PascalCase | `NewValue`, `MustGet` |
| Unexported identifiers | camelCase | `loader`, `initErr` |
| Acronyms | All-caps | `initErr` not `initError` when abbreviating |
| Test functions | `Test<Type>_<Method>_<Scenario>` | `TestValue_Get_Error` |

### Error handling

- Always check returned errors. Never use `_` to discard an error in production code.
- Wrap errors with context using `fmt.Errorf`:

```go
return fmt.Errorf("open config file: %w", err)
```

- Do not use `errors.New` for errors that wrap an underlying cause.

### Commit message format

This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

**Types**

| Type | When to use |
|------|-------------|
| `feat` | A new exported function, method, or type |
| `fix` | A bug fix in existing behaviour |
| `docs` | Documentation changes only |
| `test` | Adding or updating tests without changing production code |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf` | Performance improvement |
| `chore` | Dependency bumps, tooling changes |

**Examples**

```
feat(lazy): add Peek() method to Value[T] for non-blocking cache inspection

fix(lazy): prevent double-close race on Reset() under high concurrency

docs(lazy): add reconnection example to user guide

test(lazy): add 500-goroutine concurrent Reset stress test
```

- Subject line: imperative mood, no trailing period, 72 characters or fewer.
- Body: explain _why_, not _what_. Reference issues with `Fixes #123` or `Closes #456`.

---

## Testing Requirements

### Table-driven tests

All new tests must be table-driven where there are two or more related scenarios:

```go
func TestValue_Get(t *testing.T) {
    tests := []struct {
        name    string
        loader  func() (string, error)
        want    string
        wantErr bool
    }{
        {
            name:   "success",
            loader: func() (string, error) { return "ok", nil },
            want:   "ok",
        },
        {
            name:    "error",
            loader:  func() (string, error) { return "", errors.New("fail") },
            wantErr: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            v := lazy.NewValue(tc.loader)
            got, err := v.Get()
            if tc.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

### Race detector

Every pull request must pass `go test ./... -race`. Do not submit changes that introduce data
races. The CI pipeline runs the race detector on all commits.

### Assertions

Use `github.com/stretchr/testify/assert` for non-fatal assertions and
`github.com/stretchr/testify/require` for assertions that must stop the test immediately on
failure. Do not use `t.Fatal` or `t.Error` directly.

### Coverage

New code paths must be covered by tests. There is no hard coverage percentage threshold, but
reviewers will request tests for any untested branch, including error paths.

---

## Documentation Requirements

- Every exported identifier (type, function, method, constant) must have a Go doc comment.
- Doc comments must start with the name of the identifier and be written in complete sentences.
- If you add a new exported method or type, update `docs/API_REFERENCE.md` to include the
  signature, parameter table, return value table, and at least one example.
- If your change is user-facing, add or update the relevant section in `docs/USER_GUIDE.md`.
- Every notable change must have an entry in `docs/CHANGELOG.md` under `[Unreleased]`.

---

## Pre-commit Checklist

Before opening a pull request, verify each item:

- [ ] `gofmt -w ./...` — no formatting changes remain
- [ ] `go vet ./...` — zero warnings
- [ ] `go test ./... -count=1 -race` — all tests pass, no race conditions
- [ ] All new exported identifiers have Go doc comments
- [ ] New behaviour is covered by tests (including error paths)
- [ ] `docs/CHANGELOG.md` updated under `[Unreleased]`
- [ ] `docs/API_REFERENCE.md` updated if a new method or type was added
- [ ] Commit messages follow the Conventional Commits format
- [ ] No secrets, credentials, or `.env` files are staged
- [ ] Branch is up to date with `main` (rebase preferred over merge)
