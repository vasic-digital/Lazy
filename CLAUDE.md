# CLAUDE.md - Lazy Module

## Overview

`digital.vasic.lazy` is a generic, reusable Go module for lazy-loading primitives. It defers expensive initialization until first access, guaranteeing the loader runs exactly once even under concurrent access.

**Module**: `digital.vasic.lazy` (Go 1.24+)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
go test ./... -short              # Unit tests only
go test -bench=. ./...            # Benchmarks
```

## Code Style

- Standard Go conventions, `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line separated)
- Line length <= 100 chars
- Naming: `camelCase` private, `PascalCase` exported, acronyms all-caps
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`
- Tests: table-driven, `testify`, naming `Test<Struct>_<Method>_<Scenario>`

## Package Structure

| Package | Purpose |
|---------|---------|
| `pkg/lazy` | Generic `Value[T]` and `Service[T]` types with `sync.Once` lazy initialization |

## Key Types

- `Value[T]` -- Lazy value with `Get()`, `MustGet()`, and `Reset()` for cache invalidation
- `Service[T]` -- Lazy service initialization with `Get()` and `Initialized()` check

## Design Patterns

- **Proxy**: Value[T] and Service[T] defer initialization until first `Get()` call
- **Singleton**: `sync.Once` guarantees loader executes at most once per instance
- **Factory**: `NewValue()`, `NewService()` constructors

## Commit Style

Conventional Commits: `feat(lazy): add batch lazy loading`
