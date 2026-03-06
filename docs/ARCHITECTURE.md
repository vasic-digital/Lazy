# Lazy -- Architecture

## Purpose

`digital.vasic.lazy` provides generic lazy-loading primitives for Go. It defers expensive initialization (value computation, service startup) until first access, guaranteeing that the loader runs exactly once even under concurrent access. The module uses Go generics (`any` constraint) so it works with any type without casting.

## Package Overview

| Package | Import Path | Responsibility |
|---------|-------------|----------------|
| `lazy` | `digital.vasic.lazy/pkg/lazy` | Generic `Value[T]` and `Service[T]` types that wrap `sync.Once` to provide lazy initialization with error handling |

## Design Patterns

### Proxy (Deferred Initialization)

Both `Value[T]` and `Service[T]` act as proxies for the underlying value or service. Callers interact with the proxy via `Get()` and the real initialization is deferred until that first call. This is the classic virtual proxy pattern -- the expensive work happens transparently on demand.

### Singleton (Once Semantics)

`sync.Once` guarantees the loader/init function executes at most once, regardless of how many goroutines call `Get()` concurrently. This gives each `Value` or `Service` instance singleton semantics for its computed result.

## Dependency Diagram

```
+---------------------------+
|       Consumer code       |
+---------------------------+
            |
            | calls Get() / MustGet()
            v
+---------------------------+
|    pkg/lazy               |
|                           |
|  Value[T]    Service[T]   |
|    |              |       |
|    +------+-------+       |
|           |               |
|       sync.Once           |
+---------------------------+
```

No external dependencies beyond the Go standard library.

## Key Interfaces

The module does not define interfaces -- it exposes two concrete generic structs. Their public API is intentionally small:

### Value[T]

```go
type Value[T any] struct { /* unexported */ }

func NewValue[T any](loader func() (T, error)) *Value[T]
func (v *Value[T]) Get() (T, error)      // Load on first call, return cached thereafter
func (v *Value[T]) MustGet() T           // Like Get but panics on error
func (v *Value[T]) Reset()               // Clear cache; next Get() re-runs the loader
```

### Service[T]

```go
type Service[T any] struct { /* unexported */ }

func NewService[T any](init func() (T, error)) *Service[T]
func (s *Service[T]) Get() (T, error)    // Initialize on first call
func (s *Service[T]) Initialized() bool  // True if init succeeded (no error)
```

`Value[T]` supports `Reset()` for scenarios where a cached value must be invalidated (e.g., configuration reload). `Service[T]` omits `Reset()` because services are typically initialized once for the lifetime of the process.

## Usage Example

```go
package main

import (
    "database/sql"
    "fmt"

    "digital.vasic.lazy/pkg/lazy"
)

func main() {
    // Lazy database connection -- opened only when first queried
    db := lazy.NewValue(func() (*sql.DB, error) {
        return sql.Open("sqlite3", "app.db")
    })

    // Nothing happens until Get() is called
    conn, err := db.Get()
    if err != nil {
        panic(err)
    }

    // Subsequent calls return the same *sql.DB without re-opening
    conn2, _ := db.Get()
    fmt.Println(conn == conn2) // true

    // Lazy service initialization
    svc := lazy.NewService(func() (string, error) {
        // Expensive startup logic
        return "ready", nil
    })

    name, _ := svc.Get()
    fmt.Println(name, svc.Initialized()) // "ready" true
}
```
