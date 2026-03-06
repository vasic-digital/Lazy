# AGENTS.md - Lazy Module Multi-Agent Coordination Guide

## Overview

This document provides guidance for AI agents (Claude Code, Copilot, Cursor, etc.) working on the `digital.vasic.lazy` module. It defines responsibilities, boundaries, and coordination protocols to prevent conflicts when multiple agents operate concurrently.

## Module Identity

- **Module path**: `digital.vasic.lazy`
- **Language**: Go 1.24+
- **Dependencies**: `github.com/stretchr/testify`
- **Packages**: `pkg/lazy`

## Package Ownership Boundaries

### `pkg/lazy` -- Lazy Loading Primitives

- **Scope**: Generic `Value[T]` and `Service[T]` types, `sync.Once` lazy initialization with error handling.
- **Owner concern**: Self-contained single package. No cross-package dependencies within this module.
- **Thread safety**: Both types use `sync.Once` (and `sync.Mutex` for `Reset()`). All new methods MUST maintain concurrent safety.

## Dependency Graph

```
pkg/lazy (independent, no internal deps)
```

No external dependencies beyond the Go standard library and testify for tests.

## Agent Coordination Rules

### 1. Type Changes

If you modify `Value[T]`:
- Ensure `Get()`, `MustGet()`, and `Reset()` remain thread-safe
- Update corresponding tests in `lazy_test.go`
- Do NOT break the `sync.Once` guarantee

If you modify `Service[T]`:
- Ensure `Get()` and `Initialized()` remain thread-safe
- Update corresponding tests in `lazy_test.go`

### 2. Concurrency Safety

Both types are designed for concurrent access:
- `Value[T]`: `sync.Once` for initialization, `sync.Mutex` for `Reset()`
- `Service[T]`: `sync.Once` for initialization

Rules:
- Never hold a lock while calling an external function that might also lock
- Always return copies of internal data when appropriate

### 3. Testing Standards

- **Framework**: `github.com/stretchr/testify` (assert + require)
- **Naming**: `Test<Struct>_<Method>_<Scenario>` (e.g., `TestValue_Get_ConcurrentAccess`)
- **Style**: Table-driven tests with `tests` slice and `t.Run` subtests
- **Concurrency**: Include concurrent access tests for all public methods
- **Run all tests**: `go test ./... -count=1 -race`

### 4. Adding New Types

To add a new lazy primitive:
1. Add to `pkg/lazy/lazy.go` (single package module)
2. Add tests to `pkg/lazy/lazy_test.go`
3. Include concurrent access tests
4. Maintain the generic `[T any]` pattern

### 5. File Ownership

| File | Primary Concern | Cross-Package Impact |
|------|----------------|---------------------|
| `pkg/lazy/lazy.go` | Value[T], Service[T] types | NONE |
| `pkg/lazy/lazy_test.go` | All tests | NONE |

## Build and Validation Commands

```bash
# Full validation
go build ./...
go test ./... -count=1 -race
go vet ./...
gofmt -l .

# Single package
go test -v ./pkg/lazy/...

# Benchmarks
go test -bench=. ./...
```

## Commit Conventions

- Use Conventional Commits: `feat(lazy): add batch lazy loading`
- Scope: `lazy`
- Use `docs` scope for documentation-only changes
- Run `gofmt` and `go vet` before every commit
