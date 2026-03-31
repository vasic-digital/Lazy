# Architecture -- Lazy

## Purpose

Generic, reusable Go module for lazy-loading primitives. Defers expensive initialization until first access, guaranteeing the loader function runs exactly once even under concurrent access. Uses Go generics for type-safe operation.

## Structure

```
pkg/
  lazy/   Generic Value[T] and Service[T] types wrapping sync.Once for lazy initialization
```

## Key Components

- **`lazy.Value[T]`** -- Lazy value with `Get() (T, error)`, `MustGet() T` (panics on error), and `Reset()` for cache invalidation. Thread-safe via `sync.Once`
- **`lazy.Service[T]`** -- Lazy service initialization with `Get() (T, error)` and `Initialized() bool` check. Same once-only guarantee as Value[T]
- **`lazy.NewValue(loader func() (T, error))`** -- Constructor accepting a loader function
- **`lazy.NewService(loader func() (T, error))`** -- Constructor for service-style lazy loading

## Data Flow

```
NewValue(loader) -> Value[T]{loader, sync.Once}
    |
    Get() -> sync.Once.Do(loader) -> cache result
    Get() -> return cached result (no re-execution)
    |
    Reset() -> new sync.Once -> next Get() re-executes loader
```

## Dependencies

- `github.com/stretchr/testify` -- Test assertions (only dependency)

## Testing Strategy

Table-driven tests with `testify` and race detection. Tests verify single execution guarantee under concurrent access, error propagation, Reset behavior, and MustGet panic on error.
