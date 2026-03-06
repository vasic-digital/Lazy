# Lazy Module

`digital.vasic.lazy` is a standalone Go module providing generic lazy-loading primitives using `sync.Once`. It offers `Value[T]` for deferred value computation and `Service[T]` for one-time service initialization, both safe for concurrent use.

## Key Features

- **Lazy value loading** -- `Value[T]` calls a loader function at most once on first access, caching the result for subsequent calls
- **Lazy service initialization** -- `Service[T]` wraps one-time service setup with error handling
- **Generic types** -- Uses Go generics (`any` constraint) for type-safe lazy loading of any type
- **Thread safety** -- Built on `sync.Once` for safe concurrent access without explicit locking
- **Reset support** -- `Value[T].Reset()` clears the cached value so the loader runs again on next access
- **Panic on error** -- `Value[T].MustGet()` provides a convenience accessor that panics on loader failure

## API Overview

| Type | Method | Description |
|------|--------|-------------|
| `Value[T]` | `NewValue(loader)` | Create a lazy value with a loader function |
| `Value[T]` | `Get() (T, error)` | Load and return the value (loader runs once) |
| `Value[T]` | `MustGet() T` | Load and return the value, panicking on error |
| `Value[T]` | `Reset()` | Clear cached value; loader will run again |
| `Service[T]` | `NewService(init)` | Create a lazy service with an init function |
| `Service[T]` | `Get() (T, error)` | Initialize and return the service (init runs once) |
| `Service[T]` | `Initialized() bool` | Check if the service initialized without error |

## Installation

```bash
go get digital.vasic.lazy
```

Requires Go 1.21 or later (generics support). Only external dependency is `testify` for tests.
