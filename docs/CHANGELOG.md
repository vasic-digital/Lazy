# Changelog

All notable changes to `digital.vasic.lazy` are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [1.1.0] - 2026-04-03

### Added

- Edge case tests for concurrent access patterns covering 200-goroutine races on `Value[T]`
- Tests for panic propagation in initializers (`TestValue_PanicInInitializer`)
- Tests for nil pointer return values (`TestValue_NilReturnValue`)
- Tests for zero-value scalar types (`TestValue_ZeroValueType`)
- Tests for struct, slice, and map type parameters
- Tests confirming error caching — loader is not retried after a failed call without `Reset()`
  (`TestValue_ErrorCachedOnSubsequentCalls`)
- Tests for `Reset()` under concurrent access (`TestValue_ResetThenConcurrentGet`)
- Tests confirming `Reset()` clears a previously cached error (`TestValue_Reset_ClearsError`)
- Tests for `Service[T]` initialization state before first `Get()` call
  (`TestService_Initialized_BeforeGet_Edge`)
- Tests confirming `Service[T]` preserves error state across repeated `Get()` calls
  (`TestService_ErrorPreservesState`)
- Version alignment with Catalogizer v2.2.0

---

## [1.0.0] - 2026-03-06

### Added

- Initial release of `digital.vasic.lazy`
- `Value[T]` generic lazy value type with deferred initialization via `sync.Once`
- `Service[T]` generic lazy service type with deferred initialization via `sync.Once`
- `NewValue[T](loader func() (T, error)) *Value[T]` constructor
- `NewService[T](initializer func() (T, error)) *Service[T]` constructor
- `(*Value[T]).Get() (T, error)` — load on first call, return cached result on subsequent calls
- `(*Value[T]).MustGet() T` — convenience wrapper that panics on error
- `(*Value[T]).Reset()` — clear cached value and error; next `Get()` re-runs the loader
- `(*Service[T]).Get() (T, error)` — initialize on first call, return cached result thereafter
- `(*Service[T]).Initialized() bool` — report whether initialization succeeded
- Thread-safe initialization: all internal state protected by `sync.Mutex`
- `sync.Once` guarantees loader/initializer executes at most once per instance (per reset cycle
  for `Value[T]`)
- Zero external dependencies — standard library only
- Full test suite: table-driven tests with `testify`, concurrent access tests with `sync.WaitGroup`
- Race detector verified: `go test ./... -race`
- `ARCHITECTURE.md` documenting Proxy and Singleton design patterns
- `CLAUDE.md` with build, test, and code style guidance

---

[Unreleased]: https://github.com/vasic-digital/Lazy/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/vasic-digital/Lazy/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/vasic-digital/Lazy/releases/tag/v1.0.0
